package store

import (
	"context"

	"im/einterfaces"
	"im/mlog"
	"im/model"
)

const (
	ENABLE_EXPERIMENTAL_REDIS = false
)

type LayeredStoreDatabaseLayer interface {
	LayeredStoreSupplier
	Store
}

type LayeredStore struct {
	TmpContext context.Context

	RoleStore       RoleStore
	SchemeStore     SchemeStore
	DatabaseLayer   LayeredStoreDatabaseLayer
	LocalCacheLayer *LocalCacheSupplier
	RedisLayer      *RedisSupplier
	LayerChainHead  LayeredStoreSupplier
}

func NewLayeredStore(db LayeredStoreDatabaseLayer, cluster einterfaces.ClusterInterface) Store {
	store := &LayeredStore{
		TmpContext:      context.TODO(),
		DatabaseLayer:   db,
		LocalCacheLayer: NewLocalCacheSupplier(cluster),
	}

	store.RoleStore = &LayeredRoleStore{store}
	store.SchemeStore = &LayeredSchemeStore{store}

	// Setup the chain
	if ENABLE_EXPERIMENTAL_REDIS {
		mlog.Debug("Experimental redis enabled.")
		store.RedisLayer = NewRedisSupplier()
		store.RedisLayer.SetChainNext(store.DatabaseLayer)
		store.LayerChainHead = store.RedisLayer
	} else {
		store.LocalCacheLayer.SetChainNext(store.DatabaseLayer)
		store.LayerChainHead = store.LocalCacheLayer
	}

	return store
}

type QueryFunction func(LayeredStoreSupplier) *LayeredStoreSupplierResult

func (s *LayeredStore) RunQuery(queryFunction QueryFunction) StoreChannel {
	storeChannel := make(StoreChannel)

	go func() {
		result := queryFunction(s.LayerChainHead)
		storeChannel <- result.StoreResult
	}()

	return storeChannel
}

func (s *LayeredStore) Team() TeamStore {
	return s.DatabaseLayer.Team()
}

func (s *LayeredStore) Channel() ChannelStore {
	return s.DatabaseLayer.Channel()
}

func (s *LayeredStore) Post() PostStore {
	return s.DatabaseLayer.Post()
}

func (s *LayeredStore) User() UserStore {
	return s.DatabaseLayer.User()
}

func (s *LayeredStore) Audit() AuditStore {
	return s.DatabaseLayer.Audit()
}

func (s *LayeredStore) ClusterDiscovery() ClusterDiscoveryStore {
	return s.DatabaseLayer.ClusterDiscovery()
}

func (s *LayeredStore) Session() SessionStore {
	return s.DatabaseLayer.Session()
}

func (s *LayeredStore) OAuth() OAuthStore {
	return s.DatabaseLayer.OAuth()
}

func (s *LayeredStore) System() SystemStore {
	return s.DatabaseLayer.System()
}

func (s *LayeredStore) Preference() PreferenceStore {
	return s.DatabaseLayer.Preference()
}

func (s *LayeredStore) Token() TokenStore {
	return s.DatabaseLayer.Token()
}
func (s *LayeredStore) Section() SectionStore {
	return s.DatabaseLayer.Section()
}

func (s *LayeredStore) Status() StatusStore {
	return s.DatabaseLayer.Status()
}

func (s *LayeredStore) FileInfo() FileInfoStore {
	return s.DatabaseLayer.FileInfo()
}

func (s *LayeredStore) Job() JobStore {
	return s.DatabaseLayer.Job()
}

func (s *LayeredStore) UserAccessToken() UserAccessTokenStore {
	return s.DatabaseLayer.UserAccessToken()
}

func (s *LayeredStore) ChannelMemberHistory() ChannelMemberHistoryStore {
	return s.DatabaseLayer.ChannelMemberHistory()
}

func (s *LayeredStore) Role() RoleStore {
	return s.RoleStore
}

func (s *LayeredStore) Scheme() SchemeStore {
	return s.SchemeStore
}

func (s *LayeredStore) Product() ProductStore {
	return s.DatabaseLayer.Product()
}
func (s *LayeredStore) Promo() PromoStore {
	return s.DatabaseLayer.Promo()
}
func (s *LayeredStore) Category() CategoryStore {
	return s.DatabaseLayer.Category()
}
func (s *LayeredStore) LinkMetadata() LinkMetadataStore {
	return s.DatabaseLayer.LinkMetadata()
}

func (s *LayeredStore) MarkSystemRanUnitTests() {
	s.DatabaseLayer.MarkSystemRanUnitTests()
}
func (s *LayeredStore) Office() OfficeStore {
	return s.DatabaseLayer.Office()
}
func (s *LayeredStore) Application() ApplicationStore {
	return s.DatabaseLayer.Application()
}
func (s *LayeredStore) Order() OrderStore {
	return s.DatabaseLayer.Order()
}
func (s *LayeredStore) Basket() BasketStore {
	return s.DatabaseLayer.Basket()
}
func (s *LayeredStore) Transaction() TransactionStore {
	return s.DatabaseLayer.Transaction()
}
func (s *LayeredStore) Level() LevelStore {
	return s.DatabaseLayer.Level()
}

