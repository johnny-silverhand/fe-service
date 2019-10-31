package app

import (
	"im/model"
	"net/http"
)

func (a *App) GetSingleCategory(categoryId string) (*model.Post, *model.AppError) {
	//todo single
	result := <-a.Srv.Store.Category().Get(categoryId)
	if result.Err != nil {
		return nil, result.Err
	}
	return result.Data.(*model.Post), nil
}

func (a *App) GetCategoriesByClientIdPage(clientId string, page int, perPage int) ([]*model.Category, *model.AppError) {
	return a.GetCategoriesByClientId(clientId, page*perPage, perPage)
}
func (a *App) GetCategoriesByClientId(clientId string, offset int, limit int) ([]*model.Category, *model.AppError) {
	result := <-a.Srv.Store.Category().GetAllByClientIdPage(clientId, offset, limit)
	if result.Err != nil {
		return nil, result.Err
	}
	return result.Data.([]*model.Category), nil
}

func (a *App) GetCategory(categoryId string) (*model.Category, *model.AppError) {
	result := <-a.Srv.Store.Category().Get(categoryId)
	if result.Err != nil {
		return nil, result.Err
	}

	return result.Data.(*model.Category), nil
}

func (a *App) GetAllCategories() ([]*model.Category, *model.AppError) {
	result := <-a.Srv.Store.Category().GetAll()
	if result.Err != nil {
		return nil, result.Err
	}
	return result.Data.([]*model.Category), nil
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
	result := <-a.Srv.Store.Category().CreateCategoryBySp(category)
	if result.Err != nil {
		return nil, result.Err
	}
	return result.Data.(*model.Category), nil
}

func (a *App) CreateCategoryBySp(category *model.Category) (*model.Category, *model.AppError) {
	result := <-a.Srv.Store.Category().CreateCategoryBySp(category)
	if result.Err != nil {
		return nil, result.Err
	}
	return result.Data.(*model.Category), nil
}

func (a *App) DeleteOneCategory(category *model.Category) *model.AppError {
	result := <-a.Srv.Store.Category().DeleteCategoryBySp(category)
	return result.Err
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
	return nil, nil
}

func (a *App) GetDescendants(category *model.Category) ([]*model.Category, *model.AppError) {
	result := <-a.Srv.Store.Category().GetDescendants(category)
	if result.Err != nil {
		return nil, result.Err
	}
	return result.Data.([]*model.Category), nil
}

func (a *App) MoveClientCategory(category *model.Category, parentCategory *model.Category) *model.AppError {
	result := <-a.Srv.Store.Category().MoveCategoryBySp(category)
	return result.Err
}

func (a *App) MoveClientCategoryBySp(category *model.Category) *model.AppError {
	result := <-a.Srv.Store.Category().MoveCategoryBySp(category)
	return result.Err
}

func (a *App) OrderCategoryBySp(category *model.Category, destinationId string) *model.AppError {
	result := <-a.Srv.Store.Category().OrderCategoryBySp(category, destinationId)
	return result.Err
}

func (a *App) UpdateCategory(category *model.Category, safeUpdate bool) (*model.Category, *model.AppError) {
	//category.SanitizeProps()

	result := <-a.Srv.Store.Category().Get(category.Id)
	if result.Err != nil {
		return nil, result.Err
	}

	oldCategory := result.Data.(*model.Category)

	if oldCategory == nil {
		err := model.NewAppError("UpdateCategory",
			"api.category.update_category.find.app_error", nil,
			"id="+category.Id, http.StatusBadRequest)
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

	result = <-a.Srv.Store.Category().Update(newCategory)

	if result.Err != nil {
		return nil, result.Err
	}

	payload := result.Data.(*model.Category)

	//a.InvalidateCacheForChannelCategorys(payload.ChannelId)

	return payload, nil
}

func (a *App) GetCategoryPath(categoryId string) ([]*model.Category, *model.AppError) {
	result := <-a.Srv.Store.Category().GetCategoryPath(categoryId)
	if result.Err != nil {
		return nil, result.Err
	}
	return result.Data.([]*model.Category), nil
}
