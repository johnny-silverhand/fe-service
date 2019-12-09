package api4

import (
	"im/model"
	"net/http"
	"strconv"
)

func (api *API) InitApplication() {

	api.BaseRoutes.Applications.Handle("", api.ApiHandler(getAllApplications)).Methods("GET")
	api.BaseRoutes.Applications.Handle("", api.ApiHandler(createApplication)).Methods("POST")

	api.BaseRoutes.Application.Handle("", api.ApiHandler(getApplication)).Methods("GET")
	api.BaseRoutes.Application.Handle("", api.ApiHandler(updateApplication)).Methods("PUT")
	api.BaseRoutes.Application.Handle("", api.ApiHandler(deleteApplication)).Methods("DELETE")

	api.BaseRoutes.Application.Handle("/offices", api.ApiHandler(getApplicationOffices)).Methods("GET")
	api.BaseRoutes.Application.Handle("/products", api.ApiHandler(getApplicationProducts)).Methods("GET")
	api.BaseRoutes.Application.Handle("/promos", api.ApiHandler(getApplicationPromos)).Methods("GET")
	api.BaseRoutes.Application.Handle("/levels", api.ApiHandler(getApplicationLevels)).Methods("GET")
}

func getApplicationLevels(c *Context, w http.ResponseWriter, r *http.Request) {
	c.RequireAppId()

	if c.Err != nil {
		return
	}

	afterLevel := r.URL.Query().Get("after")
	beforeLevel := r.URL.Query().Get("before")
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

	var list *model.LevelList
	var err *model.AppError
	//etag := ""

	if since > 0 {
		list, err = c.App.GetAllLevelsSince(since, &c.Params.AppId)
	} else if len(afterLevel) > 0 {

		list, err = c.App.GetAllLevelsAfterLevel(afterLevel, c.Params.Page, c.Params.PerPage, &c.Params.AppId)
	} else if len(beforeLevel) > 0 {

		list, err = c.App.GetAllLevelsBeforeLevel(beforeLevel, c.Params.Page, c.Params.PerPage, &c.Params.AppId)
	} else {
		list, err = c.App.GetAllLevelsPage(c.Params.Page, c.Params.PerPage, &c.Params.AppId)
	}

	if err != nil {
		c.Err = err
		return
	}

	w.Write([]byte(list.ToJson()))

}

func getApplicationPromos(c *Context, w http.ResponseWriter, r *http.Request) {
	c.RequireAppId()

	if c.Err != nil {
		return
	}

	if products, err := c.App.GetPromosPageByApp(c.Params.Page, c.Params.PerPage, c.Params.Sort, c.Params.AppId); err != nil {
		c.Err = err
		return
	} else {
		w.Write([]byte(products.ToJson()))
	}
}

func getApplicationProducts(c *Context, w http.ResponseWriter, r *http.Request) {
	c.RequireAppId()

	if c.Err != nil {
		return
	}

	if products, err := c.App.GetProductsPageByApp(c.Params.Page, c.Params.PerPage, c.Params.Sort, c.Params.AppId); err != nil {
		c.Err = err
		return
	} else {
		w.Write([]byte(products.ToJson()))
	}
}

func getApplicationOffices(c *Context, w http.ResponseWriter, r *http.Request) {
	c.RequireAppId()

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
		list, err = c.App.GetAllOfficesSince(since, &c.Params.AppId)
	} else if len(afterOffice) > 0 {

		list, err = c.App.GetAllOfficesAfterOffice(afterOffice, c.Params.Page, c.Params.PerPage, &c.Params.AppId)
	} else if len(beforeOffice) > 0 {

		list, err = c.App.GetAllOfficesBeforeOffice(beforeOffice, c.Params.Page, c.Params.PerPage, &c.Params.AppId)
	} else {
		list, err = c.App.GetAllOfficesPage(c.Params.Page, c.Params.PerPage, &c.Params.AppId)
	}

	if err != nil {
		c.Err = err
		return
	}

	w.Write([]byte(list.ToJson()))
}

func getAllApplications(c *Context, w http.ResponseWriter, r *http.Request) {
	//c.RequireUserId()
	if c.Err != nil {
		return
	}

	afterApplication := r.URL.Query().Get("after")
	beforeApplication := r.URL.Query().Get("before")
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

	var list *model.ApplicationList
	var err *model.AppError
	//etag := ""

	if since > 0 {
		list, err = c.App.GetAllApplicationsSince(since)
	} else if len(afterApplication) > 0 {

		list, err = c.App.GetAllApplicationsAfter(afterApplication, c.Params.Page, c.Params.PerPage)
	} else if len(beforeApplication) > 0 {

		list, err = c.App.GetAllApplicationsBefore(beforeApplication, c.Params.Page, c.Params.PerPage)
	} else {
		list, err = c.App.GetAllApplicationsPage(c.Params.Page, c.Params.PerPage)
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

func createApplication(c *Context, w http.ResponseWriter, r *http.Request) {

	app := model.ApplicationFromJson(r.Body)

	if app == nil {
		c.SetInvalidParam("app")
		return
	}

	if app.Email == "" {
		c.SetInvalidParam("email")
		return
	}

	result, err := c.App.CreateApplication(app)
	if err != nil {
		c.Err = err
		return
	}

	newUser := &model.User{
		Nickname:      result.Name,
		Email:         result.Email,
		EmailVerified: true,
	}

	c.App.AutoCreateUser(newUser)

	w.Write([]byte(result.ToJson()))
}

func getApplication(c *Context, w http.ResponseWriter, r *http.Request) {
	c.RequireAppId()
	if c.Err != nil {
		return
	}

	app, err := c.App.GetApplication(c.Params.AppId)

	if err != nil {
		c.Err = err
		return
	}

	w.Write([]byte(app.ToJson()))

}

func updateApplication(c *Context, w http.ResponseWriter, r *http.Request) {
	c.RequireAppId()
	if c.Err != nil {
		return
	}

	app := model.ApplicationFromJson(r.Body)

	if app == nil {
		c.SetInvalidParam("app")
		return
	}

	// The app being updated in the payload must be the same one as indicated in the URL.
	if app.Id != c.Params.AppId {
		c.SetInvalidParam("id")
		return
	}

	app.Id = c.Params.AppId

	rapp, err := c.App.UpdateApplication(app, false)
	if err != nil {
		c.Err = err
		return
	}

	w.Write([]byte(rapp.ToJson()))
}

func deleteApplication(c *Context, w http.ResponseWriter, r *http.Request) {
	c.RequireAppId()
	if c.Err != nil {
		return
	}

	_, err := c.App.GetApplication(c.Params.AppId)
	if err != nil {
		//c.SetPermissionError(model.PERMISSION_DELETE_APP)
		return
	}

	/*if c.App.Session.UserId == app.UserId {
		if !c.App.SessionHasPermissionToChannel(c.App.Session, app.ChannelId, model.PERMISSION_DELETE_APP) {
			c.SetPermissionError(model.PERMISSION_DELETE_APP)
			return
		}
	} else {
		if !c.App.SessionHasPermissionToChannel(c.App.Session, app.ChannelId, model.PERMISSION_DELETE_OTHERS_APPS) {
			c.SetPermissionError(model.PERMISSION_DELETE_OTHERS_APPS)
			return
		}
	}*/

	if _, err := c.App.DeleteApplication(c.Params.AppId, c.App.Session.UserId); err != nil {
		c.Err = err
		return
	}

	ReturnStatusOK(w)
}
