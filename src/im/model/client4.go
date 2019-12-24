// Copyright (c) 2017-present Mattermost, Inc. All Rights Reserved.
// See License.txt for license information.

package model

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const (
	HEADER_REQUEST_ID         = "X-Request-ID"
	HEADER_VERSION_ID         = "X-Version-ID"
	HEADER_CLUSTER_ID         = "X-Cluster-ID"
	HEADER_ETAG_SERVER        = "ETag"
	HEADER_ETAG_CLIENT        = "If-None-Match"
	HEADER_FORWARDED          = "X-Forwarded-For"
	HEADER_REAL_IP            = "X-Real-IP"
	HEADER_FORWARDED_PROTO    = "X-Forwarded-Proto"
	HEADER_TOKEN              = "token"
	HEADER_CSRF_TOKEN         = "X-CSRF-Token"
	HEADER_BEARER             = "BEARER"
	HEADER_AUTH               = "Authorization"
	HEADER_REQUESTED_WITH     = "X-Requested-With"
	HEADER_REQUESTED_WITH_XML = "XMLHttpRequest"
	STATUS                    = "status"
	STATUS_OK                 = "OK"
	STATUS_FAIL               = "FAIL"
	STATUS_REMOVE             = "REMOVE"

	STAGE_TOKEN               = "stage_token"
	INVITE_TOKEN               = "invite_token"

	CLIENT_DIR = "client"

	API_URL_SUFFIX_V1 = "/api/v1"
	API_URL_SUFFIX_V4 = "/api/v4"
	API_URL_SUFFIX    = API_URL_SUFFIX_V4
)

type Response struct {
	StatusCode    int
	Error         *AppError
	RequestId     string
	Etag          string
	ServerVersion string
	Header        http.Header
}

type Client4 struct {
	Url        string       // The location of the server, for example  "http://localhost:8065"
	ApiUrl     string       // The api location of the server, for example "http://localhost:8065/api/v4"
	HttpClient *http.Client // The http client
	AuthToken  string
	AuthType   string
	HttpHeader map[string]string // Headers to be copied over for each request
}

func closeBody(r *http.Response) {
	if r.Body != nil {
		_, _ = ioutil.ReadAll(r.Body)
		_ = r.Body.Close()
	}
}

// Must is a convenience function used for testing.
func (c *Client4) Must(result interface{}, resp *Response) interface{} {
	if resp.Error != nil {
		time.Sleep(time.Second)
		panic(resp.Error)
	}

	return result
}

func NewAPIv4Client(url string) *Client4 {
	return &Client4{url, url + API_URL_SUFFIX, &http.Client{}, "", "", map[string]string{}}
}

func BuildErrorResponse(r *http.Response, err *AppError) *Response {
	var statusCode int
	var header http.Header
	if r != nil {
		statusCode = r.StatusCode
		header = r.Header
	} else {
		statusCode = 0
		header = make(http.Header)
	}

	return &Response{
		StatusCode: statusCode,
		Error:      err,
		Header:     header,
	}
}

func BuildResponse(r *http.Response) *Response {
	return &Response{
		StatusCode:    r.StatusCode,
		RequestId:     r.Header.Get(HEADER_REQUEST_ID),
		Etag:          r.Header.Get(HEADER_ETAG_SERVER),
		ServerVersion: r.Header.Get(HEADER_VERSION_ID),
		Header:        r.Header,
	}
}

func (c *Client4) MockSession(sessionToken string) {
	c.AuthToken = sessionToken
	c.AuthType = HEADER_BEARER
}

func (c *Client4) SetOAuthToken(token string) {
	c.AuthToken = token
	c.AuthType = HEADER_TOKEN
}

func (c *Client4) ClearOAuthToken() {
	c.AuthToken = ""
	c.AuthType = HEADER_BEARER
}

func (c *Client4) GetUsersRoute() string {
	return fmt.Sprintf("/users")
}

func (c *Client4) GetUserRoute(userId string) string {
	return fmt.Sprintf(c.GetUsersRoute()+"/%v", userId)
}

func (c *Client4) GetUserAccessTokensRoute() string {
	return fmt.Sprintf(c.GetUsersRoute() + "/tokens")
}

func (c *Client4) GetUserAccessTokenRoute(tokenId string) string {
	return fmt.Sprintf(c.GetUsersRoute()+"/tokens/%v", tokenId)
}

func (c *Client4) GetUserByUsernameRoute(userName string) string {
	return fmt.Sprintf(c.GetUsersRoute()+"/username/%v", userName)
}

func (c *Client4) GetUserByEmailRoute(email string) string {
	return fmt.Sprintf(c.GetUsersRoute()+"/email/%v", email)
}

func (c *Client4) GetBotsRoute() string {
	return fmt.Sprintf("/bots")
}

func (c *Client4) GetBotRoute(botUserId string) string {
	return fmt.Sprintf("%s/%s", c.GetBotsRoute(), botUserId)
}

func (c *Client4) GetTeamsRoute() string {
	return fmt.Sprintf("/teams")
}

func (c *Client4) GetTeamRoute(teamId string) string {
	return fmt.Sprintf(c.GetTeamsRoute()+"/%v", teamId)
}

func (c *Client4) GetTeamAutoCompleteCommandsRoute(teamId string) string {
	return fmt.Sprintf(c.GetTeamsRoute()+"/%v/commands/autocomplete", teamId)
}

func (c *Client4) GetTeamByNameRoute(teamName string) string {
	return fmt.Sprintf(c.GetTeamsRoute()+"/name/%v", teamName)
}

func (c *Client4) GetTeamMemberRoute(teamId, userId string) string {
	return fmt.Sprintf(c.GetTeamRoute(teamId)+"/members/%v", userId)
}

func (c *Client4) GetTeamMembersRoute(teamId string) string {
	return fmt.Sprintf(c.GetTeamRoute(teamId) + "/members")
}

func (c *Client4) GetTeamStatsRoute(teamId string) string {
	return fmt.Sprintf(c.GetTeamRoute(teamId) + "/stats")
}

func (c *Client4) GetTeamImportRoute(teamId string) string {
	return fmt.Sprintf(c.GetTeamRoute(teamId) + "/import")
}

func (c *Client4) GetChannelsRoute() string {
	return fmt.Sprintf("/channels")
}

func (c *Client4) GetChannelsForTeamRoute(teamId string) string {
	return fmt.Sprintf(c.GetTeamRoute(teamId) + "/channels")
}

func (c *Client4) GetChannelRoute(channelId string) string {
	return fmt.Sprintf(c.GetChannelsRoute()+"/%v", channelId)
}

func (c *Client4) GetChannelByNameRoute(channelName, teamId string) string {
	return fmt.Sprintf(c.GetTeamRoute(teamId)+"/channels/name/%v", channelName)
}

func (c *Client4) GetChannelByNameForTeamNameRoute(channelName, teamName string) string {
	return fmt.Sprintf(c.GetTeamByNameRoute(teamName)+"/channels/name/%v", channelName)
}

func (c *Client4) GetChannelMembersRoute(channelId string) string {
	return fmt.Sprintf(c.GetChannelRoute(channelId) + "/members")
}

func (c *Client4) GetChannelMemberRoute(channelId, userId string) string {
	return fmt.Sprintf(c.GetChannelMembersRoute(channelId)+"/%v", userId)
}

func (c *Client4) GetPostsRoute() string {
	return fmt.Sprintf("/posts")
}

func (c *Client4) GetPostsEphemeralRoute() string {
	return fmt.Sprintf("/posts/ephemeral")
}

func (c *Client4) GetConfigRoute() string {
	return fmt.Sprintf("/config")
}

func (c *Client4) GetLicenseRoute() string {
	return fmt.Sprintf("/license")
}

func (c *Client4) GetPostRoute(postId string) string {
	return fmt.Sprintf(c.GetPostsRoute()+"/%v", postId)
}

func (c *Client4) GetFilesRoute() string {
	return fmt.Sprintf("/files")
}

func (c *Client4) GetFileRoute(fileId string) string {
	return fmt.Sprintf(c.GetFilesRoute()+"/%v", fileId)
}

func (c *Client4) GetPluginsRoute() string {
	return fmt.Sprintf("/plugins")
}

func (c *Client4) GetPluginRoute(pluginId string) string {
	return fmt.Sprintf(c.GetPluginsRoute()+"/%v", pluginId)
}

func (c *Client4) GetSystemRoute() string {
	return fmt.Sprintf("/system")
}

func (c *Client4) GetTestEmailRoute() string {
	return fmt.Sprintf("/email/test")
}

func (c *Client4) GetTestS3Route() string {
	return fmt.Sprintf("/file/s3_test")
}

func (c *Client4) GetDatabaseRoute() string {
	return fmt.Sprintf("/database")
}

func (c *Client4) GetCacheRoute() string {
	return fmt.Sprintf("/caches")
}

func (c *Client4) GetClusterRoute() string {
	return fmt.Sprintf("/cluster")
}

func (c *Client4) GetIncomingWebhooksRoute() string {
	return fmt.Sprintf("/hooks/incoming")
}

func (c *Client4) GetIncomingWebhookRoute(hookID string) string {
	return fmt.Sprintf(c.GetIncomingWebhooksRoute()+"/%v", hookID)
}

func (c *Client4) GetComplianceReportsRoute() string {
	return fmt.Sprintf("/compliance/reports")
}

func (c *Client4) GetComplianceReportRoute(reportId string) string {
	return fmt.Sprintf("/compliance/reports/%v", reportId)
}

func (c *Client4) GetOutgoingWebhooksRoute() string {
	return fmt.Sprintf("/hooks/outgoing")
}

func (c *Client4) GetOutgoingWebhookRoute(hookID string) string {
	return fmt.Sprintf(c.GetOutgoingWebhooksRoute()+"/%v", hookID)
}

func (c *Client4) GetPreferencesRoute(userId string) string {
	return fmt.Sprintf(c.GetUserRoute(userId) + "/preferences")
}

func (c *Client4) GetUserStatusRoute(userId string) string {
	return fmt.Sprintf(c.GetUserRoute(userId) + "/status")
}

func (c *Client4) GetUserStatusesRoute() string {
	return fmt.Sprintf(c.GetUsersRoute() + "/status")
}

func (c *Client4) GetSamlRoute() string {
	return fmt.Sprintf("/saml")
}

func (c *Client4) GetLdapRoute() string {
	return fmt.Sprintf("/ldap")
}

func (c *Client4) GetBrandRoute() string {
	return fmt.Sprintf("/brand")
}

func (c *Client4) GetDataRetentionRoute() string {
	return fmt.Sprintf("/data_retention")
}

func (c *Client4) GetElasticsearchRoute() string {
	return fmt.Sprintf("/elasticsearch")
}

func (c *Client4) GetCommandsRoute() string {
	return fmt.Sprintf("/commands")
}

func (c *Client4) GetCommandRoute(commandId string) string {
	return fmt.Sprintf(c.GetCommandsRoute()+"/%v", commandId)
}

func (c *Client4) GetEmojisRoute() string {
	return fmt.Sprintf("/emoji")
}

func (c *Client4) GetEmojiRoute(emojiId string) string {
	return fmt.Sprintf(c.GetEmojisRoute()+"/%v", emojiId)
}

func (c *Client4) GetEmojiByNameRoute(name string) string {
	return fmt.Sprintf(c.GetEmojisRoute()+"/name/%v", name)
}

func (c *Client4) GetReactionsRoute() string {
	return fmt.Sprintf("/reactions")
}

func (c *Client4) GetOAuthAppsRoute() string {
	return fmt.Sprintf("/oauth/apps")
}

func (c *Client4) GetOAuthAppRoute(appId string) string {
	return fmt.Sprintf("/oauth/apps/%v", appId)
}

func (c *Client4) GetOpenGraphRoute() string {
	return fmt.Sprintf("/opengraph")
}

func (c *Client4) GetJobsRoute() string {
	return fmt.Sprintf("/jobs")
}

func (c *Client4) GetRolesRoute() string {
	return fmt.Sprintf("/roles")
}

func (c *Client4) GetSchemesRoute() string {
	return fmt.Sprintf("/schemes")
}

func (c *Client4) GetSchemeRoute(id string) string {
	return c.GetSchemesRoute() + fmt.Sprintf("/%v", id)
}

func (c *Client4) GetAnalyticsRoute() string {
	return fmt.Sprintf("/analytics")
}

func (c *Client4) GetTimezonesRoute() string {
	return fmt.Sprintf(c.GetSystemRoute() + "/timezones")
}

func (c *Client4) GetChannelSchemeRoute(channelId string) string {
	return fmt.Sprintf(c.GetChannelsRoute()+"/%v/scheme", channelId)
}

func (c *Client4) GetTeamSchemeRoute(teamId string) string {
	return fmt.Sprintf(c.GetTeamsRoute()+"/%v/scheme", teamId)
}

func (c *Client4) GetTotalUsersStatsRoute() string {
	return fmt.Sprintf(c.GetUsersRoute() + "/stats")
}

func (c *Client4) GetRedirectLocationRoute() string {
	return fmt.Sprintf("/redirect_location")
}

func (c *Client4) GetUserTermsOfServiceRoute(userId string) string {
	return c.GetUserRoute(userId) + "/terms_of_service"
}

func (c *Client4) GetTermsOfServiceRoute() string {
	return "/terms_of_service"
}

func (c *Client4) GetGroupsRoute() string {
	return "/groups"
}

func (c *Client4) GetGroupRoute(groupID string) string {
	return fmt.Sprintf("%s/%s", c.GetGroupsRoute(), groupID)
}

func (c *Client4) DoApiGet(url string, etag string) (*http.Response, *AppError) {
	return c.DoApiRequest(http.MethodGet, c.ApiUrl+url, "", etag)
}

func (c *Client4) DoApiPost(url string, data string) (*http.Response, *AppError) {
	return c.DoApiRequest(http.MethodPost, c.ApiUrl+url, data, "")
}

func (c *Client4) doApiPostBytes(url string, data []byte) (*http.Response, *AppError) {
	return c.doApiRequestBytes(http.MethodPost, c.ApiUrl+url, data, "")
}

func (c *Client4) DoApiPut(url string, data string) (*http.Response, *AppError) {
	return c.DoApiRequest(http.MethodPut, c.ApiUrl+url, data, "")
}

func (c *Client4) doApiPutBytes(url string, data []byte) (*http.Response, *AppError) {
	return c.doApiRequestBytes(http.MethodPut, c.ApiUrl+url, data, "")
}

func (c *Client4) DoApiDelete(url string) (*http.Response, *AppError) {
	return c.DoApiRequest(http.MethodDelete, c.ApiUrl+url, "", "")
}

func (c *Client4) DoApiRequest(method, url, data, etag string) (*http.Response, *AppError) {
	return c.doApiRequestReader(method, url, strings.NewReader(data), etag)
}

func (c *Client4) doApiRequestBytes(method, url string, data []byte, etag string) (*http.Response, *AppError) {
	return c.doApiRequestReader(method, url, bytes.NewReader(data), etag)
}

func (c *Client4) doApiRequestReader(method, url string, data io.Reader, etag string) (*http.Response, *AppError) {
	rq, err := http.NewRequest(method, url, data)
	if err != nil {
		return nil, NewAppError(url, "model.client.connecting.app_error", nil, err.Error(), http.StatusBadRequest)
	}

	if len(etag) > 0 {
		rq.Header.Set(HEADER_ETAG_CLIENT, etag)
	}

	if len(c.AuthToken) > 0 {
		rq.Header.Set(HEADER_AUTH, c.AuthType+" "+c.AuthToken)
	}

	if c.HttpHeader != nil && len(c.HttpHeader) > 0 {
		for k, v := range c.HttpHeader {
			rq.Header.Set(k, v)
		}
	}

	rp, err := c.HttpClient.Do(rq)
	if err != nil || rp == nil {
		return nil, NewAppError(url, "model.client.connecting.app_error", nil, err.Error(), 0)
	}

	if rp.StatusCode == 304 {
		return rp, nil
	}

	if rp.StatusCode >= 300 {
		defer closeBody(rp)
		return rp, AppErrorFromJson(rp.Body)
	}

	return rp, nil
}

func (c *Client4) DoUploadFile(url string, data []byte, contentType string) (*FileUploadResponse, *Response) {
	return c.doUploadFile(url, bytes.NewReader(data), contentType, 0)
}

