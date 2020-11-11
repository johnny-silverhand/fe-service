package store

import (
	"context"

	"im/einterfaces"
	"im/model"
	"im/utils"
)

const (
	REACTION_CACHE_SIZE = 20000
	REACTION_CACHE_SEC  = 30 * 60

	ROLE_CACHE_SIZE = 20000
	ROLE_CACHE_SEC  = 30 * 60

	SCHEME_CACHE_SIZE = 20000
	SCHEME_CACHE_SEC  = 30 * 60

	GROUP_CACHE_SIZE = 20000
	GROUP_CACHE_SEC  = 30 * 60

	CLEAR_CACHE_MESSAGE_DATA = ""
)

type LocalCacheSupplier struct {
	next        LayeredStoreSupplier
	roleCache   *utils.Cache
	schemeCache *utils.Cache
	cluster     einterfaces.ClusterInterface
}

// Caching Interface
type ObjectCache interface {
	AddWithExpiresInSecs(key, value interface{}, expireAtSecs int64)
	AddWithDefaultExpires(key, value interface{})
	Purge()
	Get(key interface{}) (value interface{}, ok bool)
	Remove(key interface{})
	Len() int
	Name() string
	GetInvalidateClusterEvent() string
}

func NewLocalCacheSupplier(cluster einterfaces.ClusterInterface) *LocalCacheSupplier {
	supplier := &LocalCacheSupplier{
		roleCache:   utils.NewLruWithParams(ROLE_CACHE_SIZE, "Role", ROLE_CACHE_SEC, model.CLUSTER_EVENT_INVALIDATE_CACHE_FOR_ROLES),
		schemeCache: utils.NewLruWithParams(SCHEME_CACHE_SIZE, "Scheme", SCHEME_CACHE_SEC, model.CLUSTER_EVENT_INVALIDATE_CACHE_FOR_SCHEMES),

		cluster: cluster,
	}

	if cluster != nil {
		cluster.RegisterClusterMessageHandler(model.CLUSTER_EVENT_INVALIDATE_CACHE_FOR_ROLES, supplier.handleClusterInvalidateRole)

	}

	return supplier
}

func (s *LocalCacheSupplier) SetChainNext(next LayeredStoreSupplier) {
	s.next = next
}

func (s *LocalCacheSupplier) Next() LayeredStoreSupplier {
	return s.next
}

func (s *LocalCacheSupplier) doStandardReadCache(ctx context.Context, cache ObjectCache, key string, hints ...LayeredStoreHint) *LayeredStoreSupplierResult {
	if hintsContains(hints, LSH_NO_CACHE) {

		return nil
	}

	if cacheItem, ok := cache.Get(key); ok {

		result := NewSupplierResult()
		result.Data = cacheItem
		return result
	}

	return nil
}

func (s *LocalCacheSupplier) doStandardAddToCache(ctx context.Context, cache ObjectCache, key string, result *LayeredStoreSupplierResult, hints ...LayeredStoreHint) {
	if result.Err == nil && result.Data != nil {
		cache.AddWithDefaultExpires(key, result.Data)
	}
}

func (s *LocalCacheSupplier) doInvalidateCacheCluster(cache ObjectCache, key string) {
	cache.Remove(key)
	if s.cluster != nil {
		msg := &model.ClusterMessage{
			Event:    cache.GetInvalidateClusterEvent(),
			SendType: model.CLUSTER_SEND_BEST_EFFORT,
			Data:     key,
		}
		s.cluster.SendClusterMessage(msg)
	}
}

func (s *LocalCacheSupplier) doClearCacheCluster(cache ObjectCache) {
	cache.Purge()
	if s.cluster != nil {
		msg := &model.ClusterMessage{
			Event:    cache.GetInvalidateClusterEvent(),
			SendType: model.CLUSTER_SEND_BEST_EFFORT,
			Data:     CLEAR_CACHE_MESSAGE_DATA,
		}
		s.cluster.SendClusterMessage(msg)
	}
}

func (s *LocalCacheSupplier) Invalidate() {
	s.doClearCacheCluster(s.roleCache)
	s.doClearCacheCluster(s.schemeCache)
}
