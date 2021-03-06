package store

import (
	"time"

	"im/model"
)

type StoreResult struct {
	Data interface{}
	Err  *model.AppError
}

type StoreChannel chan StoreResult

func Do(f func(result *StoreResult)) StoreChannel {
	storeChannel := make(StoreChannel, 1)
	go func() {
		result := StoreResult{}
		f(&result)
		storeChannel <- result
		close(storeChannel)
	}()
	return storeChannel
}

func Must(sc StoreChannel) interface{} {
	r := <-sc
	if r.Err != nil {

		time.Sleep(time.Second)
		panic(r.Err)
	}

	return r.Data
}

type Store interface {
	Team() TeamStore
	Channel() ChannelStore
	Post() PostStore
	User() UserStore
	Audit() AuditStore
	ClusterDiscovery() ClusterDiscoveryStore

	Session() SessionStore
	OAuth() OAuthStore
	System() SystemStore

	Preference() PreferenceStore

	Token() TokenStore
	Section() SectionStore
	Status() StatusStore
	FileInfo() FileInfoStore

	Role() RoleStore
	Scheme() SchemeStore
	Job() JobStore
	UserAccessToken() UserAccessTokenStore
	ChannelMemberHistory() ChannelMemberHistoryStore

	LinkMetadata() LinkMetadataStore
	MarkSystemRanUnitTests()
	Close()
	LockToMaster()
	UnlockFromMaster()
	DropAllTables()
	TotalMasterDbConnections() int
	TotalReadDbConnections() int
	TotalSearchDbConnections() int

	Product() ProductStore
	Category() CategoryStore
	Promo() PromoStore
	Office() OfficeStore
	Application() ApplicationStore
	Order() OrderStore
	Basket() BasketStore
	Transaction() TransactionStore
	Level() LevelStore
	Extra() ExtraStore

	ProductOffice() ProductOfficeStore
}

type TeamStore interface {
	Save(team *model.Team) StoreChannel
	Update(team *model.Team) StoreChannel
	UpdateDisplayName(name string, teamId string) StoreChannel
	Get(id string) StoreChannel
	GetByName(name string) StoreChannel
	SearchByName(name string) StoreChannel
	SearchAll(term string) StoreChannel
	SearchOpen(term string) StoreChannel
	SearchPrivate(term string) StoreChannel
	GetAll() StoreChannel
	GetAllPage(offset int, limit int) StoreChannel
	GetAllPrivateTeamListing() StoreChannel
	GetAllPrivateTeamPageListing(offset int, limit int) StoreChannel
	GetAllTeamListing() StoreChannel
	GetAllTeamPageListing(offset int, limit int) StoreChannel
	GetTeamsByUserId(userId string) StoreChannel
	GetByInviteId(inviteId string) StoreChannel
	PermanentDelete(teamId string) StoreChannel
	AnalyticsTeamCount() StoreChannel
	SaveMember(member *model.TeamMember, maxUsersPerTeam int) StoreChannel
	UpdateMember(member *model.TeamMember) StoreChannel
	GetMember(teamId string, userId string) StoreChannel
	GetMembers(teamId string, offset int, limit int) StoreChannel
	GetMembersByIds(teamId string, userIds []string) StoreChannel
	GetTotalMemberCount(teamId string) StoreChannel
	GetActiveMemberCount(teamId string) StoreChannel
	GetTeamsForUser(userId string) StoreChannel
	GetTeamsForUserWithPagination(userId string, page, perPage int) StoreChannel
	GetChannelUnreadsForAllTeams(excludeTeamId, userId string) StoreChannel
	GetChannelUnreadsForTeam(teamId, userId string) StoreChannel
	RemoveMember(teamId string, userId string) StoreChannel
	RemoveAllMembersByTeam(teamId string) StoreChannel
	RemoveAllMembersByUser(userId string) StoreChannel
	UpdateLastTeamIconUpdate(teamId string, curTime int64) StoreChannel
	GetTeamsByScheme(schemeId string, offset int, limit int) StoreChannel
	MigrateTeamMembers(fromTeamId string, fromUserId string) StoreChannel
	ResetAllTeamSchemes() StoreChannel
	ClearAllCustomRoleAssignments() StoreChannel
	AnalyticsGetTeamCountForScheme(schemeId string) StoreChannel
	GetAllForExportAfter(limit int, afterId string) StoreChannel
	GetTeamMembersForExport(userId string) StoreChannel
}