func (c *Client4) doUploadFile(url string, body io.Reader, contentType string, contentLength int64) (*FileUploadResponse, *Response) {
	rq, err := http.NewRequest("POST", c.ApiUrl+url, body)
	if err != nil {
		return nil, &Response{Error: NewAppError(url, "model.client.connecting.app_error", nil, err.Error(), http.StatusBadRequest)}
	}
	if contentLength != 0 {
		rq.ContentLength = contentLength
	}
	rq.Header.Set("Content-Type", contentType)

	if len(c.AuthToken) > 0 {
		rq.Header.Set(HEADER_AUTH, c.AuthType+" "+c.AuthToken)
	}

	rp, err := c.HttpClient.Do(rq)
	if err != nil || rp == nil {
		return nil, BuildErrorResponse(rp, NewAppError(url, "model.client.connecting.app_error", nil, err.Error(), 0))
	}
	defer closeBody(rp)

	if rp.StatusCode >= 300 {
		return nil, BuildErrorResponse(rp, AppErrorFromJson(rp.Body))
	}

	return FileUploadResponseFromJson(rp.Body), BuildResponse(rp)
}

// CheckStatusOK is a convenience function for checking the standard OK response
// from the web service.
func CheckStatusOK(r *http.Response) bool {
	m := MapFromJson(r.Body)
	defer closeBody(r)

	if m != nil && m[STATUS] == STATUS_OK {
		return true
	}

	return false
}

// Authentication Section

// LoginById authenticates a user by user id and password.
func (c *Client4) LoginById(id string, password string) (*User, *Response) {
	m := make(map[string]string)
	m["id"] = id
	m["password"] = password
	return c.login(m)
}

// Login authenticates a user by login id, which can be username, email or some sort
// of SSO identifier based on server configuration, and a password.
func (c *Client4) Login(loginId string, password string) (*User, *Response) {
	m := make(map[string]string)
	m["login_id"] = loginId
	m["password"] = password
	return c.login(m)
}

// LoginByLdap authenticates a user by LDAP id and password.
func (c *Client4) LoginByLdap(loginId string, password string) (*User, *Response) {
	m := make(map[string]string)
	m["login_id"] = loginId
	m["password"] = password
	m["ldap_only"] = "true"
	return c.login(m)
}

// LoginWithDevice authenticates a user by login id (username, email or some sort
// of SSO identifier based on configuration), password and attaches a device id to
// the session.
func (c *Client4) LoginWithDevice(loginId string, password string, deviceId string) (*User, *Response) {
	m := make(map[string]string)
	m["login_id"] = loginId
	m["password"] = password
	m["device_id"] = deviceId
	return c.login(m)
}

// LoginWithMFA logs a user in with a MFA token
func (c *Client4) LoginWithMFA(loginId, password, mfaToken string) (*User, *Response) {
	m := make(map[string]string)
	m["login_id"] = loginId
	m["password"] = password
	m["token"] = mfaToken
	return c.login(m)
}

func (c *Client4) login(m map[string]string) (*User, *Response) {
	r, err := c.DoApiPost("/users/login", MapToJson(m))
	if err != nil {
		return nil, BuildErrorResponse(r, err)
	}
	defer closeBody(r)
	c.AuthToken = r.Header.Get(HEADER_TOKEN)
	c.AuthType = HEADER_BEARER
	return UserFromJson(r.Body), BuildResponse(r)
}

// Logout terminates the current user's session.
func (c *Client4) Logout() (bool, *Response) {
	r, err := c.DoApiPost("/users/logout", "")
	if err != nil {
		return false, BuildErrorResponse(r, err)
	}
	defer closeBody(r)
	c.AuthToken = ""
	c.AuthType = HEADER_BEARER
	return CheckStatusOK(r), BuildResponse(r)
}

// SwitchAccountType changes a user's login type from one type to another.
func (c *Client4) SwitchAccountType(switchRequest *SwitchRequest) (string, *Response) {
	r, err := c.DoApiPost(c.GetUsersRoute()+"/login/switch", switchRequest.ToJson())
	if err != nil {
		return "", BuildErrorResponse(r, err)
	}
	defer closeBody(r)
	return MapFromJson(r.Body)["follow_link"], BuildResponse(r)
}

// User Section

// CreateUser creates a user in the system based on the provided user struct.
func (c *Client4) CreateUser(user *User) (*User, *Response) {
	r, err := c.DoApiPost(c.GetUsersRoute(), user.ToJson())
	if err != nil {
		return nil, BuildErrorResponse(r, err)
	}
	defer closeBody(r)
	return UserFromJson(r.Body), BuildResponse(r)
}

// CreateUserWithToken creates a user in the system based on the provided tokenId.
func (c *Client4) CreateUserWithToken(user *User, tokenId string) (*User, *Response) {
	if tokenId == "" {
		err := NewAppError("MissingHashOrData", "api.user.create_user.missing_token.app_error", nil, "", http.StatusBadRequest)
		return nil, &Response{StatusCode: err.StatusCode, Error: err}
	}

	query := fmt.Sprintf("?t=%v", tokenId)
	r, err := c.DoApiPost(c.GetUsersRoute()+query, user.ToJson())
	if err != nil {
		return nil, BuildErrorResponse(r, err)
	}
	defer closeBody(r)

	return UserFromJson(r.Body), BuildResponse(r)
}

// CreateUserWithInviteId creates a user in the system based on the provided invited id.
func (c *Client4) CreateUserWithInviteId(user *User, inviteId string) (*User, *Response) {
	if inviteId == "" {
		err := NewAppError("MissingInviteId", "api.user.create_user.missing_invite_id.app_error", nil, "", http.StatusBadRequest)
		return nil, &Response{StatusCode: err.StatusCode, Error: err}
	}

	query := fmt.Sprintf("?iid=%v", url.QueryEscape(inviteId))
	r, err := c.DoApiPost(c.GetUsersRoute()+query, user.ToJson())
	if err != nil {
		return nil, BuildErrorResponse(r, err)
	}
	defer closeBody(r)

	return UserFromJson(r.Body), BuildResponse(r)
}

// GetMe returns the logged in user.
func (c *Client4) GetMe(etag string) (*User, *Response) {
	r, err := c.DoApiGet(c.GetUserRoute(ME), etag)
	if err != nil {
		return nil, BuildErrorResponse(r, err)
	}
	defer closeBody(r)
	return UserFromJson(r.Body), BuildResponse(r)
}

// GetUser returns a user based on the provided user id string.
func (c *Client4) GetUser(userId, etag string) (*User, *Response) {
	r, err := c.DoApiGet(c.GetUserRoute(userId), etag)
	if err != nil {
		return nil, BuildErrorResponse(r, err)
	}
	defer closeBody(r)
	return UserFromJson(r.Body), BuildResponse(r)
}

// GetUserByUsername returns a user based on the provided user name string.
func (c *Client4) GetUserByUsername(userName, etag string) (*User, *Response) {
	r, err := c.DoApiGet(c.GetUserByUsernameRoute(userName), etag)
	if err != nil {
		return nil, BuildErrorResponse(r, err)
	}
	defer closeBody(r)
	return UserFromJson(r.Body), BuildResponse(r)
}

// GetUserByEmail returns a user based on the provided user email string.
func (c *Client4) GetUserByEmail(email, etag string) (*User, *Response) {
	r, err := c.DoApiGet(c.GetUserByEmailRoute(email), etag)
	if err != nil {
		return nil, BuildErrorResponse(r, err)
	}
	defer closeBody(r)
	return UserFromJson(r.Body), BuildResponse(r)
}

// AutocompleteUsersInTeam returns the users on a team based on search term.
func (c *Client4) AutocompleteUsersInTeam(teamId string, username string, limit int, etag string) (*UserAutocomplete, *Response) {
	query := fmt.Sprintf("?in_team=%v&name=%v&limit=%d", teamId, username, limit)
	r, err := c.DoApiGet(c.GetUsersRoute()+"/autocomplete"+query, etag)
	if err != nil {
		return nil, BuildErrorResponse(r, err)
	}
	defer closeBody(r)
	return UserAutocompleteFromJson(r.Body), BuildResponse(r)
}

// AutocompleteUsersInChannel returns the users in a channel based on search term.
func (c *Client4) AutocompleteUsersInChannel(teamId string, channelId string, username string, limit int, etag string) (*UserAutocomplete, *Response) {
	query := fmt.Sprintf("?in_team=%v&in_channel=%v&name=%v&limit=%d", teamId, channelId, username, limit)
	r, err := c.DoApiGet(c.GetUsersRoute()+"/autocomplete"+query, etag)
	if err != nil {
		return nil, BuildErrorResponse(r, err)
	}
	defer closeBody(r)
	return UserAutocompleteFromJson(r.Body), BuildResponse(r)
}

// AutocompleteUsers returns the users in the system based on search term.
func (c *Client4) AutocompleteUsers(username string, limit int, etag string) (*UserAutocomplete, *Response) {
	query := fmt.Sprintf("?name=%v&limit=%d", username, limit)
	r, err := c.DoApiGet(c.GetUsersRoute()+"/autocomplete"+query, etag)
	if err != nil {
		return nil, BuildErrorResponse(r, err)
	}
	defer closeBody(r)
	return UserAutocompleteFromJson(r.Body), BuildResponse(r)
}

// GetDefaultProfileImage gets the default user's profile image. Must be logged in.
func (c *Client4) GetDefaultProfileImage(userId string) ([]byte, *Response) {
	r, appErr := c.DoApiGet(c.GetUserRoute(userId)+"/image/default", "")
	if appErr != nil {
		return nil, BuildErrorResponse(r, appErr)
	}
	defer closeBody(r)

	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return nil, BuildErrorResponse(r, NewAppError("GetDefaultProfileImage", "model.client.read_file.app_error", nil, err.Error(), r.StatusCode))
	}

	return data, BuildResponse(r)
}

// GetProfileImage gets user's profile image. Must be logged in.
func (c *Client4) GetProfileImage(userId, etag string) ([]byte, *Response) {
	r, appErr := c.DoApiGet(c.GetUserRoute(userId)+"/image", etag)
	if appErr != nil {
		return nil, BuildErrorResponse(r, appErr)
	}
	defer closeBody(r)

	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return nil, BuildErrorResponse(r, NewAppError("GetProfileImage", "model.client.read_file.app_error", nil, err.Error(), r.StatusCode))
	}
	return data, BuildResponse(r)
}

// GetUsers returns a page of users on the system. Page counting starts at 0.
func (c *Client4) GetUsers(page int, perPage int, etag string) ([]*User, *Response) {
	query := fmt.Sprintf("?page=%v&per_page=%v", page, perPage)
	r, err := c.DoApiGet(c.GetUsersRoute()+query, etag)
	if err != nil {
		return nil, BuildErrorResponse(r, err)
	}
	defer closeBody(r)
	return UserListFromJson(r.Body), BuildResponse(r)
}

// GetUsersInTeam returns a page of users on a team. Page counting starts at 0.
func (c *Client4) GetUsersInTeam(teamId string, page int, perPage int, etag string) ([]*User, *Response) {
	query := fmt.Sprintf("?in_team=%v&page=%v&per_page=%v", teamId, page, perPage)
	r, err := c.DoApiGet(c.GetUsersRoute()+query, etag)
	if err != nil {
		return nil, BuildErrorResponse(r, err)
	}
	defer closeBody(r)
	return UserListFromJson(r.Body), BuildResponse(r)
}

// GetNewUsersInTeam returns a page of users on a team. Page counting starts at 0.
func (c *Client4) GetNewUsersInTeam(teamId string, page int, perPage int, etag string) ([]*User, *Response) {
	query := fmt.Sprintf("?sort=create_at&in_team=%v&page=%v&per_page=%v", teamId, page, perPage)
	r, err := c.DoApiGet(c.GetUsersRoute()+query, etag)
	if err != nil {
		return nil, BuildErrorResponse(r, err)
	}
	defer closeBody(r)
	return UserListFromJson(r.Body), BuildResponse(r)
}

// GetRecentlyActiveUsersInTeam returns a page of users on a team. Page counting starts at 0.
func (c *Client4) GetRecentlyActiveUsersInTeam(teamId string, page int, perPage int, etag string) ([]*User, *Response) {
	query := fmt.Sprintf("?sort=last_activity_at&in_team=%v&page=%v&per_page=%v", teamId, page, perPage)
	r, err := c.DoApiGet(c.GetUsersRoute()+query, etag)
	if err != nil {
		return nil, BuildErrorResponse(r, err)
	}
	defer closeBody(r)
	return UserListFromJson(r.Body), BuildResponse(r)
}

// GetUsersNotInTeam returns a page of users who are not in a team. Page counting starts at 0.
func (c *Client4) GetUsersNotInTeam(teamId string, page int, perPage int, etag string) ([]*User, *Response) {
	query := fmt.Sprintf("?not_in_team=%v&page=%v&per_page=%v", teamId, page, perPage)
	r, err := c.DoApiGet(c.GetUsersRoute()+query, etag)
	if err != nil {
		return nil, BuildErrorResponse(r, err)
	}
	defer closeBody(r)
	return UserListFromJson(r.Body), BuildResponse(r)
}

// GetUsersInChannel returns a page of users in a channel. Page counting starts at 0.
func (c *Client4) GetUsersInChannel(channelId string, page int, perPage int, etag string) ([]*User, *Response) {
	query := fmt.Sprintf("?in_channel=%v&page=%v&per_page=%v", channelId, page, perPage)
	r, err := c.DoApiGet(c.GetUsersRoute()+query, etag)
	if err != nil {
		return nil, BuildErrorResponse(r, err)
	}
	defer closeBody(r)
	return UserListFromJson(r.Body), BuildResponse(r)
}

// GetUsersInChannelByStatus returns a page of users in a channel. Page counting starts at 0. Sorted by Status
func (c *Client4) GetUsersInChannelByStatus(channelId string, page int, perPage int, etag string) ([]*User, *Response) {
	query := fmt.Sprintf("?in_channel=%v&page=%v&per_page=%v&sort=status", channelId, page, perPage)
	r, err := c.DoApiGet(c.GetUsersRoute()+query, etag)
	if err != nil {
		return nil, BuildErrorResponse(r, err)
	}
	defer closeBody(r)
	return UserListFromJson(r.Body), BuildResponse(r)
}

// GetUsersNotInChannel returns a page of users not in a channel. Page counting starts at 0.
func (c *Client4) GetUsersNotInChannel(teamId, channelId string, page int, perPage int, etag string) ([]*User, *Response) {
	query := fmt.Sprintf("?in_team=%v&not_in_channel=%v&page=%v&per_page=%v", teamId, channelId, page, perPage)
	r, err := c.DoApiGet(c.GetUsersRoute()+query, etag)
	if err != nil {
		return nil, BuildErrorResponse(r, err)
	}
	defer closeBody(r)
	return UserListFromJson(r.Body), BuildResponse(r)
}

// GetUsersWithoutTeam returns a page of users on the system that aren't on any teams. Page counting starts at 0.
func (c *Client4) GetUsersWithoutTeam(page int, perPage int, etag string) ([]*User, *Response) {
	query := fmt.Sprintf("?without_team=1&page=%v&per_page=%v", page, perPage)
	r, err := c.DoApiGet(c.GetUsersRoute()+query, etag)
	if err != nil {
		return nil, BuildErrorResponse(r, err)
	}
	defer closeBody(r)
	return UserListFromJson(r.Body), BuildResponse(r)
}

// GetUsersByIds returns a list of users based on the provided user ids.
func (c *Client4) GetUsersByIds(userIds []string) ([]*User, *Response) {
	r, err := c.DoApiPost(c.GetUsersRoute()+"/ids", ArrayToJson(userIds))
	if err != nil {
		return nil, BuildErrorResponse(r, err)
	}
	defer closeBody(r)
	return UserListFromJson(r.Body), BuildResponse(r)
}

// GetUsersByUsernames returns a list of users based on the provided usernames.
func (c *Client4) GetUsersByUsernames(usernames []string) ([]*User, *Response) {
	r, err := c.DoApiPost(c.GetUsersRoute()+"/usernames", ArrayToJson(usernames))
	if err != nil {
		return nil, BuildErrorResponse(r, err)
	}
	defer closeBody(r)
	return UserListFromJson(r.Body), BuildResponse(r)
}

// SearchUsers returns a list of users based on some search criteria.
func (c *Client4) SearchUsers(search *UserSearch) ([]*User, *Response) {
	r, err := c.doApiPostBytes(c.GetUsersRoute()+"/search", search.ToJson())
	if err != nil {
		return nil, BuildErrorResponse(r, err)
	}
	defer closeBody(r)
	return UserListFromJson(r.Body), BuildResponse(r)
}

