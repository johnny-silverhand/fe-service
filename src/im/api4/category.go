package api4

import (
	"fmt"
	"im/model"
	"net/http"
)

func (api *API) InitCategory() {
	api.BaseRoutes.Categories.Handle("", api.ApiHandler(getAllCategories)).Methods("GET")
	api.BaseRoutes.Categories.Handle("", api.ApiHandler(createCategory)).Methods("POST")
	api.BaseRoutes.Category.Handle("", api.ApiHandler(getCategory)).Methods("GET")
	api.BaseRoutes.Category.Handle("", api.ApiHandler(updateCategory)).Methods("PUT")
	api.BaseRoutes.Category.Handle("", api.ApiHandler(deleteCategory)).Methods("DELETE")
	api.BaseRoutes.MoveCategory.Handle("", api.ApiHandler(moveCategory)).Methods("PUT")
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
		err = c.App.MoveClientCategoryBySp(storedCategory)
	} else {
		c.App.DeleteOneCategory(storedCategory)
		_, err = c.App.CreateCategoryBySp(storedCategory)
	}
}

func getCategory(c *Context, w http.ResponseWriter, r *http.Request) {
	c.RequireCategoryId()
	if c.Err != nil {
		return
	}
	if categories, err := c.App.GetCategoryPath(c.Params.CategoryId); err != nil {
		c.Err = err
		return
	} else {
		w.Write([]byte(model.CategoriesAllToJson(categories)))
	}
}

func getAllCategories(c *Context, w http.ResponseWriter, r *http.Request) {
	if categories, err := c.App.GetAllCategories(); err != nil {
		c.Err = err
		return
	} else {
		w.Write([]byte(model.CategoriesAllToJson(categories)))
	}
}

func getCategories(c *Context, w http.ResponseWriter, r *http.Request) {
	categories, err := c.App.GetCategoriesPage(0, c.Params.PerPage)
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
	result, err := c.App.CreateCategoryBySp(category)
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
	if category, err := c.App.GetCategory(c.Params.CategoryId); err != nil {
		c.Err = err
		return
	} else {
		if r := c.App.DeleteOneCategory(category); r != nil {
			c.Err = r
			return
		}
	}

	/*
		if err != nil {
			c.Err = err
			ReturnStatusOK(w)
		}
	*/
	ReturnStatusOK(w)

	//w.Write([]byte(model.MapToJson(map[string]string{"status": strconv.Itoa(result["status"])})))
}