type ChannelStore interface {
	Save(channel *model.Channel, maxChannelsPerTeam int64) StoreChannel
	CreateDirectChannel(userId string, otherUserId string) StoreChannel
	SaveDirectChannel(channel *model.Channel, member1 *model.ChannelMember, member2 *model.ChannelMember) StoreChannel
	Update(channel *model.Channel) StoreChannel
	Get(id string, allowFromCache bool) StoreChannel
	InvalidateChannel(id string)
	InvalidateChannelByName(teamId, name string)
	GetFromMaster(id string) StoreChannel
	Delete(channelId string, time int64) StoreChannel
	Restore(channelId string, time int64) StoreChannel
	SetDeleteAt(channelId string, deleteAt int64, updateAt int64) StoreChannel
	PermanentDeleteByTeam(teamId string) StoreChannel
	PermanentDelete(channelId string) StoreChannel
	GetByName(team_id string, name string, allowFromCache bool) StoreChannel
	GetByNames(team_id string, names []string, allowFromCache bool) StoreChannel
	GetByNameIncludeDeleted(team_id string, name string, allowFromCache bool) StoreChannel
	GetDeletedByName(team_id string, name string) StoreChannel
	GetDeleted(team_id string, offset int, limit int) StoreChannel
	GetDeletedForUser(userId string, offset int, limit int) StoreChannel
	GetChannels(teamId string, userId string, includeDeleted bool) StoreChannel
	GetChannelsForUser(userId string, includeDeleted bool) StoreChannel
	GetAllChannels(page, perPage int, includeDeleted bool) StoreChannel
	GetMoreChannels(teamId string, userId string, offset int, limit int) StoreChannel
	GetPublicChannelsForTeam(teamId string, offset int, limit int) StoreChannel
	GetPublicChannelsByIdsForTeam(teamId string, channelIds []string) StoreChannel
	GetChannelCounts(teamId string, userId string) StoreChannel
	GetTeamChannels(teamId string) StoreChannel
	GetAll(teamId string) StoreChannel
	GetChannelsByIds(channelIds []string) StoreChannel
	GetForPost(postId string) StoreChannel
	SaveMember(member *model.ChannelMember) StoreChannel
	UpdateMember(member *model.ChannelMember) StoreChannel
	GetMembers(channelId string, offset, limit int) StoreChannel
	GetMember(channelId string, userId string) (*model.ChannelMember, *model.AppError)
	GetChannelMembersTimezones(channelId string) StoreChannel
	GetAllChannelMembersForUser(userId string, allowFromCache bool, includeDeleted bool) StoreChannel
	InvalidateAllChannelMembersForUser(userId string)
	IsUserInChannelUseCache(userId string, channelId string) bool
	GetAllChannelMembersNotifyPropsForChannel(channelId string, allowFromCache bool) StoreChannel
	InvalidateCacheForChannelMembersNotifyProps(channelId string)
	GetMemberForPost(postId string, userId string) StoreChannel
	InvalidateMemberCount(channelId string)
	GetMemberCountFromCache(channelId string) int64
	GetMemberCount(channelId string, allowFromCache bool) StoreChannel
	GetPinnedPosts(channelId string) StoreChannel
	RemoveMember(channelId string, userId string) StoreChannel
	PermanentDeleteMembersByUser(userId string) StoreChannel
	PermanentDeleteMembersByChannel(channelId string) StoreChannel
	UpdateLastViewedAt(channelIds []string, userId string) StoreChannel
	IncrementMentionCount(channelId string, userId string) StoreChannel
	AnalyticsTypeCount(teamId string, channelType string) StoreChannel
	GetMembersForUser(teamId string, userId string) StoreChannel
	GetAllMembersForUser(userId string) StoreChannel
	GetMembersForUserWithPagination(teamId, userId string, page, perPage int) StoreChannel
	AutocompleteInTeam(teamId string, term string, includeDeleted bool) StoreChannel
	AutocompleteInTeamForSearch(teamId string, userId string, term string, includeDeleted bool) StoreChannel
	SearchAllChannels(term string, includeDeleted bool) StoreChannel
	SearchInTeam(teamId string, term string, includeDeleted bool) StoreChannel
	SearchMore(userId string, teamId string, term string) StoreChannel
	GetMembersByIds(channelId string, userIds []string) StoreChannel
	AnalyticsDeletedTypeCount(teamId string, channelType string) StoreChannel
	GetChannelUnread(channelId, userId string) StoreChannel
	ClearCaches()
	GetChannelsByScheme(schemeId string, offset int, limit int) StoreChannel
	MigrateChannelMembers(fromChannelId string, fromUserId string) StoreChannel
	ResetAllChannelSchemes() StoreChannel
	ClearAllCustomRoleAssignments() StoreChannel
	MigratePublicChannels() error
	GetAllChannelsForExportAfter(limit int, afterId string) StoreChannel
	GetAllDirectChannelsForExportAfter(limit int, afterId string) StoreChannel
	GetChannelMembersForExport(userId string, teamId string) StoreChannel
	RemoveAllDeactivatedMembers(channelId string) StoreChannel
	GetChannelsBatchForIndexing(startTime, endTime int64, limit int) StoreChannel

	FindOpennedChannel(userId string) StoreChannel
	CreateUnresolvedChannel(user *model.User, additionalMemeber []string, nn int64) StoreChannel

	GetDeferredChannelForUser(userId string) StoreChannel
	CreateDeferredChannel(user *model.User, channelMemberIds []string) StoreChannel

	GetByOrderId(orderId string) StoreChannel
}

