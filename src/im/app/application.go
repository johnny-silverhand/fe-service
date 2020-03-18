package app

import (
	"fmt"
	"im/mlog"
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

func (a *App) CreateSingleApplication(application *model.Application) (*model.Application, *model.AppError) {

	result := <-a.Srv.Store.Application().Save(application)
	if result.Err != nil {
		return nil, result.Err
	}

	rapplication := result.Data.(*model.Application)

	return rapplication, nil
}

func (a *App) CreateApplication(application *model.Application) (*model.Application, *model.AppError) {

	result := <-a.Srv.Store.Application().Save(application)
	if result.Err != nil {
		return nil, result.Err
	}

	rapplication := result.Data.(*model.Application)

	team := &model.Team{
		DisplayName:     rapplication.Name,
		Name:            rapplication.Id,
		Description:     rapplication.Description,
		Email:           rapplication.Email,
		Type:            "O",
		CompanyName:     rapplication.Name,
		AllowOpenInvite: true,
	}

	if rteam, err := a.CreateTeam(team); err != nil {

		fmt.Println(err.Error())

	} else {
		user := &model.User{
			Username:      rapplication.Email,
			Password:      "123",
			Email:         rapplication.Email,
			EmailVerified: true,
			Nickname:      rapplication.Email,
			FirstName:     "Оператор",
			LastName:      rapplication.Name,
			Locale:        "ru",
			AppId:         rapplication.Id,
			Roles:         "channel_admin",
		}

		if ruser, err := a.CreateUserWithInviteId(user, rteam.InviteId); err == nil {
			a.UpdateTeamMemberRoles(rteam.Id, ruser.Id, "team_user team_admin channel_user")
		}
	}

	return rapplication, nil
}

func (a *App) PatchApplication(id string, patch *model.ApplicationPatch) (*model.Application, *model.AppError) {
	result := <-a.Srv.Store.Application().Get(id)
	if result.Err != nil {
		return nil, result.Err
	}
	resultTeam := <-a.Srv.Store.Team().GetByName(id)
	if resultTeam.Err != nil {
		return nil, resultTeam.Err
	}

	oldApplication := result.Data.(*model.Application)
	oldTeam := resultTeam.Data.(*model.Team)

	if oldApplication == nil {
		err := model.NewAppError("UpdateApplication", "api.application.update_application.find.app_error", nil, "id="+id, http.StatusBadRequest)
		return nil, err
	}

	if oldApplication.DeleteAt != 0 {
		err := model.NewAppError("UpdateApplication", "api.application.update_application.permissions_details.app_error", map[string]interface{}{"ApplicationId": id}, "", http.StatusBadRequest)
		return nil, err
	}

	if oldTeam == nil {
		err := model.NewAppError("UpdateApplication", "api.application.update_application.find.app_error", nil, "id="+id, http.StatusBadRequest)
		return nil, err
	}

	if oldTeam.DeleteAt != 0 {
		err := model.NewAppError("UpdateApplication", "api.application.update_application.permissions_details.app_error", map[string]interface{}{"ApplicationId": id}, "", http.StatusBadRequest)
		return nil, err
	}

	newApplication := &model.Application{}
	*newApplication = *oldApplication
	newApplication.Patch(patch)

	result = <-a.Srv.Store.Application().Update(newApplication)
	if result.Err != nil {
		return nil, result.Err
	}

	if newApplication.Email != oldApplication.Email {
		if ruser, err := a.GetUserApplicationByEmail(oldApplication.Email, oldApplication.Id); err != nil {
			return nil, err
		} else if _, err := a.PatchUser(ruser.Id, &model.UserPatch{Email: model.NewString(newApplication.Email)}, true); err != nil {
			return nil, err
		}
	}

	if len(newApplication.Password) > 0 {
		if ruser, err := a.GetUserApplicationByEmail(newApplication.Email, newApplication.Id); err != nil {
			return nil, err
		} else if err := a.UpdatePassword(ruser, newApplication.Password); err != nil {
			return nil, err
		}
	}

	if newApplication.Name != oldApplication.Name ||
		newApplication.Description != oldApplication.Description {

		teamPatch := &model.TeamPatch{
			DisplayName: model.NewString(newApplication.Name),
			Description: model.NewString(newApplication.Description),
			CompanyName: model.NewString(newApplication.Name),
		}

		newTeam := &model.Team{}
		*newTeam = *oldTeam
		newTeam.Patch(teamPatch)

		resultTeam = <-a.Srv.Store.Team().Update(newTeam)
		if resultTeam.Err != nil {
			return nil, resultTeam.Err
		}
	}

	rapplication := result.Data.(*model.Application)
	rapplication = a.PrepareApplicationForClient(rapplication, false)

	return rapplication, nil
}

func (a *App) UpdateApplication(id string, patch *model.ApplicationPatch, safeUpdate bool) (*model.Application, *model.AppError) {
	//application.SanitizeProps()

	result := <-a.Srv.Store.Application().Get(id)
	if result.Err != nil {
		return nil, result.Err
	}

	oldApplication := result.Data.(*model.Application)

	if oldApplication == nil {
		err := model.NewAppError("UpdateApplication", "api.application.update_application.find.app_error", nil, "id="+id, http.StatusBadRequest)
		return nil, err
	}

	if oldApplication.DeleteAt != 0 {
		err := model.NewAppError("UpdateApplication", "api.application.update_application.permissions_details.app_error", map[string]interface{}{"ApplicationId": id}, "", http.StatusBadRequest)
		return nil, err
	}

	newApplication := &model.Application{}
	*newApplication = *oldApplication
	newApplication.Patch(patch)

	if newApplication.Email != oldApplication.Email {
		if ruser, err := a.GetUserByEmail(oldApplication.Email); err != nil {
			return nil, err
		} else {
			ruser.Username = newApplication.Email
			ruser.Nickname = newApplication.Email
			ruser.Email = newApplication.Email
			if _, err := a.UpdateUser(ruser, false); err != nil {
				return nil, err
			}
		}
	}

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
	application.ModerationCount = 0
	if productsForModeration, err := a.GetProductsForModeration(&model.ProductGetOptions{AppId: application.Id}); err != nil {
		mlog.Warn("Failed to get products for a moderation", mlog.String("application_id", application.Id), mlog.Any("err", err))
	} else {
		application.ModerationCount += len(productsForModeration.Products)
	}

	if promosForModeration, err := a.GetPromosForModeration(&model.PromoGetOptions{AppId: application.Id}); err != nil {
		mlog.Warn("Failed to get promos for a moderation", mlog.String("application_id", application.Id), mlog.Any("err", err))
	} else {
		application.ModerationCount += len(promosForModeration.Promos)
	}

	//application.ModerationCount = 999

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

	result = <-a.Srv.Store.Team().GetByName(appId)
	if result.Err != nil {
		result.Err.StatusCode = http.StatusBadRequest
		return nil, result.Err
	}
	team := result.Data.(*model.Team)
	if err := a.SoftDeleteTeam(team.Id); err != nil {
		return nil, err
	}

	if users, err := a.GetUsers(&model.UserGetOptions{AppId: appId, Page: 0, PerPage: 100000}); err != nil {
		return nil, err
	} else {
		for _, user := range users {
			a.UpdateActive(user, false)
		}
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

func (a *App) GetAllApplicationsPage(options *model.ApplicationGetOptions) (*model.ApplicationList, *model.AppError) {
	if result := <-a.Srv.Store.Application().GetApplications(options); result.Err != nil {
		return nil, result.Err
	} else {
		return result.Data.(*model.ApplicationList), nil
	}
}
