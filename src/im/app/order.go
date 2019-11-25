package app

import (
	"fmt"
	"im/model"
	"im/store"
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
			Value:       -newOrder.DiscountValue,
		}

		a.DeductionTransaction(transaction)
	}

	accural := newOrder.Price * 0.1

	transaction := &model.Transaction{
		UserId:      newOrder.UserId,
		OrderId:     newOrder.Id,
		Description: fmt.Sprintf("Начисление по заказу № %s \n", newOrder.FormatOrderNumber()),
		Value:       accural,
	}

	a.AccrualTransaction(transaction)
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

	newOrder := &model.Order{}
	*newOrder = *oldOrder

	result = <-a.Srv.Store.Order().Update(newOrder)
	if result.Err != nil {
		return nil, result.Err
	}

	rorder := result.Data.(*model.Order)
	rorder = a.PrepareOrderForClient(rorder, false)

	//a.InvalidateCacheForChannelOrders(rorder.ChannelId)

	return rorder, nil
}

func (a *App) PrepareOrderForClient(originalOrder *model.Order, isNewOrder bool) *model.Order {
	order := originalOrder.Clone()

	order.Positions = a.GetBasketForOrder(order)

	return order
}

func (a *App) PrepareOrderListForClient(originalList *model.OrderList) *model.OrderList {
	list := &model.OrderList{
		Orders: make(map[string]*model.Order, len(originalList.Orders)),
		Order:  originalList.Order, // Note that this uses the original Order array, so it isn't a deep copy
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

func (a *App) GetAllOrdersBeforeOrder(orderId string, page, perPage int) (*model.OrderList, *model.AppError) {

	if result := <-a.Srv.Store.Order().GetAllOrdersBefore(orderId, perPage, page*perPage); result.Err != nil {
		return nil, result.Err
	} else {
		return result.Data.(*model.OrderList), nil
	}
}

func (a *App) GetAllOrdersAfterOrder(orderId string, page, perPage int) (*model.OrderList, *model.AppError) {

	if result := <-a.Srv.Store.Order().GetAllOrdersAfter(orderId, perPage, page*perPage); result.Err != nil {
		return nil, result.Err
	} else {
		return result.Data.(*model.OrderList), nil
	}
}

func (a *App) GetAllOrdersAroundOrder(orderId string, offset, limit int, before bool) (*model.OrderList, *model.AppError) {
	var pchan store.StoreChannel

	if before {
		pchan = a.Srv.Store.Order().GetAllOrdersBefore(orderId, limit, offset)
	} else {
		pchan = a.Srv.Store.Order().GetAllOrdersAfter(orderId, limit, offset)
	}

	if result := <-pchan; result.Err != nil {
		return nil, result.Err
	} else {
		return result.Data.(*model.OrderList), nil
	}
}

func (a *App) GetAllOrdersSince(time int64) (*model.OrderList, *model.AppError) {
	if result := <-a.Srv.Store.Order().GetAllOrdersSince(time, true); result.Err != nil {
		return nil, result.Err
	} else {
		return result.Data.(*model.OrderList), nil
	}
}

func (a *App) GetAllOrdersPage(page int, perPage int) (*model.OrderList, *model.AppError) {
	if result := <-a.Srv.Store.Order().GetAllOrders(page*perPage, perPage, true); result.Err != nil {
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

func (a *App) SetOrderPayed(orderId string) *model.AppError {

	result := <-a.Srv.Store.Order().Get(orderId)
	if result.Err != nil {
		result.Err.StatusCode = http.StatusBadRequest
		return result.Err
	}
	order := result.Data.(*model.Order)

	if result := <-a.Srv.Store.Order().SetOrderPayed(order.Id); result.Err != nil {
		return result.Err
	} else {

		a.AccrualTransaction(&model.Transaction{
			OrderId: order.Id,
			UserId:  order.UserId,
			Value:   math.Round(order.Price * 0.05),
		})

		post := &model.Post{
			UserId:   order.UserId,
			Message:  "Оплата банковской картой № транзакции " + strconv.FormatInt(order.CreateAt, 10),
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

	if result := <-a.Srv.Store.Order().SetOrderCancel(order.Id); result.Err != nil {
		return result.Err
	} else {

		a.DeductionTransaction(&model.Transaction{
			OrderId:     order.Id,
			UserId:      order.UserId,
			Value:       math.Round(order.Price * 0.05),
			Description: "Отмена транзакции № " + strconv.FormatInt(order.CreateAt, 10),
		})

		post := &model.Post{
			UserId:   order.UserId,
			Message:  "Отмена транзакции № " + strconv.FormatInt(order.CreateAt, 10),
			CreateAt: model.GetMillis() + 1,
			Type:     model.POST_WITH_TRANSACTION,
		}

		a.CreatePostWithTransaction(post, false)

		return nil
	}
}
