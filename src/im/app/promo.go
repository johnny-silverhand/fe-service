package app

import (
	"im/mlog"
	"im/model"
	"im/store"
	"net/http"
)

func (a *App) GetPromo(promoId string) (*model.Promo, *model.AppError) {

	result := <-a.Srv.Store.Promo().Get(promoId)
	if result.Err != nil {
		return nil, result.Err
	}

	rpromo := result.Data.(*model.Promo)

	rpromo = a.PreparePromoForClient(rpromo, false)

	return rpromo, nil
}

func (a *App) GetPromosPage(page int, perPage int, sort string) (*model.PromoList, *model.AppError) {
	return a.GetPromos(page*perPage, perPage, sort)
}

func (a *App) GetPromosPageByApp(page int, perPage int, sort string, appId string) (*model.PromoList, *model.AppError) {
	return a.GetPromosByApp(page*perPage, perPage, sort, appId)
}

func (a *App) GetPromosByApp(offset int, limit int, sort string, appId string) (*model.PromoList, *model.AppError) {

	result := <-a.Srv.Store.Promo().GetAllPageByApp(offset, limit, model.GetOrder(sort), appId)

	if result.Err != nil {
		return nil, result.Err
	}
	list := a.PreparePromoListForClient(result.Data.(*model.PromoList))
	return list, nil
}

func (a *App) GetPromosForModeration(options *model.PromoGetOptions) (*model.PromoList, *model.AppError) {

	result := <-a.Srv.Store.Promo().GetForModeration(options)

	if result.Err != nil {
		return nil, result.Err
	}

	list := a.PreparePromoListForClient(result.Data.(*model.PromoList))

	return list, nil
}

func (a *App) GetPromos(offset int, limit int, sort string) (*model.PromoList, *model.AppError) {

	result := <-a.Srv.Store.Promo().GetAllPage(offset, limit, model.GetOrder(sort))

	if result.Err != nil {
		return nil, result.Err
	}
	list := a.PreparePromoListForClient(result.Data.(*model.PromoList))
	return list, nil
}

func (a *App) CreatePromo(promo *model.Promo) (*model.Promo, *model.AppError) {

	result := <-a.Srv.Store.Promo().Save(promo)
	if result.Err != nil {
		return nil, result.Err
	}

	if len(promo.Media) > 0 {
		if err := a.attachMediaToPromo(promo); err != nil {
			mlog.Error("Encountered error attaching files to post", mlog.String("product_id", promo.Id), mlog.Any("files_ids", promo.FileIds), mlog.Err(result.Err))
		}
	}

	rpromo := result.Data.(*model.Promo)

	return rpromo, nil
}

func (a *App) attachMediaToPromo(promo *model.Promo) *model.AppError {
	var attachedIds []string
	for _, media := range promo.Media {
		result := <-a.Srv.Store.FileInfo().AttachTo(media.Id, promo.Id, model.METADATA_TYPE_PROMO)
		if result.Err != nil {
			mlog.Warn("Failed to attach file to post", mlog.String("file_id", media.Id), mlog.String("product_id", promo.Id), mlog.Err(result.Err))
			continue
		}

		attachedIds = append(attachedIds, media.Id)
	}

	return nil
}

