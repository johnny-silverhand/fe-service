
package web

import (
	"net/http"
	"strings"

	"im/app"
	"im/mlog"
	"im/model"
	"im/utils"
)

type Context struct {
	App           *app.App
	Log           *mlog.Logger
	Params        *Params
	Err           *model.AppError
	siteURLHeader string
}

func (c *Context) LogAudit(extraInfo string) {
	audit := &model.Audit{UserId: c.App.Session.UserId, IpAddress: c.App.IpAddress, Action: c.App.Path, ExtraInfo: extraInfo, SessionId: c.App.Session.Id}
	if err := c.App.Srv.Store.Audit().Save(audit); err != nil {
		c.LogError(err)
	}
}

func (c *Context) LogAuditWithUserId(userId, extraInfo string) {

	if len(c.App.Session.UserId) > 0 {
		extraInfo = strings.TrimSpace(extraInfo + " session_user=" + c.App.Session.UserId)
	}

	audit := &model.Audit{UserId: userId, IpAddress: c.App.IpAddress, Action: c.App.Path, ExtraInfo: extraInfo, SessionId: c.App.Session.Id}
	if err := c.App.Srv.Store.Audit().Save(audit); err != nil {
		c.LogError(err)
	}
}

func (c *Context) LogError(err *model.AppError) {
	// Filter out 404s, endless reconnects and browser compatibility errors
	if err.StatusCode == http.StatusNotFound ||
		(c.App.Path == "/api/v3/users/websocket" && err.StatusCode == http.StatusUnauthorized) ||
		err.Id == "web.check_browser_compatibility.app_error" {
		c.LogDebug(err)
	} else {
		c.Log.Error(
			err.SystemMessage(utils.TDefault),
			mlog.String("err_where", err.Where),
			mlog.Int("http_code", err.StatusCode),
			mlog.String("err_details", err.DetailedError),
		)
	}
}

func (c *Context) LogInfo(err *model.AppError) {
	// Filter out 401s
	if err.StatusCode == http.StatusUnauthorized {
		c.LogDebug(err)
	} else {
		c.Log.Info(
			err.SystemMessage(utils.TDefault),
			mlog.String("err_where", err.Where),
			mlog.Int("http_code", err.StatusCode),
			mlog.String("err_details", err.DetailedError),
		)
	}
}

func (c *Context) LogDebug(err *model.AppError) {
	c.Log.Debug(
		err.SystemMessage(utils.TDefault),
		mlog.String("err_where", err.Where),
		mlog.Int("http_code", err.StatusCode),
		mlog.String("err_details", err.DetailedError),
	)
}

func (c *Context) IsSystemAdmin() bool {
	return c.App.SessionHasPermissionTo(c.App.Session, model.PERMISSION_MANAGE_SYSTEM)
}

func (c *Context) SessionRequired() {
	if !*c.App.Config().ServiceSettings.EnableUserAccessTokens && c.App.Session.Props[model.SESSION_PROP_TYPE] == model.SESSION_TYPE_USER_ACCESS_TOKEN {
		c.Err = model.NewAppError("", "api.context.session_expired.app_error", nil, "UserAccessToken", http.StatusUnauthorized)
		return
	}

	if len(c.App.Session.UserId) == 0 {
		c.Err = model.NewAppError("", "api.context.session_expired.app_error", nil, "UserRequired", http.StatusUnauthorized)
		return
	}
}

func (c *Context) RemoveSessionCookie(w http.ResponseWriter, r *http.Request) {
	cookie := &http.Cookie{
		Name:     model.SESSION_COOKIE_TOKEN,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
	}

	http.SetCookie(w, cookie)
}

func (c *Context) SetInvalidParam(parameter string) {
	c.Err = NewInvalidParamError(parameter)
}

func (c *Context) SetInvalidUrlParam(parameter string) {
	c.Err = NewInvalidUrlParamError(parameter)
}

func (c *Context) HandleEtag(etag string, routeName string, w http.ResponseWriter, r *http.Request) bool {

	if et := r.Header.Get(model.HEADER_ETAG_CLIENT); len(etag) > 0 {
		if et == etag {
			w.Header().Set(model.HEADER_ETAG_SERVER, etag)
			w.WriteHeader(http.StatusNotModified)

			return true
		}
	}

	return false
}

