package web

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
)

const (
	PAGE_DEFAULT          = 0
	PER_PAGE_DEFAULT      = 60
	PER_PAGE_MAXIMUM      = 200
	LOGS_PER_PAGE_DEFAULT = 10000
	LOGS_PER_PAGE_MAXIMUM = 10000
)

type Params struct {
	UserId           string
	TeamId           string
	InviteId         string
	TokenId          string
	ChannelId        string
	PromoId          string
	OfficeId         string
	OrderId          string
	TransactionId    string
	LevelId          string
	ExtraId          string
	ProductId        string
	CategoryId       string
	CategoryParentId string
	CategoryDepth    string
	DestinationId    string
	PostId           string
	FileId           string
	Filename         string
	PluginId         string
	CommandId        string
	HookId           string
	ReportId         string
	EmojiId          string
	AppId            string
	Email            string
	Username         string
	TeamName         string
	ChannelName      string
	PreferenceName   string
	EmojiName        string
	Category         string
	Service          string
	JobId            string
	JobType          string
	ActionId         string
	RoleId           string
	RoleName         string
	SchemeId         string
	Scope            string
	GroupId          string
	MessageId        string
	Page             int
	PerPage          int
	LogsPerPage      int
	Permanent        bool
	RemoteId         string
	Status           string
	Active           bool

	Sort         string
	BotUserId    string
	Q            string
	IsLinked     *bool
	IsConfigured *bool

	BuildId       string
	BuildType     string
	BuildSettings string
}

func ParamsFromRequest(r *http.Request) *Params {
	params := &Params{}

	props := mux.Vars(r)
	query := r.URL.Query()

	if val, ok := props["user_id"]; ok {
		params.UserId = val
	}

	if val := query.Get("sort"); val == "" {
		params.Sort = "Id"
	} else {
		params.Sort = val
	}

	if val, ok := props["team_id"]; ok {
		params.TeamId = val
	}

	if val, ok := props["invite_id"]; ok {
		params.InviteId = val
	} else if val := query.Get("invite_id"); val != "" {
		params.InviteId = val
	}

	if val, ok := props["token_id"]; ok {
		params.TokenId = val
	}

	if val, ok := props["channel_id"]; ok {
		params.ChannelId = val
	} else {
		params.ChannelId = query.Get("channel_id")
	}

	if val, ok := props["post_id"]; ok {
		params.PostId = val
	}

	if val, ok := props["file_id"]; ok {
		params.FileId = val
	}

	params.Filename = query.Get("filename")

	if val, ok := props["plugin_id"]; ok {
		params.PluginId = val
	}

	if val, ok := props["command_id"]; ok {
		params.CommandId = val
	}

	if val, ok := props["hook_id"]; ok {
		params.HookId = val
	}

	if val, ok := props["report_id"]; ok {
		params.ReportId = val
	}

	if val, ok := props["product_id"]; ok {
		params.ProductId = val
	}
	if val, ok := props["promo_id"]; ok {
		params.PromoId = val
	}
	if val, ok := props["office_id"]; ok {
		params.OfficeId = val
	} else if val := query.Get("office_id"); val != "" {
		params.OfficeId = val
	} else {
		params.OfficeId = r.Header.Get("OfficeId")
	}

	if val, ok := props["transaction_id"]; ok {
		params.TransactionId = val
	}
	if val, ok := props["order_id"]; ok {
		params.OrderId = val
	}
	if val, ok := props["level_id"]; ok {
		params.LevelId = val
	}

	if val, ok := props["extra_id"]; ok {
		params.ExtraId = val
	}
	if val, ok := props["app_id"]; ok {
		params.AppId = val
	} else if val = r.Header.Get("AppId"); val != "" {
		params.AppId = val
	} else {
		params.AppId = query.Get("app_id")
	}

	if val, ok := props["status"]; ok {
		params.Status = val
	} else {
		params.Status = query.Get("status")
	}

	if val, ok := props["email"]; ok {
		params.Email = val
	}

	if val, ok := props["username"]; ok {
		params.Username = val
	}

	if val, ok := props["team_name"]; ok {
		params.TeamName = strings.ToLower(val)
	}

	if val, ok := props["channel_name"]; ok {
		params.ChannelName = strings.ToLower(val)
	}
	if val, ok := props["category_id"]; ok {
		params.CategoryId = val
	}
	if val, ok := props["parent_category_id"]; ok {
		params.CategoryParentId = val
	}
	if val, ok := props["depth"]; ok {
		params.CategoryDepth = val
	}
	if val, ok := props["destination_id"]; ok {
		params.DestinationId = val
	}
	if val, ok := props["category"]; ok {
		params.Category = val
	}
	if val, ok := props["message_id"]; ok {
		params.MessageId = val
	}
	if val, ok := props["service"]; ok {
		params.Service = val
	}

	if val, ok := props["preference_name"]; ok {
		params.PreferenceName = val
	}

	if val, ok := props["emoji_name"]; ok {
		params.EmojiName = val
	}

	if val, ok := props["job_id"]; ok {
		params.JobId = val
	}

	if val, ok := props["job_type"]; ok {
		params.JobType = val
	}

	if val, ok := props["action_id"]; ok {
		params.ActionId = val
	}

	if val, ok := props["role_id"]; ok {
		params.RoleId = val
	}

	if val, ok := props["role_name"]; ok {
		params.RoleName = val
	}

	if val, ok := props["scheme_id"]; ok {
		params.SchemeId = val
	}

	if val, ok := props["group_id"]; ok {
		params.GroupId = val
	}

	if val, ok := props["remote_id"]; ok {
		params.RemoteId = val
	}

	if val := query.Get("build_id"); val != "" {
		params.BuildId = val
	}

	if val := query.Get("build_type"); val != "" {
		params.BuildType = val
	}

	params.Scope = query.Get("scope")

	if val, err := strconv.Atoi(query.Get("page")); err != nil || val < 0 {
		params.Page = PAGE_DEFAULT
	} else {
		params.Page = val
	}

	if val, err := strconv.ParseBool(query.Get("permanent")); err == nil {
		params.Permanent = val
	}

	if val, err := strconv.ParseBool(query.Get("active")); err == nil {
		params.Active = val
	}

	if val, err := strconv.Atoi(query.Get("per_page")); err != nil || val < 0 {
		params.PerPage = PER_PAGE_DEFAULT
	} else if val > PER_PAGE_MAXIMUM {
		params.PerPage = PER_PAGE_MAXIMUM
	} else {
		params.PerPage = val
	}

	if val, err := strconv.Atoi(query.Get("logs_per_page")); err != nil || val < 0 {
		params.LogsPerPage = LOGS_PER_PAGE_DEFAULT
	} else if val > LOGS_PER_PAGE_MAXIMUM {
		params.LogsPerPage = LOGS_PER_PAGE_MAXIMUM
	} else {
		params.LogsPerPage = val
	}

	params.Q = query.Get("q")

	if val, err := strconv.ParseBool(query.Get("is_linked")); err == nil {
		params.IsLinked = &val
	}

	if val, err := strconv.ParseBool(query.Get("is_configured")); err == nil {
		params.IsConfigured = &val
	}

	return params
}
