package api4

import (
	"fmt"
	"im/mlog"
	"im/model"
	"im/utils"
	"net/http"
)

func (api *API) InitProduct() {
	api.BaseRoutes.Products.Handle("/status", api.ApiHandler(updateProductsStatuses)).Methods("PUT")
	api.BaseRoutes.Products.Handle("/{product_id:[A-Za-z0-9]+}", api.ApiHandler(updateProduct)).Methods("PUT")

	api.BaseRoutes.Products.Handle("/extra", api.ApiHandler(getExtraProducts)).Methods("GET")
	api.BaseRoutes.Products.Handle("", api.ApiHandler(getProducts)).Methods("GET")

	api.BaseRoutes.Products.Handle("/search", api.ApiHandler(searchProducts)).Methods("POST")
	api.BaseRoutes.Products.Handle("", api.ApiHandler(createProduct)).Methods("POST")

	api.BaseRoutes.Products.Handle("/{product_id:[A-Za-z0-9_-]+}", api.ApiHandler(getProduct)).Methods("GET")

	api.BaseRoutes.Product.Handle("", api.ApiHandler(deleteProduct)).Methods("DELETE")

	api.BaseRoutes.Product.Handle("/status", api.ApiHandler(updateProductStatus)).Methods("PUT")

	api.BaseRoutes.ProductsForCategory.Handle("", api.ApiHandler(getProductsForCategory)).Methods("GET")
}

func updateProductsStatuses(c *Context, w http.ResponseWriter, r *http.Request) {

	if c.Err != nil {
		return
	}

	status := model.ProductStatusFromJson(r.Body)
	if status == nil {
		c.SetInvalidParam("status")
		return
	}

	// The user being updated in the payload must be the same one as indicated in the URL.
	if len(status.ProductIds) == 0 {
		c.SetInvalidParam("product_ids")
		return
	}

	//product, err := c.App.GetProduct(c.Params.ProductId)
	/*if err == nil && product.Status == model.STATUS_OUT_OF_OFFICE && status.Status != model.STATUS_OUT_OF_OFFICE {
		//c.App.DisableAutoResponder(c.Params.UserId, c.IsSystemAdmin())
	}*/

	c.App.Srv.Go(func() {
		for _, productId := range status.ProductIds {
			if _, err := c.App.UpdateProductStatus(productId, status); err != nil {
				mlog.Warn(fmt.Sprintf("Failed to update Product Status %v", err.Error()))
			}
		}
	})

	ReturnStatusOK(w)
}

func updateProductStatus(c *Context, w http.ResponseWriter, r *http.Request) {
	c.RequireProductId()
	if c.Err != nil {
		return
	}

	status := model.ProductStatusFromJson(r.Body)
	if status == nil {
		c.SetInvalidParam("status")
		return
	}

	// The user being updated in the payload must be the same one as indicated in the URL.
	if status.ProductId != c.Params.ProductId {
		c.SetInvalidParam("product_id")
		return
	}

	//product, err := c.App.GetProduct(c.Params.ProductId)
	/*if err == nil && product.Status == model.STATUS_OUT_OF_OFFICE && status.Status != model.STATUS_OUT_OF_OFFICE {
		//c.App.DisableAutoResponder(c.Params.UserId, c.IsSystemAdmin())
	}*/

	if product, err := c.App.UpdateProductStatus(c.Params.ProductId, status); err != nil {
		c.Err = err
		return
	} else {
		w.Write([]byte(product.ToJson()))
	}
}

func deleteProduct(c *Context, w http.ResponseWriter, r *http.Request) {
	c.RequireProductId()
	if c.Err != nil {
		return
	}

	if _, err := c.App.GetProduct(c.Params.ProductId); err != nil {
		// TODO CHANGE TO PERMISSION_DELETE_PRODUCT
		c.SetPermissionError(model.PERMISSION_DELETE_POST)
		return
	}

	if _, err := c.App.DeleteProduct(c.Params.ProductId, c.App.Session.UserId); err != nil {
		c.Err = err
		return
	}

	ReturnStatusOK(w)
}

