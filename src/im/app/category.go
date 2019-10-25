package app

import (
	"im/model"
)


func (a *App) GetSingleCategory(categoryId string) (*model.Post, *model.AppError) {
	//todo single
	result := <-a.Srv.Store.Category().Get(categoryId)
	if result.Err != nil {
		return nil, result.Err
	}
	return result.Data.(*model.Post), nil
}

func (a *App) GetCategory(categoryId string) (*model.Category, *model.AppError) {
	result := <-a.Srv.Store.Category().Get(categoryId)
	if result.Err != nil {
		return nil, result.Err
	}

	/*	rcat := result.Data.(*model.Category)
		if (rcat.CountChildren > 0) {
			result := <-a.Srv.Store.Category().GetWithChildren(rcat.Id)
			if result.Err == nil {
				rcat.Children = result.Data.([]*model.Category)
			}
		}*/
	return result.Data.(*model.Category), nil
}

func (a *App) GetCategoriesPage(page int, perPage int) ([]*model.Category, *model.AppError) {
	return a.GetCategories(page*perPage, perPage)
}

func (a *App) GetCategories(offset int, limit int) ([]*model.Category, *model.AppError) {
	result := <-a.Srv.Store.Category().GetAllPage(offset, limit)
	if result.Err != nil {
		return nil, result.Err
	}
	return result.Data.([]*model.Category), nil
}

func (a *App) CreateCategory(category *model.Category) (*model.Category, *model.AppError) {
	result := <-a.Srv.Store.Category().Save(category)
	if result.Err != nil {
		return nil, result.Err
	}

	return result.Data.(*model.Category), nil
}

func (a *App) DeleteCategory(category *model.Category) (map[string]int, *model.AppError) {
	result := <-a.Srv.Store.Category().Delete(category)
	if result.Err != nil {
		return nil, result.Err
	}
	descendants, _ := a.GetDescendants(category)
	for _, descendant := range descendants {
		a.DeleteCategory(descendant)
	}
	return result.Data.(map[string]int), nil
}

func (a *App) GetDescendants(category *model.Category) ([]*model.Category, *model.AppError) {
	result := <-a.Srv.Store.Category().GetDescendants(category)
	if result.Err != nil {
		return nil, result.Err
	}
	return result.Data.([]*model.Category), nil
}

func (a *App) UpdateCategory(category *model.Category, safeUpdate bool) (*model.Category, *model.AppError) {
	//category.SanitizeProps()

	/*
	result := <-a.Srv.Store.Category().Get(category.Id)
	if result.Err != nil {
		return nil, result.Err
	}

	oldCategory := result.Data.(*model.Category)

	if oldCategory == nil {
		err := model.NewAppError("UpdateCategory", "api.category.update_category.find.app_error", nil, "id="+category.Id, http.StatusBadRequest)
		return nil, err
	}

	if oldCategory.DeleteAt != 0 {
		err := model.NewAppError("UpdateCategory", "api.category.update_category.permissions_details.app_error", map[string]interface{}{"CategoryId": category.Id}, "", http.StatusBadRequest)
		return nil, err
	}

	newCategory := &model.Category{}
	*newCategory = *oldCategory

	if newCategory.Name != category.Name {
		newCategory.Name = category.Name
	}
	if newCategory.ParentId != category.ParentId {
		newCategory.ParentId = category.ParentId
	}
	if newCategory.Depth != category.Depth {
		newCategory.Depth = category.Depth
	}
	if newCategory.UpdateAt != category.UpdateAt {
		newCategory.UpdateAt = category.UpdateAt
	}

	/*if !safeUpdate {

	}*/

	/*
	result = <-a.Srv.Store.Category().Update(newCategory)
	if result.Err != nil {
		return nil, result.Err
	}


	rcategory := result.Data.(*model.Category)
	*/

	//a.InvalidateCacheForChannelCategorys(rcategory.ChannelId)

	return category, nil
}