type ChannelMemberHistoryStore interface {
	LogJoinEvent(userId string, channelId string, joinTime int64) StoreChannel
	LogLeaveEvent(userId string, channelId string, leaveTime int64) StoreChannel
	GetUsersInChannelDuring(startTime int64, endTime int64, channelId string) StoreChannel
	PermanentDeleteBatch(endTime int64, limit int64) StoreChannel
}

type PostStore interface {
	Save(post *model.Post) StoreChannel
	Update(newPost *model.Post, oldPost *model.Post) StoreChannel
	Get(id string) StoreChannel
	GetSingle(id string) StoreChannel
	Delete(postId string, time int64, deleteByID string) StoreChannel
	PermanentDeleteByUser(userId string) StoreChannel
	PermanentDeleteByChannel(channelId string) StoreChannel
	PermanentDelete(postId string) StoreChannel
	GetPosts(channelId string, offset int, limit int, allowFromCache bool) StoreChannel
	GetFlaggedPosts(userId string, offset int, limit int) StoreChannel
	GetFlaggedPostsForTeam(userId, teamId string, offset int, limit int) StoreChannel
	GetFlaggedPostsForChannel(userId, channelId string, offset int, limit int) StoreChannel
	GetDeferredPosts(userId string, offset int, limit int) StoreChannel
	GetDeferredPostsForTeam(userId, teamId string, offset int, limit int) StoreChannel
	GetDeferredPostsForChannel(userId, channelId string, offset int, limit int) StoreChannel
	GetPostsBefore(channelId string, postId string, numPosts int, offset int) StoreChannel
	GetPostsAfter(channelId string, postId string, numPosts int, offset int) StoreChannel
	GetPostsSince(channelId string, time int64, allowFromCache bool) StoreChannel
	GetEtag(channelId string, allowFromCache bool) StoreChannel
	Search(teamId string, userId string, params *model.SearchParams) StoreChannel
	AnalyticsUserCountsWithPostsByDay(teamId string) StoreChannel
	AnalyticsPostCountsByDay(teamId string) StoreChannel
	AnalyticsPostCount(teamId string, mustHaveFile bool, mustHaveHashtag bool) StoreChannel
	ClearCaches()
	InvalidateLastPostTimeCache(channelId string)
	GetPostsCreatedAt(channelId string, time int64) StoreChannel
	Overwrite(post *model.Post) StoreChannel
	GetPostsByIds(postIds []string) StoreChannel
	GetPostsBatchForIndexing(startTime int64, endTime int64, limit int) StoreChannel
	PermanentDeleteBatch(endTime int64, limit int64) StoreChannel
	GetOldest() StoreChannel
	GetMaxPostSize() StoreChannel
	GetParentsForExportAfter(limit int, afterId string) StoreChannel
	GetRepliesForExport(parentId string) StoreChannel
	GetDirectPostParentsForExportAfter(limit int, afterId string) StoreChannel

	GetAllMessages(userId string, offset int, limit int, allowFromCache bool, limitMin int64) StoreChannel
	GetAllMessagesBefore(userId string, postId string, numMessages int, offset int, limitMin int64) StoreChannel
	GetAllMessagesAfter(userId string, postId string, numMessages int, offset int, limitMin int64) StoreChannel
	GetAllMessagesSince(userId string, time int64, allowFromCache bool, limitMin int64) StoreChannel

	FindPostWithOrder(orderId string) StoreChannel
}