func (a *App) UpdatePromo(promo *model.Promo, safeUpdate bool) (*model.Promo, *model.AppError) {
	//promo.SanitizeProps()

	result := <-a.Srv.Store.Promo().Get(promo.Id)
	if result.Err != nil {
		return nil, result.Err
	}

	oldPromo := result.Data.(*model.Promo)

	if oldPromo == nil {
		err := model.NewAppError("UpdatePromo", "api.promo.update_promo.find.app_error", nil, "id="+promo.Id, http.StatusBadRequest)
		return nil, err
	}

	if oldPromo.DeleteAt != 0 {
		err := model.NewAppError("UpdatePromo", "api.promo.update_promo.permissions_details.app_error", map[string]interface{}{"PromoId": promo.Id}, "", http.StatusBadRequest)
		return nil, err
	}

	newPromo := &model.Promo{}
	*newPromo = *oldPromo

	if newPromo.Name != promo.Name {
		newPromo.Name = promo.Name
	}

	newPromo.Preview = promo.Preview
	newPromo.Description = promo.Description
	newPromo.BeginAt = promo.BeginAt
	newPromo.ExpireAt = promo.ExpireAt
	//if !safeUpdate {
	newPromo.Media = promo.Media
	//}

	newPromo.Status = model.PROMO_STATUS_DRAFT

	a.deleteMediaFromPromo(oldPromo, newPromo)

	if len(newPromo.Media) > 0 {
		if err := a.attachMediaToPromo(newPromo); err != nil {
			mlog.Error("Encountered error attaching files to post", mlog.String("post_id", newPromo.Id), mlog.Any("file_ids", newPromo.FileIds), mlog.Err(result.Err))
		}

	}

	result = <-a.Srv.Store.Promo().Update(newPromo)
	if result.Err != nil {
		return nil, result.Err
	}

	rpromo := result.Data.(*model.Promo)

	rpromo = a.PreparePromoForClient(rpromo, false)

	//a.InvalidateCacheForChannelPromos(rpromo.ChannelId)

	return rpromo, nil
}

func (a *App) PreparePromoForClient(originalPromo *model.Promo, isNewPromo bool) *model.Promo {
	promo := originalPromo.Clone()

	if fileInfos, err := a.getMediaForPromo(originalPromo); err != nil {
		mlog.Warn("Failed to get files for a product", mlog.String("product_id", originalPromo.Id), mlog.Any("err", err))
	} else {
		promo.Media = fileInfos

	}
	//promo.Metadata.Images = a.getCategoryForPromo(promo)

	return promo
}

func (a *App) deleteMediaFromPromo(oldPromo, newPromo *model.Promo) *model.AppError {
	promo := a.PreparePromoForClient(oldPromo, false)

	var mediaIds []string
	var newIds []string
	var diff []string

	for _, media := range promo.Media {
		mediaIds = append(mediaIds, media.Id)
	}
	for _, media := range newPromo.Media {
		newIds = append(newIds, media.Id)
	}

	for _, s1 := range mediaIds {
		found := false
		for _, s2 := range newIds {
			if s1 == s2 {
				found = true
				break
			}
		}
		// String not found. We add it to return slice
		if !found {
			diff = append(diff, s1)
		}
	}

	for _, mediaId := range diff {
		result := <-a.Srv.Store.FileInfo().PermanentDelete(mediaId)
		if result.Err != nil {
			mlog.Warn("Failed to delete media from promo", mlog.String("promo_id", promo.Id), mlog.String("promo_id", promo.Id), mlog.Err(result.Err))
		}
	}
	a.Srv.Store.FileInfo().ClearCaches() // RemoveFromCache dont work???

	return nil

	/*result := <-a.Srv.Store.FileInfo().DeleteForProduct(product.Id)
	if result.Err != nil {
		mlog.Warn("Failed to delete offices from product", mlog.String("product_id", product.Id), mlog.String("product_id", product.Id), mlog.Err(result.Err))
	}

	return nil*/
}

func (a *App) getMediaForPromo(promo *model.Promo) ([]*model.FileInfo, *model.AppError) {
	/*	if len(promo.FileIds) == 0 {
			return nil, nil
		}
	*/
	return a.GetFileInfosForMetadata(promo.Id)
}

func (a *App) PreparePromoListForClient(originalList *model.PromoList) *model.PromoList {
	list := &model.PromoList{
		Promos: make(map[string]*model.Promo, len(originalList.Promos)),
		Order:  originalList.Order, // Note that this uses the original Order array, so it isn't a deep copy
	}

	for id, originalPromo := range originalList.Promos {
		promo := a.PreparePromoForClient(originalPromo, false)

		list.Promos[id] = promo
	}

	return list
}

