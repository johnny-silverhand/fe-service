package app

import (
	"im/model"
	"im/store"
	"net/http"
)

func (a *App) GetApplication(appId string) (*model.Application, *model.AppError) {

	result := <-a.Srv.Store.Application().Get(appId)
	if result.Err != nil {
		return nil, result.Err
	}

	rapplication := result.Data.(*model.Application)

	rapplication = a.PrepareApplicationForClient(rapplication, false)

	return rapplication, nil
}

func (a *App) GetApplicationsPage(page int, perPage int, sort string) (*model.ApplicationList, *model.AppError) {
	return a.GetApplications(page*perPage, perPage, sort)
}

func (a *App) GetApplications(offset int, limit int, sort string) (*model.ApplicationList, *model.AppError) {

	result := <-a.Srv.Store.Application().GetAllPage(offset, limit, model.GetOrder(sort))

	if result.Err != nil {
		return nil, result.Err
	}

	return result.Data.(*model.ApplicationList), nil
}

func (a *App) CreateApplication(application *model.Application) (*model.Application, *model.AppError) {

	result := <-a.Srv.Store.Application().Save(application)
	if result.Err != nil {
		return nil, result.Err
	}

	rapplication := result.Data.(*model.Application)

	return rapplication, nil
}

func (a *App) UpdateApplication(application *model.Application, safeUpdate bool) (*model.Application, *model.AppError) {
	//application.SanitizeProps()

	result := <-a.Srv.Store.Application().Get(application.Id)
	if result.Err != nil {
		return nil, result.Err
	}

	oldApplication := result.Data.(*model.Application)

	if oldApplication == nil {
		err := model.NewAppError("UpdateApplication", "api.application.update_application.find.app_error", nil, "id="+application.Id, http.StatusBadRequest)
		return nil, err
	}

	if oldApplication.DeleteAt != 0 {
		err := model.NewAppError("UpdateApplication", "api.application.update_application.permissions_details.app_error", map[string]interface{}{"ApplicationId": application.Id}, "", http.StatusBadRequest)
		return nil, err
	}

	newApplication := &model.Application{}
	*newApplication = *oldApplication

	if newApplication.Name != application.Name {
		newApplication.Name = application.Name
	}

	newApplication.Preview = application.Preview
	newApplication.Description = application.Description
	newApplication.Phone = application.Phone
	newApplication.PaymentDetails = application.PaymentDetails
	newApplication.Email = application.Email

	result = <-a.Srv.Store.Application().Update(newApplication)
	if result.Err != nil {
		return nil, result.Err
	}

	rapplication := result.Data.(*model.Application)
	rapplication = a.PrepareApplicationForClient(rapplication, false)

	//a.InvalidateCacheForChannelApplications(rapplication.ChannelId)

	return rapplication, nil
}

func (a *App) PrepareApplicationForClient(originalApplication *model.Application, isNewApplication bool) *model.Application {
	application := originalApplication.Clone()

	//application.Metadata.Images = a.getCategoryForApplication(application)

	return application
}

func (a *App) PrepareApplicationListForClient(originalList *model.ApplicationList) *model.ApplicationList {
	list := &model.ApplicationList{
		Applications: make(map[string]*model.Application, len(originalList.Applications)),
		Order:        originalList.Order, // Note that this uses the original Order array, so it isn't a deep copy
	}

	for id, originalApplication := range originalList.Applications {
		application := a.PrepareApplicationForClient(originalApplication, false)

		list.Applications[id] = application
	}

	return list
}

func (a *App) DeleteApplication(appId, deleteByID string) (*model.Application, *model.AppError) {
	result := <-a.Srv.Store.Application().Get(appId)
	if result.Err != nil {
		result.Err.StatusCode = http.StatusBadRequest
		return nil, result.Err
	}
	application := result.Data.(*model.Application)

	if result := <-a.Srv.Store.Application().Delete(appId, model.GetMillis(), deleteByID); result.Err != nil {
		return nil, result.Err
	}

	return application, nil
}

func (a *App) GetAllApplicationsBefore(appId string, page, perPage int) (*model.ApplicationList, *model.AppError) {
	if result := <-a.Srv.Store.Application().GetAllApplicationsBefore(appId, perPage, page*perPage); result.Err != nil {
		return nil, result.Err
	} else {
		return result.Data.(*model.ApplicationList), nil
	}
}

func (a *App) GetAllApplicationsAfter(appId string, page, perPage int) (*model.ApplicationList, *model.AppError) {
	if result := <-a.Srv.Store.Application().GetAllApplicationsAfter(appId, perPage, page*perPage); result.Err != nil {
		return nil, result.Err
	} else {
		return result.Data.(*model.ApplicationList), nil
	}
}

func (a *App) GetAllApplicationsAroundApplication(appId string, offset, limit int, before bool) (*model.ApplicationList, *model.AppError) {
	var pchan store.StoreChannel

	if before {
		pchan = a.Srv.Store.Application().GetAllApplicationsBefore(appId, limit, offset)
	} else {
		pchan = a.Srv.Store.Application().GetAllApplicationsAfter(appId, limit, offset)
	}

	if result := <-pchan; result.Err != nil {
		return nil, result.Err
	} else {
		return result.Data.(*model.ApplicationList), nil
	}
}

func (a *App) GetAllApplicationsSince(time int64) (*model.ApplicationList, *model.AppError) {
	if result := <-a.Srv.Store.Application().GetAllApplicationsSince(time, true); result.Err != nil {
		return nil, result.Err
	} else {
		return result.Data.(*model.ApplicationList), nil
	}
}

func (a *App) GetAllApplicationsPage(page int, perPage int) (*model.ApplicationList, *model.AppError) {
	if result := <-a.Srv.Store.Application().GetAllApplications(page*perPage, perPage, true); result.Err != nil {
		return nil, result.Err
	} else {
		return result.Data.(*model.ApplicationList), nil
	}
}