type UserStore interface {
	Save(user *model.User) StoreChannel
	Update(user *model.User, allowRoleUpdate bool) StoreChannel
	UpdateLastPictureUpdate(userId string) StoreChannel
	ResetLastPictureUpdate(userId string) StoreChannel
	UpdateUpdateAt(userId string) StoreChannel
	UpdatePassword(userId, newPassword string) StoreChannel
	UpdateAuthData(userId string, service string, authData *string, email string, resetMfa bool) StoreChannel
	UpdateMfaSecret(userId, secret string) StoreChannel
	UpdateMfaActive(userId string, active bool) StoreChannel
	Get(id string) (*model.User, *model.AppError)
	GetAll() StoreChannel
	ClearCaches()
	InvalidateProfilesInChannelCacheByUser(userId string)
	InvalidateProfilesInChannelCache(channelId string)
	GetProfilesInChannel(channelId string, offset int, limit int) StoreChannel
	GetProfilesInChannelByStatus(channelId string, offset int, limit int) StoreChannel
	GetAllProfilesInChannel(channelId string, allowFromCache bool) StoreChannel
	GetProfilesNotInChannel(teamId string, channelId string, offset int, limit int) StoreChannel
	GetProfilesWithoutTeam(offset int, limit int) StoreChannel
	GetProfilesByUsernames(usernames []string, teamId string) StoreChannel
	GetAllProfiles(options *model.UserGetOptions) StoreChannel
	GetProfiles(options *model.UserGetOptions) StoreChannel
	GetProfileByIds(userId []string, allowFromCache bool) StoreChannel
	InvalidatProfileCacheForUser(userId string)
	GetByEmail(email string) StoreChannel
	GetByEmailAppId(email string, appId string) StoreChannel
	GetByAuth(authData *string, authService string) StoreChannel
	GetAllUsingAuthService(authService string) StoreChannel
	GetByUsername(username string) StoreChannel
	GetForLogin(loginId string, allowSignInWithUsername, allowSignInWithEmail bool) StoreChannel
	VerifyEmail(userId, email string) StoreChannel
	GetEtagForAllProfiles() StoreChannel
	GetEtagForProfiles(teamId string) StoreChannel
	UpdateFailedPasswordAttempts(userId string, attempts int) StoreChannel
	GetSystemAdminProfiles() StoreChannel
	PermanentDelete(userId string) StoreChannel
	AnalyticsActiveCount(time int64) StoreChannel
	GetUnreadCount(userId string) StoreChannel
	GetUnreadCountForChannel(userId string, channelId string) StoreChannel
	GetAnyUnreadPostCountForChannel(userId string, channelId string) StoreChannel
	GetRecentlyActiveUsersForTeam(teamId string, offset, limit int) StoreChannel
	GetNewUsersForTeam(teamId string, offset, limit int) StoreChannel
	Search(appId string, term string, options *model.UserSearchOptions) StoreChannel
	SearchNotInTeam(notInTeamId string, term string, options *model.UserSearchOptions) StoreChannel
	SearchInChannel(channelId string, term string, options *model.UserSearchOptions) StoreChannel
	SearchNotInChannel(teamId string, channelId string, term string, options *model.UserSearchOptions) StoreChannel
	SearchWithoutTeam(term string, options *model.UserSearchOptions) StoreChannel
	AnalyticsGetInactiveUsersCount() StoreChannel
	AnalyticsGetSystemAdminCount() StoreChannel
	GetProfilesNotInTeam(teamId string, offset int, limit int) StoreChannel
	GetEtagForProfilesNotInTeam(teamId string) StoreChannel
	ClearAllCustomRoleAssignments() StoreChannel
	InferSystemInstallDate() StoreChannel
	GetAllAfter(limit int, afterId string) StoreChannel
	GetUsersBatchForIndexing(startTime, endTime int64, limit int) StoreChannel
	Count(options model.UserCountOptions) StoreChannel
	GetTeamGroupUsers(teamID string) StoreChannel
	GetChannelGroupUsers(channelID string) StoreChannel
	SaveWithCheckExist(user *model.User) StoreChannel
	VerifyPhone(userId string) StoreChannel
	VerifyPhoneNew(userId string) StoreChannel

	GetByPhone(phone string) StoreChannel
	GetByPhoneApp(phone string, appId string) StoreChannel
	AccrualBalance(userId string, value float64) StoreChannel
	DeductionBalance(userId string, value float64) StoreChannel
	GetInvitedUsers(userId string) StoreChannel

	GetMetricsForRegister(appId string, beginAt int64, expireAt int64) StoreChannel
	GetMetricsForRating(options model.UserGetOptions) StoreChannel
	GetMetricsForBonuses(options model.UserGetOptions) StoreChannel

	GetByInviteId(inviteId string) StoreChannel
}

