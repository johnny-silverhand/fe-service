package app

import (
	"fmt"
	"im/model"
	"im/services/payment/sberbank/schema"
	"math"
	"net/http"
	"strconv"
)

func NewBasketFromModel(pr *model.Product, order *model.Order, q int) *model.Basket {
	return &model.Basket{
		OrderId:   order.Id,
		ProductId: pr.Id,
		Price:     pr.Price,
		Currency:  pr.Currency,
		Quantity:  q,
	}
}

func (a *App) GetOrder(orderId string) (*model.Order, *model.AppError) {

	result := <-a.Srv.Store.Order().Get(orderId)
	if result.Err != nil {
		return nil, result.Err
	}

	rorder := result.Data.(*model.Order)

	rorder = a.PrepareOrderForClient(rorder, false)

	return rorder, nil
}

func (a *App) GetOrdersPage(page int, perPage int, sort string) (*model.OrderList, *model.AppError) {
	return a.GetOrders(page*perPage, perPage, sort)
}

func (a *App) GetOrders(offset int, limit int, sort string) (*model.OrderList, *model.AppError) {

	result := <-a.Srv.Store.Order().GetAllPage(offset, limit, model.GetOrder(sort))

	if result.Err != nil {
		return nil, result.Err
	}

	return result.Data.(*model.OrderList), nil
}

func (a *App) RecalculateOrder(order *model.Order) (*model.Order, *model.AppError) {

	basket := make([]*model.Basket, 0)
	var total float64

	if order.Positions != nil {
		order.NormalizePositions()
		for _, ps := range order.Positions {
			pr := <-a.Srv.Store.Product().Get(ps.ProductId)

			if pr.Err == nil {
				product := pr.Data.(*model.Product)

				ps.Price = product.Price
				ps.Currency = product.Currency
				ps.Name = product.Name

				total += ps.Price * float64(ps.Quantity)
				basket = append(basket, ps)
			}
		}
		order.Price = total - order.DiscountValue
		order.Positions = basket
	}

	return order, nil
}

func (a *App) CreateOrderInvoice(order *model.Order, user *model.User) (*model.Order, *model.AppError) {
	result := <-a.Srv.Store.Order().Save(order)
	if result.Err != nil {
		return nil, result.Err
	}

	newOrder := result.Data.(*model.Order)
	var msg string
	msg += fmt.Sprintf("Счет на оплату № %s \n", newOrder.FormatOrderNumber())

	post := &model.Post{
		UserId:   user.Id,
		Message:  msg,
		CreateAt: model.GetMillis() + 1,
		Type:     model.POST_WITH_INVOICE,
	}

	a.CreatePostWithOrder(post, newOrder, false)

	return newOrder, nil
}

func (a *App) CreateOrder(order *model.Order) (*model.Order, *model.AppError) {
	// проверка DiscountValue на превышение допустимого лимита оплаты бонусами в заказе
	var productIds = make([]string, 0)
	for _, position := range order.Positions {
		productIds = append(productIds, position.ProductId)
	}
	if discountLimit, err := a.GetDiscountLimits(productIds); err != nil {
		return nil, err
	} else if int64(order.DiscountValue) > discountLimit.Total {
		return nil, model.NewAppError("CreateOrder", "api.order.create_order.discount_limit.app_error", nil, "id="+order.Id, http.StatusBadRequest)
	}

	a.RecalculateOrder(order)

	for i, position := range order.Positions {
		position = a.PrepareBasketForClient(position, true)
		order.Positions[i] = position
	}

	result := <-a.Srv.Store.Order().SaveWithBasket(order)

	if result.Err != nil {
		return nil, result.Err
	}

	newOrder := result.Data.(*model.Order)
	var msg string
	msg += fmt.Sprintf("Заказ № %s \n", newOrder.FormatOrderNumber())

	post := &model.Post{
		UserId:   newOrder.UserId,
		Message:  msg,
		CreateAt: model.GetMillis() + 1,
		Type:     model.POST_WITH_METADATA,
	}

	if newOrder.DiscountValue > 0 {
		transaction := &model.Transaction{
			UserId:      newOrder.UserId,
			OrderId:     newOrder.Id,
			Description: fmt.Sprintf("Списание по заказу № %s \n", newOrder.FormatOrderNumber()),
			Value:       -math.Floor(newOrder.DiscountValue),
		}

		a.DeductionTransaction(transaction)
	}

	a.CreatePostWithOrder(post, newOrder, false)

	return newOrder, nil
}