// UpdateUser updates a user in the system based on the provided user struct.
func (c *Client4) UpdateUser(user *User) (*User, *Response) {
	r, err := c.DoApiPut(c.GetUserRoute(user.Id), user.ToJson())
	if err != nil {
		return nil, BuildErrorResponse(r, err)
	}
	defer closeBody(r)
	return UserFromJson(r.Body), BuildResponse(r)
}

// PatchUser partially updates a user in the system. Any missing fields are not updated.
func (c *Client4) PatchUser(userId string, patch *UserPatch) (*User, *Response) {
	r, err := c.DoApiPut(c.GetUserRoute(userId)+"/patch", patch.ToJson())
	if err != nil {
		return nil, BuildErrorResponse(r, err)
	}
	defer closeBody(r)
	return UserFromJson(r.Body), BuildResponse(r)
}

// UpdateUserAuth updates a user AuthData (uthData, authService and password) in the system.
func (c *Client4) UpdateUserAuth(userId string, userAuth *UserAuth) (*UserAuth, *Response) {
	r, err := c.DoApiPut(c.GetUserRoute(userId)+"/auth", userAuth.ToJson())
	if err != nil {
		return nil, BuildErrorResponse(r, err)
	}
	defer closeBody(r)
	return UserAuthFromJson(r.Body), BuildResponse(r)
}

// UpdateUserMfa activates multi-factor authentication for a user if activate
// is true and a valid code is provided. If activate is false, then code is not
// required and multi-factor authentication is disabled for the user.
func (c *Client4) UpdateUserMfa(userId, code string, activate bool) (bool, *Response) {
	requestBody := make(map[string]interface{})
	requestBody["activate"] = activate
	requestBody["code"] = code

	r, err := c.DoApiPut(c.GetUserRoute(userId)+"/mfa", StringInterfaceToJson(requestBody))
	if err != nil {
		return false, BuildErrorResponse(r, err)
	}
	defer closeBody(r)
	return CheckStatusOK(r), BuildResponse(r)
}

// CheckUserMfa checks whether a user has MFA active on their account or not based on the
// provided login id.
// Deprecated: Clients should use Login method and check for MFA Error
func (c *Client4) CheckUserMfa(loginId string) (bool, *Response) {
	requestBody := make(map[string]interface{})
	requestBody["login_id"] = loginId
	r, err := c.DoApiPost(c.GetUsersRoute()+"/mfa", StringInterfaceToJson(requestBody))
	if err != nil {
		return false, BuildErrorResponse(r, err)
	}
	defer closeBody(r)

	data := StringInterfaceFromJson(r.Body)
	mfaRequired, ok := data["mfa_required"].(bool)
	if !ok {
		return false, BuildResponse(r)
	}
	return mfaRequired, BuildResponse(r)
}

// UpdateUserPassword updates a user's password. Must be logged in as the user or be a system administrator.
func (c *Client4) UpdateUserPassword(userId, currentPassword, newPassword string) (bool, *Response) {
	requestBody := map[string]string{"current_password": currentPassword, "new_password": newPassword}
	r, err := c.DoApiPut(c.GetUserRoute(userId)+"/password", MapToJson(requestBody))
	if err != nil {
		return false, BuildErrorResponse(r, err)
	}
	defer closeBody(r)
	return CheckStatusOK(r), BuildResponse(r)
}

// UpdateUserRoles updates a user's roles in the system. A user can have "system_user" and "system_admin" roles.
func (c *Client4) UpdateUserRoles(userId, roles string) (bool, *Response) {
	requestBody := map[string]string{"roles": roles}
	r, err := c.DoApiPut(c.GetUserRoute(userId)+"/roles", MapToJson(requestBody))
	if err != nil {
		return false, BuildErrorResponse(r, err)
	}
	defer closeBody(r)
	return CheckStatusOK(r), BuildResponse(r)
}

// UpdateUserActive updates status of a user whether active or not.
func (c *Client4) UpdateUserActive(userId string, active bool) (bool, *Response) {
	requestBody := make(map[string]interface{})
	requestBody["active"] = active
	r, err := c.DoApiPut(c.GetUserRoute(userId)+"/active", StringInterfaceToJson(requestBody))
	if err != nil {
		return false, BuildErrorResponse(r, err)
	}
	defer closeBody(r)

	return CheckStatusOK(r), BuildResponse(r)
}

// DeleteUser deactivates a user in the system based on the provided user id string.
func (c *Client4) DeleteUser(userId string) (bool, *Response) {
	r, err := c.DoApiDelete(c.GetUserRoute(userId))
	if err != nil {
		return false, BuildErrorResponse(r, err)
	}
	defer closeBody(r)
	return CheckStatusOK(r), BuildResponse(r)
}

// SendPasswordResetEmail will send a link for password resetting to a user with the
// provided email.
func (c *Client4) SendPasswordResetEmail(email string) (bool, *Response) {
	requestBody := map[string]string{"email": email}
	r, err := c.DoApiPost(c.GetUsersRoute()+"/password/reset/send", MapToJson(requestBody))
	if err != nil {
		return false, BuildErrorResponse(r, err)
	}
	defer closeBody(r)
	return CheckStatusOK(r), BuildResponse(r)
}

// ResetPassword uses a recovery code to update reset a user's password.
func (c *Client4) ResetPassword(token, newPassword string) (bool, *Response) {
	requestBody := map[string]string{"token": token, "new_password": newPassword}
	r, err := c.DoApiPost(c.GetUsersRoute()+"/password/reset", MapToJson(requestBody))
	if err != nil {
		return false, BuildErrorResponse(r, err)
	}
	defer closeBody(r)
	return CheckStatusOK(r), BuildResponse(r)
}

// GetSessions returns a list of sessions based on the provided user id string.
func (c *Client4) GetSessions(userId, etag string) ([]*Session, *Response) {
	r, err := c.DoApiGet(c.GetUserRoute(userId)+"/sessions", etag)
	if err != nil {
		return nil, BuildErrorResponse(r, err)
	}
	defer closeBody(r)
	return SessionsFromJson(r.Body), BuildResponse(r)
}

// RevokeSession revokes a user session based on the provided user id and session id strings.
func (c *Client4) RevokeSession(userId, sessionId string) (bool, *Response) {
	requestBody := map[string]string{"session_id": sessionId}
	r, err := c.DoApiPost(c.GetUserRoute(userId)+"/sessions/revoke", MapToJson(requestBody))
	if err != nil {
		return false, BuildErrorResponse(r, err)
	}
	defer closeBody(r)
	return CheckStatusOK(r), BuildResponse(r)
}

// RevokeAllSessions revokes all sessions for the provided user id string.
func (c *Client4) RevokeAllSessions(userId string) (bool, *Response) {
	r, err := c.DoApiPost(c.GetUserRoute(userId)+"/sessions/revoke/all", "")
	if err != nil {
		return false, BuildErrorResponse(r, err)
	}
	defer closeBody(r)
	return CheckStatusOK(r), BuildResponse(r)
}

// AttachDeviceId attaches a mobile device ID to the current session.
func (c *Client4) AttachDeviceId(deviceId string) (bool, *Response) {
	requestBody := map[string]string{"device_id": deviceId}
	r, err := c.DoApiPut(c.GetUsersRoute()+"/sessions/device", MapToJson(requestBody))
	if err != nil {
		return false, BuildErrorResponse(r, err)
	}
	defer closeBody(r)
	return CheckStatusOK(r), BuildResponse(r)
}

// GetTeamsUnreadForUser will return an array with TeamUnread objects that contain the amount
// of unread messages and mentions the current user has for the teams it belongs to.
// An optional team ID can be set to exclude that team from the results. Must be authenticated.
func (c *Client4) GetTeamsUnreadForUser(userId, teamIdToExclude string) ([]*TeamUnread, *Response) {
	var optional string
	if teamIdToExclude != "" {
		optional += fmt.Sprintf("?exclude_team=%s", url.QueryEscape(teamIdToExclude))
	}

	r, err := c.DoApiGet(c.GetUserRoute(userId)+"/teams/unread"+optional, "")
	if err != nil {
		return nil, BuildErrorResponse(r, err)
	}
	defer closeBody(r)
	return TeamsUnreadFromJson(r.Body), BuildResponse(r)
}

// GetUserAudits returns a list of audit based on the provided user id string.
func (c *Client4) GetUserAudits(userId string, page int, perPage int, etag string) (Audits, *Response) {
	query := fmt.Sprintf("?page=%v&per_page=%v", page, perPage)
	r, err := c.DoApiGet(c.GetUserRoute(userId)+"/audits"+query, etag)
	if err != nil {
		return nil, BuildErrorResponse(r, err)
	}
	defer closeBody(r)
	return AuditsFromJson(r.Body), BuildResponse(r)
}

// VerifyUserEmail will verify a user's email using the supplied token.
func (c *Client4) VerifyUserEmail(token string) (bool, *Response) {
	requestBody := map[string]string{"token": token}
	r, err := c.DoApiPost(c.GetUsersRoute()+"/email/verify", MapToJson(requestBody))
	if err != nil {
		return false, BuildErrorResponse(r, err)
	}
	defer closeBody(r)
	return CheckStatusOK(r), BuildResponse(r)
}

// SendVerificationEmail will send an email to the user with the provided email address, if
// that user exists. The email will contain a link that can be used to verify the user's
// email address.
func (c *Client4) SendVerificationEmail(email string) (bool, *Response) {
	requestBody := map[string]string{"email": email}
	r, err := c.DoApiPost(c.GetUsersRoute()+"/email/verify/send", MapToJson(requestBody))
	if err != nil {
		return false, BuildErrorResponse(r, err)
	}
	defer closeBody(r)
	return CheckStatusOK(r), BuildResponse(r)
}

// SetDefaultProfileImage resets the profile image to a default generated one.
func (c *Client4) SetDefaultProfileImage(userId string) (bool, *Response) {
	r, err := c.DoApiDelete(c.GetUserRoute(userId) + "/image")
	if err != nil {
		return false, BuildErrorResponse(r, err)
	}
	return CheckStatusOK(r), BuildResponse(r)
}

// SetProfileImage sets profile image of the user.
func (c *Client4) SetProfileImage(userId string, data []byte) (bool, *Response) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	part, err := writer.CreateFormFile("image", "profile.png")
	if err != nil {
		return false, &Response{Error: NewAppError("SetProfileImage", "model.client.set_profile_user.no_file.app_error", nil, err.Error(), http.StatusBadRequest)}
	}

	if _, err = io.Copy(part, bytes.NewBuffer(data)); err != nil {
		return false, &Response{Error: NewAppError("SetProfileImage", "model.client.set_profile_user.no_file.app_error", nil, err.Error(), http.StatusBadRequest)}
	}

	if err = writer.Close(); err != nil {
		return false, &Response{Error: NewAppError("SetProfileImage", "model.client.set_profile_user.writer.app_error", nil, err.Error(), http.StatusBadRequest)}
	}

	rq, err := http.NewRequest("POST", c.ApiUrl+c.GetUserRoute(userId)+"/image", bytes.NewReader(body.Bytes()))
	if err != nil {
		return false, &Response{Error: NewAppError("SetProfileImage", "model.client.connecting.app_error", nil, err.Error(), http.StatusBadRequest)}
	}
	rq.Header.Set("Content-Type", writer.FormDataContentType())

	if len(c.AuthToken) > 0 {
		rq.Header.Set(HEADER_AUTH, c.AuthType+" "+c.AuthToken)
	}

	rp, err := c.HttpClient.Do(rq)
	if err != nil || rp == nil {
		return false, &Response{StatusCode: http.StatusForbidden, Error: NewAppError(c.GetUserRoute(userId)+"/image", "model.client.connecting.app_error", nil, err.Error(), http.StatusForbidden)}
	}
	defer closeBody(rp)

	if rp.StatusCode >= 300 {
		return false, BuildErrorResponse(rp, AppErrorFromJson(rp.Body))
	}

	return CheckStatusOK(rp), BuildResponse(rp)
}

// CreateUserAccessToken will generate a user access token that can be used in place
// of a session token to access the REST API. Must have the 'create_user_access_token'
// permission and if generating for another user, must have the 'edit_other_users'
// permission. A non-blank description is required.
func (c *Client4) CreateUserAccessToken(userId, description string) (*UserAccessToken, *Response) {
	requestBody := map[string]string{"description": description}
	r, err := c.DoApiPost(c.GetUserRoute(userId)+"/tokens", MapToJson(requestBody))
	if err != nil {
		return nil, BuildErrorResponse(r, err)
	}
	defer closeBody(r)
	return UserAccessTokenFromJson(r.Body), BuildResponse(r)
}

// GetUserAccessTokens will get a page of access tokens' id, description, is_active
// and the user_id in the system. The actual token will not be returned. Must have
// the 'manage_system' permission.
func (c *Client4) GetUserAccessTokens(page int, perPage int) ([]*UserAccessToken, *Response) {
	query := fmt.Sprintf("?page=%v&per_page=%v", page, perPage)
	r, err := c.DoApiGet(c.GetUserAccessTokensRoute()+query, "")
	if err != nil {
		return nil, BuildErrorResponse(r, err)
	}
	defer closeBody(r)
	return UserAccessTokenListFromJson(r.Body), BuildResponse(r)
}

// GetUserAccessToken will get a user access tokens' id, description, is_active
// and the user_id of the user it is for. The actual token will not be returned.
// Must have the 'read_user_access_token' permission and if getting for another
// user, must have the 'edit_other_users' permission.
func (c *Client4) GetUserAccessToken(tokenId string) (*UserAccessToken, *Response) {
	r, err := c.DoApiGet(c.GetUserAccessTokenRoute(tokenId), "")
	if err != nil {
		return nil, BuildErrorResponse(r, err)
	}
	defer closeBody(r)
	return UserAccessTokenFromJson(r.Body), BuildResponse(r)
}

// GetUserAccessTokensForUser will get a paged list of user access tokens showing id,
// description and user_id for each. The actual tokens will not be returned. Must have
// the 'read_user_access_token' permission and if getting for another user, must have the
// 'edit_other_users' permission.
func (c *Client4) GetUserAccessTokensForUser(userId string, page, perPage int) ([]*UserAccessToken, *Response) {
	query := fmt.Sprintf("?page=%v&per_page=%v", page, perPage)
	r, err := c.DoApiGet(c.GetUserRoute(userId)+"/tokens"+query, "")
	if err != nil {
		return nil, BuildErrorResponse(r, err)
	}
	defer closeBody(r)
	return UserAccessTokenListFromJson(r.Body), BuildResponse(r)
}

// RevokeUserAccessToken will revoke a user access token by id. Must have the
// 'revoke_user_access_token' permission and if revoking for another user, must have the
// 'edit_other_users' permission.
func (c *Client4) RevokeUserAccessToken(tokenId string) (bool, *Response) {
	requestBody := map[string]string{"token_id": tokenId}
	r, err := c.DoApiPost(c.GetUsersRoute()+"/tokens/revoke", MapToJson(requestBody))
	if err != nil {
		return false, BuildErrorResponse(r, err)
	}
	defer closeBody(r)
	return CheckStatusOK(r), BuildResponse(r)
}

// SearchUserAccessTokens returns user access tokens matching the provided search term.
func (c *Client4) SearchUserAccessTokens(search *UserAccessTokenSearch) ([]*UserAccessToken, *Response) {
	r, err := c.DoApiPost(c.GetUsersRoute()+"/tokens/search", search.ToJson())
	if err != nil {
		return nil, BuildErrorResponse(r, err)
	}
	defer closeBody(r)
	return UserAccessTokenListFromJson(r.Body), BuildResponse(r)
}

// DisableUserAccessToken will disable a user access token by id. Must have the
// 'revoke_user_access_token' permission and if disabling for another user, must have the
// 'edit_other_users' permission.
func (c *Client4) DisableUserAccessToken(tokenId string) (bool, *Response) {
	requestBody := map[string]string{"token_id": tokenId}
	r, err := c.DoApiPost(c.GetUsersRoute()+"/tokens/disable", MapToJson(requestBody))
	if err != nil {
		return false, BuildErrorResponse(r, err)
	}
	defer closeBody(r)
	return CheckStatusOK(r), BuildResponse(r)
}

// EnableUserAccessToken will enable a user access token by id. Must have the
// 'create_user_access_token' permission and if enabling for another user, must have the
// 'edit_other_users' permission.
func (c *Client4) EnableUserAccessToken(tokenId string) (bool, *Response) {
	requestBody := map[string]string{"token_id": tokenId}
	r, err := c.DoApiPost(c.GetUsersRoute()+"/tokens/enable", MapToJson(requestBody))
	if err != nil {
		return false, BuildErrorResponse(r, err)
	}
	defer closeBody(r)
	return CheckStatusOK(r), BuildResponse(r)
}