func NewInvalidParamError(parameter string) *model.AppError {
	err := model.NewAppError("Context", "api.context.invalid_body_param.app_error", map[string]interface{}{"Name": parameter}, "", http.StatusBadRequest)
	return err
}
func NewInvalidUrlParamError(parameter string) *model.AppError {
	err := model.NewAppError("Context", "api.context.invalid_url_param.app_error", map[string]interface{}{"Name": parameter}, "", http.StatusBadRequest)
	return err
}

func (c *Context) SetPermissionError(permission *model.Permission) {
	c.Err = c.App.MakePermissionError(permission)
}

func (c *Context) SetSiteURLHeader(url string) {
	c.siteURLHeader = strings.TrimRight(url, "/")
}

func (c *Context) GetSiteURLHeader() string {
	return c.siteURLHeader
}

func (c *Context) RequireUserId() *Context {
	if c.Err != nil {
		return c
	}

	if c.Params.UserId == model.ME {
		c.Params.UserId = c.App.Session.UserId
	}

	if len(c.Params.UserId) != 26 {
		c.SetInvalidUrlParam("user_id")
	}
	return c
}

func (c *Context) RequireTeamId() *Context {
	if c.Err != nil {
		return c
	}

	if len(c.Params.TeamId) != 26 {
		c.SetInvalidUrlParam("team_id")
	}
	return c
}

func (c *Context) RequireInviteId() *Context {
	if c.Err != nil {
		return c
	}

	if len(c.Params.InviteId) == 0 {
		c.SetInvalidUrlParam("invite_id")
	}
	return c
}

func (c *Context) RequireTokenId() *Context {
	if c.Err != nil {
		return c
	}

	if len(c.Params.TokenId) != 26 {
		c.SetInvalidUrlParam("token_id")
	}
	return c
}
func (c *Context) RequireCategoryId() *Context {
	if c.Err != nil {
		return c
	}

	if len(c.Params.CategoryId) != 26 {
		c.SetInvalidUrlParam("category_id")
	}
	return c
}
func (c *Context) RequireProductId() *Context {
	if c.Err != nil {
		return c
	}

	if len(c.Params.ProductId) != 26 {
		c.SetInvalidUrlParam("product_id")
	}
	return c
}
func (c *Context) RequireChannelId() *Context {
	if c.Err != nil {
		return c
	}

	if len(c.Params.ChannelId) != 26 {
		c.SetInvalidUrlParam("channel_id")
	}
	return c
}
func (c *Context) RequireMessageId() *Context {
	if c.Err != nil {
		return c
	}

	if len(c.Params.MessageId) != 26 {
		c.SetInvalidUrlParam("message_id")
	}
	return c
}
func (c *Context) RequireSessionUserId() *Context {
	if c.Err != nil {
		return c
	}

	if len(c.App.Session.UserId) != 26 {
		c.SetInvalidUrlParam("user_id")
	}

	return c
}

func (c *Context) RequireUsername() *Context {
	if c.Err != nil {
		return c
	}

	if !model.IsValidUsername(c.Params.Username) {
		c.SetInvalidParam("username")
	}

	return c
}

func (c *Context) RequirePostId() *Context {
	if c.Err != nil {
		return c
	}

	if len(c.Params.PostId) != 26 {
		c.SetInvalidUrlParam("post_id")
	}
	return c
}

func (c *Context) RequireAppId() *Context {
	if c.Err != nil {
		return c
	}

	if len(c.Params.AppId) != 26 {
		c.SetInvalidUrlParam("app_id")
	}
	return c
}

func (c *Context) RequireFileId() *Context {
	if c.Err != nil {
		return c
	}

	if len(c.Params.FileId) != 26 {
		c.SetInvalidUrlParam("file_id")
	}

	return c
}

func (c *Context) RequireFilename() *Context {
	if c.Err != nil {
		return c
	}

	if len(c.Params.Filename) == 0 {
		c.SetInvalidUrlParam("filename")
	}

	return c
}

