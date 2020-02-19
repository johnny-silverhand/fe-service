package app

import (
	"fmt"
	"html/template"
	"net/http"
	"strconv"

	goi18n "github.com/nicksnyder/go-i18n/i18n"
	"im/einterfaces"
	"im/jobs"
	"im/mlog"
	"im/model"
	"im/services/httpservice"
	"im/services/imageproxy"
	"im/services/timezones"
	"im/utils"
)

type App struct {
	Srv *Server

	Log *mlog.Logger

	T              goi18n.TranslateFunc
	Session        model.Session
	RequestId      string
	IpAddress      string
	Path           string
	UserAgent      string
	AcceptLanguage string

	Cluster einterfaces.ClusterInterface

	Elasticsearch einterfaces.ElasticsearchInterface

	HTTPService httpservice.HTTPService
	ImageProxy  *imageproxy.ImageProxy
	Timezones   *timezones.Timezones
}

func New(options ...AppOption) *App {
	app := &App{}

	for _, option := range options {
		option(app)
	}

	return app
}

// DO NOT CALL THIS.
// This is to avoid having to change all the code in cmd/mattermost/commands/* for now
// shutdown should be called directly on the server
func (a *App) Shutdown() {
	a.Srv.Shutdown()
	a.Srv = nil
}

func (a *App) configOrLicenseListener() {
	a.regenerateClientConfig()
}

func (s *Server) initJobs() {
	s.Jobs = jobs.NewJobServer(s, s.Store)

	if jobsElasticsearchAggregatorInterface != nil {
		s.Jobs.ElasticsearchAggregator = jobsElasticsearchAggregatorInterface(s.FakeApp())
	}
	if jobsElasticsearchIndexerInterface != nil {
		s.Jobs.ElasticsearchIndexer = jobsElasticsearchIndexerInterface(s.FakeApp())
	}

	s.Jobs.Workers = s.Jobs.InitWorkers()
	s.Jobs.Schedulers = s.Jobs.InitSchedulers()
}

func (a *App) DiagnosticId() string {
	return a.Srv.diagnosticId
}

func (a *App) SetDiagnosticId(id string) {
	a.Srv.diagnosticId = id
}

func (a *App) EnsureDiagnosticId() {
	if a.Srv.diagnosticId != "" {
		return
	}
	if result := <-a.Srv.Store.System().Get(); result.Err == nil {
		props := result.Data.(model.StringMap)

		id := props[model.SYSTEM_DIAGNOSTIC_ID]
		if len(id) == 0 {
			id = model.NewId()
			systemId := &model.System{Name: model.SYSTEM_DIAGNOSTIC_ID, Value: id}
			<-a.Srv.Store.System().Save(systemId)
		}

		a.Srv.diagnosticId = id
	}
}

func (a *App) MasterKey() string {
	return a.Srv.masterKey
}

func (a *App) SetMasterKey(id string) {
	a.Srv.masterKey = id
}

func (a *App) ResetMasterKey() {
	if result := <-a.Srv.Store.System().Get(); result.Err == nil {
		props := result.Data.(model.StringMap)

		id := props[model.SYSTEM_MASTER_KEY]
		if len(id) != 0 {
			id = model.NewId() + model.NewId()
			systemId := &model.System{Name: model.SYSTEM_MASTER_KEY, Value: id}
			<-a.Srv.Store.System().Update(systemId)
		}

		a.Srv.masterKey = id
	}
}

func (a *App) EnsureMasterKey() {
	if a.Srv.masterKey != "" {
		return
	}
	if result := <-a.Srv.Store.System().Get(); result.Err == nil {
		props := result.Data.(model.StringMap)

		id := props[model.SYSTEM_MASTER_KEY]
		if len(id) == 0 {
			id = model.NewId() + model.NewId()
			systemId := &model.System{Name: model.SYSTEM_MASTER_KEY, Value: id}
			<-a.Srv.Store.System().Save(systemId)
		}

		a.Srv.masterKey = id
	}
}

func (a *App) HTMLTemplates() *template.Template {
	if a.Srv.htmlTemplateWatcher != nil {
		return a.Srv.htmlTemplateWatcher.Templates()
	}

	return nil
}

func (a *App) Handle404(w http.ResponseWriter, r *http.Request) {
	err := model.NewAppError("Handle404", "api.context.404.app_error", nil, "", http.StatusNotFound)

	mlog.Debug(fmt.Sprintf("%v: code=404 ip=%v", r.URL.Path, utils.GetIpAddress(r)))

	utils.RenderWebAppError(a.Config(), w, r, err, a.AsymmetricSigningKey())
}

func (a *App) getSystemInstallDate() (int64, *model.AppError) {
	result := <-a.Srv.Store.System().GetByName(model.SYSTEM_INSTALLATION_DATE_KEY)
	if result.Err != nil {
		return 0, result.Err
	}
	systemData := result.Data.(*model.System)
	value, err := strconv.ParseInt(systemData.Value, 10, 64)
	if err != nil {
		return 0, model.NewAppError("getSystemInstallDate", "app.system_install_date.parse_int.app_error", nil, err.Error(), http.StatusInternalServerError)
	}
	return value, nil
}