// Team Section

// CreateTeam creates a team in the system based on the provided team struct.
func (c *Client4) CreateTeam(team *Team) (*Team, *Response) {
	r, err := c.DoApiPost(c.GetTeamsRoute(), team.ToJson())
	if err != nil {
		return nil, BuildErrorResponse(r, err)
	}
	defer closeBody(r)
	return TeamFromJson(r.Body), BuildResponse(r)
}

// GetTeam returns a team based on the provided team id string.
func (c *Client4) GetTeam(teamId, etag string) (*Team, *Response) {
	r, err := c.DoApiGet(c.GetTeamRoute(teamId), etag)
	if err != nil {
		return nil, BuildErrorResponse(r, err)
	}
	defer closeBody(r)
	return TeamFromJson(r.Body), BuildResponse(r)
}

// GetAllTeams returns all teams based on permissions.
func (c *Client4) GetAllTeams(etag string, page int, perPage int) ([]*Team, *Response) {
	query := fmt.Sprintf("?page=%v&per_page=%v", page, perPage)
	r, err := c.DoApiGet(c.GetTeamsRoute()+query, etag)
	if err != nil {
		return nil, BuildErrorResponse(r, err)
	}
	defer closeBody(r)
	return TeamListFromJson(r.Body), BuildResponse(r)
}

// GetTeamByName returns a team based on the provided team name string.
func (c *Client4) GetTeamByName(name, etag string) (*Team, *Response) {
	r, err := c.DoApiGet(c.GetTeamByNameRoute(name), etag)
	if err != nil {
		return nil, BuildErrorResponse(r, err)
	}
	defer closeBody(r)
	return TeamFromJson(r.Body), BuildResponse(r)
}

// SearchTeams returns teams matching the provided search term.
func (c *Client4) SearchTeams(search *TeamSearch) ([]*Team, *Response) {
	r, err := c.DoApiPost(c.GetTeamsRoute()+"/search", search.ToJson())
	if err != nil {
		return nil, BuildErrorResponse(r, err)
	}
	defer closeBody(r)
	return TeamListFromJson(r.Body), BuildResponse(r)
}

// TeamExists returns true or false if the team exist or not.
func (c *Client4) TeamExists(name, etag string) (bool, *Response) {
	r, err := c.DoApiGet(c.GetTeamByNameRoute(name)+"/exists", etag)
	if err != nil {
		return false, BuildErrorResponse(r, err)
	}
	defer closeBody(r)
	return MapBoolFromJson(r.Body)["exists"], BuildResponse(r)
}

// GetTeamsForUser returns a list of teams a user is on. Must be logged in as the user
// or be a system administrator.
func (c *Client4) GetTeamsForUser(userId, etag string) ([]*Team, *Response) {
	r, err := c.DoApiGet(c.GetUserRoute(userId)+"/teams", etag)
	if err != nil {
		return nil, BuildErrorResponse(r, err)
	}
	defer closeBody(r)
	return TeamListFromJson(r.Body), BuildResponse(r)
}

// GetTeamMember returns a team member based on the provided team and user id strings.
func (c *Client4) GetTeamMember(teamId, userId, etag string) (*TeamMember, *Response) {
	r, err := c.DoApiGet(c.GetTeamMemberRoute(teamId, userId), etag)
	if err != nil {
		return nil, BuildErrorResponse(r, err)
	}
	defer closeBody(r)
	return TeamMemberFromJson(r.Body), BuildResponse(r)
}

// UpdateTeamMemberRoles will update the roles on a team for a user.
func (c *Client4) UpdateTeamMemberRoles(teamId, userId, newRoles string) (bool, *Response) {
	requestBody := map[string]string{"roles": newRoles}
	r, err := c.DoApiPut(c.GetTeamMemberRoute(teamId, userId)+"/roles", MapToJson(requestBody))
	if err != nil {
		return false, BuildErrorResponse(r, err)
	}
	defer closeBody(r)
	return CheckStatusOK(r), BuildResponse(r)
}

// UpdateTeamMemberSchemeRoles will update the scheme-derived roles on a team for a user.
func (c *Client4) UpdateTeamMemberSchemeRoles(teamId string, userId string, schemeRoles *SchemeRoles) (bool, *Response) {
	r, err := c.DoApiPut(c.GetTeamMemberRoute(teamId, userId)+"/schemeRoles", schemeRoles.ToJson())
	if err != nil {
		return false, BuildErrorResponse(r, err)
	}
	defer closeBody(r)
	return CheckStatusOK(r), BuildResponse(r)
}

// UpdateTeam will update a team.
func (c *Client4) UpdateTeam(team *Team) (*Team, *Response) {
	r, err := c.DoApiPut(c.GetTeamRoute(team.Id), team.ToJson())
	if err != nil {
		return nil, BuildErrorResponse(r, err)
	}
	defer closeBody(r)
	return TeamFromJson(r.Body), BuildResponse(r)
}

// PatchTeam partially updates a team. Any missing fields are not updated.
func (c *Client4) PatchTeam(teamId string, patch *TeamPatch) (*Team, *Response) {
	r, err := c.DoApiPut(c.GetTeamRoute(teamId)+"/patch", patch.ToJson())
	if err != nil {
		return nil, BuildErrorResponse(r, err)
	}
	defer closeBody(r)
	return TeamFromJson(r.Body), BuildResponse(r)
}

// SoftDeleteTeam deletes the team softly (archive only, not permanent delete).
func (c *Client4) SoftDeleteTeam(teamId string) (bool, *Response) {
	r, err := c.DoApiDelete(c.GetTeamRoute(teamId))
	if err != nil {
		return false, BuildErrorResponse(r, err)
	}
	defer closeBody(r)
	return CheckStatusOK(r), BuildResponse(r)
}

// PermanentDeleteTeam deletes the team, should only be used when needed for
// compliance and the like.
func (c *Client4) PermanentDeleteTeam(teamId string) (bool, *Response) {
	r, err := c.DoApiDelete(c.GetTeamRoute(teamId) + "?permanent=true")
	if err != nil {
		return false, BuildErrorResponse(r, err)
	}
	defer closeBody(r)
	return CheckStatusOK(r), BuildResponse(r)
}

// GetTeamMembers returns team members based on the provided team id string.
func (c *Client4) GetTeamMembers(teamId string, page int, perPage int, etag string) ([]*TeamMember, *Response) {
	query := fmt.Sprintf("?page=%v&per_page=%v", page, perPage)
	r, err := c.DoApiGet(c.GetTeamMembersRoute(teamId)+query, etag)
	if err != nil {
		return nil, BuildErrorResponse(r, err)
	}
	defer closeBody(r)
	return TeamMembersFromJson(r.Body), BuildResponse(r)
}

// GetTeamMembersForUser returns the team members for a user.
func (c *Client4) GetTeamMembersForUser(userId string, etag string) ([]*TeamMember, *Response) {
	r, err := c.DoApiGet(c.GetUserRoute(userId)+"/teams/members", etag)
	if err != nil {
		return nil, BuildErrorResponse(r, err)
	}
	defer closeBody(r)
	return TeamMembersFromJson(r.Body), BuildResponse(r)
}

// GetTeamMembersByIds will return an array of team members based on the
// team id and a list of user ids provided. Must be authenticated.
func (c *Client4) GetTeamMembersByIds(teamId string, userIds []string) ([]*TeamMember, *Response) {
	r, err := c.DoApiPost(fmt.Sprintf("/teams/%v/members/ids", teamId), ArrayToJson(userIds))
	if err != nil {
		return nil, BuildErrorResponse(r, err)
	}
	defer closeBody(r)
	return TeamMembersFromJson(r.Body), BuildResponse(r)
}

// AddTeamMember adds user to a team and return a team member.
func (c *Client4) AddTeamMember(teamId, userId string) (*TeamMember, *Response) {
	member := &TeamMember{TeamId: teamId, UserId: userId}
	r, err := c.DoApiPost(c.GetTeamMembersRoute(teamId), member.ToJson())
	if err != nil {
		return nil, BuildErrorResponse(r, err)
	}
	defer closeBody(r)
	return TeamMemberFromJson(r.Body), BuildResponse(r)
}

// AddTeamMemberFromInvite adds a user to a team and return a team member using an invite id
// or an invite token/data pair.
func (c *Client4) AddTeamMemberFromInvite(token, inviteId string) (*TeamMember, *Response) {
	var query string

	if inviteId != "" {
		query += fmt.Sprintf("?invite_id=%v", inviteId)
	}

	if token != "" {
		query += fmt.Sprintf("?token=%v", token)
	}

	r, err := c.DoApiPost(c.GetTeamsRoute()+"/members/invite"+query, "")
	if err != nil {
		return nil, BuildErrorResponse(r, err)
	}
	defer closeBody(r)
	return TeamMemberFromJson(r.Body), BuildResponse(r)
}

// AddTeamMembers adds a number of users to a team and returns the team members.
func (c *Client4) AddTeamMembers(teamId string, userIds []string) ([]*TeamMember, *Response) {
	var members []*TeamMember
	for _, userId := range userIds {
		member := &TeamMember{TeamId: teamId, UserId: userId}
		members = append(members, member)
	}

	r, err := c.DoApiPost(c.GetTeamMembersRoute(teamId)+"/batch", TeamMembersToJson(members))
	if err != nil {
		return nil, BuildErrorResponse(r, err)
	}
	defer closeBody(r)
	return TeamMembersFromJson(r.Body), BuildResponse(r)
}

// RemoveTeamMember will remove a user from a team.
func (c *Client4) RemoveTeamMember(teamId, userId string) (bool, *Response) {
	r, err := c.DoApiDelete(c.GetTeamMemberRoute(teamId, userId))
	if err != nil {
		return false, BuildErrorResponse(r, err)
	}
	defer closeBody(r)
	return CheckStatusOK(r), BuildResponse(r)
}

// GetTeamStats returns a team stats based on the team id string.
// Must be authenticated.
func (c *Client4) GetTeamStats(teamId, etag string) (*TeamStats, *Response) {
	r, err := c.DoApiGet(c.GetTeamStatsRoute(teamId), etag)
	if err != nil {
		return nil, BuildErrorResponse(r, err)
	}
	defer closeBody(r)
	return TeamStatsFromJson(r.Body), BuildResponse(r)
}

// GetTotalUsersStats returns a total system user stats.
// Must be authenticated.
func (c *Client4) GetTotalUsersStats(etag string) (*UsersStats, *Response) {
	r, err := c.DoApiGet(c.GetTotalUsersStatsRoute(), etag)
	if err != nil {
		return nil, BuildErrorResponse(r, err)
	}
	defer closeBody(r)
	return UsersStatsFromJson(r.Body), BuildResponse(r)
}

// GetTeamUnread will return a TeamUnread object that contains the amount of
// unread messages and mentions the user has for the specified team.
// Must be authenticated.
func (c *Client4) GetTeamUnread(teamId, userId string) (*TeamUnread, *Response) {
	r, err := c.DoApiGet(c.GetUserRoute(userId)+c.GetTeamRoute(teamId)+"/unread", "")
	if err != nil {
		return nil, BuildErrorResponse(r, err)
	}
	defer closeBody(r)
	return TeamUnreadFromJson(r.Body), BuildResponse(r)
}

// InviteUsersToTeam invite users by email to the team.
func (c *Client4) InviteUsersToTeam(teamId string, userEmails []string) (bool, *Response) {
	r, err := c.DoApiPost(c.GetTeamRoute(teamId)+"/invite/email", ArrayToJson(userEmails))
	if err != nil {
		return false, BuildErrorResponse(r, err)
	}
	defer closeBody(r)
	return CheckStatusOK(r), BuildResponse(r)
}

// InvalidateEmailInvites will invalidate active email invitations that have not been accepted by the user.
func (c *Client4) InvalidateEmailInvites() (bool, *Response) {
	r, err := c.DoApiDelete(c.GetTeamsRoute() + "/invites/email")
	if err != nil {
		return false, BuildErrorResponse(r, err)
	}
	defer closeBody(r)
	return CheckStatusOK(r), BuildResponse(r)
}

// GetTeamInviteInfo returns a team object from an invite id containing sanitized information.
func (c *Client4) GetTeamInviteInfo(inviteId string) (*Team, *Response) {
	r, err := c.DoApiGet(c.GetTeamsRoute()+"/invite/"+inviteId, "")
	if err != nil {
		return nil, BuildErrorResponse(r, err)
	}
	defer closeBody(r)
	return TeamFromJson(r.Body), BuildResponse(r)
}

// SetTeamIcon sets team icon of the team.
func (c *Client4) SetTeamIcon(teamId string, data []byte) (bool, *Response) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	part, err := writer.CreateFormFile("image", "teamIcon.png")
	if err != nil {
		return false, &Response{Error: NewAppError("SetTeamIcon", "model.client.set_team_icon.no_file.app_error", nil, err.Error(), http.StatusBadRequest)}
	}

	if _, err = io.Copy(part, bytes.NewBuffer(data)); err != nil {
		return false, &Response{Error: NewAppError("SetTeamIcon", "model.client.set_team_icon.no_file.app_error", nil, err.Error(), http.StatusBadRequest)}
	}

	if err = writer.Close(); err != nil {
		return false, &Response{Error: NewAppError("SetTeamIcon", "model.client.set_team_icon.writer.app_error", nil, err.Error(), http.StatusBadRequest)}
	}

	rq, err := http.NewRequest("POST", c.ApiUrl+c.GetTeamRoute(teamId)+"/image", bytes.NewReader(body.Bytes()))
	if err != nil {
		return false, &Response{Error: NewAppError("SetTeamIcon", "model.client.connecting.app_error", nil, err.Error(), http.StatusBadRequest)}
	}
	rq.Header.Set("Content-Type", writer.FormDataContentType())

	if len(c.AuthToken) > 0 {
		rq.Header.Set(HEADER_AUTH, c.AuthType+" "+c.AuthToken)
	}

	rp, err := c.HttpClient.Do(rq)
	if err != nil || rp == nil {
		// set to http.StatusForbidden(403)
		return false, &Response{StatusCode: http.StatusForbidden, Error: NewAppError(c.GetTeamRoute(teamId)+"/image", "model.client.connecting.app_error", nil, err.Error(), 403)}
	}
	defer closeBody(rp)

	if rp.StatusCode >= 300 {
		return false, BuildErrorResponse(rp, AppErrorFromJson(rp.Body))
	}

	return CheckStatusOK(rp), BuildResponse(rp)
}

// GetTeamIcon gets the team icon of the team.
func (c *Client4) GetTeamIcon(teamId, etag string) ([]byte, *Response) {
	r, appErr := c.DoApiGet(c.GetTeamRoute(teamId)+"/image", etag)
	if appErr != nil {
		return nil, BuildErrorResponse(r, appErr)
	}
	defer closeBody(r)

	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return nil, BuildErrorResponse(r, NewAppError("GetTeamIcon", "model.client.get_team_icon.app_error", nil, err.Error(), r.StatusCode))
	}
	return data, BuildResponse(r)
}

// RemoveTeamIcon updates LastTeamIconUpdate to 0 which indicates team icon is removed.
func (c *Client4) RemoveTeamIcon(teamId string) (bool, *Response) {
	r, err := c.DoApiDelete(c.GetTeamRoute(teamId) + "/image")
	if err != nil {
		return false, BuildErrorResponse(r, err)
	}
	defer closeBody(r)
	return CheckStatusOK(r), BuildResponse(r)
}

// Channel Section

// GetAllChannels get all the channels. Must be a system administrator.
func (c *Client4) GetAllChannels(page int, perPage int, etag string) (*ChannelListWithTeamData, *Response) {
	query := fmt.Sprintf("?page=%v&per_page=%v", page, perPage)
	r, err := c.DoApiGet(c.GetChannelsRoute()+query, etag)
	if err != nil {
		return nil, BuildErrorResponse(r, err)
	}
	defer closeBody(r)
	return ChannelListWithTeamDataFromJson(r.Body), BuildResponse(r)
}

// CreateChannel creates a channel based on the provided channel struct.
func (c *Client4) CreateChannel(channel *Channel) (*Channel, *Response) {
	r, err := c.DoApiPost(c.GetChannelsRoute(), channel.ToJson())
	if err != nil {
		return nil, BuildErrorResponse(r, err)
	}
	defer closeBody(r)
	return ChannelFromJson(r.Body), BuildResponse(r)
}

