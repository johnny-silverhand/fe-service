package api4

import (
	"net/http"
	"strconv"
	"time"

	//l4g "github.com/alecthomas/log4go"

	"im/model"
)

func (api *API) InitMessage() {
	//l4g.Debug(utils.T("api.post.init.debug"))

	api.BaseRoutes.Messages.Handle("", api.ApiSessionRequired(getAllMessages)).Methods("GET")
	api.BaseRoutes.Messages.Handle("", api.ApiSessionRequired(createMessage)).Methods("POST")
	api.BaseRoutes.Message.Handle("", api.ApiSessionRequired(getMessage)).Methods("GET")

	api.BaseRoutes.Message.Handle("/files/info", api.ApiSessionRequired(getFileInfosForMessage)).Methods("GET")
	api.BaseRoutes.MessagesForChannel.Handle("", api.ApiSessionRequired(getMessagesForChannel)).Methods("GET")

}

func ts() string {
	now := time.Now()
	unixNano := now.UnixNano()
	umillisec := unixNano / 1000000
	return strconv.FormatInt(umillisec, 10)
}

func createMessage(c *Context, w http.ResponseWriter, r *http.Request) {
	post := model.PostFromJson(r.Body)
	if post == nil {
		c.SetInvalidParam("message")
		return
	}

	c.RequireSessionUserId()

	post.UserId = c.App.Session.UserId

	hasPermission := false

	if channel, err := c.App.FindOpennedChannel(post.UserId); err == nil {
		post.ChannelId = channel.Id

		if channel.Status == model.CHANNEL_STATUS_CLOSED {
			c.App.PatchChannel(channel, &model.ChannelPatch{Status: model.NewString(model.CHANNEL_STATUS_OPEN)}, post.UserId)
		}
		c.App.AddChannelMemberIfNeeded(post.UserId, channel)
	} else {

		if channel, err = c.App.CreateUnresolvedChannel(post.UserId); err != nil {
			c.Err = err
			return
		} else {
			if cmhjResult := <-c.App.Srv.Store.ChannelMemberHistory().LogJoinEvent(post.UserId, channel.Id, model.GetMillis()); cmhjResult.Err != nil {
				//l4g.Warn(cmhjResult.Err.Error())
			}
			post.ChannelId = channel.Id
		}
	}

	hasPermission = true

	if !hasPermission {
		c.SetPermissionError(model.PERMISSION_CREATE_POST)
		return
	}

	if post.CreateAt != 0 && !c.App.SessionHasPermissionTo(c.App.Session, model.PERMISSION_MANAGE_SYSTEM) {
		post.CreateAt = 0
	}

	rp, err := c.App.CreatePostAsExternal(post)
	if err != nil {
		c.Err = err
		return
	}

	//c.App.SetStatusOnline(c.Session.UserId, c.Session.Id, false, "")
	c.App.UpdateLastActivityAtIfNeeded(c.App.Session)

	//c.App.SetLastViewedChannel(view, c.Session.UserId, !c.Session.IsMobileApp())

	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(rp.ToJson()))
}

func getMessagesForChannel(c *Context, w http.ResponseWriter, r *http.Request) {
	c.RequireChannelId()
	if c.Err != nil {
		return
	}

	afterMessage := r.URL.Query().Get("after")
	beforeMessage := r.URL.Query().Get("before")
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

	/*if !c.App.SessionHasPermissionToChannel(c.Session, c.Params.ChannelId, model.PERMISSION_READ_CHANNEL) {
		c.SetPermissionError(model.PERMISSION_READ_CHANNEL)
		return
	}*/

	var list *model.PostList
	var err *model.AppError
	etag := ""

	if since > 0 {
		list, err = c.App.GetPostsSince(c.Params.ChannelId, since)
	} else if len(afterMessage) > 0 {
		etag = c.App.GetPostsEtag(c.Params.ChannelId)

		if c.HandleEtag(etag, "Get Messages After", w, r) {
			return
		}

		list, err = c.App.GetPostsAfterPost(c.Params.ChannelId, afterMessage, c.Params.Page, c.Params.PerPage)
	} else if len(beforeMessage) > 0 {
		etag = c.App.GetPostsEtag(c.Params.ChannelId)

		if c.HandleEtag(etag, "Get Messages Before", w, r) {
			return
		}

		list, err = c.App.GetPostsBeforePost(c.Params.ChannelId, beforeMessage, c.Params.Page, c.Params.PerPage)
	} else {
		etag = c.App.GetPostsEtag(c.Params.ChannelId)

		if c.HandleEtag(etag, "Get Messages", w, r) {
			return
		}

		list, err = c.App.GetPostsPage(c.Params.ChannelId, c.Params.Page, c.Params.PerPage)
	}

	if err != nil {
		c.Err = err
		return
	}

	if len(etag) > 0 {
		w.Header().Set(model.HEADER_ETAG_SERVER, etag)
	}
	w.Write([]byte(c.App.PreparePostListForClient(list).ToJson()))
}

func getAllMessages(c *Context, w http.ResponseWriter, r *http.Request) {
	//c.RequireUserId()
	if c.Err != nil {
		return
	}

	afterMessage := r.URL.Query().Get("after")
	beforeMessage := r.URL.Query().Get("before")
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

	var list *model.MessageArray
	var err *model.AppError
	etag := ""

	if since > 0 {
		list, err = c.App.GetAllMessagesSince(c.App.Session.UserId, since)
	} else if len(afterMessage) > 0 {

		list, err = c.App.GetAllMessagesAfterMessage(c.App.Session.UserId, afterMessage, c.Params.Page, c.Params.PerPage)
	} else if len(beforeMessage) > 0 {

		list, err = c.App.GetAllMessagesBeforeMessage(c.App.Session.UserId, beforeMessage, c.Params.Page, c.Params.PerPage)
	} else {

		list, err = c.App.GetAllMessagesPage(c.App.Session.UserId, c.Params.Page, c.Params.PerPage)
	}

	if err != nil {
		c.Err = err
		return
	}

	if len(etag) > 0 {
		w.Header().Set(model.HEADER_ETAG_SERVER, etag)
	}

	w.Write([]byte(c.App.PrepareMessageListForClient(list).ToJson()))
}

func getMessage(c *Context, w http.ResponseWriter, r *http.Request) {
	c.RequireMessageId()
	if c.Err != nil {
		return
	}

	var post *model.Post
	var err *model.AppError
	if post, err = c.App.GetSinglePost(c.Params.MessageId); err != nil {
		c.Err = err
		return
	}

	if c.HandleEtag(post.Etag(), "Get Message", w, r) {
		return
	} else {
		w.Header().Set(model.HEADER_ETAG_SERVER, post.Etag())
		w.Write([]byte(c.App.PreparePostForClient(post, false).ToJson()))
	}
}

func getFileInfosForMessage(c *Context, w http.ResponseWriter, r *http.Request) {
	c.RequireMessageId()
	if c.Err != nil {
		return
	}

	if infos, err := c.App.GetFileInfosForPost(c.Params.MessageId); err != nil {
		c.Err = err
		return
	} else if c.HandleEtag(model.GetEtagForFileInfos(infos), "Get File Infos For Message", w, r) {
		return
	} else {
		w.Header().Set("Cache-Control", "max-age=2592000, public")
		w.Header().Set(model.HEADER_ETAG_SERVER, model.GetEtagForFileInfos(infos))
		w.Write([]byte(model.FileInfosToJson(infos)))
	}
}
