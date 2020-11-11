package api4

import (
	"im/model"
	"net/http"
	"strconv"
)

func (api *API) InitExtra() {

	api.BaseRoutes.Extras.Handle("", api.ApiHandler(getAllExtras)).Methods("GET")

	api.BaseRoutes.Extras.Handle("", api.ApiHandler(createExtra)).Methods("POST")
	api.BaseRoutes.Extra.Handle("", api.ApiHandler(getExtra)).Methods("GET")
	api.BaseRoutes.Extra.Handle("", api.ApiHandler(updateExtra)).Methods("PUT")
	api.BaseRoutes.Extra.Handle("", api.ApiHandler(deleteExtra)).Methods("DELETE")

}

func getAllExtras(c *Context, w http.ResponseWriter, r *http.Request) {
	//c.RequireUserId()
	if c.Err != nil {
		return
	}

	afterExtra := r.URL.Query().Get("after")
	beforeExtra := r.URL.Query().Get("before")
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

	var list *model.ExtraList
	var err *model.AppError
	//etag := ""

	if since > 0 {
		list, err = c.App.GetAllExtrasSince(since)
	} else if len(afterExtra) > 0 {

		list, err = c.App.GetAllExtrasAfterExtra(afterExtra, c.Params.Page, c.Params.PerPage)
	} else if len(beforeExtra) > 0 {

		list, err = c.App.GetAllExtrasBeforeExtra(beforeExtra, c.Params.Page, c.Params.PerPage)
	} else {
		list, err = c.App.GetAllExtrasPage(c.Params.Page, c.Params.PerPage)
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

func getExtra(c *Context, w http.ResponseWriter, r *http.Request) {
	c.RequireExtraId()
	if c.Err != nil {
		return
	}

	extra, err := c.App.GetExtra(c.Params.ExtraId)

	if err != nil {
		c.Err = err
		return
	}

	w.Write([]byte(extra.ToJson()))

}

func updateExtra(c *Context, w http.ResponseWriter, r *http.Request) {
	c.RequireExtraId()
	if c.Err != nil {
		return
	}

	extra := model.ExtraFromJson(r.Body)

	if extra == nil {
		c.SetInvalidParam("extra")
		return
	}

	// The extra being updated in the payload must be the same one as indicated in the URL.
	if extra.Id != c.Params.ExtraId {
		c.SetInvalidParam("id")
		return
	}

	extra.Id = c.Params.ExtraId

	rextra, err := c.App.UpdateExtra(extra, false)
	if err != nil {
		c.Err = err
		return
	}

	w.Write([]byte(rextra.ToJson()))
}

func createExtra(c *Context, w http.ResponseWriter, r *http.Request) {

	extra := model.ExtraFromJson(r.Body)

	if extra == nil {
		c.SetInvalidParam("extra")
		return
	}

	result, err := c.App.CreateExtra(extra)
	if err != nil {
		c.Err = err
		return
	}
	w.Write([]byte(result.ToJson()))
}

func deleteExtra(c *Context, w http.ResponseWriter, r *http.Request) {
	c.RequireExtraId()
	if c.Err != nil {
		return
	}

	_, err := c.App.GetExtra(c.Params.ExtraId)
	if err != nil {
		c.SetPermissionError(model.PERMISSION_DELETE_POST)
		return
	}

	/*if c.App.Session.UserId == extra.UserId {
		if !c.App.SessionHasPermissionToChannel(c.App.Session, extra.ChannelId, model.PERMISSION_DELETE_POST) {
			c.SetPermissionError(model.PERMISSION_DELETE_POST)
			return
		}
	} else {
		if !c.App.SessionHasPermissionToChannel(c.App.Session, extra.ChannelId, model.PERMISSION_DELETE_OTHERS_POSTS) {
			c.SetPermissionError(model.PERMISSION_DELETE_OTHERS_POSTS)
			return
		}
	}*/

	if _, err := c.App.DeleteExtra(c.Params.ExtraId, c.App.Session.UserId); err != nil {
		c.Err = err
		return
	}

	ReturnStatusOK(w)
}
