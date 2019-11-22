package app

import (
	"im/model"
	"im/store"
	"net/http"
)

func (a *App) GetClient(clientId string) (*model.Client, *model.AppError) {

	result := <-a.Srv.Store.Client().Get(clientId)
	if result.Err != nil {
		return nil, result.Err
	}

	rclient := result.Data.(*model.Client)

	rclient = a.PrepareClientForClient(rclient, false)

	return rclient, nil
}

func (a *App) GetClientsPage(page int, perPage int, sort string) (*model.ClientList, *model.AppError) {
	return a.GetClients(page*perPage, perPage, sort)
}

func (a *App) GetClients(offset int, limit int, sort string) (*model.ClientList, *model.AppError) {

	result := <-a.Srv.Store.Client().GetAllPage(offset, limit, model.GetOrder(sort))

	if result.Err != nil {
		return nil, result.Err
	}

	return result.Data.(*model.ClientList), nil
}

func (a *App) CreateClient(client *model.Client) (*model.Client, *model.AppError) {

	result := <-a.Srv.Store.Client().Save(client)
	if result.Err != nil {
		return nil, result.Err
	}

	rclient := result.Data.(*model.Client)

	return rclient, nil
}

func (a *App) UpdateClient(client *model.Client, safeUpdate bool) (*model.Client, *model.AppError) {
	//client.SanitizeProps()

	result := <-a.Srv.Store.Client().Get(client.Id)
	if result.Err != nil {
		return nil, result.Err
	}

	oldClient := result.Data.(*model.Client)

	if oldClient == nil {
		err := model.NewAppError("UpdateClient", "api.client.update_client.find.app_error", nil, "id="+client.Id, http.StatusBadRequest)
		return nil, err
	}

	if oldClient.DeleteAt != 0 {
		err := model.NewAppError("UpdateClient", "api.client.update_client.permissions_details.app_error", map[string]interface{}{"ClientId": client.Id}, "", http.StatusBadRequest)
		return nil, err
	}

	newClient := &model.Client{}
	*newClient = *oldClient

	if newClient.Name != client.Name {
		newClient.Name = client.Name
	}


	newClient.Preview = client.Preview
	newClient.Description = client.Description


	result = <-a.Srv.Store.Client().Update(newClient)
	if result.Err != nil {
		return nil, result.Err
	}

	rclient := result.Data.(*model.Client)
	rclient = a.PrepareClientForClient(rclient, false)

	//a.InvalidateCacheForChannelClients(rclient.ChannelId)

	return rclient, nil
}

func (a *App) PrepareClientForClient(originalClient *model.Client, isNewClient bool) *model.Client {
	client := originalClient.Clone()





	//client.Metadata.Images = a.getCategoryForClient(client)

	return client
}

func (a *App) PrepareClientListForClient(originalList *model.ClientList) *model.ClientList {
	list := &model.ClientList{
		Clients: make(map[string]*model.Client, len(originalList.Clients)),
		Order: originalList.Order, // Note that this uses the original Order array, so it isn't a deep copy
	}

	for id, originalClient := range originalList.Clients {
		client := a.PrepareClientForClient(originalClient, false)

		list.Clients[id] = client
	}

	return list
}

func (a *App) DeleteClient(clientId, deleteByID string) (*model.Client, *model.AppError) {
	result := <-a.Srv.Store.Client().Get(clientId)
	if result.Err != nil {
		result.Err.StatusCode = http.StatusBadRequest
		return nil, result.Err
	}
	client := result.Data.(*model.Client)


	if result := <-a.Srv.Store.Client().Delete(clientId, model.GetMillis(), deleteByID); result.Err != nil {
		return nil, result.Err
	}


	return client, nil
}


func (a *App) GetAllClientsBeforeClient(clientId string, page, perPage int) (*model.ClientList, *model.AppError) {

	if result := <-a.Srv.Store.Client().GetAllClientsBefore( clientId, perPage, page*perPage); result.Err != nil {
		return nil, result.Err
	} else {
		return result.Data.(*model.ClientList), nil
	}
}

func (a *App) GetAllClientsAfterClient( clientId string, page, perPage int) (*model.ClientList, *model.AppError) {


	if result := <-a.Srv.Store.Client().GetAllClientsAfter(clientId, perPage, page*perPage); result.Err != nil {
		return nil, result.Err
	} else {
		return result.Data.(*model.ClientList), nil
	}
}

func (a *App) GetAllClientsAroundClient(clientId string, offset, limit int, before bool) (*model.ClientList, *model.AppError) {
	var pchan store.StoreChannel

	if before {
		pchan = a.Srv.Store.Client().GetAllClientsBefore(clientId, limit, offset)
	} else {
		pchan = a.Srv.Store.Client().GetAllClientsAfter(clientId, limit, offset)
	}

	if result := <-pchan; result.Err != nil {
		return nil, result.Err
	} else {
		return result.Data.(*model.ClientList), nil
	}
}

func (a *App) GetAllClientsSince(time int64) (*model.ClientList, *model.AppError) {
	if result := <-a.Srv.Store.Client().GetAllClientsSince(time, true); result.Err != nil {
		return nil, result.Err
	} else {
		return result.Data.(*model.ClientList), nil
	}
}

func (a *App) GetAllClientsPage(page int, perPage int) (*model.ClientList, *model.AppError) {
	if result := <-a.Srv.Store.Client().GetAllClients(page*perPage, perPage, true); result.Err != nil {
		return nil, result.Err
	} else {
		return result.Data.(*model.ClientList), nil
	}
}