type SessionStore interface {
	Save(session *model.Session) StoreChannel
	Get(sessionIdOrToken string) StoreChannel
	GetSessions(userId string) StoreChannel
	GetSessionsWithActiveDeviceIds(userId string) StoreChannel
	Remove(sessionIdOrToken string) StoreChannel
	RemoveAllSessions() StoreChannel
	PermanentDeleteSessionsByUser(teamId string) StoreChannel
	UpdateLastActivityAt(sessionId string, time int64) StoreChannel
	UpdateRoles(userId string, roles string) StoreChannel
	UpdateDeviceId(id string, deviceId string, expiresAt int64) StoreChannel
	UpdateOfficeId(id string, officeId string, expiresAt int64) StoreChannel
	UpdateAppId(id string, appId string, expiresAt int64) StoreChannel
	AnalyticsSessionCount() StoreChannel
	Cleanup(expiryTime int64, batchSize int64)
	GetSessionByProps(props string, appId string) StoreChannel
}

type AuditStore interface {
	Save(audit *model.Audit) *model.AppError
	Get(user_id string, offset int, limit int) (model.Audits, *model.AppError)
	PermanentDeleteByUser(userId string) *model.AppError
	PermanentDeleteBatch(endTime int64, limit int64) (int64, *model.AppError)
}

type ClusterDiscoveryStore interface {
	Save(discovery *model.ClusterDiscovery) StoreChannel
	Delete(discovery *model.ClusterDiscovery) StoreChannel
	Exists(discovery *model.ClusterDiscovery) StoreChannel
	GetAll(discoveryType, clusterName string) StoreChannel
	SetLastPingAt(discovery *model.ClusterDiscovery) StoreChannel
	Cleanup() StoreChannel
}

type OAuthStore interface {
	SaveApp(app *model.OAuthApp) StoreChannel
	UpdateApp(app *model.OAuthApp) StoreChannel
	GetApp(id string) StoreChannel
	GetAppByUser(userId string, offset, limit int) StoreChannel
	GetApps(offset, limit int) StoreChannel
	GetAuthorizedApps(userId string, offset, limit int) StoreChannel
	DeleteApp(id string) StoreChannel
	SaveAuthData(authData *model.AuthData) StoreChannel
	GetAuthData(code string) StoreChannel
	RemoveAuthData(code string) StoreChannel
	PermanentDeleteAuthDataByUser(userId string) StoreChannel
	SaveAccessData(accessData *model.AccessData) StoreChannel
	UpdateAccessData(accessData *model.AccessData) StoreChannel
	GetAccessData(token string) StoreChannel
	GetAccessDataByUserForApp(userId, appId string) StoreChannel
	GetAccessDataByRefreshToken(token string) StoreChannel
	GetPreviousAccessData(userId, appId string) StoreChannel
	RemoveAccessData(token string) StoreChannel
}

type SystemStore interface {
	Save(system *model.System) StoreChannel
	SaveOrUpdate(system *model.System) StoreChannel
	Update(system *model.System) StoreChannel
	Get() StoreChannel
	GetByName(name string) StoreChannel
	PermanentDeleteByName(name string) StoreChannel
}

type PreferenceStore interface {
	Save(preferences *model.Preferences) StoreChannel
	Get(userId string, category string, name string) StoreChannel
	GetCategory(userId string, category string) StoreChannel
	GetAll(userId string) StoreChannel
	Delete(userId, category, name string) StoreChannel
	DeleteCategory(userId string, category string) StoreChannel
	DeleteCategoryAndName(category string, name string) StoreChannel
	PermanentDeleteByUser(userId string) StoreChannel
	IsFeatureEnabled(feature, userId string) StoreChannel
	CleanupFlagsBatch(limit int64) StoreChannel
}

type TokenStore interface {
	Save(recovery *model.Token) StoreChannel
	Delete(token string) StoreChannel
	GetByToken(token string) StoreChannel
	GetByUserInviteToken(userId string) StoreChannel
	Cleanup()
	RemoveAllTokensByType(tokenType string) StoreChannel
	RemoveUserTokensByType(tokenType string, userId string) StoreChannel
	UpdateExtra(token, newExtra string) StoreChannel
	GetByApplicationInviteCode(appId string, code string) StoreChannel
}

type SectionStore interface {
	GetAll(parentId string) StoreChannel
	Insert(section *model.Section) StoreChannel
}