func (a *App) UpdateOrder(order *model.Order, safeUpdate bool) (*model.Order, *model.AppError) {
	//order.SanitizeProps()

	result := <-a.Srv.Store.Order().Get(order.Id)
	if result.Err != nil {
		return nil, result.Err
	}

	oldOrder := result.Data.(*model.Order)

	if oldOrder == nil {
		err := model.NewAppError("UpdateOrder", "api.order.update_order.find.app_error", nil, "id="+order.Id, http.StatusBadRequest)
		return nil, err
	}

	if oldOrder.DeleteAt != 0 {
		err := model.NewAppError("UpdateOrder", "api.order.update_order.permissions_details.app_error", map[string]interface{}{"OrderId": order.Id}, "", http.StatusBadRequest)
		return nil, err
	}

	/*if order.Status == oldOrder.Status && order.DeliveryAt == oldOrder.DeliveryAt {
		err := model.NewAppError("UpdateOrder", "api.order.update_order.status.app_error", map[string]interface{}{"OrderId": order.Id}, "", http.StatusBadRequest)
		return nil, err
	}*/

	oldOrder = a.PrepareOrderForClient(oldOrder, false)

	if oldOrder.User == nil {
		err := model.NewAppError("UpdateOrder", "api.order.update_order.user.app_error", map[string]interface{}{"OrderId": order.Id}, "", http.StatusBadRequest)
		return nil, err
	}

	switch order.Status {
	case model.ORDER_STATUS_AWAITING_PAYMENT:
	case model.ORDER_STATUS_AWAITING_FULFILLMENT:
	case model.ORDER_STATUS_AWAITING_PICKUP:
	case model.ORDER_STATUS_AWAITING_SHIPMENT:
	case model.ORDER_STATUS_DECLINED:
		if oldOrder.Status == model.ORDER_STATUS_SHIPPED {
			a.SetOrderCancel(order.Id)
		}
	case model.ORDER_STATUS_REFUNDED:
	case model.ORDER_STATUS_SHIPPED:
		if oldOrder.Status != model.ORDER_STATUS_SHIPPED {
			if err := a.SetOrderShipped(order.Id); err != nil {
				return nil, err
			}
		}
	default:
		// If not part of the scheme for this channel, then it is not allowed to apply it as an explicit role.
		return nil, model.NewAppError("UpdateOrder", "app.order.update_order.status.not_found.app_error", map[string]interface{}{"OrderId": order.Id}, "", http.StatusBadRequest)
	}

	newOrder := &model.Order{}
	*newOrder = *oldOrder

	newOrder.PaySystemCode = order.PaySystemCode
	newOrder.Status = order.Status
	newOrder.DeliveryAt = order.DeliveryAt

	result = <-a.Srv.Store.Order().Update(newOrder)
	if result.Err != nil {
		return nil, result.Err
	}

	rorder := result.Data.(*model.Order)
	rorder = a.PrepareOrderForClient(rorder, false)

	a.UpdatePostWithOrder(order, false)

	//a.InvalidateCacheForChannelOrders(rorder.ChannelId)

	return rorder, nil
}

func (a *App) PrepareOrderForClient(originalOrder *model.Order, isNewOrder bool) *model.Order {
	order := originalOrder.Clone()

	order.Positions = a.GetBasketForOrder(order)
	if post, err := a.FindPostWithOrder(order.Id); err == nil {
		order.Post = post
	}
	if users, err := a.GetUsersByIds([]string{order.UserId}, true); err == nil {
		a.SanitizeProfile(users[len(users)-1], false)
		order.User = users[len(users)-1]
	}

	return order
}

func (a *App) PrepareOrderListForClient(originalList *model.OrderList) *model.OrderList {
	list := &model.OrderList{
		Orders: make(map[string]*model.Order, len(originalList.Orders)),
		Order:  originalList.Order, // Note that this uses the original Order array, so it isn't a deep copy
		Total:  originalList.Total,
	}

	for id, originalOrder := range originalList.Orders {
		order := a.PrepareOrderForClient(originalOrder, false)

		list.Orders[id] = order
	}

	return list
}

