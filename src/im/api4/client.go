package api4

import (
	"im/model"
	"net/http"
	"strconv"
)

func (api *API) InitClient() {

	api.BaseRoutes.Clients.Handle("", api.ApiHandler(getAllClients)).Methods("GET")
	api.BaseRoutes.Clients.Handle("", api.ApiHandler(createClient)).Methods("POST")

	api.BaseRoutes.Client.Handle("", api.ApiHandler(getClient)).Methods("GET")
	api.BaseRoutes.Client.Handle("", api.ApiHandler(updateClient)).Methods("PUT")
	api.BaseRoutes.Client.Handle("", api.ApiHandler(deleteClient)).Methods("DELETE")

	api.BaseRoutes.Client.Handle("/offices", api.ApiHandler(getClientOffices)).Methods("GET")
	api.BaseRoutes.Client.Handle("/products", api.ApiHandler(getClientProducts)).Methods("GET")
	api.BaseRoutes.Client.Handle("/promos", api.ApiHandler(getClientPromos)).Methods("GET")
}

func getClientPromos(c *Context, w http.ResponseWriter, r *http.Request) {
	c.RequireClientId()

	if c.Err != nil {
		return
	}

	if products, err := c.App.GetPromosPageByClient(c.Params.Page, c.Params.PerPage, c.Params.Sort, c.Params.ClientId); err != nil {
		c.Err = err
		return
	} else {
		w.Write([]byte(products.ToJson()))
	}
}

func getClientProducts(c *Context, w http.ResponseWriter, r *http.Request) {
	c.RequireClientId()

	if c.Err != nil {
		return
	}

	if products, err := c.App.GetProductsPageByClient(c.Params.Page, c.Params.PerPage, c.Params.Sort, c.Params.ClientId); err != nil {
		c.Err = err
		return
	} else {
		w.Write([]byte(products.ToJson()))
	}
}

func getClientOffices(c *Context, w http.ResponseWriter, r *http.Request) {
	c.RequireClientId()

	if c.Err != nil {
		return
	}

	afterOffice := r.URL.Query().Get("after")
	beforeOffice := r.URL.Query().Get("before")
	sinceString := r.URL.Query().Get("since")

	var since int64
	var parseError error

	if len(sinceString) > 0 {
		since, parseError = strconv.ParseInt(sinceString, 10, 64)
		if parseError != nil {
			c.SetInvalidParam("since")
			return
		}
	}

	var list *model.OfficeList
	var err *model.AppError
	//etag := ""

	if since > 0 {
		list, err = c.App.GetAllOfficesSince(since, &c.Params.ClientId)
	} else if len(afterOffice) > 0 {

		list, err = c.App.GetAllOfficesAfterOffice(afterOffice, c.Params.Page, c.Params.PerPage, &c.Params.ClientId)
	} else if len(beforeOffice) > 0 {

		list, err = c.App.GetAllOfficesBeforeOffice(beforeOffice, c.Params.Page, c.Params.PerPage, &c.Params.ClientId)
	} else {
		list, err = c.App.GetAllOfficesPage(c.Params.Page, c.Params.PerPage, &c.Params.ClientId)
	}

	if err != nil {
		c.Err = err
		return
	}

	w.Write([]byte(list.ToJson()))
}

func getAllClients(c *Context, w http.ResponseWriter, r *http.Request) {
	//c.RequireUserId()
	if c.Err != nil {
		return
	}

	afterClient := r.URL.Query().Get("after")
	beforeClient := r.URL.Query().Get("before")
	sinceString := r.URL.Query().Get("since")

	var since int64
	var parseError error

	if len(sinceString) > 0 {
		since, parseError = strconv.ParseInt(sinceString, 10, 64)
		if parseError != nil {
			c.SetInvalidParam("since")
			return
		}
	}

	/*	if !c.App.SessionHasPermissionToChannel(c.Session, c.Params.ChannelId, model.PERMISSION_READ_CHANNEL) {
		c.SetPermissionError(model.PERMISSION_READ_CHANNEL)
		return
	}*/

	var list *model.ClientList
	var err *model.AppError
	//etag := ""

	if since > 0 {
		list, err = c.App.GetAllClientsSince(since)
	} else if len(afterClient) > 0 {

		list, err = c.App.GetAllClientsAfterClient(afterClient, c.Params.Page, c.Params.PerPage)
	} else if len(beforeClient) > 0 {

		list, err = c.App.GetAllClientsBeforeClient(beforeClient, c.Params.Page, c.Params.PerPage)
	} else {
		list, err = c.App.GetAllClientsPage(c.Params.Page, c.Params.PerPage)
	}

	if err != nil {
		c.Err = err
		return
	}

	/*	if len(etag) > 0 {
		w.Header().Set(model.HEADER_ETAG_SERVER, etag)
	}*/

	w.Write([]byte(list.ToJson()))
}

func createClient(c *Context, w http.ResponseWriter, r *http.Request) {

	client := model.ClientFromJson(r.Body)

	if client == nil {
		c.SetInvalidParam("client")
		return
	}

	result, err := c.App.CreateClient(client)
	if err != nil {
		c.Err = err
		return
	}
	w.Write([]byte(result.ToJson()))
}

func getClient(c *Context, w http.ResponseWriter, r *http.Request) {
	c.RequireClientId()
	if c.Err != nil {
		return
	}

	client, err := c.App.GetClient(c.Params.ClientId)

	if err != nil {
		c.Err = err
		return
	}

	w.Write([]byte(client.ToJson()))

}

func updateClient(c *Context, w http.ResponseWriter, r *http.Request) {
	c.RequireClientId()
	if c.Err != nil {
		return
	}

	client := model.ClientFromJson(r.Body)

	if client == nil {
		c.SetInvalidParam("client")
		return
	}

	// The client being updated in the payload must be the same one as indicated in the URL.
	if client.Id != c.Params.ClientId {
		c.SetInvalidParam("id")
		return
	}

	client.Id = c.Params.ClientId

	rclient, err := c.App.UpdateClient(client, false)
	if err != nil {
		c.Err = err
		return
	}

	w.Write([]byte(rclient.ToJson()))
}

func deleteClient(c *Context, w http.ResponseWriter, r *http.Request) {
	c.RequireClientId()
	if c.Err != nil {
		return
	}

	_, err := c.App.GetClient(c.Params.ClientId)
	if err != nil {
		c.SetPermissionError(model.PERMISSION_DELETE_POST)
		return
	}

	/*if c.App.Session.UserId == client.UserId {
		if !c.App.SessionHasPermissionToChannel(c.App.Session, client.ChannelId, model.PERMISSION_DELETE_POST) {
			c.SetPermissionError(model.PERMISSION_DELETE_POST)
			return
		}
	} else {
		if !c.App.SessionHasPermissionToChannel(c.App.Session, client.ChannelId, model.PERMISSION_DELETE_OTHERS_POSTS) {
			c.SetPermissionError(model.PERMISSION_DELETE_OTHERS_POSTS)
			return
		}
	}*/

	if _, err := c.App.DeleteClient(c.Params.ClientId, c.App.Session.UserId); err != nil {
		c.Err = err
		return
	}

	ReturnStatusOK(w)
}