func (a *App) DeletePromo(promoId, deleteByID string) (*model.Promo, *model.AppError) {
	result := <-a.Srv.Store.Promo().Get(promoId)
	if result.Err != nil {
		result.Err.StatusCode = http.StatusBadRequest
		return nil, result.Err
	}
	promo := result.Data.(*model.Promo)

	if result := <-a.Srv.Store.Promo().Delete(promoId, model.GetMillis(), deleteByID); result.Err != nil {
		return nil, result.Err
	}

	return promo, nil
}

func (a *App) GetAllPromosBeforePromo(promoId string, page, perPage int, options *model.PromoGetOptions) (*model.PromoList, *model.AppError) {

	if result := <-a.Srv.Store.Promo().GetAllPromosBefore(promoId, perPage, page*perPage, options); result.Err != nil {
		return nil, result.Err
	} else {
		list := a.PreparePromoListForClient(result.Data.(*model.PromoList))

		return list, nil
	}
}

func (a *App) GetAllPromosAfterPromo(promoId string, page, perPage int, options *model.PromoGetOptions) (*model.PromoList, *model.AppError) {

	if result := <-a.Srv.Store.Promo().GetAllPromosAfter(promoId, perPage, page*perPage, options); result.Err != nil {
		return nil, result.Err
	} else {
		list := a.PreparePromoListForClient(result.Data.(*model.PromoList))

		return list, nil
	}
}

func (a *App) GetAllPromosAroundPromo(promoId string, offset, limit int, before bool) (*model.PromoList, *model.AppError) {
	var pchan store.StoreChannel

	/*if before {
		pchan = a.Srv.Store.Promo().GetAllPromosBefore(promoId, limit, offset)
	} else {
		pchan = a.Srv.Store.Promo().GetAllPromosAfter(promoId, limit, offset)
	}*/

	if result := <-pchan; result.Err != nil {
		return nil, result.Err
	} else {
		list := a.PreparePromoListForClient(result.Data.(*model.PromoList))

		return list, nil
	}
}

func (a *App) GetAllPromosSince(time int64, options *model.PromoGetOptions) (*model.PromoList, *model.AppError) {
	options.AllowFromCache = true
	if result := <-a.Srv.Store.Promo().GetAllPromosSince(time, options); result.Err != nil {
		return nil, result.Err
	} else {
		list := a.PreparePromoListForClient(result.Data.(*model.PromoList))

		return list, nil
	}
}

func (a *App) GetAllPromosPage(page int, perPage int, options *model.PromoGetOptions) (*model.PromoList, *model.AppError) {
	options.AllowFromCache = true
	if result := <-a.Srv.Store.Promo().GetAllPromos(page*perPage, perPage, options); result.Err != nil {
		return nil, result.Err
	} else {
		list := a.PreparePromoListForClient(result.Data.(*model.PromoList))

		return list, nil
	}
}

func (a *App) UpdatePromoStatus(promoId string, status *model.PromoStatus) (*model.Promo, *model.AppError) {
	switch status.Status {
	case model.PROMO_STATUS_DRAFT:
	case model.PROMO_STATUS_MODERATION:
	case model.PROMO_STATUS_ACCEPTED:
	case model.PROMO_STATUS_REJECTED:
	default:
		return nil, model.NewAppError("UpdatePromoStatus", "api.promo.update_promo_status.status_validate.app_error", nil, status.Status, http.StatusBadRequest)
	}
	promo, err := a.GetPromo(promoId)
	if err != nil {
		return nil, err
	}
	/*if err == nil && promo.Status == status.Status {
		return promo, nil
	}*/

	promo.Status = status.Status
	promo.Active = status.Active
	result := <-a.Srv.Store.Promo().Update(promo)

	if result.Err != nil {
		return nil, result.Err
	}

	return result.Data.(*model.Promo), nil
}