type StatusStore interface {
	SaveOrUpdate(status *model.Status) StoreChannel
	Get(userId string) StoreChannel
	GetByIds(userIds []string) StoreChannel
	GetOnlineAway() StoreChannel
	GetOnline() StoreChannel
	GetAllFromTeam(teamId string) StoreChannel
	ResetAll() StoreChannel
	GetTotalActiveUsersCount() StoreChannel
	UpdateLastActivityAt(userId string, lastActivityAt int64) StoreChannel
}

type FileInfoStore interface {
	Save(info *model.FileInfo) StoreChannel
	Get(id string) StoreChannel
	GetByPath(path string) StoreChannel
	GetForPost(postId string, readFromMaster bool, allowFromCache bool) StoreChannel
	GetForMetadata(metadataId string, readFromMaster bool, allowFromCache bool) StoreChannel

	GetForUser(userId string) StoreChannel
	InvalidateFileInfosForPostCache(postId string)
	AttachToPost(fileId string, postId string, creatorId string) StoreChannel
	AttachTo(fileId string, metadataId string, metadataType string) StoreChannel
	DeleteForPost(postId string) StoreChannel
	DeleteForProduct(productId string) StoreChannel
	DeleteForPromo(promoId string) StoreChannel
	PermanentDelete(fileId string) StoreChannel
	PermanentDeleteBatch(endTime int64, limit int64) StoreChannel
	PermanentDeleteByUser(userId string) StoreChannel
	ClearCaches()
	RemoveFromCache(metadataId string)
}

type ProductOfficeStore interface {
	Save(productOffice *model.ProductOffice) StoreChannel
	Get(id string) StoreChannel
	GetForProduct(productId string, readFromMaster bool, allowFromCache bool) StoreChannel
	DeleteForProduct(productId string) StoreChannel

	/*AttachToPost(productOfficeId string, postId string, creatorId string) StoreChannel
	AttachTo(productOfficeId string, metadataId string, metadataType string) StoreChannel
	DeleteForPost(postId string) StoreChannel
	PermanentDelete(productOfficeId string) StoreChannel
	PermanentDeleteBatch(endTime int64, limit int64) StoreChannel
	PermanentDeleteByUser(userId string) StoreChannel
	ClearCaches()*/
}

type JobStore interface {
	Save(job *model.Job) StoreChannel
	UpdateOptimistically(job *model.Job, currentStatus string) StoreChannel
	UpdateStatus(id string, status string) StoreChannel
	UpdateStatusOptimistically(id string, currentStatus string, newStatus string) StoreChannel
	Get(id string) StoreChannel
	GetAllPage(offset int, limit int) StoreChannel
	GetAllByType(jobType string) StoreChannel
	GetAllByTypePage(jobType string, offset int, limit int) StoreChannel
	GetAllByStatus(status string) StoreChannel
	GetNewestJobByStatusAndType(status string, jobType string) StoreChannel
	GetCountByStatusAndType(status string, jobType string) StoreChannel
	Delete(id string) StoreChannel
}

type UserAccessTokenStore interface {
	Save(token *model.UserAccessToken) StoreChannel
	Delete(tokenId string) StoreChannel
	DeleteAllForUser(userId string) StoreChannel
	Get(tokenId string) StoreChannel
	GetAll(offset int, limit int) StoreChannel
	GetByToken(tokenString string) StoreChannel
	GetByUser(userId string, page, perPage int) StoreChannel
	Search(term string) StoreChannel
	UpdateTokenEnable(tokenId string) StoreChannel
	UpdateTokenDisable(tokenId string) StoreChannel
}

type RoleStore interface {
	Save(role *model.Role) StoreChannel
	Get(roleId string) StoreChannel
	GetAll() StoreChannel
	GetByName(name string) StoreChannel
	GetByNames(names []string) StoreChannel
	Delete(roldId string) StoreChannel
	PermanentDeleteAll() StoreChannel
}

type SchemeStore interface {
	Save(scheme *model.Scheme) StoreChannel
	Get(schemeId string) StoreChannel
	GetByName(schemeName string) StoreChannel
	GetAllPage(scope string, offset int, limit int) StoreChannel
	Delete(schemeId string) StoreChannel
	PermanentDeleteAll() StoreChannel
}

type LinkMetadataStore interface {
	Save(linkMetadata *model.LinkMetadata) StoreChannel
	Get(url string, timestamp int64) StoreChannel
}

