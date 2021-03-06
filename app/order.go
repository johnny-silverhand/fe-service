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
	var price float64 = 0
	if order.Positions != nil {
		order.NormalizePositions()
		order.Positions = a.PrepareBasketListForClient(order.Positions, true)

		for _, position := range order.Positions {
			price += position.Price * float64(position.Quantity)
		}
		order.Price = price - order.DiscountValue
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
		for i := 1; i <= position.Quantity; i++ {
			productIds = append(productIds, position.ProductId)
		}
	}
	if discountLimit, err := a.GetDiscountLimits(productIds); err != nil {
		return nil, err
	} else if int64(order.DiscountValue) > discountLimit.Total {
		return nil, model.NewAppError("CreateOrder", "api.order.create_order.discount_limit.app_error", nil, "id="+order.Id, http.StatusBadRequest)
	}

	var err *model.AppError
	order, err = a.RecalculateOrder(order)
	if err != nil {
		return nil, err
	}

	for i, position := range order.Positions {
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
			Value:       math.Floor(newOrder.DiscountValue),
		}

		a.DeductionTransaction(transaction)
	}

	a.CreatePostWithOrder(post, newOrder, false)

	return newOrder, nil
}

func (a *App) UpdateOrder(id string, patch *model.OrderPatch, safeUpdate bool) (*model.Order, *model.AppError) {
	//order.SanitizeProps()

	result := <-a.Srv.Store.Order().Get(id)
	if result.Err != nil {
		return nil, result.Err
	}

	oldOrder := result.Data.(*model.Order)

	if oldOrder == nil {
		err := model.NewAppError("UpdateOrder", "api.order.update_order.find.app_error", nil, "id="+id, http.StatusBadRequest)
		return nil, err
	}

	if oldOrder.DeleteAt != 0 {
		err := model.NewAppError("UpdateOrder", "api.order.update_order.permissions_details.app_error", map[string]interface{}{"OrderId": id}, "", http.StatusBadRequest)
		return nil, err
	}

	newOrder := &model.Order{}
	*newOrder = *oldOrder
	newOrder.Patch(patch)

	if oldOrder.Status != newOrder.Status {
		switch newOrder.Status {
		case model.ORDER_STATUS_AWAITING_FULFILLMENT:
		case model.ORDER_STATUS_AWAITING_PICKUP:
		case model.ORDER_STATUS_AWAITING_SHIPMENT:
		case model.ORDER_STATUS_DECLINED:
			if err := a.SetOrderCancel(id); err != nil {
				return nil, err
			}
			result = <-a.Srv.Store.Order().Get(id)
			if result.Err != nil {
				return nil, result.Err
			}
			rorder := result.Data.(*model.Order)
			rorder = a.PrepareOrderForClient(rorder, false)
			a.UpdatePostWithOrder(rorder, false)
			return rorder, nil
		case model.ORDER_STATUS_REFUNDED:
		case model.ORDER_STATUS_SHIPPED:
			if err := a.SetOrderShipped(id); err != nil {
				return nil, err
			}
			result = <-a.Srv.Store.Order().Get(id)
			if result.Err != nil {
				return nil, result.Err
			}
			rorder := result.Data.(*model.Order)
			rorder = a.PrepareOrderForClient(rorder, false)
			a.UpdatePostWithOrder(rorder, false)
			return rorder, nil
		default:
			return nil, model.NewAppError("UpdateOrder", "app.order.update_order.status.not_found.app_error", map[string]interface{}{"OrderId": id}, "", http.StatusBadRequest)
		}
	}

	result = <-a.Srv.Store.Order().Update(newOrder)
	if result.Err != nil {
		return nil, result.Err
	}
	rorder := result.Data.(*model.Order)
	rorder = a.PrepareOrderForClient(rorder, false)

	a.UpdatePostWithOrder(rorder, false)

	return rorder, nil
}