// UpdateChannel updates a channel based on the provided channel struct.
func (c *Client4) UpdateChannel(channel *Channel) (*Channel, *Response) {
	r, err := c.DoApiPut(c.GetChannelRoute(channel.Id), channel.ToJson())
	if err != nil {
		return nil, BuildErrorResponse(r, err)
	}
	defer closeBody(r)
	return ChannelFromJson(r.Body), BuildResponse(r)
}

// PatchChannel partially updates a channel. Any missing fields are not updated.
func (c *Client4) PatchChannel(channelId string, patch *ChannelPatch) (*Channel, *Response) {
	r, err := c.DoApiPut(c.GetChannelRoute(channelId)+"/patch", patch.ToJson())
	if err != nil {
		return nil, BuildErrorResponse(r, err)
	}
	defer closeBody(r)
	return ChannelFromJson(r.Body), BuildResponse(r)
}

// ConvertChannelToPrivate converts public to private channel.
func (c *Client4) ConvertChannelToPrivate(channelId string) (*Channel, *Response) {
	r, err := c.DoApiPost(c.GetChannelRoute(channelId)+"/convert", "")
	if err != nil {
		return nil, BuildErrorResponse(r, err)
	}
	defer closeBody(r)
	return ChannelFromJson(r.Body), BuildResponse(r)
}

// RestoreChannel restores a previously deleted channel. Any missing fields are not updated.
func (c *Client4) RestoreChannel(channelId string) (*Channel, *Response) {
	r, err := c.DoApiPost(c.GetChannelRoute(channelId)+"/restore", "")
	if err != nil {
		return nil, BuildErrorResponse(r, err)
	}
	defer closeBody(r)
	return ChannelFromJson(r.Body), BuildResponse(r)
}

// CreateDirectChannel creates a direct message channel based on the two user
// ids provided.
func (c *Client4) CreateDirectChannel(userId1, userId2 string) (*Channel, *Response) {
	requestBody := []string{userId1, userId2}
	r, err := c.DoApiPost(c.GetChannelsRoute()+"/direct", ArrayToJson(requestBody))
	if err != nil {
		return nil, BuildErrorResponse(r, err)
	}
	defer closeBody(r)
	return ChannelFromJson(r.Body), BuildResponse(r)
}

// CreateGroupChannel creates a group message channel based on userIds provided.
func (c *Client4) CreateGroupChannel(userIds []string) (*Channel, *Response) {
	r, err := c.DoApiPost(c.GetChannelsRoute()+"/group", ArrayToJson(userIds))
	if err != nil {
		return nil, BuildErrorResponse(r, err)
	}
	defer closeBody(r)
	return ChannelFromJson(r.Body), BuildResponse(r)
}

// GetChannel returns a channel based on the provided channel id string.
func (c *Client4) GetChannel(channelId, etag string) (*Channel, *Response) {
	r, err := c.DoApiGet(c.GetChannelRoute(channelId), etag)
	if err != nil {
		return nil, BuildErrorResponse(r, err)
	}
	defer closeBody(r)
	return ChannelFromJson(r.Body), BuildResponse(r)
}

// GetChannelStats returns statistics for a channel.
func (c *Client4) GetChannelStats(channelId string, etag string) (*ChannelStats, *Response) {
	r, err := c.DoApiGet(c.GetChannelRoute(channelId)+"/stats", etag)
	if err != nil {
		return nil, BuildErrorResponse(r, err)
	}
	defer closeBody(r)
	return ChannelStatsFromJson(r.Body), BuildResponse(r)
}

// GetChannelMembersTimezones gets a list of timezones for a channel.
func (c *Client4) GetChannelMembersTimezones(channelId string) ([]string, *Response) {
	r, err := c.DoApiGet(c.GetChannelRoute(channelId)+"/timezones", "")
	if err != nil {
		return nil, BuildErrorResponse(r, err)
	}
	defer closeBody(r)
	return ArrayFromJson(r.Body), BuildResponse(r)
}

// GetPinnedPosts gets a list of pinned posts.
func (c *Client4) GetPinnedPosts(channelId string, etag string) (*PostList, *Response) {
	r, err := c.DoApiGet(c.GetChannelRoute(channelId)+"/pinned", etag)
	if err != nil {
		return nil, BuildErrorResponse(r, err)
	}
	defer closeBody(r)
	return PostListFromJson(r.Body), BuildResponse(r)
}

// GetPublicChannelsForTeam returns a list of public channels based on the provided team id string.
func (c *Client4) GetPublicChannelsForTeam(teamId string, page int, perPage int, etag string) ([]*Channel, *Response) {
	query := fmt.Sprintf("?page=%v&per_page=%v", page, perPage)
	r, err := c.DoApiGet(c.GetChannelsForTeamRoute(teamId)+query, etag)
	if err != nil {
		return nil, BuildErrorResponse(r, err)
	}
	defer closeBody(r)
	return ChannelSliceFromJson(r.Body), BuildResponse(r)
}

// GetDeletedChannelsForTeam returns a list of public channels based on the provided team id string.
func (c *Client4) GetDeletedChannelsForTeam(teamId string, page int, perPage int, etag string) ([]*Channel, *Response) {
	query := fmt.Sprintf("/deleted?page=%v&per_page=%v", page, perPage)
	r, err := c.DoApiGet(c.GetChannelsForTeamRoute(teamId)+query, etag)
	if err != nil {
		return nil, BuildErrorResponse(r, err)
	}
	defer closeBody(r)
	return ChannelSliceFromJson(r.Body), BuildResponse(r)
}

// GetPublicChannelsByIdsForTeam returns a list of public channels based on provided team id string.
func (c *Client4) GetPublicChannelsByIdsForTeam(teamId string, channelIds []string) ([]*Channel, *Response) {
	r, err := c.DoApiPost(c.GetChannelsForTeamRoute(teamId)+"/ids", ArrayToJson(channelIds))
	if err != nil {
		return nil, BuildErrorResponse(r, err)
	}
	defer closeBody(r)
	return ChannelSliceFromJson(r.Body), BuildResponse(r)
}

// GetChannelsForTeamForUser returns a list channels of on a team for a user.
func (c *Client4) GetChannelsForTeamForUser(teamId, userId, etag string) ([]*Channel, *Response) {
	r, err := c.DoApiGet(c.GetUserRoute(userId)+c.GetTeamRoute(teamId)+"/channels", etag)
	if err != nil {
		return nil, BuildErrorResponse(r, err)
	}
	defer closeBody(r)
	return ChannelSliceFromJson(r.Body), BuildResponse(r)
}

// SearchChannels returns the channels on a team matching the provided search term.
func (c *Client4) SearchChannels(teamId string, search *ChannelSearch) ([]*Channel, *Response) {
	r, err := c.DoApiPost(c.GetChannelsForTeamRoute(teamId)+"/search", search.ToJson())
	if err != nil {
		return nil, BuildErrorResponse(r, err)
	}
	defer closeBody(r)
	return ChannelSliceFromJson(r.Body), BuildResponse(r)
}

// SearchAllChannels search in all the channels. Must be a system administrator.
func (c *Client4) SearchAllChannels(search *ChannelSearch) (*ChannelListWithTeamData, *Response) {
	r, err := c.DoApiPost(c.GetChannelsRoute()+"/search", search.ToJson())
	if err != nil {
		return nil, BuildErrorResponse(r, err)
	}
	defer closeBody(r)
	return ChannelListWithTeamDataFromJson(r.Body), BuildResponse(r)
}

// DeleteChannel deletes channel based on the provided channel id string.
func (c *Client4) DeleteChannel(channelId string) (bool, *Response) {
	r, err := c.DoApiDelete(c.GetChannelRoute(channelId))
	if err != nil {
		return false, BuildErrorResponse(r, err)
	}
	defer closeBody(r)
	return CheckStatusOK(r), BuildResponse(r)
}

// GetChannelByName returns a channel based on the provided channel name and team id strings.
func (c *Client4) GetChannelByName(channelName, teamId string, etag string) (*Channel, *Response) {
	r, err := c.DoApiGet(c.GetChannelByNameRoute(channelName, teamId), etag)
	if err != nil {
		return nil, BuildErrorResponse(r, err)
	}
	defer closeBody(r)
	return ChannelFromJson(r.Body), BuildResponse(r)
}

// GetChannelByNameIncludeDeleted returns a channel based on the provided channel name and team id strings. Other then GetChannelByName it will also return deleted channels.
func (c *Client4) GetChannelByNameIncludeDeleted(channelName, teamId string, etag string) (*Channel, *Response) {
	r, err := c.DoApiGet(c.GetChannelByNameRoute(channelName, teamId)+"?include_deleted=true", etag)
	if err != nil {
		return nil, BuildErrorResponse(r, err)
	}
	defer closeBody(r)
	return ChannelFromJson(r.Body), BuildResponse(r)
}

// GetChannelByNameForTeamName returns a channel based on the provided channel name and team name strings.
func (c *Client4) GetChannelByNameForTeamName(channelName, teamName string, etag string) (*Channel, *Response) {
	r, err := c.DoApiGet(c.GetChannelByNameForTeamNameRoute(channelName, teamName), etag)
	if err != nil {
		return nil, BuildErrorResponse(r, err)
	}
	defer closeBody(r)
	return ChannelFromJson(r.Body), BuildResponse(r)
}

// GetChannelByNameForTeamNameIncludeDeleted returns a channel based on the provided channel name and team name strings. Other then GetChannelByNameForTeamName it will also return deleted channels.
func (c *Client4) GetChannelByNameForTeamNameIncludeDeleted(channelName, teamName string, etag string) (*Channel, *Response) {
	r, err := c.DoApiGet(c.GetChannelByNameForTeamNameRoute(channelName, teamName)+"?include_deleted=true", etag)
	if err != nil {
		return nil, BuildErrorResponse(r, err)
	}
	defer closeBody(r)
	return ChannelFromJson(r.Body), BuildResponse(r)
}

// GetChannelMembers gets a page of channel members.
func (c *Client4) GetChannelMembers(channelId string, page, perPage int, etag string) (*ChannelMembers, *Response) {
	query := fmt.Sprintf("?page=%v&per_page=%v", page, perPage)
	r, err := c.DoApiGet(c.GetChannelMembersRoute(channelId)+query, etag)
	if err != nil {
		return nil, BuildErrorResponse(r, err)
	}
	defer closeBody(r)
	return ChannelMembersFromJson(r.Body), BuildResponse(r)
}

// GetChannelMembersByIds gets the channel members in a channel for a list of user ids.
func (c *Client4) GetChannelMembersByIds(channelId string, userIds []string) (*ChannelMembers, *Response) {
	r, err := c.DoApiPost(c.GetChannelMembersRoute(channelId)+"/ids", ArrayToJson(userIds))
	if err != nil {
		return nil, BuildErrorResponse(r, err)
	}
	defer closeBody(r)
	return ChannelMembersFromJson(r.Body), BuildResponse(r)
}

// GetChannelMember gets a channel member.
func (c *Client4) GetChannelMember(channelId, userId, etag string) (*ChannelMember, *Response) {
	r, err := c.DoApiGet(c.GetChannelMemberRoute(channelId, userId), etag)
	if err != nil {
		return nil, BuildErrorResponse(r, err)
	}
	defer closeBody(r)
	return ChannelMemberFromJson(r.Body), BuildResponse(r)
}

// GetChannelMembersForUser gets all the channel members for a user on a team.
func (c *Client4) GetChannelMembersForUser(userId, teamId, etag string) (*ChannelMembers, *Response) {
	r, err := c.DoApiGet(fmt.Sprintf(c.GetUserRoute(userId)+"/teams/%v/channels/members", teamId), etag)
	if err != nil {
		return nil, BuildErrorResponse(r, err)
	}
	defer closeBody(r)
	return ChannelMembersFromJson(r.Body), BuildResponse(r)
}

// ViewChannel performs a view action for a user. Synonymous with switching channels or marking channels as read by a user.
func (c *Client4) ViewChannel(userId string, view *ChannelView) (*ChannelViewResponse, *Response) {
	url := fmt.Sprintf(c.GetChannelsRoute()+"/members/%v/view", userId)
	r, err := c.DoApiPost(url, view.ToJson())
	if err != nil {
		return nil, BuildErrorResponse(r, err)
	}
	defer closeBody(r)
	return ChannelViewResponseFromJson(r.Body), BuildResponse(r)
}

// GetChannelUnread will return a ChannelUnread object that contains the number of
// unread messages and mentions for a user.
func (c *Client4) GetChannelUnread(channelId, userId string) (*ChannelUnread, *Response) {
	r, err := c.DoApiGet(c.GetUserRoute(userId)+c.GetChannelRoute(channelId)+"/unread", "")
	if err != nil {
		return nil, BuildErrorResponse(r, err)
	}
	defer closeBody(r)
	return ChannelUnreadFromJson(r.Body), BuildResponse(r)
}

// UpdateChannelRoles will update the roles on a channel for a user.
func (c *Client4) UpdateChannelRoles(channelId, userId, roles string) (bool, *Response) {
	requestBody := map[string]string{"roles": roles}
	r, err := c.DoApiPut(c.GetChannelMemberRoute(channelId, userId)+"/roles", MapToJson(requestBody))
	if err != nil {
		return false, BuildErrorResponse(r, err)
	}
	defer closeBody(r)
	return CheckStatusOK(r), BuildResponse(r)
}

// UpdateChannelMemberSchemeRoles will update the scheme-derived roles on a channel for a user.
func (c *Client4) UpdateChannelMemberSchemeRoles(channelId string, userId string, schemeRoles *SchemeRoles) (bool, *Response) {
	r, err := c.DoApiPut(c.GetChannelMemberRoute(channelId, userId)+"/schemeRoles", schemeRoles.ToJson())
	if err != nil {
		return false, BuildErrorResponse(r, err)
	}
	defer closeBody(r)
	return CheckStatusOK(r), BuildResponse(r)
}

// UpdateChannelNotifyProps will update the notification properties on a channel for a user.
func (c *Client4) UpdateChannelNotifyProps(channelId, userId string, props map[string]string) (bool, *Response) {
	r, err := c.DoApiPut(c.GetChannelMemberRoute(channelId, userId)+"/notify_props", MapToJson(props))
	if err != nil {
		return false, BuildErrorResponse(r, err)
	}
	defer closeBody(r)
	return CheckStatusOK(r), BuildResponse(r)
}

// AddChannelMember adds user to channel and return a channel member.
func (c *Client4) AddChannelMember(channelId, userId string) (*ChannelMember, *Response) {
	requestBody := map[string]string{"user_id": userId}
	r, err := c.DoApiPost(c.GetChannelMembersRoute(channelId)+"", MapToJson(requestBody))
	if err != nil {
		return nil, BuildErrorResponse(r, err)
	}
	defer closeBody(r)
	return ChannelMemberFromJson(r.Body), BuildResponse(r)
}

// AddChannelMemberWithRootId adds user to channel and return a channel member. Post add to channel message has the postRootId.
func (c *Client4) AddChannelMemberWithRootId(channelId, userId, postRootId string) (*ChannelMember, *Response) {
	requestBody := map[string]string{"user_id": userId, "post_root_id": postRootId}
	r, err := c.DoApiPost(c.GetChannelMembersRoute(channelId)+"", MapToJson(requestBody))
	if err != nil {
		return nil, BuildErrorResponse(r, err)
	}
	defer closeBody(r)
	return ChannelMemberFromJson(r.Body), BuildResponse(r)
}

// RemoveUserFromChannel will delete the channel member object for a user, effectively removing the user from a channel.
func (c *Client4) RemoveUserFromChannel(channelId, userId string) (bool, *Response) {
	r, err := c.DoApiDelete(c.GetChannelMemberRoute(channelId, userId))
	if err != nil {
		return false, BuildErrorResponse(r, err)
	}
	defer closeBody(r)
	return CheckStatusOK(r), BuildResponse(r)
}

// AutocompleteChannelsForTeam will return an ordered list of channels autocomplete suggestions.
func (c *Client4) AutocompleteChannelsForTeam(teamId, name string) (*ChannelList, *Response) {
	query := fmt.Sprintf("?name=%v", name)
	r, err := c.DoApiGet(c.GetChannelsForTeamRoute(teamId)+"/autocomplete"+query, "")
	if err != nil {
		return nil, BuildErrorResponse(r, err)
	}
	defer closeBody(r)
	return ChannelListFromJson(r.Body), BuildResponse(r)
}

