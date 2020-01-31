package app

import "im/model"

func (a *App) AddBasketItem(item *model.Basket) (*model.Basket, *model.AppError) {

	result := <-a.Srv.Store.Basket().Save(item)
	if result.Err != nil {
		return nil, result.Err
	}

	ritem := result.Data.(*model.Basket)

	return ritem, nil
}

func (a *App) GetBasketForOrder(order *model.Order) []*model.Basket {
	var basket []*model.Basket

	result := <-a.Srv.Store.Basket().GetByOrderId(order.Id)
	if result.Err == nil {
		basket = result.Data.([]*model.Basket)
	}

	return basket
}

func (a *App) PrepareBasketForClient(originalBasket *model.Basket, isNewBasket bool) *model.Basket {
	basket := originalBasket.Clone()

	if product, err := a.GetProduct(basket.ProductId); err == nil {
		basket.Product = product
	}

	if isNewBasket && basket.Product != nil {
		basket.DiscountValue = int64(basket.Product.Cashback)
	}

	return basket
}
