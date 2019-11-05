package app

import (
	"im/model"
	"im/store"
	"net/http"
)

func (a *App) GetOffice(officeId string) (*model.Office, *model.AppError) {

	result := <-a.Srv.Store.Office().Get(officeId)
	if result.Err != nil {
		return nil, result.Err
	}

	roffice := result.Data.(*model.Office)

	roffice = a.PrepareOfficeForClient(roffice, false)

	return roffice, nil
}

func (a *App) GetOfficesPage(page int, perPage int, sort string) (*model.OfficeList, *model.AppError) {
	return a.GetOffices(page*perPage, perPage, sort)
}

func (a *App) GetOffices(offset int, limit int, sort string) (*model.OfficeList, *model.AppError) {

	result := <-a.Srv.Store.Office().GetAllPage(offset, limit, model.GetOrder(sort))

	if result.Err != nil {
		return nil, result.Err
	}

	return result.Data.(*model.OfficeList), nil
}

func (a *App) CreateOffice(office *model.Office) (*model.Office, *model.AppError) {

	result := <-a.Srv.Store.Office().Save(office)
	if result.Err != nil {
		return nil, result.Err
	}

	roffice := result.Data.(*model.Office)

	return roffice, nil
}

func (a *App) UpdateOffice(office *model.Office, safeUpdate bool) (*model.Office, *model.AppError) {
	//office.SanitizeProps()

	result := <-a.Srv.Store.Office().Get(office.Id)
	if result.Err != nil {
		return nil, result.Err
	}

	oldOffice := result.Data.(*model.Office)

	if oldOffice == nil {
		err := model.NewAppError("UpdateOffice", "api.office.update_office.find.app_error", nil, "id="+office.Id, http.StatusBadRequest)
		return nil, err
	}

	if oldOffice.DeleteAt != 0 {
		err := model.NewAppError("UpdateOffice", "api.office.update_office.permissions_details.app_error", map[string]interface{}{"OfficeId": office.Id}, "", http.StatusBadRequest)
		return nil, err
	}

	newOffice := &model.Office{}
	*newOffice = *oldOffice

	if newOffice.Name != office.Name {
		newOffice.Name = office.Name
	}

	
	newOffice.Preview = office.Preview
	newOffice.Description = office.Description


	result = <-a.Srv.Store.Office().Update(newOffice)
	if result.Err != nil {
		return nil, result.Err
	}

	roffice := result.Data.(*model.Office)
	roffice = a.PrepareOfficeForClient(roffice, false)

	//a.InvalidateCacheForChannelOffices(roffice.ChannelId)

	return roffice, nil
}

func (a *App) PrepareOfficeForClient(originalOffice *model.Office, isNewOffice bool) *model.Office {
	office := originalOffice.Clone()





	//office.Metadata.Images = a.getCategoryForOffice(office)

	return office
}

func (a *App) PrepareOfficeListForClient(originalList *model.OfficeList) *model.OfficeList {
	list := &model.OfficeList{
		Offices: make(map[string]*model.Office, len(originalList.Offices)),
		Order: originalList.Order, // Note that this uses the original Order array, so it isn't a deep copy
	}

	for id, originalOffice := range originalList.Offices {
		office := a.PrepareOfficeForClient(originalOffice, false)

		list.Offices[id] = office
	}

	return list
}

func (a *App) DeleteOffice(officeId, deleteByID string) (*model.Office, *model.AppError) {
	result := <-a.Srv.Store.Office().Get(officeId)
	if result.Err != nil {
		result.Err.StatusCode = http.StatusBadRequest
		return nil, result.Err
	}
	office := result.Data.(*model.Office)


	if result := <-a.Srv.Store.Office().Delete(officeId, model.GetMillis(), deleteByID); result.Err != nil {
		return nil, result.Err
	}


	return office, nil
}


func (a *App) GetAllOfficesBeforeOffice(officeId string, page, perPage int) (*model.OfficeList, *model.AppError) {

	if result := <-a.Srv.Store.Office().GetAllOfficesBefore( officeId, perPage, page*perPage); result.Err != nil {
		return nil, result.Err
	} else {
		return result.Data.(*model.OfficeList), nil
	}
}

func (a *App) GetAllOfficesAfterOffice( officeId string, page, perPage int) (*model.OfficeList, *model.AppError) {


	if result := <-a.Srv.Store.Office().GetAllOfficesAfter(officeId, perPage, page*perPage); result.Err != nil {
		return nil, result.Err
	} else {
		return result.Data.(*model.OfficeList), nil
	}
}

func (a *App) GetAllOfficesAroundOffice(officeId string, offset, limit int, before bool) (*model.OfficeList, *model.AppError) {
	var pchan store.StoreChannel

	if before {
		pchan = a.Srv.Store.Office().GetAllOfficesBefore(officeId, limit, offset)
	} else {
		pchan = a.Srv.Store.Office().GetAllOfficesAfter(officeId, limit, offset)
	}

	if result := <-pchan; result.Err != nil {
		return nil, result.Err
	} else {
		return result.Data.(*model.OfficeList), nil
	}
}

func (a *App) GetAllOfficesSince(time int64) (*model.OfficeList, *model.AppError) {
	if result := <-a.Srv.Store.Office().GetAllOfficesSince(time, true); result.Err != nil {
		return nil, result.Err
	} else {
		return result.Data.(*model.OfficeList), nil
	}
}

func (a *App) GetAllOfficesPage(page int, perPage int) (*model.OfficeList, *model.AppError) {
	if result := <-a.Srv.Store.Office().GetAllOffices(page*perPage, perPage, true); result.Err != nil {
		return nil, result.Err
	} else {
		return result.Data.(*model.OfficeList), nil
	}
}