func (a *App) DeleteOrder(orderId, deleteByID string) (*model.Order, *model.AppError) {
	result := <-a.Srv.Store.Order().Get(orderId)
	if result.Err != nil {
		result.Err.StatusCode = http.StatusBadRequest
		return nil, result.Err
	}
	order := result.Data.(*model.Order)

	if result := <-a.Srv.Store.Order().Delete(orderId, model.GetMillis(), deleteByID); result.Err != nil {
		return nil, result.Err
	}

	return order, nil
}

func (a *App) GetAllOrdersBeforeOrder(orderId string, options *model.OrderGetOptions) (*model.OrderList, *model.AppError) {

	//if result := <-a.Srv.Store.Order().GetAllOrdersBefore(orderId, perPage, page*perPage, appId); result.Err != nil {
	if result := <-a.Srv.Store.Order().GetAllOrdersBefore(orderId, *options); result.Err != nil {
		return nil, result.Err
	} else {
		return result.Data.(*model.OrderList), nil
	}
}

func (a *App) GetAllOrdersAfterOrder(orderId string, options *model.OrderGetOptions) (*model.OrderList, *model.AppError) {

	//if result := <-a.Srv.Store.Order().GetAllOrdersAfter(orderId, perPage, page*perPage, appId); result.Err != nil {
	if result := <-a.Srv.Store.Order().GetAllOrdersAfter(orderId, *options); result.Err != nil {
		return nil, result.Err
	} else {
		return result.Data.(*model.OrderList), nil
	}
}

func (a *App) GetAllOrdersSince(time int64, options *model.OrderGetOptions) (*model.OrderList, *model.AppError) {
	//if result := <-a.Srv.Store.Order().GetAllOrdersSince(time, true, options); result.Err != nil {
	if result := <-a.Srv.Store.Order().GetAllOrdersSince(time, *options); result.Err != nil {
		return nil, result.Err
	} else {
		return result.Data.(*model.OrderList), nil
	}
}

func (a *App) GetAllOrdersPage(options *model.OrderGetOptions) (*model.OrderList, *model.AppError) {
	//if result := <-a.Srv.Store.Order().GetAllOrders(page*perPage, perPage, true, appId); result.Err != nil {
	if result := <-a.Srv.Store.Order().GetAllOrders(*options); result.Err != nil {
		return nil, result.Err
	} else {
		return result.Data.(*model.OrderList), nil
	}
}

func (a *App) GetUserOrders(userId string, page int, perPage int, sort string) (*model.OrderList, *model.AppError) {
	if result := <-a.Srv.Store.Order().GetByUserId(userId, page*perPage, perPage, model.GetOrder(sort)); result.Err != nil {
		return nil, result.Err
	} else {
		orderList := result.Data.(*model.OrderList)
		return a.PrepareOrderListForClient(orderList), nil
	}
}

