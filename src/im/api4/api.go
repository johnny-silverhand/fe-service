package api4

import (
	"net/http"

	"github.com/gorilla/mux"
	"im/app"
	"im/model"
	"im/services/configservice"
	"im/web"

	_ "github.com/nicksnyder/go-i18n/i18n"
)

type Routes struct {
	Root    *mux.Router // ''
	ApiRoot *mux.Router // 'api/v4'

	Users          *mux.Router // 'api/v4/users'
	User           *mux.Router // 'api/v4/users/{user_id:[A-Za-z0-9]+}'
	UserByUsername *mux.Router // 'api/v4/users/username/{username:[A-Za-z0-9_-\.]+}'
	UserByEmail    *mux.Router // 'api/v4/users/email/{email}'

	Teams              *mux.Router // 'api/v4/teams'
	TeamsForUser       *mux.Router // 'api/v4/users/{user_id:[A-Za-z0-9]+}/teams'
	Team               *mux.Router // 'api/v4/teams/{team_id:[A-Za-z0-9]+}'
	TeamForUser        *mux.Router // 'api/v4/users/{user_id:[A-Za-z0-9]+}/teams/{team_id:[A-Za-z0-9]+}'
	TeamByName         *mux.Router // 'api/v4/teams/name/{team_name:[A-Za-z0-9_-]+}'
	TeamMembers        *mux.Router // 'api/v4/teams/{team_id:[A-Za-z0-9_-]+}/members'
	TeamMember         *mux.Router // 'api/v4/teams/{team_id:[A-Za-z0-9_-]+}/members/{user_id:[A-Za-z0-9_-]+}'
	TeamMembersForUser *mux.Router // 'api/v4/users/{user_id:[A-Za-z0-9]+}/teams/members'

	Channels                 *mux.Router // 'api/v4/channels'
	Channel                  *mux.Router // 'api/v4/channels/{channel_id:[A-Za-z0-9]+}'
	ChannelForUser           *mux.Router // 'api/v4/users/{user_id:[A-Za-z0-9]+}/channels/{channel_id:[A-Za-z0-9]+}'
	ChannelByName            *mux.Router // 'api/v4/teams/{team_id:[A-Za-z0-9]+}/channels/name/{channel_name:[A-Za-z0-9_-]+}'
	ChannelByNameForTeamName *mux.Router // 'api/v4/teams/name/{team_name:[A-Za-z0-9_-]+}/channels/name/{channel_name:[A-Za-z0-9_-]+}'
	ChannelsForTeam          *mux.Router // 'api/v4/teams/{team_id:[A-Za-z0-9]+}/channels'
	ChannelMembers           *mux.Router // 'api/v4/channels/{channel_id:[A-Za-z0-9]+}/members'
	ChannelMember            *mux.Router // 'api/v4/channels/{channel_id:[A-Za-z0-9]+}/members/{user_id:[A-Za-z0-9]+}'
	ChannelMembersForUser    *mux.Router // 'api/v4/users/{user_id:[A-Za-z0-9]+}/teams/{team_id:[A-Za-z0-9]+}/channels/members'
	ChannelAllMembersForUser *mux.Router // 'api/v4/users/{user_id:[A-Za-z0-9]+}/channels/members'

	Posts           *mux.Router // 'api/v4/posts'
	Post            *mux.Router // 'api/v4/posts/{post_id:[A-Za-z0-9]+}'
	PostsForChannel *mux.Router // 'api/v4/channels/{channel_id:[A-Za-z0-9]+}/posts'
	PostsForUser    *mux.Router // 'api/v4/users/{user_id:[A-Za-z0-9]+}/posts'
	PostForUser     *mux.Router // 'api/v4/users/{user_id:[A-Za-z0-9]+}/posts/{post_id:[A-Za-z0-9]+}'

	Messages           *mux.Router // 'api/v4/messages'
	Message            *mux.Router // 'api/v4/messages/{message_id:[A-Za-z0-9]+}'
	MessagesForChannel *mux.Router // 'api/v4/channels/{channel_id:[A-Za-z0-9]+}/messages'
	MessagesForUser    *mux.Router // 'api/v4/users/{user_id:[A-Za-z0-9]+}/messages'
	MessageForUser     *mux.Router // 'api/v4/users/{user_id:[A-Za-z0-9]+}/messages/{message_id:[A-Za-z0-9]+}'

	Sections *mux.Router // 'api/v4/sections'
	Section  *mux.Router // 'api/v4/sections/{sections_id:[A-Za-z0-9]+}'

	Files *mux.Router // 'api/v4/files'
	File  *mux.Router // 'api/v4/files/{file_id:[A-Za-z0-9]+}'

	Plugins *mux.Router // 'api/v4/plugins'
	Plugin  *mux.Router // 'api/v4/plugins/{plugin_id:[A-Za-z0-9_-]+}'

	PublicFile *mux.Router // 'files/{file_id:[A-Za-z0-9]+}/public'

	OAuth     *mux.Router // 'api/v4/oauth'
	OAuthApps *mux.Router // 'api/v4/oauth/apps'
	OAuthApp  *mux.Router // 'api/v4/oauth/apps/{app_id:[A-Za-z0-9]+}'

	OpenGraph *mux.Router // 'api/v4/opengraph'

	Cluster *mux.Router // 'api/v4/cluster'

	Image *mux.Router // 'api/v4/image'

	Elasticsearch *mux.Router // 'api/v4/elasticsearch'

	Brand *mux.Router // 'api/v4/brand'

	System *mux.Router // 'api/v4/system'

	Jobs *mux.Router // 'api/v4/jobs'

	Preferences *mux.Router // 'api/v4/users/{user_id:[A-Za-z0-9]+}/preferences'

	Public *mux.Router // 'api/v4/public'

	Roles   *mux.Router // 'api/v4/roles'
	Schemes *mux.Router // 'api/v4/schemes'

	Categories         *mux.Router // 'api/v4/categories'
	Category           *mux.Router // 'api/v4/categories/{category_id:[A-Za-z0-9_-]+}'
	CategoriesByClient *mux.Router // 'api/v4/categories/client/{client_id:[A-Za-z0-9_-]+}'

	ProductsForCategory *mux.Router // 'api/v4/categories/{category_id:[A-Za-z0-9]+}/products'

	Products *mux.Router // 'api/v4/products'
	Product  *mux.Router // 'api/v4/products/{product_id:[A-Za-z0-9_-]+}'

	ProductsByClient *mux.Router // 'api/v4/products/client/{client_id:[A-Za-z0-9_-]+}'

	Promos *mux.Router // 'api/v4/promos'
	Promo  *mux.Router // 'api/v4/promos/{promo_id:[A-Za-z0-9_-]+}'

	Offices *mux.Router // 'api/v4/offices'
	Office  *mux.Router // 'api/v4/offices/{office_id:[A-Za-z0-9_-]+}'

	Orders *mux.Router // 'api/v4/orders'
	Order  *mux.Router // 'api/v4/orders/{order_id:[A-Za-z0-9_-]+}'

	Transactions *mux.Router // 'api/v4/transactions'
	Transaction  *mux.Router // 'api/v4/transactions/{transaction_id:[A-Za-z0-9_-]+}'

	Basket *mux.Router // 'api/v4/basket/{office_id:[A-Za-z0-9_-]+}'

	Levels *mux.Router // 'api/v4/levels'
	Level  *mux.Router // 'api/v4/levels/{level_id:[A-Za-z0-9_-]+}'

	Extras *mux.Router // 'api/v4/extras'
	Extra  *mux.Router // 'api/v4/extras/{extra_id:[A-Za-z0-9_-]+}'

}