func (a *App) PrepareOrderForClient(originalOrder *model.Order, isNewOrder bool) *model.Order {
	order := originalOrder.Clone()

	basketList := a.GetBasketForOrder(order)
	order.Positions = a.PrepareBasketListForClient(basketList, isNewOrder)
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

func (a *App) GetUserOrders(options *model.OrderGetOptions) (*model.OrderList, *model.AppError) {
	//if result := <-a.Srv.Store.Order().GetByUserId(userId, page*perPage, perPage, model.GetOrder(sort)); result.Err != nil {
	if result := <-a.Srv.Store.Order().GetByUserId(*options); result.Err != nil {
		return nil, result.Err
	} else {
		orderList := result.Data.(*model.OrderList)
		return a.PrepareOrderListForClient(orderList), nil
	}
}

func (a *App) CalculateBonusForOrder(order *model.Order) *model.AppError {
	var application *model.Application
	order = a.PrepareOrderForClient(order, false)
	if order.User == nil {
		return model.NewAppError("CalculateBonusForOrder", "api.order.calculate_bonus_for_order.get_user.app_error", nil, "id="+order.Id, http.StatusBadRequest)
	}
	if result := <-a.Srv.Store.Application().Get(order.User.AppId); result.Err != nil {
		return result.Err
	} else if application = result.Data.(*model.Application); application == nil {
		return model.NewAppError("CalculateBonusForOrder", "api.order.calculate_bonus_for_order.get_application.app_error", nil, "id="+order.Id, http.StatusBadRequest)
	}

	var cashback float64 = 0
	var price float64 = 0
	for _, position := range order.Positions {
		cashback += float64(position.Quantity) * position.Cashback
		if position.Product != nil && position.Product.PrivateRule == false {
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
			if user, _ := a.GetUser(order.User.InvitedBy); user != nil {
				for _, id := range levels.Order {
					value := math.Floor(price * (levels.Levels[id].Value / 100))
					transaction := &model.Transaction{
						UserId:      user.Id,
						OrderId:     order.Id,
						Description: fmt.Sprintf("Начисление по заказу друга \n"),
						Value:       value,
						Type:        model.TRANSACTION_TYPE_BONUS,
					}
					if transaction.Value > 0 {
						a.AccrualTransaction(transaction)
					}
					if user, _ = a.GetUser(user.InvitedBy); user == nil {
						break
					}
				}
			}
		}
	})

	return nil
}

func (a *App) SetOrderShipped(orderId string) *model.AppError {
	if result := <-a.Srv.Store.Order().SetOrderPayed(orderId); result.Err != nil {
		return result.Err
	}

	result := <-a.Srv.Store.Order().Get(orderId)
	if result.Err != nil {
		result.Err.StatusCode = http.StatusBadRequest
		return result.Err
	}
	order := result.Data.(*model.Order)

	if order.DiscountValue == 0 {
		if err := a.CalculateBonusForOrder(order); err != nil {
			return err
		}
	}

	//order = a.PrepareOrderForClient(order, false)
	order.Status = model.ORDER_STATUS_SHIPPED
	order.UpdateAt = model.GetMillis()

	if result := <-a.Srv.Store.Order().Update(order); result.Err != nil {
		return result.Err
	}
	a.UpdatePostWithOrder(order, false)

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

		if order.DiscountValue > 0 {
			transaction := &model.Transaction{
				UserId:      order.UserId,
				OrderId:     order.Id,
				Description: fmt.Sprintf("Возврат по заказу № %s \n", order.FormatOrderNumber()),
				Value:       math.Floor(order.DiscountValue),
			}

			_, err := a.AccrualTransaction(transaction)
			fmt.Println(err)
		}

		/*a.DeductionTransaction(&model.Transaction{
			OrderId:     order.Id,
			UserId:      order.UserId,
			Value:       -math.Floor(order.Price * 0.05),
			Description: "Отмена транзакции № " + strconv.FormatInt(order.CreateAt, 10),
		})*/

		/*if utils.StringInSlice(order.PaySystemId, []string{model.PAYMENT_SYSTEM_ALFABANK, model.PAYMENT_SYSTEM_SBERBANK}) {

			post := &model.Post{
				UserId:   order.UserId,
				Message:  "Отмена оплаты по транзакции № " + strconv.FormatInt(order.CreateAt, 10) + ". Заказ № " + order.FormatOrderNumber(),
				CreateAt: model.GetMillis() + 1,
				Type:     model.POST_WITH_TRANSACTION,
			}

			a.CreatePostWithTransaction(post, false)

			// TODO отправка запроса в эквайринг банка об возврате денежных средств
		} else {
			post := &model.Post{
				UserId:   order.UserId,
				Message:  "Заказ № " + order.FormatOrderNumber() + " отменен.",
				CreateAt: model.GetMillis() + 1,
				Type:     model.POST_WITH_TRANSACTION,
			}

			a.CreatePostWithTransaction(post, false)
		}*/

		a.UpdatePostWithOrder(order, false)

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