func searchProducts(c *Context, w http.ResponseWriter, r *http.Request) {
	if c.Err != nil {
		return
	}
	params := model.ProductSearchFromJson(r.Body)
	if params.Terms == nil || len(*params.Terms) == 0 {
		c.SetInvalidParam("terms")
		return
	}
	terms := *params.Terms

	var categoryId string = ""
	if params.CategoryId != nil {
		categoryId = *params.CategoryId
	}

	timeZoneOffset := 0
	if params.TimeZoneOffset != nil {
		timeZoneOffset = *params.TimeZoneOffset
	}

	page := 0
	if params.Page != nil {
		page = *params.Page
	}

	perPage := 60
	if params.PerPage != nil {
		perPage = *params.PerPage
	}

	if results, err := c.App.SearchProducts(terms, categoryId, timeZoneOffset, page, perPage); err != nil {
		var pl model.ProductList
		pl.MakeNonNil()
		w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
		w.Write([]byte(pl.ToJson()))

		/*c.Err = err
		return*/
	} else {
		w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
		w.Write([]byte(results.ToJson()))
	}
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
	//c.RequireCategoryId()
	/*if c.Err != nil {
		return
	}
	productGetOptions := &model.ProductGetOptions{
		CategoryId: c.Params.CategoryId,
		OfficeId:   c.Params.OfficeId,
		AppId:      c.Params.AppId,
		Status:     c.Params.Status,
	}
	if len(productGetOptions.AppId) == 0 {
		productGetOptions.AppId = c.App.Session.AppId
	}
	if products, err := c.App.GetProductsPage(c.Params.Page, c.Params.PerPage, productGetOptions); err != nil {
		c.Err = err
		return
	} else {
		w.Write([]byte(products.ToJson()))
	}*/

	c.RequireAppId()

	if c.Err != nil {
		return
	}

	productGetOptions := &model.ProductGetOptions{
		AppId:      c.Params.AppId,
		CategoryId: c.Params.CategoryId,
		OfficeId:   c.Params.OfficeId,
		Status:     c.Params.Status,
	}

	if active := r.URL.Query().Get("active"); active != "" {
		productGetOptions.Active = &c.Params.Active
	}

	if utils.StringInSlice(c.App.Session.Roles, []string{model.CHANNEL_USER_ROLE_ID, ""}) {
		productGetOptions.Status = model.PRODUCT_STATUS_ACCEPTED
		productGetOptions.Active = model.NewBool(true)
	}

	var list *model.ProductList
	var err *model.AppError
	etag := ""

	list, err = c.App.GetProductsPage(c.Params.Page, c.Params.PerPage, productGetOptions)

	if err != nil {
		c.Err = err
		return
	}

	if len(etag) > 0 {
		w.Header().Set(model.HEADER_ETAG_SERVER, etag)
	}

	w.Write([]byte(c.App.PrepareProductListForClient(list).ToJson()))
}

func getExtraProducts(c *Context, w http.ResponseWriter, r *http.Request) {
	//c.RequireCategoryId()
	if c.Err != nil {
		return
	}
	productGetOptions := &model.ProductGetOptions{
		AppId: c.Params.AppId,
		Extra: true,
	}
	if len(productGetOptions.AppId) == 0 {
		productGetOptions.AppId = c.App.Session.AppId
	}
	if active := r.URL.Query().Get("active"); active != "" {
		productGetOptions.Active = &c.Params.Active
	}
	if products, err := c.App.GetProductsPage(c.Params.Page, c.Params.PerPage, productGetOptions); err != nil {
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

	productGetOptions := &model.ProductGetOptions{
		AppId:      c.Params.AppId,
		CategoryId: c.Params.CategoryId,
		OfficeId:   c.Params.OfficeId,
		Status:     c.Params.Status,
	}

	if active := r.URL.Query().Get("active"); active != "" {
		productGetOptions.Active = &c.Params.Active
	}

	if utils.StringInSlice(c.App.Session.Roles, []string{model.CHANNEL_USER_ROLE_ID, ""}) {
		productGetOptions.Status = model.PRODUCT_STATUS_ACCEPTED
		productGetOptions.Active = model.NewBool(true)
	}

	var list *model.ProductList
	var err *model.AppError
	etag := ""

	list, err = c.App.GetProductsPage(c.Params.Page, c.Params.PerPage, productGetOptions)

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

	patch := model.ProductPatchFromJson(r.Body)

	if patch == nil {
		c.SetInvalidParam("product")
		return
	}

	rproduct, err := c.App.UpdateProduct(c.Params.ProductId, patch, false)
	if err != nil {
		c.Err = err
		return
	}

	w.Write([]byte(rproduct.ToJson()))
}
