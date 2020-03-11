package api4

import (
	"fmt"
	"github.com/mssola/user_agent"
	"im/model"
	"net/http"
)

func (api *API) InitInvite() {
	api.BaseRoutes.Invite.Handle("", api.ApiSessionRequired(getInviteLink)).Methods("GET")
	api.BaseRoutes.Invite.Handle("/check_info", api.ApiHandler(getInviteInfoByIp)).Methods("GET")
	api.BaseRoutes.Invite.Handle("/{invite_id:[A-Za-z0-9]+}", api.ApiHandler(redirectToStore)).Methods("GET")
}

func getInviteLink(c *Context, w http.ResponseWriter, r *http.Request) {
	c.RequireAppId()
	if c.Err != nil {
		return
	}
	var token *model.Token
	user, err := c.App.GetUser(c.App.Session.UserId)
	if err != nil {
		c.Err = err
		return
	}
	token, err = c.App.GetUserInviteToken(user)
	if err != nil {
		c.Err = err
		return
	}
	w.Write([]byte(token.Token))
}

func redirectToStore(c *Context, w http.ResponseWriter, r *http.Request) {
	c.RequireInviteId()
	if c.Err != nil {
		return
	}

	var user *model.User
	var err *model.AppError

	if token, err := c.App.GetStageToken(c.Params.InviteId); err != nil {
		c.Err = err
		return
	} else {
		user, err = c.App.GetUser(token.UserId)
		if err != nil {
			c.Err = err
			return
		}
		user.SanitizeProfile(nil)
	}

	session := &model.Session{UserId: user.Id, Roles: user.GetRawRoles(), IsOAuth: true, AppId: user.AppId}
	session.GenerateCSRF()
	session.AddProp("IP", fmt.Sprintf("%s", c.App.IpAddress))
	session.SetExpireInDays(1)

	if session, err = c.App.CreateSession(session); err != nil {
		err.StatusCode = http.StatusInternalServerError
		c.Err = err
		return
	}

	ua := user_agent.New(r.UserAgent())
	os := ua.OSInfo().Name
	if os == "iPhone OS" || os == "iPhone" {
		http.Redirect(w, r, "itms-apps://itunes.apple.com/app/id1492900077", 301)
	} else if os == "Android" {
		// TODO брать ид из сущности Application
		http.Redirect(w, r, "market://details?id=ID", 301)
	} else {
		c.SetInvalidParam("invite_id")
	}

	//itms-apps://itunes.apple.com/app/<<App ID>>
	//market://details?id=<<Package id>>
}

func getInviteInfoByIp(c *Context, w http.ResponseWriter, r *http.Request) {
	c.RequireAppId()

	if session, _ := c.App.GetSessionByProps(c.App.IpAddress, c.Params.AppId); session != nil {
		var token *model.Token
		user, err := c.App.GetUser(session.UserId)

		if err != nil {
			c.Err = err
			return
		}

		token, err = c.App.GetUserInviteToken(user)

		if err != nil {
			c.Err = err
			return
		}

		w.Write([]byte(token.Extra))
	}
}
