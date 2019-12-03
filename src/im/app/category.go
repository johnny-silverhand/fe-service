package app

import (
	"im/model"
	"net/http"
)

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
	result := <-a.Srv.Store.Category().Create(category)
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
	return nil, nil
}

func (a *App) MoveCategory(category *model.Category) *model.AppError {
	result := <-a.Srv.Store.Category().Move(category)
	return result.Err
}

func (a *App) OrderCategory(category *model.Category) *model.AppError {
	result := <-a.Srv.Store.Category().Order(category)
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

func (a *App) GetCategoriesByIds(categoryIds []string) ([]*model.Category, *model.AppError) {
	result := <-a.Srv.Store.Category().GetCategoriesByIds(categoryIds)
	if result.Err != nil {
		return nil, result.Err
	}
	return result.Data.([]*model.Category), nil
}
