package api4

import (
	"im/model"
	"net/http"
)

func (api *API) InitBasket() {

	api.BaseRoutes.Basket.Handle("/extras", api.ApiHandler(getExtraBasket)).Methods("POST")
	api.BaseRoutes.Basket.Handle("/limits", api.ApiHandler(getDiscountLimits)).Methods("POST")
	api.BaseRoutes.Basket.Handle("/update", api.ApiHandler(getUpdatedBasket)).Methods("POST")

}

func getUpdatedBasket(c *Context, w http.ResponseWriter, r *http.Request) {
	productIds := model.ArrayFromJson(r.Body)

	if len(productIds) == 0 {
		c.SetInvalidParam("product_ids")
		return
	}

	result, err := c.App.GetUpdatedBasket(productIds)
	if err != nil {
		c.Err = err
		return
	}
	w.Write([]byte(result.ToJson()))
}

func getExtraBasket(c *Context, w http.ResponseWriter, r *http.Request) {
	productIds := model.ArrayFromJson(r.Body)

	if len(productIds) == 0 {
		c.SetInvalidParam("product_ids")
		return
	}

	result, err := c.App.GetExtrasBasket(productIds)
	if err != nil {
		c.Err = err
		return
	}
	w.Write([]byte(result.ToJson()))
}

func getDiscountLimits(c *Context, w http.ResponseWriter, r *http.Request) {
	productIds := model.ArrayFromJson(r.Body)

	if len(productIds) == 0 {
		c.SetInvalidParam("product_ids")
		return
	}

	discountLimits, err := c.App.GetDiscountLimits(productIds)
	if err != nil {
		c.Err = err
		return
	}

	w.Write([]byte(discountLimits.ToJson()))
}