// AutocompleteChannelsForTeamForSearch will return an ordered list of your channels autocomplete suggestions.
func (c *Client4) AutocompleteChannelsForTeamForSearch(teamId, name string) (*ChannelList, *Response) {
	query := fmt.Sprintf("?name=%v", name)
	r, err := c.DoApiGet(c.GetChannelsForTeamRoute(teamId)+"/search_autocomplete"+query, "")
	if err != nil {
		return nil, BuildErrorResponse(r, err)
	}
	defer closeBody(r)
	return ChannelListFromJson(r.Body), BuildResponse(r)
}

// Post Section

// CreatePost creates a post based on the provided post struct.
func (c *Client4) CreatePost(post *Post) (*Post, *Response) {
	r, err := c.DoApiPost(c.GetPostsRoute(), post.ToUnsanitizedJson())
	if err != nil {
		return nil, BuildErrorResponse(r, err)
	}
	defer closeBody(r)
	return PostFromJson(r.Body), BuildResponse(r)
}

// CreatePostEphemeral creates a ephemeral post based on the provided post struct which is send to the given user id.
func (c *Client4) CreatePostEphemeral(post *PostEphemeral) (*Post, *Response) {
	r, err := c.DoApiPost(c.GetPostsEphemeralRoute(), post.ToUnsanitizedJson())
	if err != nil {
		return nil, BuildErrorResponse(r, err)
	}
	defer closeBody(r)
	return PostFromJson(r.Body), BuildResponse(r)
}

// UpdatePost updates a post based on the provided post struct.
func (c *Client4) UpdatePost(postId string, post *Post) (*Post, *Response) {
	r, err := c.DoApiPut(c.GetPostRoute(postId), post.ToUnsanitizedJson())
	if err != nil {
		return nil, BuildErrorResponse(r, err)
	}
	defer closeBody(r)
	return PostFromJson(r.Body), BuildResponse(r)
}

// PatchPost partially updates a post. Any missing fields are not updated.
func (c *Client4) PatchPost(postId string, patch *PostPatch) (*Post, *Response) {
	r, err := c.DoApiPut(c.GetPostRoute(postId)+"/patch", patch.ToJson())
	if err != nil {
		return nil, BuildErrorResponse(r, err)
	}
	defer closeBody(r)
	return PostFromJson(r.Body), BuildResponse(r)
}

// PinPost pin a post based on provided post id string.
func (c *Client4) PinPost(postId string) (bool, *Response) {
	r, err := c.DoApiPost(c.GetPostRoute(postId)+"/pin", "")
	if err != nil {
		return false, BuildErrorResponse(r, err)
	}
	defer closeBody(r)
	return CheckStatusOK(r), BuildResponse(r)
}

// UnpinPost unpin a post based on provided post id string.
func (c *Client4) UnpinPost(postId string) (bool, *Response) {
	r, err := c.DoApiPost(c.GetPostRoute(postId)+"/unpin", "")
	if err != nil {
		return false, BuildErrorResponse(r, err)
	}
	defer closeBody(r)
	return CheckStatusOK(r), BuildResponse(r)
}

// GetPost gets a single post.
func (c *Client4) GetPost(postId string, etag string) (*Post, *Response) {
	r, err := c.DoApiGet(c.GetPostRoute(postId), etag)
	if err != nil {
		return nil, BuildErrorResponse(r, err)
	}
	defer closeBody(r)
	return PostFromJson(r.Body), BuildResponse(r)
}

// DeletePost deletes a post from the provided post id string.
func (c *Client4) DeletePost(postId string) (bool, *Response) {
	r, err := c.DoApiDelete(c.GetPostRoute(postId))
	if err != nil {
		return false, BuildErrorResponse(r, err)
	}
	defer closeBody(r)
	return CheckStatusOK(r), BuildResponse(r)
}

// GetPostThread gets a post with all the other posts in the same thread.
func (c *Client4) GetPostThread(postId string, etag string) (*PostList, *Response) {
	r, err := c.DoApiGet(c.GetPostRoute(postId)+"/thread", etag)
	if err != nil {
		return nil, BuildErrorResponse(r, err)
	}
	defer closeBody(r)
	return PostListFromJson(r.Body), BuildResponse(r)
}

// GetPostsForChannel gets a page of posts with an array for ordering for a channel.
func (c *Client4) GetPostsForChannel(channelId string, page, perPage int, etag string) (*PostList, *Response) {
	query := fmt.Sprintf("?page=%v&per_page=%v", page, perPage)
	r, err := c.DoApiGet(c.GetChannelRoute(channelId)+"/posts"+query, etag)
	if err != nil {
		return nil, BuildErrorResponse(r, err)
	}
	defer closeBody(r)
	return PostListFromJson(r.Body), BuildResponse(r)
}

// GetFlaggedPostsForUser returns flagged posts of a user based on user id string.
func (c *Client4) GetFlaggedPostsForUser(userId string, page int, perPage int) (*PostList, *Response) {
	query := fmt.Sprintf("?page=%v&per_page=%v", page, perPage)
	r, err := c.DoApiGet(c.GetUserRoute(userId)+"/posts/flagged"+query, "")
	if err != nil {
		return nil, BuildErrorResponse(r, err)
	}
	defer closeBody(r)
	return PostListFromJson(r.Body), BuildResponse(r)
}

// GetFlaggedPostsForUserInTeam returns flagged posts in team of a user based on user id string.
func (c *Client4) GetFlaggedPostsForUserInTeam(userId string, teamId string, page int, perPage int) (*PostList, *Response) {
	if len(teamId) == 0 || len(teamId) != 26 {
		return nil, &Response{StatusCode: http.StatusBadRequest, Error: NewAppError("GetFlaggedPostsForUserInTeam", "model.client.get_flagged_posts_in_team.missing_parameter.app_error", nil, "", http.StatusBadRequest)}
	}

	query := fmt.Sprintf("?team_id=%v&page=%v&per_page=%v", teamId, page, perPage)
	r, err := c.DoApiGet(c.GetUserRoute(userId)+"/posts/flagged"+query, "")
	if err != nil {
		return nil, BuildErrorResponse(r, err)
	}
	defer closeBody(r)
	return PostListFromJson(r.Body), BuildResponse(r)
}

// GetFlaggedPostsForUserInChannel returns flagged posts in channel of a user based on user id string.
func (c *Client4) GetFlaggedPostsForUserInChannel(userId string, channelId string, page int, perPage int) (*PostList, *Response) {
	if len(channelId) == 0 || len(channelId) != 26 {
		return nil, &Response{StatusCode: http.StatusBadRequest, Error: NewAppError("GetFlaggedPostsForUserInChannel", "model.client.get_flagged_posts_in_channel.missing_parameter.app_error", nil, "", http.StatusBadRequest)}
	}

	query := fmt.Sprintf("?channel_id=%v&page=%v&per_page=%v", channelId, page, perPage)
	r, err := c.DoApiGet(c.GetUserRoute(userId)+"/posts/flagged"+query, "")
	if err != nil {
		return nil, BuildErrorResponse(r, err)
	}
	defer closeBody(r)
	return PostListFromJson(r.Body), BuildResponse(r)
}

// GetPostsSince gets posts created after a specified time as Unix time in milliseconds.
func (c *Client4) GetPostsSince(channelId string, time int64) (*PostList, *Response) {
	query := fmt.Sprintf("?since=%v", time)
	r, err := c.DoApiGet(c.GetChannelRoute(channelId)+"/posts"+query, "")
	if err != nil {
		return nil, BuildErrorResponse(r, err)
	}
	defer closeBody(r)
	return PostListFromJson(r.Body), BuildResponse(r)
}

// GetPostsAfter gets a page of posts that were posted after the post provided.
func (c *Client4) GetPostsAfter(channelId, postId string, page, perPage int, etag string) (*PostList, *Response) {
	query := fmt.Sprintf("?page=%v&per_page=%v&after=%v", page, perPage, postId)
	r, err := c.DoApiGet(c.GetChannelRoute(channelId)+"/posts"+query, etag)
	if err != nil {
		return nil, BuildErrorResponse(r, err)
	}
	defer closeBody(r)
	return PostListFromJson(r.Body), BuildResponse(r)
}

// GetPostsBefore gets a page of posts that were posted before the post provided.
func (c *Client4) GetPostsBefore(channelId, postId string, page, perPage int, etag string) (*PostList, *Response) {
	query := fmt.Sprintf("?page=%v&per_page=%v&before=%v", page, perPage, postId)
	r, err := c.DoApiGet(c.GetChannelRoute(channelId)+"/posts"+query, etag)
	if err != nil {
		return nil, BuildErrorResponse(r, err)
	}
	defer closeBody(r)
	return PostListFromJson(r.Body), BuildResponse(r)
}

// SearchPosts returns any posts with matching terms string.
func (c *Client4) SearchPosts(teamId string, terms string, isOrSearch bool) (*PostList, *Response) {
	params := SearchParameter{
		Terms:      &terms,
		IsOrSearch: &isOrSearch,
	}
	return c.SearchPostsWithParams(teamId, &params)
}

// SearchPostsWithParams returns any posts with matching terms string.
func (c *Client4) SearchPostsWithParams(teamId string, params *SearchParameter) (*PostList, *Response) {
	r, err := c.DoApiPost(c.GetTeamRoute(teamId)+"/posts/search", params.SearchParameterToJson())
	if err != nil {
		return nil, BuildErrorResponse(r, err)
	}
	defer closeBody(r)
	return PostListFromJson(r.Body), BuildResponse(r)
}

// SearchPostsWithMatches returns any posts with matching terms string, including.
func (c *Client4) SearchPostsWithMatches(teamId string, terms string, isOrSearch bool) (*PostSearchResults, *Response) {
	requestBody := map[string]interface{}{"terms": terms, "is_or_search": isOrSearch}
	r, err := c.DoApiPost(c.GetTeamRoute(teamId)+"/posts/search", StringInterfaceToJson(requestBody))
	if err != nil {
		return nil, BuildErrorResponse(r, err)
	}
	defer closeBody(r)
	return PostSearchResultsFromJson(r.Body), BuildResponse(r)
}

// DoPostAction performs a post action.
func (c *Client4) DoPostAction(postId, actionId string) (bool, *Response) {
	r, err := c.DoApiPost(c.GetPostRoute(postId)+"/actions/"+actionId, "")
	if err != nil {
		return false, BuildErrorResponse(r, err)
	}
	defer closeBody(r)
	return CheckStatusOK(r), BuildResponse(r)
}


// UploadFile will upload a file to a channel using a multipart request, to be later attached to a post.
// This method is functionally equivalent to Client4.UploadFileAsRequestBody.
func (c *Client4) UploadFile(data []byte, channelId string, filename string) (*FileUploadResponse, *Response) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	part, err := writer.CreateFormField("channel_id")
	if err != nil {
		return nil, &Response{Error: NewAppError("UploadPostAttachment", "model.client.upload_post_attachment.channel_id.app_error", nil, err.Error(), http.StatusBadRequest)}
	}

	_, err = io.Copy(part, strings.NewReader(channelId))
	if err != nil {
		return nil, &Response{Error: NewAppError("UploadPostAttachment", "model.client.upload_post_attachment.channel_id.app_error", nil, err.Error(), http.StatusBadRequest)}
	}

	part, err = writer.CreateFormFile("files", filename)
	if err != nil {
		return nil, &Response{Error: NewAppError("UploadPostAttachment", "model.client.upload_post_attachment.file.app_error", nil, err.Error(), http.StatusBadRequest)}
	}
	_, err = io.Copy(part, bytes.NewBuffer(data))
	if err != nil {
		return nil, &Response{Error: NewAppError("UploadPostAttachment", "model.client.upload_post_attachment.file.app_error", nil, err.Error(), http.StatusBadRequest)}
	}

	err = writer.Close()
	if err != nil {
		return nil, &Response{Error: NewAppError("UploadPostAttachment", "model.client.upload_post_attachment.writer.app_error", nil, err.Error(), http.StatusBadRequest)}
	}

	return c.DoUploadFile(c.GetFilesRoute(), body.Bytes(), writer.FormDataContentType())
}

// UploadFileAsRequestBody will upload a file to a channel as the body of a request, to be later attached
// to a post. This method is functionally equivalent to Client4.UploadFile.
func (c *Client4) UploadFileAsRequestBody(data []byte, channelId string, filename string) (*FileUploadResponse, *Response) {
	return c.DoUploadFile(c.GetFilesRoute()+fmt.Sprintf("?channel_id=%v&filename=%v", url.QueryEscape(channelId), url.QueryEscape(filename)), data, http.DetectContentType(data))
}

// GetFile gets the bytes for a file by id.
func (c *Client4) GetFile(fileId string) ([]byte, *Response) {
	r, appErr := c.DoApiGet(c.GetFileRoute(fileId), "")
	if appErr != nil {
		return nil, BuildErrorResponse(r, appErr)
	}
	defer closeBody(r)

	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return nil, BuildErrorResponse(r, NewAppError("GetFile", "model.client.read_file.app_error", nil, err.Error(), r.StatusCode))
	}
	return data, BuildResponse(r)
}

// DownloadFile gets the bytes for a file by id, optionally adding headers to force the browser to download it.
func (c *Client4) DownloadFile(fileId string, download bool) ([]byte, *Response) {
	r, appErr := c.DoApiGet(c.GetFileRoute(fileId)+fmt.Sprintf("?download=%v", download), "")
	if appErr != nil {
		return nil, BuildErrorResponse(r, appErr)
	}
	defer closeBody(r)

	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return nil, BuildErrorResponse(r, NewAppError("DownloadFile", "model.client.read_file.app_error", nil, err.Error(), r.StatusCode))
	}
	return data, BuildResponse(r)
}

// GetFileThumbnail gets the bytes for a file by id.
func (c *Client4) GetFileThumbnail(fileId string) ([]byte, *Response) {
	r, appErr := c.DoApiGet(c.GetFileRoute(fileId)+"/thumbnail", "")
	if appErr != nil {
		return nil, BuildErrorResponse(r, appErr)
	}
	defer closeBody(r)

	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return nil, BuildErrorResponse(r, NewAppError("GetFileThumbnail", "model.client.read_file.app_error", nil, err.Error(), r.StatusCode))
	}
	return data, BuildResponse(r)
}

// DownloadFileThumbnail gets the bytes for a file by id, optionally adding headers to force the browser to download it.
func (c *Client4) DownloadFileThumbnail(fileId string, download bool) ([]byte, *Response) {
	r, appErr := c.DoApiGet(c.GetFileRoute(fileId)+fmt.Sprintf("/thumbnail?download=%v", download), "")
	if appErr != nil {
		return nil, BuildErrorResponse(r, appErr)
	}
	defer closeBody(r)

	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return nil, BuildErrorResponse(r, NewAppError("DownloadFileThumbnail", "model.client.read_file.app_error", nil, err.Error(), r.StatusCode))
	}
	return data, BuildResponse(r)
}

// GetFileLink gets the public link of a file by id.
func (c *Client4) GetFileLink(fileId string) (string, *Response) {
	r, err := c.DoApiGet(c.GetFileRoute(fileId)+"/link", "")
	if err != nil {
		return "", BuildErrorResponse(r, err)
	}
	defer closeBody(r)
	return MapFromJson(r.Body)["link"], BuildResponse(r)
}

// GetFilePreview gets the bytes for a file by id.
func (c *Client4) GetFilePreview(fileId string) ([]byte, *Response) {
	r, appErr := c.DoApiGet(c.GetFileRoute(fileId)+"/preview", "")
	if appErr != nil {
		return nil, BuildErrorResponse(r, appErr)
	}
	defer closeBody(r)

	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return nil, BuildErrorResponse(r, NewAppError("GetFilePreview", "model.client.read_file.app_error", nil, err.Error(), r.StatusCode))
	}
	return data, BuildResponse(r)
}

// DownloadFilePreview gets the bytes for a file by id.
func (c *Client4) DownloadFilePreview(fileId string, download bool) ([]byte, *Response) {
	r, appErr := c.DoApiGet(c.GetFileRoute(fileId)+fmt.Sprintf("/preview?download=%v", download), "")
	if appErr != nil {
		return nil, BuildErrorResponse(r, appErr)
	}
	defer closeBody(r)

	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return nil, BuildErrorResponse(r, NewAppError("DownloadFilePreview", "model.client.read_file.app_error", nil, err.Error(), r.StatusCode))
	}
	return data, BuildResponse(r)
}

// GetFileInfo gets all the file info objects.
func (c *Client4) GetFileInfo(fileId string) (*FileInfo, *Response) {
	r, err := c.DoApiGet(c.GetFileRoute(fileId)+"/info", "")
	if err != nil {
		return nil, BuildErrorResponse(r, err)
	}
	defer closeBody(r)
	return FileInfoFromJson(r.Body), BuildResponse(r)
}

