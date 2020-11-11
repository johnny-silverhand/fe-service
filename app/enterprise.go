package app

import (
	"im/einterfaces"
	ejobs "im/einterfaces/jobs"
)

var clusterInterface func(*Server) einterfaces.ClusterInterface

func RegisterClusterInterface(f func(*Server) einterfaces.ClusterInterface) {
	clusterInterface = f
}

var elasticsearchInterface func(*App) einterfaces.ElasticsearchInterface

func RegisterElasticsearchInterface(f func(*App) einterfaces.ElasticsearchInterface) {
	elasticsearchInterface = f
}

var jobsElasticsearchAggregatorInterface func(*App) ejobs.ElasticsearchAggregatorInterface

func RegisterJobsElasticsearchAggregatorInterface(f func(*App) ejobs.ElasticsearchAggregatorInterface) {
	jobsElasticsearchAggregatorInterface = f
}

var jobsElasticsearchIndexerInterface func(*App) ejobs.ElasticsearchIndexerInterface

func RegisterJobsElasticsearchIndexerInterface(f func(*App) ejobs.ElasticsearchIndexerInterface) {
	jobsElasticsearchIndexerInterface = f
}

func (s *Server) initEnterprise() {

	if elasticsearchInterface != nil {
		s.Elasticsearch = elasticsearchInterface(s.FakeApp())
	}

	if clusterInterface != nil {
		s.Cluster = clusterInterface(s)
	}
}
