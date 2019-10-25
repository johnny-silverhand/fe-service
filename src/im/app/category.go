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
	return category, nil
}
