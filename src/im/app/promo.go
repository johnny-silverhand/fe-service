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
	//if !safeUpdate {
	newPromo.Media = promo.Media
	//}

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

func (a *App) GetAllPromosBeforePromo(promoId string, page, perPage int) (*model.PromoList, *model.AppError) {

	if result := <-a.Srv.Store.Promo().GetAllPromosBefore(promoId, perPage, page*perPage); result.Err != nil {
		return nil, result.Err
	} else {
		list := a.PreparePromoListForClient(result.Data.(*model.PromoList))

		return list, nil
	}
}

func (a *App) GetAllPromosAfterPromo(promoId string, page, perPage int) (*model.PromoList, *model.AppError) {

	if result := <-a.Srv.Store.Promo().GetAllPromosAfter(promoId, perPage, page*perPage); result.Err != nil {
		return nil, result.Err
	} else {
		list := a.PreparePromoListForClient(result.Data.(*model.PromoList))

		return list, nil
	}
}

func (a *App) GetAllPromosAroundPromo(promoId string, offset, limit int, before bool) (*model.PromoList, *model.AppError) {
	var pchan store.StoreChannel

	if before {
		pchan = a.Srv.Store.Promo().GetAllPromosBefore(promoId, limit, offset)
	} else {
		pchan = a.Srv.Store.Promo().GetAllPromosAfter(promoId, limit, offset)
	}

	if result := <-pchan; result.Err != nil {
		return nil, result.Err
	} else {
		list := a.PreparePromoListForClient(result.Data.(*model.PromoList))

		return list, nil
	}
}

func (a *App) GetAllPromosSince(time int64) (*model.PromoList, *model.AppError) {
	if result := <-a.Srv.Store.Promo().GetAllPromosSince(time, true); result.Err != nil {
		return nil, result.Err
	} else {
		list := a.PreparePromoListForClient(result.Data.(*model.PromoList))

		return list, nil
	}
}

func (a *App) GetAllPromosPage(page int, perPage int) (*model.PromoList, *model.AppError) {
	if result := <-a.Srv.Store.Promo().GetAllPromos(page*perPage, perPage, true); result.Err != nil {
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
	default:
		return nil, model.NewAppError("UpdatePromoStatus", "api.promo.update_promo_status.status_validate.app_error", nil, status.Status, http.StatusBadRequest)
	}
	promo, err := a.GetPromo(promoId)
	if err == nil && promo.Status == status.Status {
		return promo, nil
	}

	promo.Status = status.Status
	result := <-a.Srv.Store.Promo().Update(promo)

	if result.Err != nil {
		return nil, result.Err
	}

	return result.Data.(*model.Promo), nil
}
