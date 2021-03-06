package api4

import (
	"fmt"
	"im/model"
	"net/http"
)

func (api *API) InitCategory() {
	api.BaseRoutes.Category.Handle("", api.ApiHandler(getCategory)).Methods("GET")
	api.BaseRoutes.Category.Handle("/path", api.ApiHandler(getCategoryPath)).Methods("GET")
	api.BaseRoutes.Categories.Handle("", api.ApiHandler(getCategories)).Methods("GET")
	api.BaseRoutes.Categories.Handle("", api.ApiHandler(createCategory)).Methods("POST")
	api.BaseRoutes.Category.Handle("", api.ApiHandler(updateCategory)).Methods("PUT")
	api.BaseRoutes.Category.Handle("/move", api.ApiHandler(moveCategory)).Methods("PUT")
	api.BaseRoutes.Category.Handle("/order", api.ApiHandler(orderCategory)).Methods("PUT")
	api.BaseRoutes.Category.Handle("", api.ApiHandler(deleteCategory)).Methods("DELETE")
	api.BaseRoutes.Categories.Handle("/search", api.ApiHandler(searchCategories)).Methods("POST")
}

func searchCategories(c *Context, w http.ResponseWriter, r *http.Request) {
	categoryIds := model.ArrayFromJson(r.Body)
	if len(categoryIds) == 0 {
		c.SetInvalidParam("category_ids")
		return
	}
	if categories, err := c.App.GetCategoriesByIds(categoryIds); err != nil {
		c.Err = err
		return
	} else {
		w.Write([]byte(model.CategoriesToJson(categories)))
	}
}

func moveCategory(c *Context, w http.ResponseWriter, r *http.Request) {
	c.RequireCategoryId()
	if c.Err != nil {
		return
	}
	category := model.CategoryFromJson(r.Body)
	storedCategory, err := c.App.GetCategory(category.Id)
	if err != nil {
		return
	}

	storedCategory.ParentId = category.ParentId
	if len(category.ParentId) > 0 {
		err = c.App.MoveCategory(storedCategory)
	} else {
		_, err = c.App.DeleteCategory(storedCategory)
		_, err = c.App.CreateCategory(storedCategory)
	}

	if err == nil {
		ReturnStatusOK(w)
	}
}

func orderCategory(c *Context, w http.ResponseWriter, r *http.Request) {
	c.RequireCategoryId()
	if c.Err != nil {
		return
	}
	category := model.CategoryFromJson(r.Body)
	storedCategory, err := c.App.GetCategory(category.Id)
	if err != nil {
		return
	}

	storedCategory.ParentId = category.ParentId
	if len(category.DestinationId) > 0 {
		storedCategory.DestinationId = category.DestinationId
		err = c.App.OrderCategory(storedCategory)
	}

	if err == nil {
		ReturnStatusOK(w)
	}

}

func getCategoryPath(c *Context, w http.ResponseWriter, r *http.Request) {
	c.RequireCategoryId()
	if c.Err != nil {
		return
	}
	if categories, err := c.App.GetCategoryPath(c.Params.CategoryId); err != nil {
		c.Err = err
		return
	} else {

		// load products for categories
		/*	for i, category := range categories {
			if productList, err := c.App.GetProductsPage(c.Params.Page, c.Params.PerPage, c.Params.Sort, category.Id); err == nil {
				categories[i].ProductList = productList
			}
		}*/

		w.Write([]byte(model.CategoriesToJson(categories)))
	}
}

func getCategory(c *Context, w http.ResponseWriter, r *http.Request) {
	c.RequireCategoryId()

	if c.Err != nil {
		return
	}
	if category, err := c.App.GetCategory(c.Params.CategoryId); err != nil {
		c.Err = err
		return
	} else {
		productGetOptions := &model.ProductGetOptions{
			CategoryId: category.Id,
			OfficeId:   c.Params.OfficeId,
		}
		productList, _ := c.App.GetProductsPage(c.Params.Page, c.Params.PerPage, productGetOptions)
		category.ProductList = productList

		w.Write([]byte(category.ToJson()))
	}
}

func getCategories(c *Context, w http.ResponseWriter, r *http.Request) {
	categories, err := c.App.GetCategoriesPage(0, c.Params.PerPage, &c.Params.AppId)
	if err != nil {
		c.Err = err
		return
	}
	w.Write([]byte(model.CategoriesToJson(categories)))
}

func createCategory(c *Context, w http.ResponseWriter, r *http.Request) {
	category := model.CategoryFromJson(r.Body)
	if category == nil {
		c.Err = model.NewAppError("createCategory", "api.category", nil, "nil object", http.StatusForbidden)
		return
	}
	result, err := c.App.CreateCategory(category)
	if err != nil {
		c.Err = err
		return
	}
	w.Write([]byte(result.ToJson()))
}

func updateCategory(c *Context, w http.ResponseWriter, r *http.Request) {
	c.RequireCategoryId()
	if c.Err != nil {
		return
	}

	category := model.CategoryFromJson(r.Body)
	fmt.Print(category)
	if category == nil {
		c.SetInvalidParam("category")
		return
	}

	// The post being updated in the payload must be the same one as indicated in the URL.
	if category.Id != c.Params.CategoryId {
		c.SetInvalidParam("id")
		return
	}

	category.Id = c.Params.CategoryId

	rcategory, err := c.App.UpdateCategory(category, false)
	if err != nil {
		c.Err = err
		return
	}

	w.Write([]byte(rcategory.ToJson()))
}

func deleteCategory(c *Context, w http.ResponseWriter, r *http.Request) {
	category, err := c.App.GetCategory(c.Params.CategoryId)
	if err != nil {
		c.Err = err
		return
	}
	c.App.DeleteCategory(category)
	/*
		if err != nil {
			c.Err = err
			ReturnStatusOK(w)
		}
	*/
	ReturnStatusOK(w)

	//w.Write([]byte(model.MapToJson(map[string]string{"status": strconv.Itoa(result["status"])})))
}