// GetFileInfosForPost gets all the file info objects attached to a post.
func (c *Client4) GetFileInfosForPost(postId string, etag string) ([]*FileInfo, *Response) {
	r, err := c.DoApiGet(c.GetPostRoute(postId)+"/files/info", etag)
	if err != nil {
		return nil, BuildErrorResponse(r, err)
	}
	defer closeBody(r)
	return FileInfosFromJson(r.Body), BuildResponse(r)
}

// General/System Section

// GetPing will return ok if the running goRoutines are below the threshold and unhealthy for above.
func (c *Client4) GetPing() (string, *Response) {
	r, err := c.DoApiGet(c.GetSystemRoute()+"/ping", "")
	if r != nil && r.StatusCode == 500 {
		defer r.Body.Close()
		return "unhealthy", BuildErrorResponse(r, err)
	}
	if err != nil {
		return "", BuildErrorResponse(r, err)
	}
	defer closeBody(r)
	return MapFromJson(r.Body)["status"], BuildResponse(r)
}

// TestEmail will attempt to connect to the configured SMTP server.
func (c *Client4) TestEmail(config *Config) (bool, *Response) {
	r, err := c.DoApiPost(c.GetTestEmailRoute(), config.ToJson())
	if err != nil {
		return false, BuildErrorResponse(r, err)
	}
	defer closeBody(r)
	return CheckStatusOK(r), BuildResponse(r)
}

// TestS3Connection will attempt to connect to the AWS S3.
func (c *Client4) TestS3Connection(config *Config) (bool, *Response) {
	r, err := c.DoApiPost(c.GetTestS3Route(), config.ToJson())
	if err != nil {
		return false, BuildErrorResponse(r, err)
	}
	defer closeBody(r)
	return CheckStatusOK(r), BuildResponse(r)
}

// GetConfig will retrieve the server config with some sanitized items.
func (c *Client4) GetConfig() (*Config, *Response) {
	r, err := c.DoApiGet(c.GetConfigRoute(), "")
	if err != nil {
		return nil, BuildErrorResponse(r, err)
	}
	defer closeBody(r)
	return ConfigFromJson(r.Body), BuildResponse(r)
}

// ReloadConfig will reload the server configuration.
func (c *Client4) ReloadConfig() (bool, *Response) {
	r, err := c.DoApiPost(c.GetConfigRoute()+"/reload", "")
	if err != nil {
		return false, BuildErrorResponse(r, err)
	}
	defer closeBody(r)
	return CheckStatusOK(r), BuildResponse(r)
}

// GetOldClientConfig will retrieve the parts of the server configuration needed by the
// client, formatted in the old format.
func (c *Client4) GetOldClientConfig(etag string) (map[string]string, *Response) {
	r, err := c.DoApiGet(c.GetConfigRoute()+"/client?format=old", etag)
	if err != nil {
		return nil, BuildErrorResponse(r, err)
	}
	defer closeBody(r)
	return MapFromJson(r.Body), BuildResponse(r)
}

// GetEnvironmentConfig will retrieve a map mirroring the server configuration where fields
// are set to true if the corresponding config setting is set through an environment variable.
// Settings that haven't been set through environment variables will be missing from the map.
func (c *Client4) GetEnvironmentConfig() (map[string]interface{}, *Response) {
	r, err := c.DoApiGet(c.GetConfigRoute()+"/environment", "")
	if err != nil {
		return nil, BuildErrorResponse(r, err)
	}
	defer closeBody(r)
	return StringInterfaceFromJson(r.Body), BuildResponse(r)
}

// GetOldClientLicense will retrieve the parts of the server license needed by the
// client, formatted in the old format.
func (c *Client4) GetOldClientLicense(etag string) (map[string]string, *Response) {
	r, err := c.DoApiGet(c.GetLicenseRoute()+"/client?format=old", etag)
	if err != nil {
		return nil, BuildErrorResponse(r, err)
	}
	defer closeBody(r)
	return MapFromJson(r.Body), BuildResponse(r)
}

// DatabaseRecycle will recycle the connections. Discard current connection and get new one.
func (c *Client4) DatabaseRecycle() (bool, *Response) {
	r, err := c.DoApiPost(c.GetDatabaseRoute()+"/recycle", "")
	if err != nil {
		return false, BuildErrorResponse(r, err)
	}
	defer closeBody(r)
	return CheckStatusOK(r), BuildResponse(r)
}

// InvalidateCaches will purge the cache and can affect the performance while is cleaning.
func (c *Client4) InvalidateCaches() (bool, *Response) {
	r, err := c.DoApiPost(c.GetCacheRoute()+"/invalidate", "")
	if err != nil {
		return false, BuildErrorResponse(r, err)
	}
	defer closeBody(r)
	return CheckStatusOK(r), BuildResponse(r)
}

// UpdateConfig will update the server configuration.
func (c *Client4) UpdateConfig(config *Config) (*Config, *Response) {
	r, err := c.DoApiPut(c.GetConfigRoute(), config.ToJson())
	if err != nil {
		return nil, BuildErrorResponse(r, err)
	}
	defer closeBody(r)
	return ConfigFromJson(r.Body), BuildResponse(r)
}

// UploadLicenseFile will add a license file to the system.
func (c *Client4) UploadLicenseFile(data []byte) (bool, *Response) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	part, err := writer.CreateFormFile("license", "test-license.mattermost-license")
	if err != nil {
		return false, &Response{Error: NewAppError("UploadLicenseFile", "model.client.set_profile_user.no_file.app_error", nil, err.Error(), http.StatusBadRequest)}
	}

	if _, err = io.Copy(part, bytes.NewBuffer(data)); err != nil {
		return false, &Response{Error: NewAppError("UploadLicenseFile", "model.client.set_profile_user.no_file.app_error", nil, err.Error(), http.StatusBadRequest)}
	}

	if err = writer.Close(); err != nil {
		return false, &Response{Error: NewAppError("UploadLicenseFile", "model.client.set_profile_user.writer.app_error", nil, err.Error(), http.StatusBadRequest)}
	}

	rq, err := http.NewRequest("POST", c.ApiUrl+c.GetLicenseRoute(), bytes.NewReader(body.Bytes()))
	if err != nil {
		return false, &Response{Error: NewAppError("UploadLicenseFile", "model.client.connecting.app_error", nil, err.Error(), http.StatusBadRequest)}
	}
	rq.Header.Set("Content-Type", writer.FormDataContentType())

	if len(c.AuthToken) > 0 {
		rq.Header.Set(HEADER_AUTH, c.AuthType+" "+c.AuthToken)
	}

	rp, err := c.HttpClient.Do(rq)
	if err != nil || rp == nil {
		return false, &Response{StatusCode: http.StatusForbidden, Error: NewAppError(c.GetLicenseRoute(), "model.client.connecting.app_error", nil, err.Error(), http.StatusForbidden)}
	}
	defer closeBody(rp)

	if rp.StatusCode >= 300 {
		return false, BuildErrorResponse(rp, AppErrorFromJson(rp.Body))
	}

	return CheckStatusOK(rp), BuildResponse(rp)
}

// RemoveLicenseFile will remove the server license it exists. Note that this will
// disable all enterprise features.
func (c *Client4) RemoveLicenseFile() (bool, *Response) {
	r, err := c.DoApiDelete(c.GetLicenseRoute())
	if err != nil {
		return false, BuildErrorResponse(r, err)
	}
	defer closeBody(r)
	return CheckStatusOK(r), BuildResponse(r)
}


// Preferences Section

// GetPreferences returns the user's preferences.
func (c *Client4) GetPreferences(userId string) (Preferences, *Response) {
	r, err := c.DoApiGet(c.GetPreferencesRoute(userId), "")
	if err != nil {
		return nil, BuildErrorResponse(r, err)
	}
	defer closeBody(r)
	preferences, _ := PreferencesFromJson(r.Body)
	return preferences, BuildResponse(r)
}

// UpdatePreferences saves the user's preferences.
func (c *Client4) UpdatePreferences(userId string, preferences *Preferences) (bool, *Response) {
	r, err := c.DoApiPut(c.GetPreferencesRoute(userId), preferences.ToJson())
	if err != nil {
		return false, BuildErrorResponse(r, err)
	}
	defer closeBody(r)
	return true, BuildResponse(r)
}

// DeletePreferences deletes the user's preferences.
func (c *Client4) DeletePreferences(userId string, preferences *Preferences) (bool, *Response) {
	r, err := c.DoApiPost(c.GetPreferencesRoute(userId)+"/delete", preferences.ToJson())
	if err != nil {
		return false, BuildErrorResponse(r, err)
	}
	defer closeBody(r)
	return true, BuildResponse(r)
}

// GetPreferencesByCategory returns the user's preferences from the provided category string.
func (c *Client4) GetPreferencesByCategory(userId string, category string) (Preferences, *Response) {
	url := fmt.Sprintf(c.GetPreferencesRoute(userId)+"/%s", category)
	r, err := c.DoApiGet(url, "")
	if err != nil {
		return nil, BuildErrorResponse(r, err)
	}
	defer closeBody(r)
	preferences, _ := PreferencesFromJson(r.Body)
	return preferences, BuildResponse(r)
}

// GetPreferenceByCategoryAndName returns the user's preferences from the provided category and preference name string.
func (c *Client4) GetPreferenceByCategoryAndName(userId string, category string, preferenceName string) (*Preference, *Response) {
	url := fmt.Sprintf(c.GetPreferencesRoute(userId)+"/%s/name/%v", category, preferenceName)
	r, err := c.DoApiGet(url, "")
	if err != nil {
		return nil, BuildErrorResponse(r, err)
	}
	defer closeBody(r)
	return PreferenceFromJson(r.Body), BuildResponse(r)
}

// Cluster Section

// GetClusterStatus returns the status of all the configured cluster nodes.
func (c *Client4) GetClusterStatus() ([]*ClusterInfo, *Response) {
	r, err := c.DoApiGet(c.GetClusterRoute()+"/status", "")
	if err != nil {
		return nil, BuildErrorResponse(r, err)
	}
	defer closeBody(r)
	return ClusterInfosFromJson(r.Body), BuildResponse(r)
}


// Audits Section

// GetAudits returns a list of audits for the whole system.
func (c *Client4) GetAudits(page int, perPage int, etag string) (Audits, *Response) {
	query := fmt.Sprintf("?page=%v&per_page=%v", page, perPage)
	r, err := c.DoApiGet("/audits"+query, etag)
	if err != nil {
		return nil, BuildErrorResponse(r, err)
	}
	defer closeBody(r)
	return AuditsFromJson(r.Body), BuildResponse(r)
}

// Brand Section

// GetBrandImage retrieves the previously uploaded brand image.
func (c *Client4) GetBrandImage() ([]byte, *Response) {
	r, appErr := c.DoApiGet(c.GetBrandRoute()+"/image", "")
	if appErr != nil {
		return nil, BuildErrorResponse(r, appErr)
	}
	defer closeBody(r)

	if r.StatusCode >= 300 {
		return nil, BuildErrorResponse(r, AppErrorFromJson(r.Body))
	}

	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return nil, BuildErrorResponse(r, NewAppError("GetBrandImage", "model.client.read_file.app_error", nil, err.Error(), r.StatusCode))
	}

	return data, BuildResponse(r)
}

// DeleteBrandImage delets the brand image for the system.
func (c *Client4) DeleteBrandImage() *Response {
	r, err := c.DoApiDelete(c.GetBrandRoute() + "/image")
	if err != nil {
		return BuildErrorResponse(r, err)
	}
	return BuildResponse(r)
}

// UploadBrandImage sets the brand image for the system.
func (c *Client4) UploadBrandImage(data []byte) (bool, *Response) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	part, err := writer.CreateFormFile("image", "brand.png")
	if err != nil {
		return false, &Response{Error: NewAppError("UploadBrandImage", "model.client.set_profile_user.no_file.app_error", nil, err.Error(), http.StatusBadRequest)}
	}

	if _, err = io.Copy(part, bytes.NewBuffer(data)); err != nil {
		return false, &Response{Error: NewAppError("UploadBrandImage", "model.client.set_profile_user.no_file.app_error", nil, err.Error(), http.StatusBadRequest)}
	}

	if err = writer.Close(); err != nil {
		return false, &Response{Error: NewAppError("UploadBrandImage", "model.client.set_profile_user.writer.app_error", nil, err.Error(), http.StatusBadRequest)}
	}

	rq, err := http.NewRequest("POST", c.ApiUrl+c.GetBrandRoute()+"/image", bytes.NewReader(body.Bytes()))
	if err != nil {
		return false, &Response{Error: NewAppError("UploadBrandImage", "model.client.connecting.app_error", nil, err.Error(), http.StatusBadRequest)}
	}
	rq.Header.Set("Content-Type", writer.FormDataContentType())

	if len(c.AuthToken) > 0 {
		rq.Header.Set(HEADER_AUTH, c.AuthType+" "+c.AuthToken)
	}

	rp, err := c.HttpClient.Do(rq)
	if err != nil || rp == nil {
		return false, &Response{StatusCode: http.StatusForbidden, Error: NewAppError(c.GetBrandRoute()+"/image", "model.client.connecting.app_error", nil, err.Error(), http.StatusForbidden)}
	}
	defer closeBody(rp)

	if rp.StatusCode >= 300 {
		return false, BuildErrorResponse(rp, AppErrorFromJson(rp.Body))
	}

	return CheckStatusOK(rp), BuildResponse(rp)
}

// Logs Section

// GetLogs page of logs as a string array.
func (c *Client4) GetLogs(page, perPage int) ([]string, *Response) {
	query := fmt.Sprintf("?page=%v&logs_per_page=%v", page, perPage)
	r, err := c.DoApiGet("/logs"+query, "")
	if err != nil {
		return nil, BuildErrorResponse(r, err)
	}
	defer closeBody(r)
	return ArrayFromJson(r.Body), BuildResponse(r)
}

// PostLog is a convenience Web Service call so clients can log messages into
// the server-side logs. For example we typically log javascript error messages
// into the server-side. It returns the log message if the logging was successful.
func (c *Client4) PostLog(message map[string]string) (map[string]string, *Response) {
	r, err := c.DoApiPost("/logs", MapToJson(message))
	if err != nil {
		return nil, BuildErrorResponse(r, err)
	}
	defer closeBody(r)
	return MapFromJson(r.Body), BuildResponse(r)
}

// OAuth Section

// CreateOAuthApp will register a new OAuth 2.0 client application with Mattermost acting as an OAuth 2.0 service provider.
func (c *Client4) CreateOAuthApp(app *OAuthApp) (*OAuthApp, *Response) {
	r, err := c.DoApiPost(c.GetOAuthAppsRoute(), app.ToJson())
	if err != nil {
		return nil, BuildErrorResponse(r, err)
	}
	defer closeBody(r)
	return OAuthAppFromJson(r.Body), BuildResponse(r)
}

// UpdateOAuthApp updates a page of registered OAuth 2.0 client applications with Mattermost acting as an OAuth 2.0 service provider.
func (c *Client4) UpdateOAuthApp(app *OAuthApp) (*OAuthApp, *Response) {
	r, err := c.DoApiPut(c.GetOAuthAppRoute(app.Id), app.ToJson())
	if err != nil {
		return nil, BuildErrorResponse(r, err)
	}
	defer closeBody(r)
	return OAuthAppFromJson(r.Body), BuildResponse(r)
}

// GetOAuthApps gets a page of registered OAuth 2.0 client applications with Mattermost acting as an OAuth 2.0 service provider.
func (c *Client4) GetOAuthApps(page, perPage int) ([]*OAuthApp, *Response) {
	query := fmt.Sprintf("?page=%v&per_page=%v", page, perPage)
	r, err := c.DoApiGet(c.GetOAuthAppsRoute()+query, "")
	if err != nil {
		return nil, BuildErrorResponse(r, err)
	}
	defer closeBody(r)
	return OAuthAppListFromJson(r.Body), BuildResponse(r)
}

// GetOAuthApp gets a registered OAuth 2.0 client application with Mattermost acting as an OAuth 2.0 service provider.
func (c *Client4) GetOAuthApp(appId string) (*OAuthApp, *Response) {
	r, err := c.DoApiGet(c.GetOAuthAppRoute(appId), "")
	if err != nil {
		return nil, BuildErrorResponse(r, err)
	}
	defer closeBody(r)
	return OAuthAppFromJson(r.Body), BuildResponse(r)
}

