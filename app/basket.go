package app

import (
	"im/model"
	"math"
)

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

func (a *App) PrepareBasketListForClient(originalList []*model.Basket, isNewBasket bool) []*model.Basket {
	var list []*model.Basket
	for _, originalBasket := range originalList {
		basket := a.PrepareBasketForClient(originalBasket, isNewBasket)
		list = append(list, basket)
	}

	return list
}

func (a *App) PrepareBasketForClient(originalBasket *model.Basket, isNewBasket bool) *model.Basket {
	basket := originalBasket.Clone()

	if product, err := a.GetProduct(basket.ProductId); err == nil {
		basket.Product = product
	}

	if isNewBasket && basket.Product != nil {

		basket.Price = basket.Product.Price
		basket.Currency = basket.Product.Currency
		basket.Name = basket.Product.Name

		if application, err := a.GetApplication(basket.Product.AppId); err != nil || basket.Product.PrivateRule {
			basket.Cashback = basket.Product.Cashback
		} else {
			basket.Cashback = math.Floor(basket.Price * (application.Cashback / 100))
		}
	}

	return basket
}
