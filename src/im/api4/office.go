package api4

import (
	"im/app"
	"im/model"
	"net/http"
	"strconv"
	"time"
)

func (api *API) InitOffice() {

	api.BaseRoutes.Offices.Handle("", api.ApiHandler(getAllOffices)).Methods("GET")
	api.BaseRoutes.Offices.Handle("", api.ApiHandler(createOffice)).Methods("POST")

	api.BaseRoutes.Office.Handle("", api.ApiHandler(getOffice)).Methods("GET")
	api.BaseRoutes.Office.Handle("", api.ApiHandler(updateOffice)).Methods("PUT")
	api.BaseRoutes.Office.Handle("", api.ApiHandler(deleteOffice)).Methods("DELETE")

	api.BaseRoutes.Office.Handle("/attach_office", api.ApiSessionRequired(attachOfficeId)).Methods("POST")
}

func getAllOffices(c *Context, w http.ResponseWriter, r *http.Request) {
	//c.RequireUserId()
	c.RequireAppId()
	if c.Err != nil {
		return
	}

	var appId = c.Params.AppId

	c.App.Session.ToJson()

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

	/*	if !c.App.SessionHasPermissionToChannel(c.Session, c.Params.ChannelId, model.PERMISSION_READ_CHANNEL) {
		c.SetPermissionError(model.PERMISSION_READ_CHANNEL)
		return
	}*/

	var list *model.OfficeList
	var err *model.AppError
	//etag := ""

	if since > 0 {
		list, err = c.App.GetAllOfficesSince(since, &appId)
	} else if len(afterOffice) > 0 {

		list, err = c.App.GetAllOfficesAfterOffice(afterOffice, c.Params.Page, c.Params.PerPage, &appId)
	} else if len(beforeOffice) > 0 {

		list, err = c.App.GetAllOfficesBeforeOffice(beforeOffice, c.Params.Page, c.Params.PerPage, &appId)
	} else {
		list, err = c.App.GetAllOfficesPage(c.Params.Page, c.Params.PerPage, &appId)
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

func getOffice(c *Context, w http.ResponseWriter, r *http.Request) {
	c.RequireOfficeId()
	if c.Err != nil {
		return
	}

	office, err := c.App.GetOffice(c.Params.OfficeId)

	if err != nil {
		c.Err = err
		return
	}

	w.Write([]byte(office.ToJson()))

}

func updateOffice(c *Context, w http.ResponseWriter, r *http.Request) {
	c.RequireOfficeId()
	if c.Err != nil {
		return
	}

	office := model.OfficeFromJson(r.Body)

	if office == nil {
		c.SetInvalidParam("office")
		return
	}

	// The office being updated in the payload must be the same one as indicated in the URL.
	if office.Id != c.Params.OfficeId {
		c.SetInvalidParam("id")
		return
	}

	office.Id = c.Params.OfficeId

	roffice, err := c.App.UpdateOffice(office, false)
	if err != nil {
		c.Err = err
		return
	}

	w.Write([]byte(roffice.ToJson()))
}

func createOffice(c *Context, w http.ResponseWriter, r *http.Request) {

	office := model.OfficeFromJson(r.Body)

	if office == nil {
		c.SetInvalidParam("office")
		return
	}

	result, err := c.App.CreateOffice(office)
	if err != nil {
		c.Err = err
		return
	}
	w.Write([]byte(result.ToJson()))
}

func deleteOffice(c *Context, w http.ResponseWriter, r *http.Request) {
	c.RequireOfficeId()
	if c.Err != nil {
		return
	}

	_, err := c.App.GetOffice(c.Params.OfficeId)
	if err != nil {
		c.SetPermissionError(model.PERMISSION_DELETE_POST)
		return
	}

	/*if c.App.Session.UserId == office.UserId {
		if !c.App.SessionHasPermissionToChannel(c.App.Session, office.ChannelId, model.PERMISSION_DELETE_POST) {
			c.SetPermissionError(model.PERMISSION_DELETE_POST)
			return
		}
	} else {
		if !c.App.SessionHasPermissionToChannel(c.App.Session, office.ChannelId, model.PERMISSION_DELETE_OTHERS_POSTS) {
			c.SetPermissionError(model.PERMISSION_DELETE_OTHERS_POSTS)
			return
		}
	}*/

	if _, err := c.App.DeleteOffice(c.Params.OfficeId, c.App.Session.UserId); err != nil {
		c.Err = err
		return
	}

	ReturnStatusOK(w)
}

func attachOfficeId(c *Context, w http.ResponseWriter, r *http.Request) {
	c.RequireOfficeId()
	if c.Err != nil {
		return
	}
	/*props := model.MapFromJson(r.Body)*/

	officeId := c.Params.OfficeId
	if len(officeId) == 0 {
		c.SetInvalidParam("office_id")
		return
	}

	// A special case where we logout of all other sessions with the same office id
	/*if err := c.App.RevokeSessionsForDeviceId(c.App.Session.UserId, officeId, c.App.Session.Id); err != nil {
		c.Err = err
		return
	}*/

	c.App.ClearSessionCacheForUser(c.App.Session.UserId)
	c.App.Session.SetExpireInDays(*c.App.Config().ServiceSettings.SessionLengthMobileInDays)

	maxAge := *c.App.Config().ServiceSettings.SessionLengthMobileInDays * 60 * 60 * 24

	secure := false
	if app.GetProtocol(r) == "https" {
		secure = true
	}

	expiresAt := time.Unix(model.GetMillis()/1000+int64(maxAge), 0)
	sessionCookie := &http.Cookie{
		Name:     model.SESSION_COOKIE_TOKEN,
		Value:    c.App.Session.Token,
		Path:     "/",
		MaxAge:   maxAge,
		Expires:  expiresAt,
		HttpOnly: true,
		Domain:   c.App.GetCookieDomain(),
		Secure:   secure,
	}

	http.SetCookie(w, sessionCookie)

	if err := c.App.AttachOfficeId(c.App.Session.Id, officeId, c.App.Session.ExpiresAt); err != nil {
		c.Err = err
		return
	}

	c.LogAudit("")
	ReturnStatusOK(w)
}