func (s *LayeredStore) Extra() ExtraStore {
	return s.DatabaseLayer.Extra()
}

func (s *LayeredStore) Close() {
	s.DatabaseLayer.Close()
}

func (s *LayeredStore) LockToMaster() {
	s.DatabaseLayer.LockToMaster()
}

func (s *LayeredStore) UnlockFromMaster() {
	s.DatabaseLayer.UnlockFromMaster()
}

func (s *LayeredStore) DropAllTables() {
	defer s.LocalCacheLayer.Invalidate()
	s.DatabaseLayer.DropAllTables()
}

func (s *LayeredStore) TotalMasterDbConnections() int {
	return s.DatabaseLayer.TotalMasterDbConnections()
}

func (s *LayeredStore) TotalReadDbConnections() int {
	return s.DatabaseLayer.TotalReadDbConnections()
}

func (s *LayeredStore) TotalSearchDbConnections() int {
	return s.DatabaseLayer.TotalSearchDbConnections()
}

type LayeredRoleStore struct {
	*LayeredStore
}

func (s *LayeredRoleStore) Save(role *model.Role) StoreChannel {
	return s.RunQuery(func(supplier LayeredStoreSupplier) *LayeredStoreSupplierResult {
		return supplier.RoleSave(s.TmpContext, role)
	})
}

func (s *LayeredRoleStore) Get(roleId string) StoreChannel {
	return s.RunQuery(func(supplier LayeredStoreSupplier) *LayeredStoreSupplierResult {
		return supplier.RoleGet(s.TmpContext, roleId)
	})
}

func (s *LayeredRoleStore) GetAll() StoreChannel {
	return s.RunQuery(func(supplier LayeredStoreSupplier) *LayeredStoreSupplierResult {
		return supplier.RoleGetAll(s.TmpContext)
	})
}

func (s *LayeredRoleStore) GetByName(name string) StoreChannel {
	return s.RunQuery(func(supplier LayeredStoreSupplier) *LayeredStoreSupplierResult {
		return supplier.RoleGetByName(s.TmpContext, name)
	})
}

func (s *LayeredRoleStore) GetByNames(names []string) StoreChannel {
	return s.RunQuery(func(supplier LayeredStoreSupplier) *LayeredStoreSupplierResult {
		return supplier.RoleGetByNames(s.TmpContext, names)
	})
}

func (s *LayeredRoleStore) Delete(roldId string) StoreChannel {
	return s.RunQuery(func(supplier LayeredStoreSupplier) *LayeredStoreSupplierResult {
		return supplier.RoleDelete(s.TmpContext, roldId)
	})
}

func (s *LayeredRoleStore) PermanentDeleteAll() StoreChannel {
	return s.RunQuery(func(supplier LayeredStoreSupplier) *LayeredStoreSupplierResult {
		return supplier.RolePermanentDeleteAll(s.TmpContext)
	})
}

type LayeredSchemeStore struct {
	*LayeredStore
}

func (s *LayeredSchemeStore) Save(scheme *model.Scheme) StoreChannel {
	return s.RunQuery(func(supplier LayeredStoreSupplier) *LayeredStoreSupplierResult {
		return supplier.SchemeSave(s.TmpContext, scheme)
	})
}

func (s *LayeredSchemeStore) Get(schemeId string) StoreChannel {
	return s.RunQuery(func(supplier LayeredStoreSupplier) *LayeredStoreSupplierResult {
		return supplier.SchemeGet(s.TmpContext, schemeId)
	})
}

func (s *LayeredSchemeStore) GetByName(schemeName string) StoreChannel {
	return s.RunQuery(func(supplier LayeredStoreSupplier) *LayeredStoreSupplierResult {
		return supplier.SchemeGetByName(s.TmpContext, schemeName)
	})
}

func (s *LayeredSchemeStore) Delete(schemeId string) StoreChannel {
	return s.RunQuery(func(supplier LayeredStoreSupplier) *LayeredStoreSupplierResult {
		return supplier.SchemeDelete(s.TmpContext, schemeId)
	})
}

func (s *LayeredSchemeStore) GetAllPage(scope string, offset int, limit int) StoreChannel {
	return s.RunQuery(func(supplier LayeredStoreSupplier) *LayeredStoreSupplierResult {
		return supplier.SchemeGetAllPage(s.TmpContext, scope, offset, limit)
	})
}

func (s *LayeredSchemeStore) PermanentDeleteAll() StoreChannel {
	return s.RunQuery(func(supplier LayeredStoreSupplier) *LayeredStoreSupplierResult {
		return supplier.SchemePermanentDeleteAll(s.TmpContext)
	})
}
