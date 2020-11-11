package app

import (
	"im/model"
	"im/store"
	"net/http"
)

func (a *App) GetExtra(extraId string) (*model.Extra, *model.AppError) {

	result := <-a.Srv.Store.Extra().Get(extraId)
	if result.Err != nil {
		return nil, result.Err
	}

	rextra := result.Data.(*model.Extra)

	rextra = a.PrepareExtraForClient(rextra, false)

	return rextra, nil
}

func (a *App) GetExtrasPage(page int, perPage int, sort string) (*model.ExtraList, *model.AppError) {
	return a.GetExtras(page*perPage, perPage, sort)
}

func (a *App) GetExtras(offset int, limit int, sort string) (*model.ExtraList, *model.AppError) {

	result := <-a.Srv.Store.Extra().GetAllPage(offset, limit, model.GetOrder(sort))

	if result.Err != nil {
		return nil, result.Err
	}

	return result.Data.(*model.ExtraList), nil
}

func (a *App) CreateExtra(extra *model.Extra) (*model.Extra, *model.AppError) {

	result := <-a.Srv.Store.Extra().Save(extra)
	if result.Err != nil {
		return nil, result.Err
	}

	rextra := result.Data.(*model.Extra)

	return rextra, nil
}

func (a *App) UpdateExtra(extra *model.Extra, safeUpdate bool) (*model.Extra, *model.AppError) {
	//extra.SanitizeProps()

	result := <-a.Srv.Store.Extra().Get(extra.Id)
	if result.Err != nil {
		return nil, result.Err
	}

	oldExtra := result.Data.(*model.Extra)

	if oldExtra == nil {
		err := model.NewAppError("UpdateExtra", "api.extra.update_extra.find.app_error", nil, "id="+extra.Id, http.StatusBadRequest)
		return nil, err
	}

	if oldExtra.DeleteAt != 0 {
		err := model.NewAppError("UpdateExtra", "api.extra.update_extra.permissions_details.app_error", map[string]interface{}{"ExtraId": extra.Id}, "", http.StatusBadRequest)
		return nil, err
	}

	newExtra := &model.Extra{}
	*newExtra = *oldExtra

	result = <-a.Srv.Store.Extra().Update(newExtra)
	if result.Err != nil {
		return nil, result.Err
	}

	rextra := result.Data.(*model.Extra)
	rextra = a.PrepareExtraForClient(rextra, false)

	//a.InvalidateCacheForChannelExtras(rextra.ChannelId)

	return rextra, nil
}

func (a *App) PrepareExtraForClient(originalExtra *model.Extra, isNewExtra bool) *model.Extra {
	extra := originalExtra.Clone()

	//extra.Metadata.Images = a.getCategoryForExtra(extra)

	return extra
}

func (a *App) PrepareExtraListForClient(originalList *model.ExtraList) *model.ExtraList {
	list := &model.ExtraList{
		Extras: make(map[string]*model.Extra, len(originalList.Extras)),
		Order:  originalList.Order, // Note that this uses the original Order array, so it isn't a deep copy
	}

	for id, originalExtra := range originalList.Extras {
		extra := a.PrepareExtraForClient(originalExtra, false)

		list.Extras[id] = extra
	}

	return list
}

func (a *App) DeleteExtra(extraId, deleteByID string) (*model.Extra, *model.AppError) {
	result := <-a.Srv.Store.Extra().Get(extraId)
	if result.Err != nil {
		result.Err.StatusCode = http.StatusBadRequest
		return nil, result.Err
	}
	extra := result.Data.(*model.Extra)

	if result := <-a.Srv.Store.Extra().Delete(extraId, model.GetMillis(), deleteByID); result.Err != nil {
		return nil, result.Err
	}

	return extra, nil
}

func (a *App) GetAllExtrasBeforeExtra(extraId string, page, perPage int) (*model.ExtraList, *model.AppError) {

	if result := <-a.Srv.Store.Extra().GetAllExtrasBefore(extraId, perPage, page*perPage); result.Err != nil {
		return nil, result.Err
	} else {
		return result.Data.(*model.ExtraList), nil
	}
}

func (a *App) GetAllExtrasAfterExtra(extraId string, page, perPage int) (*model.ExtraList, *model.AppError) {

	if result := <-a.Srv.Store.Extra().GetAllExtrasAfter(extraId, perPage, page*perPage); result.Err != nil {
		return nil, result.Err
	} else {
		return result.Data.(*model.ExtraList), nil
	}
}

func (a *App) GetAllExtrasAroundExtra(extraId string, offset, limit int, before bool) (*model.ExtraList, *model.AppError) {
	var pchan store.StoreChannel

	if before {
		pchan = a.Srv.Store.Extra().GetAllExtrasBefore(extraId, limit, offset)
	} else {
		pchan = a.Srv.Store.Extra().GetAllExtrasAfter(extraId, limit, offset)
	}

	if result := <-pchan; result.Err != nil {
		return nil, result.Err
	} else {
		return result.Data.(*model.ExtraList), nil
	}
}

func (a *App) GetAllExtrasSince(time int64) (*model.ExtraList, *model.AppError) {
	if result := <-a.Srv.Store.Extra().GetAllExtrasSince(time, true); result.Err != nil {
		return nil, result.Err
	} else {
		return result.Data.(*model.ExtraList), nil
	}
}

func (a *App) GetAllExtrasPage(page int, perPage int) (*model.ExtraList, *model.AppError) {
	if result := <-a.Srv.Store.Extra().GetAllExtras(page*perPage, perPage, true); result.Err != nil {
		return nil, result.Err
	} else {
		return result.Data.(*model.ExtraList), nil
	}
}

func (a *App) GetExtrasBasket(productIds []string) (*model.ProductList, *model.AppError) {

	result := <-a.Srv.Store.Extra().GetExtraProductsByIds(productIds, false)

	if result.Err != nil {
		return nil, result.Err
	}

	return result.Data.(*model.ProductList), nil
}

func (a *App) GetUpdatedBasket(productIds []string) (*model.ProductList, *model.AppError) {

	result := <-a.Srv.Store.Product().GetProductsByIds(productIds, false)

	if result.Err != nil {
		return nil, result.Err
	}
	rproducts := result.Data.([]*model.Product)

	list := model.NewProductList()
	for _, p := range rproducts {
		list.AddProduct(p)
		list.AddOrder(p.Id)
	}
	list.MakeNonNil()

	return list, nil
}