type ProductStore interface {
	Save(product *model.Product) StoreChannel
	Get(productId string) StoreChannel
	GetAllPage(offset int, limit int, options *model.ProductGetOptions) StoreChannel

	GetAllPageByApp(offset int, limit int, order model.ColumnOrder, appId string) StoreChannel
	GetAllByAppId(appId string) StoreChannel
	//GetAllByAppPage(appId string, offset int, limit int, order model.ColumnOrder, categoryId string) StoreChannel

	//Delete(productId string) StoreChannel
	Delete(productId string, time int64, deleteByID string) StoreChannel
	GetAllByCategoryId(categoryId string, offset int, limit int, allowFromCache bool) StoreChannel
	Update(newProduct *model.Product) StoreChannel
	Overwrite(product *model.Product) StoreChannel
	GetProductsByIds(productIds []string, allowFromCache bool) StoreChannel
	GetAll() StoreChannel
	/*	Publish(product *model.Product) StoreChannel
		GetExtras(product *model.Product) StoreChannel*/
	Search(categoryId, terms string, page, perPage int) StoreChannel
	GetForModeration(options *model.ProductGetOptions) StoreChannel
}

type CategoryStore interface {
	Create(category *model.Category) StoreChannel
	Move(category *model.Category) StoreChannel
	Order(category *model.Category) StoreChannel
	Update(category *model.Category) StoreChannel
	Get(categoryId string) StoreChannel
	GetAllPage(offset int, limit int, appId *string) StoreChannel
	Delete(category *model.Category) StoreChannel
	GetCategoryPath(categoryId string) StoreChannel
	GetCategoriesByIds(categoryIds []string) StoreChannel
}

type PromoStore interface {
	Save(promo *model.Promo) StoreChannel
	Get(promoId string) StoreChannel
	GetAllPage(offset int, limit int, order model.ColumnOrder) StoreChannel
	GetAllPageByApp(offset int, limit int, order model.ColumnOrder, appId string) StoreChannel
	Activate(promoId string) StoreChannel
	Deactivate(promoId string) StoreChannel
	Update(newPromo *model.Promo) StoreChannel
	Overwrite(promo *model.Promo) StoreChannel
	Delete(promoId string, time int64, deleteByID string) StoreChannel

	GetAllPromos(offset int, limit int, options *model.PromoGetOptions) StoreChannel
	GetAllPromosSince(time int64, options *model.PromoGetOptions) StoreChannel
	GetAllPromosBefore(promoId string, numPromos int, offset int, options *model.PromoGetOptions) StoreChannel
	GetAllPromosAfter(promoId string, numPromos int, offset int, options *model.PromoGetOptions) StoreChannel

	GetForModeration(options *model.PromoGetOptions) StoreChannel
}

type OfficeStore interface {
	Save(office *model.Office) StoreChannel
	Get(officeId string) StoreChannel
	GetAllPage(offset int, limit int, order model.ColumnOrder) StoreChannel
	Activate(officeId string) StoreChannel
	Deactivate(officeId string) StoreChannel
	Update(newOffice *model.Office) StoreChannel
	Overwrite(office *model.Office) StoreChannel
	Delete(officeId string, time int64, deleteByID string) StoreChannel

	GetAllOffices(offset int, limit int, allowFromCache bool, appId *string) StoreChannel
	GetAllOfficesSince(time int64, allowFromCache bool, appId *string) StoreChannel
	GetAllOfficesBefore(officeId string, numOffices int, offset int, appId *string) StoreChannel
	GetAllOfficesAfter(officeId string, numOffices int, offset int, appId *string) StoreChannel
}

type ApplicationStore interface {
	Save(application *model.Application) StoreChannel
	Get(appId string) StoreChannel
	GetAllPage(offset int, limit int, order model.ColumnOrder) StoreChannel
	Activate(appId string) StoreChannel
	Deactivate(appId string) StoreChannel
	Update(newApplication *model.Application) StoreChannel
	Overwrite(application *model.Application) StoreChannel
	Delete(appId string, time int64, deleteByID string) StoreChannel

	GetApplications(options *model.ApplicationGetOptions) StoreChannel

	GetAllApplications(offset int, limit int, allowFromCache bool) StoreChannel
	GetAllApplicationsSince(time int64, allowFromCache bool) StoreChannel
	GetAllApplicationsBefore(appId string, numApplications int, offset int) StoreChannel
	GetAllApplicationsAfter(appId string, numApplications int, offset int) StoreChannel
}

