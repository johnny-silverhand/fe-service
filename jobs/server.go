package jobs

import (
	ejobs "im/einterfaces/jobs"
	"im/model"
	"im/services/configservice"
	"im/store"
)

type JobServer struct {
	ConfigService configservice.ConfigService
	Store         store.Store
	Workers       *Workers
	Schedulers    *Schedulers

	ElasticsearchAggregator ejobs.ElasticsearchAggregatorInterface
	ElasticsearchIndexer    ejobs.ElasticsearchIndexerInterface
}

func NewJobServer(configService configservice.ConfigService, store store.Store) *JobServer {
	return &JobServer{
		ConfigService: configService,
		Store:         store,
	}
}

func (srv *JobServer) Config() *model.Config {
	return srv.ConfigService.Config()
}

func (srv *JobServer) StartWorkers() {
	srv.Workers = srv.Workers.Start()
}

func (srv *JobServer) StartSchedulers() {
	srv.Schedulers = srv.Schedulers.Start()
}

func (srv *JobServer) StopWorkers() {
	if srv.Workers != nil {
		srv.Workers.Stop()
	}
}

func (srv *JobServer) StopSchedulers() {
	if srv.Schedulers != nil {
		srv.Schedulers.Stop()
	}
}