// GetOAuthAppInfo gets a sanitized version of a registered OAuth 2.0 client application with Mattermost acting as an OAuth 2.0 service provider.
func (c *Client4) GetOAuthAppInfo(appId string) (*OAuthApp, *Response) {
	r, err := c.DoApiGet(c.GetOAuthAppRoute(appId)+"/info", "")
	if err != nil {
		return nil, BuildErrorResponse(r, err)
	}
	defer closeBody(r)
	return OAuthAppFromJson(r.Body), BuildResponse(r)
}

// DeleteOAuthApp deletes a registered OAuth 2.0 client application.
func (c *Client4) DeleteOAuthApp(appId string) (bool, *Response) {
	r, err := c.DoApiDelete(c.GetOAuthAppRoute(appId))
	if err != nil {
		return false, BuildErrorResponse(r, err)
	}
	defer closeBody(r)
	return CheckStatusOK(r), BuildResponse(r)
}

// RegenerateOAuthAppSecret regenerates the client secret for a registered OAuth 2.0 client application.
func (c *Client4) RegenerateOAuthAppSecret(appId string) (*OAuthApp, *Response) {
	r, err := c.DoApiPost(c.GetOAuthAppRoute(appId)+"/regen_secret", "")
	if err != nil {
		return nil, BuildErrorResponse(r, err)
	}
	defer closeBody(r)
	return OAuthAppFromJson(r.Body), BuildResponse(r)
}

// GetAuthorizedOAuthAppsForUser gets a page of OAuth 2.0 client applications the user has authorized to use access their account.
func (c *Client4) GetAuthorizedOAuthAppsForUser(userId string, page, perPage int) ([]*OAuthApp, *Response) {
	query := fmt.Sprintf("?page=%v&per_page=%v", page, perPage)
	r, err := c.DoApiGet(c.GetUserRoute(userId)+"/oauth/apps/authorized"+query, "")
	if err != nil {
		return nil, BuildErrorResponse(r, err)
	}
	defer closeBody(r)
	return OAuthAppListFromJson(r.Body), BuildResponse(r)
}

// AuthorizeOAuthApp will authorize an OAuth 2.0 client application to access a user's account and provide a redirect link to follow.
func (c *Client4) AuthorizeOAuthApp(authRequest *AuthorizeRequest) (string, *Response) {
	r, err := c.DoApiRequest(http.MethodPost, c.Url+"/oauth/authorize", authRequest.ToJson(), "")
	if err != nil {
		return "", BuildErrorResponse(r, err)
	}
	defer closeBody(r)
	return MapFromJson(r.Body)["redirect"], BuildResponse(r)
}

// DeauthorizeOAuthApp will deauthorize an OAuth 2.0 client application from accessing a user's account.
func (c *Client4) DeauthorizeOAuthApp(appId string) (bool, *Response) {
	requestData := map[string]string{"client_id": appId}
	r, err := c.DoApiRequest(http.MethodPost, c.Url+"/oauth/deauthorize", MapToJson(requestData), "")
	if err != nil {
		return false, BuildErrorResponse(r, err)
	}
	defer closeBody(r)
	return CheckStatusOK(r), BuildResponse(r)
}

// GetOAuthAccessToken is a test helper function for the OAuth access token endpoint.
func (c *Client4) GetOAuthAccessToken(data url.Values) (*AccessResponse, *Response) {
	rq, err := http.NewRequest(http.MethodPost, c.Url+"/oauth/access_token", strings.NewReader(data.Encode()))
	if err != nil {
		return nil, &Response{Error: NewAppError(c.Url+"/oauth/access_token", "model.client.connecting.app_error", nil, err.Error(), http.StatusBadRequest)}
	}
	rq.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	if len(c.AuthToken) > 0 {
		rq.Header.Set(HEADER_AUTH, c.AuthType+" "+c.AuthToken)
	}

	rp, err := c.HttpClient.Do(rq)
	if err != nil || rp == nil {
		return nil, &Response{StatusCode: http.StatusForbidden, Error: NewAppError(c.Url+"/oauth/access_token", "model.client.connecting.app_error", nil, err.Error(), 403)}
	}
	defer closeBody(rp)

	if rp.StatusCode >= 300 {
		return nil, BuildErrorResponse(rp, AppErrorFromJson(rp.Body))
	}

	return AccessResponseFromJson(rp.Body), BuildResponse(rp)
}

// Elasticsearch Section

// TestElasticsearch will attempt to connect to the configured Elasticsearch server and return OK if configured.
// correctly.
func (c *Client4) TestElasticsearch() (bool, *Response) {
	r, err := c.DoApiPost(c.GetElasticsearchRoute()+"/test", "")
	if err != nil {
		return false, BuildErrorResponse(r, err)
	}
	defer closeBody(r)
	return CheckStatusOK(r), BuildResponse(r)
}

// PurgeElasticsearchIndexes immediately deletes all Elasticsearch indexes.
func (c *Client4) PurgeElasticsearchIndexes() (bool, *Response) {
	r, err := c.DoApiPost(c.GetElasticsearchRoute()+"/purge_indexes", "")
	if err != nil {
		return false, BuildErrorResponse(r, err)
	}
	defer closeBody(r)
	return CheckStatusOK(r), BuildResponse(r)
}


// Status Section

// GetUserStatus returns a user based on the provided user id string.
func (c *Client4) GetUserStatus(userId, etag string) (*Status, *Response) {
	r, err := c.DoApiGet(c.GetUserStatusRoute(userId), etag)
	if err != nil {
		return nil, BuildErrorResponse(r, err)
	}
	defer closeBody(r)
	return StatusFromJson(r.Body), BuildResponse(r)
}

// GetUsersStatusesByIds returns a list of users status based on the provided user ids.
func (c *Client4) GetUsersStatusesByIds(userIds []string) ([]*Status, *Response) {
	r, err := c.DoApiPost(c.GetUserStatusesRoute()+"/ids", ArrayToJson(userIds))
	if err != nil {
		return nil, BuildErrorResponse(r, err)
	}
	defer closeBody(r)
	return StatusListFromJson(r.Body), BuildResponse(r)
}

// UpdateUserStatus sets a user's status based on the provided user id string.
func (c *Client4) UpdateUserStatus(userId string, userStatus *Status) (*Status, *Response) {
	r, err := c.DoApiPut(c.GetUserStatusRoute(userId), userStatus.ToJson())
	if err != nil {
		return nil, BuildErrorResponse(r, err)
	}
	defer closeBody(r)
	return StatusFromJson(r.Body), BuildResponse(r)
}

// Timezone Section

// GetSupportedTimezone returns a page of supported timezones on the system.
func (c *Client4) GetSupportedTimezone() ([]string, *Response) {
	r, err := c.DoApiGet(c.GetTimezonesRoute(), "")
	if err != nil {
		return nil, BuildErrorResponse(r, err)
	}
	defer closeBody(r)
	var timezones []string
	json.NewDecoder(r.Body).Decode(&timezones)
	return timezones, BuildResponse(r)
}

// Open Graph Metadata Section

// OpenGraph return the open graph metadata for a particular url if the site have the metadata.
func (c *Client4) OpenGraph(url string) (map[string]string, *Response) {
	requestBody := make(map[string]string)
	requestBody["url"] = url

	r, err := c.DoApiPost(c.GetOpenGraphRoute(), MapToJson(requestBody))
	if err != nil {
		return nil, BuildErrorResponse(r, err)
	}
	defer closeBody(r)
	return MapFromJson(r.Body), BuildResponse(r)
}

// Jobs Section

// GetJob gets a single job.
func (c *Client4) GetJob(id string) (*Job, *Response) {
	r, err := c.DoApiGet(c.GetJobsRoute()+fmt.Sprintf("/%v", id), "")
	if err != nil {
		return nil, BuildErrorResponse(r, err)
	}
	defer closeBody(r)
	return JobFromJson(r.Body), BuildResponse(r)
}

// GetJobs gets all jobs, sorted with the job that was created most recently first.
func (c *Client4) GetJobs(page int, perPage int) ([]*Job, *Response) {
	r, err := c.DoApiGet(c.GetJobsRoute()+fmt.Sprintf("?page=%v&per_page=%v", page, perPage), "")
	if err != nil {
		return nil, BuildErrorResponse(r, err)
	}
	defer closeBody(r)
	return JobsFromJson(r.Body), BuildResponse(r)
}

// GetJobsByType gets all jobs of a given type, sorted with the job that was created most recently first.
func (c *Client4) GetJobsByType(jobType string, page int, perPage int) ([]*Job, *Response) {
	r, err := c.DoApiGet(c.GetJobsRoute()+fmt.Sprintf("/type/%v?page=%v&per_page=%v", jobType, page, perPage), "")
	if err != nil {
		return nil, BuildErrorResponse(r, err)
	}
	defer closeBody(r)
	return JobsFromJson(r.Body), BuildResponse(r)
}

// CreateJob creates a job based on the provided job struct.
func (c *Client4) CreateJob(job *Job) (*Job, *Response) {
	r, err := c.DoApiPost(c.GetJobsRoute(), job.ToJson())
	if err != nil {
		return nil, BuildErrorResponse(r, err)
	}
	defer closeBody(r)
	return JobFromJson(r.Body), BuildResponse(r)
}

// CancelJob requests the cancellation of the job with the provided Id.
func (c *Client4) CancelJob(jobId string) (bool, *Response) {
	r, err := c.DoApiPost(c.GetJobsRoute()+fmt.Sprintf("/%v/cancel", jobId), "")
	if err != nil {
		return false, BuildErrorResponse(r, err)
	}
	defer closeBody(r)
	return CheckStatusOK(r), BuildResponse(r)
}

// Roles Section

// GetRole gets a single role by ID.
func (c *Client4) GetRole(id string) (*Role, *Response) {
	r, err := c.DoApiGet(c.GetRolesRoute()+fmt.Sprintf("/%v", id), "")
	if err != nil {
		return nil, BuildErrorResponse(r, err)
	}
	defer closeBody(r)
	return RoleFromJson(r.Body), BuildResponse(r)
}

// GetRoleByName gets a single role by Name.
func (c *Client4) GetRoleByName(name string) (*Role, *Response) {
	r, err := c.DoApiGet(c.GetRolesRoute()+fmt.Sprintf("/name/%v", name), "")
	if err != nil {
		return nil, BuildErrorResponse(r, err)
	}
	defer closeBody(r)
	return RoleFromJson(r.Body), BuildResponse(r)
}

// GetRolesByNames returns a list of roles based on the provided role names.
func (c *Client4) GetRolesByNames(roleNames []string) ([]*Role, *Response) {
	r, err := c.DoApiPost(c.GetRolesRoute()+"/names", ArrayToJson(roleNames))
	if err != nil {
		return nil, BuildErrorResponse(r, err)
	}
	defer closeBody(r)
	return RoleListFromJson(r.Body), BuildResponse(r)
}

// PatchRole partially updates a role in the system. Any missing fields are not updated.
func (c *Client4) PatchRole(roleId string, patch *RolePatch) (*Role, *Response) {
	r, err := c.DoApiPut(c.GetRolesRoute()+fmt.Sprintf("/%v/patch", roleId), patch.ToJson())
	if err != nil {
		return nil, BuildErrorResponse(r, err)
	}
	defer closeBody(r)
	return RoleFromJson(r.Body), BuildResponse(r)
}

// Schemes Section

// CreateScheme creates a new Scheme.
func (c *Client4) CreateScheme(scheme *Scheme) (*Scheme, *Response) {
	r, err := c.DoApiPost(c.GetSchemesRoute(), scheme.ToJson())
	if err != nil {
		return nil, BuildErrorResponse(r, err)
	}
	defer closeBody(r)
	return SchemeFromJson(r.Body), BuildResponse(r)
}

// GetScheme gets a single scheme by ID.
func (c *Client4) GetScheme(id string) (*Scheme, *Response) {
	r, err := c.DoApiGet(c.GetSchemeRoute(id), "")
	if err != nil {
		return nil, BuildErrorResponse(r, err)
	}
	defer closeBody(r)
	return SchemeFromJson(r.Body), BuildResponse(r)
}

// GetSchemes gets all schemes, sorted with the most recently created first, optionally filtered by scope.
func (c *Client4) GetSchemes(scope string, page int, perPage int) ([]*Scheme, *Response) {
	r, err := c.DoApiGet(c.GetSchemesRoute()+fmt.Sprintf("?scope=%v&page=%v&per_page=%v", scope, page, perPage), "")
	if err != nil {
		return nil, BuildErrorResponse(r, err)
	}
	defer closeBody(r)
	return SchemesFromJson(r.Body), BuildResponse(r)
}

// DeleteScheme deletes a single scheme by ID.
func (c *Client4) DeleteScheme(id string) (bool, *Response) {
	r, err := c.DoApiDelete(c.GetSchemeRoute(id))
	if err != nil {
		return false, BuildErrorResponse(r, err)
	}
	defer closeBody(r)
	return CheckStatusOK(r), BuildResponse(r)
}

// PatchScheme partially updates a scheme in the system. Any missing fields are not updated.
func (c *Client4) PatchScheme(id string, patch *SchemePatch) (*Scheme, *Response) {
	r, err := c.DoApiPut(c.GetSchemeRoute(id)+"/patch", patch.ToJson())
	if err != nil {
		return nil, BuildErrorResponse(r, err)
	}
	defer closeBody(r)
	return SchemeFromJson(r.Body), BuildResponse(r)
}

// GetTeamsForScheme gets the teams using this scheme, sorted alphabetically by display name.
func (c *Client4) GetTeamsForScheme(schemeId string, page int, perPage int) ([]*Team, *Response) {
	r, err := c.DoApiGet(c.GetSchemeRoute(schemeId)+fmt.Sprintf("/teams?page=%v&per_page=%v", page, perPage), "")
	if err != nil {
		return nil, BuildErrorResponse(r, err)
	}
	defer closeBody(r)
	return TeamListFromJson(r.Body), BuildResponse(r)
}

// GetChannelsForScheme gets the channels using this scheme, sorted alphabetically by display name.
func (c *Client4) GetChannelsForScheme(schemeId string, page int, perPage int) (ChannelList, *Response) {
	r, err := c.DoApiGet(c.GetSchemeRoute(schemeId)+fmt.Sprintf("/channels?page=%v&per_page=%v", page, perPage), "")
	if err != nil {
		return nil, BuildErrorResponse(r, err)
	}
	defer closeBody(r)
	return *ChannelListFromJson(r.Body), BuildResponse(r)
}

// UpdateChannelScheme will update a channel's scheme.
func (c *Client4) UpdateChannelScheme(channelId, schemeId string) (bool, *Response) {
	sip := &SchemeIDPatch{SchemeID: &schemeId}
	r, err := c.DoApiPut(c.GetChannelSchemeRoute(channelId), sip.ToJson())
	if err != nil {
		return false, BuildErrorResponse(r, err)
	}
	defer closeBody(r)
	return CheckStatusOK(r), BuildResponse(r)
}

// UpdateTeamScheme will update a team's scheme.
func (c *Client4) UpdateTeamScheme(teamId, schemeId string) (bool, *Response) {
	sip := &SchemeIDPatch{SchemeID: &schemeId}
	r, err := c.DoApiPut(c.GetTeamSchemeRoute(teamId), sip.ToJson())
	if err != nil {
		return false, BuildErrorResponse(r, err)
	}
	defer closeBody(r)
	return CheckStatusOK(r), BuildResponse(r)
}

// GetRedirectLocation retrieves the value of the 'Location' header of an HTTP response for a given URL.
func (c *Client4) GetRedirectLocation(urlParam, etag string) (string, *Response) {
	url := fmt.Sprintf("%s?url=%s", c.GetRedirectLocationRoute(), url.QueryEscape(urlParam))
	r, err := c.DoApiGet(url, etag)
	if err != nil {
		return "", BuildErrorResponse(r, err)
	}
	defer closeBody(r)
	return MapFromJson(r.Body)["location"], BuildResponse(r)
}

// RegisterTermsOfServiceAction saves action performed by a user against a specific terms of service.
func (c *Client4) RegisterTermsOfServiceAction(userId, termsOfServiceId string, accepted bool) (*bool, *Response) {
	url := c.GetUserTermsOfServiceRoute(userId)
	data := map[string]interface{}{"termsOfServiceId": termsOfServiceId, "accepted": accepted}
	r, err := c.DoApiPost(url, StringInterfaceToJson(data))
	if err != nil {
		return nil, BuildErrorResponse(r, err)
	}
	defer closeBody(r)
	return NewBool(CheckStatusOK(r)), BuildResponse(r)
}
