package jobs

import (
	"sync"

	"im/mlog"
	"im/model"
	"im/services/configservice"
)

type Workers struct {
	startOnce     sync.Once
	ConfigService configservice.ConfigService
	Watcher       *Watcher

	DataRetention            model.Worker
	MessageExport            model.Worker
	ElasticsearchIndexing    model.Worker
	ElasticsearchAggregation model.Worker
	LdapSync                 model.Worker
	Migrations               model.Worker
	Plugins                  model.Worker

	listenerId string
}

func (srv *JobServer) InitWorkers() *Workers {
	workers := &Workers{
		ConfigService: srv.ConfigService,
	}
	workers.Watcher = srv.MakeWatcher(workers, DEFAULT_WATCHER_POLLING_INTERVAL)

	if elasticsearchIndexerInterface := srv.ElasticsearchIndexer; elasticsearchIndexerInterface != nil {
		workers.ElasticsearchIndexing = elasticsearchIndexerInterface.MakeWorker()
	}

	if elasticsearchAggregatorInterface := srv.ElasticsearchAggregator; elasticsearchAggregatorInterface != nil {
		workers.ElasticsearchAggregation = elasticsearchAggregatorInterface.MakeWorker()
	}

	return workers
}

func (workers *Workers) Start() *Workers {
	mlog.Info("Starting workers")

	workers.startOnce.Do(func() {
		if workers.DataRetention != nil && (*workers.ConfigService.Config().DataRetentionSettings.EnableMessageDeletion || *workers.ConfigService.Config().DataRetentionSettings.EnableFileDeletion) {
			go workers.DataRetention.Run()
		}

		if workers.ElasticsearchIndexing != nil && *workers.ConfigService.Config().ElasticsearchSettings.EnableIndexing {
			go workers.ElasticsearchIndexing.Run()
		}

		if workers.ElasticsearchAggregation != nil && *workers.ConfigService.Config().ElasticsearchSettings.EnableIndexing {
			go workers.ElasticsearchAggregation.Run()
		}

		if workers.Migrations != nil {
			go workers.Migrations.Run()
		}

		if workers.Plugins != nil {
			go workers.Plugins.Run()
		}

		go workers.Watcher.Start()
	})

	workers.listenerId = workers.ConfigService.AddConfigListener(workers.handleConfigChange)

	return workers
}

func (workers *Workers) handleConfigChange(oldConfig *model.Config, newConfig *model.Config) {
	mlog.Debug("Workers received config change.")

	if workers.DataRetention != nil {
		if (!*oldConfig.DataRetentionSettings.EnableMessageDeletion && !*oldConfig.DataRetentionSettings.EnableFileDeletion) && (*newConfig.DataRetentionSettings.EnableMessageDeletion || *newConfig.DataRetentionSettings.EnableFileDeletion) {
			go workers.DataRetention.Run()
		} else if (*oldConfig.DataRetentionSettings.EnableMessageDeletion || *oldConfig.DataRetentionSettings.EnableFileDeletion) && (!*newConfig.DataRetentionSettings.EnableMessageDeletion && !*newConfig.DataRetentionSettings.EnableFileDeletion) {
			workers.DataRetention.Stop()
		}
	}

	if workers.ElasticsearchIndexing != nil {
		if !*oldConfig.ElasticsearchSettings.EnableIndexing && *newConfig.ElasticsearchSettings.EnableIndexing {
			go workers.ElasticsearchIndexing.Run()
		} else if *oldConfig.ElasticsearchSettings.EnableIndexing && !*newConfig.ElasticsearchSettings.EnableIndexing {
			workers.ElasticsearchIndexing.Stop()
		}
	}

	if workers.ElasticsearchAggregation != nil {
		if !*oldConfig.ElasticsearchSettings.EnableIndexing && *newConfig.ElasticsearchSettings.EnableIndexing {
			go workers.ElasticsearchAggregation.Run()
		} else if *oldConfig.ElasticsearchSettings.EnableIndexing && !*newConfig.ElasticsearchSettings.EnableIndexing {
			workers.ElasticsearchAggregation.Stop()
		}
	}

}

func (workers *Workers) Stop() *Workers {
	workers.ConfigService.RemoveConfigListener(workers.listenerId)

	workers.Watcher.Stop()

	if workers.DataRetention != nil && (*workers.ConfigService.Config().DataRetentionSettings.EnableMessageDeletion || *workers.ConfigService.Config().DataRetentionSettings.EnableFileDeletion) {
		workers.DataRetention.Stop()
	}

	if workers.ElasticsearchIndexing != nil && *workers.ConfigService.Config().ElasticsearchSettings.EnableIndexing {
		workers.ElasticsearchIndexing.Stop()
	}

	if workers.ElasticsearchAggregation != nil && *workers.ConfigService.Config().ElasticsearchSettings.EnableIndexing {
		workers.ElasticsearchAggregation.Stop()
	}

	if workers.Migrations != nil {
		workers.Migrations.Stop()
	}

	if workers.Plugins != nil {
		workers.Plugins.Stop()
	}

	mlog.Info("Stopped workers")

	return workers
}