type TransactionStore interface {
	Save(transaction *model.Transaction) StoreChannel
	Get(transactionId string) StoreChannel
	GetAllPage(offset int, limit int, order model.ColumnOrder) StoreChannel
	Update(newTransaction *model.Transaction) StoreChannel
	Overwrite(transaction *model.Transaction) StoreChannel
	Delete(transactionId string, time int64, deleteByID string) StoreChannel

	GetAllTransactions(offset int, limit int, allowFromCache bool) StoreChannel
	GetAllTransactionsSince(time int64, allowFromCache bool) StoreChannel
	GetAllTransactionsBefore(transactionId string, numTransactions int, offset int) StoreChannel
	GetAllTransactionsAfter(transactionId string, numTransactions int, offset int) StoreChannel
	//GetByUserId(userId string, offset int, limit int, order model.ColumnOrder) StoreChannel
	GetByUserId(options model.TransactionGetOptions) StoreChannel

	GetBonusTransactionsForUser(orderUserId string, userId string) StoreChannel
	GetMetricsForSpy(options model.UserGetOptions, beginAt int64, expireAt int64) StoreChannel
}

type OrderStore interface {
	Save(order *model.Order) StoreChannel
	Get(orderId string) StoreChannel
	GetAllPage(offset int, limit int, order model.ColumnOrder) StoreChannel
	Cancel(orderId string) StoreChannel
	Update(newOrder *model.Order) StoreChannel
	Overwrite(order *model.Order) StoreChannel
	Delete(orderId string, time int64, deleteByID string) StoreChannel

	GetAllOrders(options model.OrderGetOptions) StoreChannel
	GetAllOrdersSince(time int64, options model.OrderGetOptions) StoreChannel
	GetAllOrdersBefore(orderId string, options model.OrderGetOptions) StoreChannel
	GetAllOrdersAfter(orderId string, options model.OrderGetOptions) StoreChannel

	GetFromMaster(id string) StoreChannel
	SaveBasket(orderId string, positions []*model.Basket) StoreChannel

	SaveWithBasket(order *model.Order) StoreChannel

	//GetByUserId(userId string, offset int, limit int, order model.ColumnOrder) StoreChannel
	GetByUserId(options model.OrderGetOptions) StoreChannel

	SetOrderPayed(orderId string) StoreChannel
	SetOrderCancel(orderId string) StoreChannel

	Count(options model.OrderCountOptions) StoreChannel

	GetMetricsForOrders(appId string, beginAt int64, expireAt int64) StoreChannel
}

type BasketStore interface {
	Save(basket *model.Basket) StoreChannel
	GetByOrderId(orderId string) StoreChannel
	GetByUserId(userId string) StoreChannel
	Update(newBasket *model.Basket) StoreChannel
	Overwrite(basket *model.Basket) StoreChannel
	Delete(basketId string, time int64, deleteByID string) StoreChannel
}

type LevelStore interface {
	Save(level *model.Level) StoreChannel
	Get(levelId string) StoreChannel
	GetAllPage(offset int, limit int, order model.ColumnOrder) StoreChannel
	Activate(levelId string) StoreChannel
	Deactivate(levelId string) StoreChannel
	Update(newLevel *model.Level) StoreChannel
	Overwrite(level *model.Level) StoreChannel
	Delete(levelId string, time int64, deleteByID string) StoreChannel

	GetAllLevels(offset int, limit int, allowFromCache bool, appId *string) StoreChannel
	GetAllLevelsSince(time int64, allowFromCache bool, appId *string) StoreChannel
	GetAllLevelsBefore(levelId string, numLevels int, offset int, appId *string) StoreChannel
	GetAllLevelsAfter(levelId string, numLevels int, offset int, appId *string) StoreChannel

	DeleteApplicationLevels(appId string) StoreChannel
}

type ExtraStore interface {
	Save(extra *model.Extra) StoreChannel
	Get(extraId string) StoreChannel
	GetAllPage(offset int, limit int, order model.ColumnOrder) StoreChannel

	Update(newExtra *model.Extra) StoreChannel
	Overwrite(extra *model.Extra) StoreChannel
	Delete(extraId string, time int64, deleteByID string) StoreChannel

	GetAllExtras(offset int, limit int, allowFromCache bool) StoreChannel
	GetAllExtrasSince(time int64, allowFromCache bool) StoreChannel
	GetAllExtrasBefore(extraId string, numExtras int, offset int) StoreChannel
	GetAllExtrasAfter(extraId string, numExtras int, offset int) StoreChannel

	GetExtraProductsByIds(productIds []string, allowFromCache bool) StoreChannel

	DeleteForProduct(productId string) StoreChannel
}