func (c *Context) RequirePluginId() *Context {
	if c.Err != nil {
		return c
	}

	if len(c.Params.PluginId) == 0 {
		c.SetInvalidUrlParam("plugin_id")
	}

	return c
}

func (c *Context) RequireReportId() *Context {
	if c.Err != nil {
		return c
	}

	if len(c.Params.ReportId) != 26 {
		c.SetInvalidUrlParam("report_id")
	}
	return c
}

func (c *Context) RequireEmojiId() *Context {
	if c.Err != nil {
		return c
	}

	if len(c.Params.EmojiId) != 26 {
		c.SetInvalidUrlParam("emoji_id")
	}
	return c
}

func (c *Context) RequireTeamName() *Context {
	if c.Err != nil {
		return c
	}

	if !model.IsValidTeamName(c.Params.TeamName) {
		c.SetInvalidUrlParam("team_name")
	}

	return c
}

func (c *Context) RequireChannelName() *Context {
	if c.Err != nil {
		return c
	}

	if !model.IsValidChannelIdentifier(c.Params.ChannelName) {
		c.SetInvalidUrlParam("channel_name")
	}

	return c
}

func (c *Context) RequireEmail() *Context {
	if c.Err != nil {
		return c
	}

	if !model.IsValidEmail(c.Params.Email) {
		c.SetInvalidUrlParam("email")
	}

	return c
}

func (c *Context) RequireCategory() *Context {
	if c.Err != nil {
		return c
	}

	if !model.IsValidAlphaNumHyphenUnderscore(c.Params.Category, true) {
		c.SetInvalidUrlParam("category")
	}

	return c
}

func (c *Context) RequireService() *Context {
	if c.Err != nil {
		return c
	}

	if len(c.Params.Service) == 0 {
		c.SetInvalidUrlParam("service")
	}

	return c
}

func (c *Context) RequirePreferenceName() *Context {
	if c.Err != nil {
		return c
	}

	if !model.IsValidAlphaNumHyphenUnderscore(c.Params.PreferenceName, true) {
		c.SetInvalidUrlParam("preference_name")
	}

	return c
}


func (c *Context) RequireHookId() *Context {
	if c.Err != nil {
		return c
	}

	if len(c.Params.HookId) != 26 {
		c.SetInvalidUrlParam("hook_id")
	}

	return c
}

func (c *Context) RequireCommandId() *Context {
	if c.Err != nil {
		return c
	}

	if len(c.Params.CommandId) != 26 {
		c.SetInvalidUrlParam("command_id")
	}
	return c
}

func (c *Context) RequireJobId() *Context {
	if c.Err != nil {
		return c
	}

	if len(c.Params.JobId) != 26 {
		c.SetInvalidUrlParam("job_id")
	}
	return c
}

func (c *Context) RequireJobType() *Context {
	if c.Err != nil {
		return c
	}

	if len(c.Params.JobType) == 0 || len(c.Params.JobType) > 32 {
		c.SetInvalidUrlParam("job_type")
	}
	return c
}

func (c *Context) RequireActionId() *Context {
	if c.Err != nil {
		return c
	}

	if len(c.Params.ActionId) != 26 {
		c.SetInvalidUrlParam("action_id")
	}
	return c
}

func (c *Context) RequireRoleId() *Context {
	if c.Err != nil {
		return c
	}

	if len(c.Params.RoleId) != 26 {
		c.SetInvalidUrlParam("role_id")
	}
	return c
}

func (c *Context) RequireSchemeId() *Context {
	if c.Err != nil {
		return c
	}

	if len(c.Params.SchemeId) != 26 {
		c.SetInvalidUrlParam("scheme_id")
	}
	return c
}

func (c *Context) RequireRoleName() *Context {
	if c.Err != nil {
		return c
	}

	if !model.IsValidRoleName(c.Params.RoleName) {
		c.SetInvalidUrlParam("role_name")
	}

	return c
}


func (c *Context) RequireRemoteId() *Context {
	if c.Err != nil {
		return c
	}

	if len(c.Params.RemoteId) == 0 {
		c.SetInvalidUrlParam("remote_id")
	}
	return c
}