func (a *App) SetOrderShipped(orderId string) *model.AppError {

	result := <-a.Srv.Store.Order().Get(orderId)
	if result.Err != nil {
		result.Err.StatusCode = http.StatusBadRequest
		return result.Err
	}
	order := result.Data.(*model.Order)
	order = a.PrepareOrderForClient(order, false)

	if result := <-a.Srv.Store.Application().Get(order.User.AppId); result.Err != nil {
		return result.Err
	} else {
		order.Status = model.ORDER_STATUS_SHIPPED
		application := result.Data.(*model.Application)

		if _, err := a.UpdatePostWithOrder(order, false); err != nil {
			fmt.Println(err)
		}

		if order.DiscountValue == 0 {

			var cashback float64 = 0
			var price float64 = 0

			for _, position := range order.Positions {
				if position.DiscountValue > 0 {
					cashback += float64(position.Quantity) * float64(position.DiscountValue)
				} else {
					cashback += math.Floor(float64(position.Quantity) * position.Price * (application.Cashback / 100))
					//cashback += float64(position.Quantity) * position.Price
					price += float64(position.Quantity) * position.Price
				}
			}

			a.Srv.Go(func() {
				value := cashback

				transaction := &model.Transaction{
					UserId:      order.User.Id,
					OrderId:     order.Id,
					Description: fmt.Sprintf("Начисление кэшбека по заказу № %s \n", order.FormatOrderNumber()),
					Value:       value,
				}

				if transaction.Value > 0 {
					a.AccrualTransaction(transaction)
				}
			})

			a.Srv.Go(func() {
				if levels, err := a.GetAllLevelsPage(0, 60, &application.Id); err == nil {
					levels.SortByLvl()
					if u, e := a.GetUser(order.User.InvitedBy); e == nil {
						for _, id := range levels.Order {
							value := math.Floor(price * (levels.Levels[id].Value / 100))
							transaction := &model.Transaction{
								UserId:      u.Id,
								OrderId:     order.Id,
								Description: fmt.Sprintf("Начисление по заказу друга \n"),
								Value:       value,
								Type:        model.TRANSACTION_TYPE_BONUS,
							}
							if transaction.Value > 0 {
								a.AccrualTransaction(transaction)
							}
							if u, e = a.GetUser(u.InvitedBy); e != nil {
								break
							}
						}
					}
				}
			})
		}

	}

	return nil
}

func (a *App) SetOrderPayed(orderId string, response *schema.OrderStatusResponse) *model.AppError {

	if response == nil {
		return model.NewAppError("", "", nil, "", http.StatusBadRequest)
	}

	result := <-a.Srv.Store.Order().Get(orderId)
	if result.Err != nil {
		result.Err.StatusCode = http.StatusBadRequest
		return result.Err
	}
	order := result.Data.(*model.Order)

	if result := <-a.Srv.Store.Order().SetOrderPayed(order.Id); result.Err != nil {
		return result.Err
	} else {

		a.UpdatePostWithOrder(order, false)

		/*a.AccrualTransaction(&model.Transaction{
			OrderId: order.Id,
			UserId:  order.UserId,
			Value:   math.Floor(order.Price * 0.05),
		})*/

		post := &model.Post{
			UserId:   order.UserId,
			Message:  "Оплата банковской картой " + response.CardAuthInfo.MaskedPan + " по транзакции № " + strconv.FormatInt(order.CreateAt, 10) + ". Заказ № " + order.FormatOrderNumber(),
			CreateAt: model.GetMillis() + 1,
			Type:     model.POST_WITH_TRANSACTION,
		}

		a.CreatePostWithTransaction(post, false)

		return nil
	}
}

func (a *App) SetOrderCancel(orderId string) *model.AppError {

	result := <-a.Srv.Store.Order().Get(orderId)
	if result.Err != nil {
		result.Err.StatusCode = http.StatusBadRequest
		return result.Err
	}

	order := result.Data.(*model.Order)

	if order.Canceled == true {
		return nil
	}

	if result := <-a.Srv.Store.Order().SetOrderCancel(order.Id); result.Err != nil {
		return result.Err
	} else {

		/*a.DeductionTransaction(&model.Transaction{
			OrderId:     order.Id,
			UserId:      order.UserId,
			Value:       -math.Floor(order.Price * 0.05),
			Description: "Отмена транзакции № " + strconv.FormatInt(order.CreateAt, 10),
		})*/

		if order.PaySystemId != model.PAYMENT_SYSTEM_CASH {

			post := &model.Post{
				UserId:   order.UserId,
				Message:  "Отмена оплаты по транзакции № " + strconv.FormatInt(order.CreateAt, 10) + ". Заказ № " + order.FormatOrderNumber(),
				CreateAt: model.GetMillis() + 1,
				Type:     model.POST_WITH_TRANSACTION,
			}

			a.CreatePostWithTransaction(post, false)

			// TODO отправка запроса в эквайринг банка об возврате денежных средств
		}

		return nil
	}
}

func (a *App) GetOrdersStats(options model.OrderCountOptions) (*model.OrdersStats, *model.AppError) {
	result := <-a.Srv.Store.Order().Count(options)
	if result.Err != nil {
		return nil, result.Err
	}
	return result.Data.(*model.OrdersStats), nil
}