type API struct {
	ConfigService       configservice.ConfigService
	GetGlobalAppOptions app.AppOptionCreator
	BaseRoutes          *Routes
}

func Init(configservice configservice.ConfigService, globalOptionsFunc app.AppOptionCreator, root *mux.Router) *API {
	api := &API{
		ConfigService:       configservice,
		GetGlobalAppOptions: globalOptionsFunc,
		BaseRoutes:          &Routes{},
	}

	api.BaseRoutes.Root = root
	api.BaseRoutes.ApiRoot = root.PathPrefix(model.API_URL_SUFFIX).Subrouter()

	api.BaseRoutes.Users = api.BaseRoutes.ApiRoot.PathPrefix("/users").Subrouter()
	api.BaseRoutes.User = api.BaseRoutes.ApiRoot.PathPrefix("/users/{user_id:[A-Za-z0-9]+}").Subrouter()
	api.BaseRoutes.UserByUsername = api.BaseRoutes.Users.PathPrefix("/username/{username:[A-Za-z0-9\\_\\-\\.]+}").Subrouter()
	api.BaseRoutes.UserByEmail = api.BaseRoutes.Users.PathPrefix("/email/{email}").Subrouter()

	api.BaseRoutes.Teams = api.BaseRoutes.ApiRoot.PathPrefix("/teams").Subrouter()
	api.BaseRoutes.TeamsForUser = api.BaseRoutes.User.PathPrefix("/teams").Subrouter()
	api.BaseRoutes.Team = api.BaseRoutes.Teams.PathPrefix("/{team_id:[A-Za-z0-9]+}").Subrouter()
	api.BaseRoutes.TeamForUser = api.BaseRoutes.TeamsForUser.PathPrefix("/{team_id:[A-Za-z0-9]+}").Subrouter()
	api.BaseRoutes.TeamByName = api.BaseRoutes.Teams.PathPrefix("/name/{team_name:[A-Za-z0-9_-]+}").Subrouter()
	api.BaseRoutes.TeamMembers = api.BaseRoutes.Team.PathPrefix("/members").Subrouter()
	api.BaseRoutes.TeamMember = api.BaseRoutes.TeamMembers.PathPrefix("/{user_id:[A-Za-z0-9]+}").Subrouter()
	api.BaseRoutes.TeamMembersForUser = api.BaseRoutes.User.PathPrefix("/teams/members").Subrouter()

	api.BaseRoutes.Channels = api.BaseRoutes.ApiRoot.PathPrefix("/channels").Subrouter()
	api.BaseRoutes.Channel = api.BaseRoutes.Channels.PathPrefix("/{channel_id:[A-Za-z0-9]+}").Subrouter()
	api.BaseRoutes.ChannelForUser = api.BaseRoutes.User.PathPrefix("/channels/{channel_id:[A-Za-z0-9]+}").Subrouter()
	api.BaseRoutes.ChannelByName = api.BaseRoutes.Team.PathPrefix("/channels/name/{channel_name:[A-Za-z0-9_-]+}").Subrouter()
	api.BaseRoutes.ChannelByNameForTeamName = api.BaseRoutes.TeamByName.PathPrefix("/channels/name/{channel_name:[A-Za-z0-9_-]+}").Subrouter()
	api.BaseRoutes.ChannelsForTeam = api.BaseRoutes.Team.PathPrefix("/channels").Subrouter()
	api.BaseRoutes.ChannelMembers = api.BaseRoutes.Channel.PathPrefix("/members").Subrouter()
	api.BaseRoutes.ChannelMember = api.BaseRoutes.ChannelMembers.PathPrefix("/{user_id:[A-Za-z0-9]+}").Subrouter()
	api.BaseRoutes.ChannelMembersForUser = api.BaseRoutes.User.PathPrefix("/teams/{team_id:[A-Za-z0-9]+}/channels/members").Subrouter()
	api.BaseRoutes.ChannelAllMembersForUser = api.BaseRoutes.User.PathPrefix("/members").Subrouter()

	api.BaseRoutes.Posts = api.BaseRoutes.ApiRoot.PathPrefix("/posts").Subrouter()
	api.BaseRoutes.Post = api.BaseRoutes.Posts.PathPrefix("/{post_id:[A-Za-z0-9]+}").Subrouter()
	api.BaseRoutes.PostsForChannel = api.BaseRoutes.Channel.PathPrefix("/posts").Subrouter()
	api.BaseRoutes.PostsForUser = api.BaseRoutes.User.PathPrefix("/posts").Subrouter()
	api.BaseRoutes.PostForUser = api.BaseRoutes.PostsForUser.PathPrefix("/{post_id:[A-Za-z0-9]+}").Subrouter()

	api.BaseRoutes.Files = api.BaseRoutes.ApiRoot.PathPrefix("/files").Subrouter()
	api.BaseRoutes.File = api.BaseRoutes.Files.PathPrefix("/{file_id:[A-Za-z0-9]+}").Subrouter()
	api.BaseRoutes.PublicFile = api.BaseRoutes.Root.PathPrefix("/files/{file_id:[A-Za-z0-9]+}/public").Subrouter()

	api.BaseRoutes.Plugins = api.BaseRoutes.ApiRoot.PathPrefix("/plugins").Subrouter()
	api.BaseRoutes.Plugin = api.BaseRoutes.Plugins.PathPrefix("/{plugin_id:[A-Za-z0-9\\_\\-\\.]+}").Subrouter()

	api.BaseRoutes.OAuth = api.BaseRoutes.ApiRoot.PathPrefix("/oauth").Subrouter()
	api.BaseRoutes.OAuthApps = api.BaseRoutes.OAuth.PathPrefix("/apps").Subrouter()
	api.BaseRoutes.OAuthApp = api.BaseRoutes.OAuthApps.PathPrefix("/{app_id:[A-Za-z0-9]+}").Subrouter()

	api.BaseRoutes.Cluster = api.BaseRoutes.ApiRoot.PathPrefix("/cluster").Subrouter()
	api.BaseRoutes.Brand = api.BaseRoutes.ApiRoot.PathPrefix("/brand").Subrouter()
	api.BaseRoutes.System = api.BaseRoutes.ApiRoot.PathPrefix("/system").Subrouter()
	api.BaseRoutes.Preferences = api.BaseRoutes.User.PathPrefix("/preferences").Subrouter()
	api.BaseRoutes.Public = api.BaseRoutes.ApiRoot.PathPrefix("/public").Subrouter()
	api.BaseRoutes.Jobs = api.BaseRoutes.ApiRoot.PathPrefix("/jobs").Subrouter()
	api.BaseRoutes.Elasticsearch = api.BaseRoutes.ApiRoot.PathPrefix("/elasticsearch").Subrouter()

	api.BaseRoutes.OpenGraph = api.BaseRoutes.ApiRoot.PathPrefix("/opengraph").Subrouter()

	api.BaseRoutes.Roles = api.BaseRoutes.ApiRoot.PathPrefix("/roles").Subrouter()
	api.BaseRoutes.Schemes = api.BaseRoutes.ApiRoot.PathPrefix("/schemes").Subrouter()

	api.BaseRoutes.Image = api.BaseRoutes.ApiRoot.PathPrefix("/image").Subrouter()

	api.BaseRoutes.Messages = api.BaseRoutes.ApiRoot.PathPrefix("/messages").Subrouter()
	api.BaseRoutes.Message = api.BaseRoutes.Messages.PathPrefix("/{message_id:[A-Za-z0-9]+}").Subrouter()
	api.BaseRoutes.MessagesForChannel = api.BaseRoutes.Channel.PathPrefix("/messages").Subrouter()
	api.BaseRoutes.MessagesForUser = api.BaseRoutes.User.PathPrefix("/messages").Subrouter()
	api.BaseRoutes.MessageForUser = api.BaseRoutes.MessagesForUser.PathPrefix("/{message_id:[A-Za-z0-9]+}").Subrouter()

	/*	api.BaseRoutes.Sections  =  api.BaseRoutes.ApiRoot.PathPrefix("/sections").Subrouter()
		api.BaseRoutes.Section  =  api.BaseRoutes.Sections.PathPrefix("/{section_id:[A-Za-z0-9]+}").Subrouter()*/

	api.BaseRoutes.Categories = api.BaseRoutes.ApiRoot.PathPrefix("/categories").Subrouter()
	api.BaseRoutes.Category = api.BaseRoutes.Categories.PathPrefix("/{category_id:[A-Za-z0-9]+}").Subrouter()
	api.BaseRoutes.CategoriesByClient = api.BaseRoutes.Categories.PathPrefix("/client/{client_id:[A-Za-z0-9]+}").Subrouter()

	api.BaseRoutes.Products = api.BaseRoutes.ApiRoot.PathPrefix("/products").Subrouter()
	api.BaseRoutes.Product = api.BaseRoutes.Products.PathPrefix("/{product_id:[A-Za-z0-9]+}").Subrouter()
	api.BaseRoutes.ProductsByClient = api.BaseRoutes.Products.PathPrefix("/client/{client_id:[A-Za-z0-9]+}").Subrouter()

	api.BaseRoutes.ProductsForCategory = api.BaseRoutes.Category.PathPrefix("/products").Subrouter()

	api.BaseRoutes.Promos = api.BaseRoutes.ApiRoot.PathPrefix("/promos").Subrouter()
	api.BaseRoutes.Promo = api.BaseRoutes.Promos.PathPrefix("/{promo_id:[A-Za-z0-9]+}").Subrouter()

	api.BaseRoutes.Offices = api.BaseRoutes.ApiRoot.PathPrefix("/offices").Subrouter()
	api.BaseRoutes.Office = api.BaseRoutes.Offices.PathPrefix("/{office_id:[A-Za-z0-9]+}").Subrouter()

	api.BaseRoutes.Orders = api.BaseRoutes.ApiRoot.PathPrefix("/orders").Subrouter()
	api.BaseRoutes.Order = api.BaseRoutes.Orders.PathPrefix("/{order_id:[A-Za-z0-9]+}").Subrouter()

	api.BaseRoutes.Transactions = api.BaseRoutes.ApiRoot.PathPrefix("/transactions").Subrouter()
	api.BaseRoutes.Transaction = api.BaseRoutes.Transactions.PathPrefix("/{transaction_id:[A-Za-z0-9]+}").Subrouter()

	api.BaseRoutes.Basket = api.BaseRoutes.ApiRoot.PathPrefix("/basket").Subrouter()

	api.BaseRoutes.Levels = api.BaseRoutes.ApiRoot.PathPrefix("/levels").Subrouter()
	api.BaseRoutes.Level = api.BaseRoutes.Levels.PathPrefix("/{level_id:[A-Za-z0-9]+}").Subrouter()

	api.BaseRoutes.Extras = api.BaseRoutes.ApiRoot.PathPrefix("/extras").Subrouter()
	api.BaseRoutes.Extra = api.BaseRoutes.Extras.PathPrefix("/{extra_id:[A-Za-z0-9]+}").Subrouter()

	api.InitUser()

	api.InitTeam()
	api.InitChannel()
	api.InitPost()
	api.InitMessage()
	api.InitFile()
	api.InitSystem()

	api.InitConfig()
	api.InitPreference()

	api.InitCluster()

	api.InitElasticsearch()

	api.InitBrand()
	api.InitJob()

	api.InitStatus()
	api.InitWebSocket()

	api.InitOAuth()

	api.InitOpenGraph()

	api.InitRole()
	api.InitScheme()
	api.InitImage()
	api.InitCategory()
	api.InitProduct()
	api.InitSection()
	api.InitPromo()
	api.InitOffice()
	api.InitOrder()
	api.InitTransaction()
	api.InitLevel()
	api.InitExtra()
	//api.InitBasket()
	root.Handle("/api/v4/{anything:.*}", http.HandlerFunc(api.Handle404))

	return api
}

func (api *API) Handle404(w http.ResponseWriter, r *http.Request) {
	web.Handle404(api.ConfigService, w, r)
}

var ReturnStatusOK = web.ReturnStatusOK

func ReturnStatusStageTokenOK(w http.ResponseWriter, token string) {
	m := make(map[string]string)
	m[model.STATUS] = model.STATUS_OK
	m[model.STAGE_TOKEN] = token
	w.Write([]byte(model.MapToJson(m)))
}
