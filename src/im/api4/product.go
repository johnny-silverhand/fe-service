package api4

import (
	"im/model"
	"net/http"
)

func (api *API) InitProduct() {
	api.BaseRoutes.Product.Handle("", api.ApiHandler(getProduct)).Methods("GET")
	api.BaseRoutes.Products.Handle("", api.ApiHandler(getProducts)).Methods("GET")
	api.BaseRoutes.Products.Handle("", api.ApiHandler(createProduct)).Methods("POST")
	api.BaseRoutes.Product.Handle("", api.ApiHandler(updateProduct)).Methods("PUT")
	api.BaseRoutes.ProductsForCategory.Handle("", api.ApiHandler(getProductsForCategory)).Methods("GET")
}

func getProduct(c *Context, w http.ResponseWriter, r *http.Request) {
	c.RequireProductId()
	if c.Err != nil {
		return
	}
	product, err := c.App.GetProduct(c.Params.ProductId)

	if err != nil {
		c.Err = err
		return
	}

	w.Write([]byte(product.ToJson()))
}

func getProducts(c *Context, w http.ResponseWriter, r *http.Request) {
	c.RequireCategoryId()
	if c.Err != nil {
		return
	}

	if products, err := c.App.GetProductsPage(c.Params.Page, c.Params.PerPage, c.Params.Sort, c.Params.CategoryId); err != nil {
		c.Err = err
		return
	} else {
		w.Write([]byte(products.ToJson()))
	}
}

func createProduct(c *Context, w http.ResponseWriter, r *http.Request) {

	product := model.ProductFromJson(r.Body)

	if product == nil {
		c.SetInvalidParam("product")
		return
	}

	result, err := c.App.CreateProduct(product)
	if err != nil {
		c.Err = err
		return
	}
	w.Write([]byte(result.ToJson()))
}

func getProductsForCategory(c *Context, w http.ResponseWriter, r *http.Request) {
	c.RequireCategoryId()
	if c.Err != nil {
		return
	}

	//afterProduct := r.URL.Query().Get("after")
	//beforeProduct := r.URL.Query().Get("before")
	//sinceString := r.URL.Query().Get("since")

	//var since int64
	//var parseError error

	/*	if len(sinceString) > 0 {
		since, parseError = strconv.ParseInt(sinceString, 10, 64)
		if parseError != nil {
			c.SetInvalidParam("since")
			return
		}
	}*/

	var list *model.ProductList
	var err *model.AppError
	etag := ""

	/*val := reflect.ValueOf(model.Product{})
	for i := 0; i < val.Type().NumField(); i++ {
		// prints empty line if there is no json tag for the field
		if ()(val.Type().Field(i).Tag.Get("json"))
	}*/

	list, err = c.App.GetProductsPage(c.Params.Page, c.Params.PerPage, c.Params.Sort, c.Params.CategoryId)
	/*if since > 0 {
		list, err = c.App.GetProductsSince(c.Params.ChannelId, since)
	} else if len(afterProduct) > 0 {
		etag = c.App.GetProductsEtag(c.Params.ChannelId)

		if c.HandleEtag(etag, "Get Products After", w, r) {
			return
		}

		list, err = c.App.GetProductsAfterProduct(c.Params.ChannelId, afterProduct, c.Params.Page, c.Params.PerPage)
	} else if len(beforeProduct) > 0 {
		etag = c.App.GetProductsEtag(c.Params.ChannelId)

		if c.HandleEtag(etag, "Get Products Before", w, r) {
			return
		}

		list, err = c.App.GetProductsBeforeProduct(c.Params.ChannelId, beforeProduct, c.Params.Page, c.Params.PerPage)
	} else {
		etag = c.App.GetProductsEtag(c.Params.ChannelId)

		if c.HandleEtag(etag, "Get Products", w, r) {
			return
		}


	}*/

	if err != nil {
		c.Err = err
		return
	}

	if len(etag) > 0 {
		w.Header().Set(model.HEADER_ETAG_SERVER, etag)
	}

	w.Write([]byte(c.App.PrepareProductListForClient(list).ToJson()))
}

func updateProduct(c *Context, w http.ResponseWriter, r *http.Request) {
	c.RequireProductId()
	if c.Err != nil {
		return
	}

	product := model.ProductFromJson(r.Body)

	if product == nil {
		c.SetInvalidParam("product")
		return
	}

	// The post being updated in the payload must be the same one as indicated in the URL.
	if product.Id != c.Params.ProductId {
		c.SetInvalidParam("id")
		return
	}

	// Updating the file_ids of a post is not a supported operation and will be ignored

	product.MoreImageIds = nil

	/*	if !c.App.SessionHasPermissionToChannelByPost(c.App.Session, c.Params.PostId, model.PERMISSION_EDIT_POST) {
		c.SetPermissionError(model.PERMISSION_EDIT_POST)
		return
	}*/

	/*originalProduct, err := c.App.GetSingleProduct(c.Params.ProductId)
	if err != nil {
		c.SetPermissionError(model.PERMISSION_EDIT_POST)
		return
	}*/

	product.Id = c.Params.ProductId

	rproduct, err := c.App.UpdateProduct(product, false)
	if err != nil {
		c.Err = err
		return
	}

	w.Write([]byte(rproduct.ToJson()))
}
