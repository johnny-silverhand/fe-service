package api4

import (
	"im/model"
	"net/http"
	"strconv"
)

func (api *API) InitLevel() {

	api.BaseRoutes.Levels.Handle("", api.ApiHandler(getAllLevels)).Methods("GET")

	api.BaseRoutes.Level.Handle("", api.ApiHandler(getLevel)).Methods("GET")
	api.BaseRoutes.Level.Handle("", api.ApiHandler(createLevel)).Methods("POST")
	api.BaseRoutes.Level.Handle("", api.ApiHandler(updateLevel)).Methods("PUT")
	api.BaseRoutes.Level.Handle("", api.ApiHandler(deleteLevel)).Methods("DELETE")

}

func getAllLevels(c *Context, w http.ResponseWriter, r *http.Request) {
	//c.RequireUserId()
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
		list, err = c.App.GetAllLevelsSince(since, nil)
	} else if len(afterLevel) > 0 {

		list, err = c.App.GetAllLevelsAfterLevel(afterLevel, c.Params.Page, c.Params.PerPage, nil)
	} else if len(beforeLevel) > 0 {

		list, err = c.App.GetAllLevelsBeforeLevel(beforeLevel, c.Params.Page, c.Params.PerPage, nil)
	} else {
		list, err = c.App.GetAllLevelsPage(c.Params.Page, c.Params.PerPage, nil)
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

func getLevel(c *Context, w http.ResponseWriter, r *http.Request) {
	c.RequireLevelId()
	if c.Err != nil {
		return
	}

	level, err := c.App.GetLevel(c.Params.LevelId)

	if err != nil {
		c.Err = err
		return
	}

	w.Write([]byte(level.ToJson()))

}

func updateLevel(c *Context, w http.ResponseWriter, r *http.Request) {
	c.RequireLevelId()
	if c.Err != nil {
		return
	}

	level := model.LevelFromJson(r.Body)

	if level == nil {
		c.SetInvalidParam("level")
		return
	}

	// The level being updated in the payload must be the same one as indicated in the URL.
	if level.Id != c.Params.LevelId {
		c.SetInvalidParam("id")
		return
	}

	level.Id = c.Params.LevelId

	rlevel, err := c.App.UpdateLevel(level, false)
	if err != nil {
		c.Err = err
		return
	}

	w.Write([]byte(rlevel.ToJson()))
}

func createLevel(c *Context, w http.ResponseWriter, r *http.Request) {

	level := model.LevelFromJson(r.Body)

	if level == nil {
		c.SetInvalidParam("level")
		return
	}

	result, err := c.App.CreateLevel(level)
	if err != nil {
		c.Err = err
		return
	}
	w.Write([]byte(result.ToJson()))
}

func deleteLevel(c *Context, w http.ResponseWriter, r *http.Request) {
	c.RequireLevelId()
	if c.Err != nil {
		return
	}

	_, err := c.App.GetLevel(c.Params.LevelId)
	if err != nil {
		c.SetPermissionError(model.PERMISSION_DELETE_POST)
		return
	}

	/*if c.App.Session.UserId == level.UserId {
		if !c.App.SessionHasPermissionToChannel(c.App.Session, level.ChannelId, model.PERMISSION_DELETE_POST) {
			c.SetPermissionError(model.PERMISSION_DELETE_POST)
			return
		}
	} else {
		if !c.App.SessionHasPermissionToChannel(c.App.Session, level.ChannelId, model.PERMISSION_DELETE_OTHERS_POSTS) {
			c.SetPermissionError(model.PERMISSION_DELETE_OTHERS_POSTS)
			return
		}
	}*/

	if _, err := c.App.DeleteLevel(c.Params.LevelId, c.App.Session.UserId); err != nil {
		c.Err = err
		return
	}

	ReturnStatusOK(w)
}
