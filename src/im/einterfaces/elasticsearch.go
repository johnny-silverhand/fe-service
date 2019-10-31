package einterfaces

import (
	"time"

	"im/model"
)

type ElasticsearchInterface interface {
	Start() *model.AppError
	Stop() *model.AppError
	IndexPost(post *model.Post, teamId string) *model.AppError
	SearchPosts(channels *model.ChannelList, searchParams []*model.SearchParams, page, perPage int) ([]string, model.PostSearchMatches, *model.AppError)
	
	SearchPostsHint( searchParams []*model.SearchParams, page, perPage int) ([]*model.Post, *model.AppError)

	DeletePost(post *model.Post) *model.AppError
	IndexChannel(channel *model.Channel) *model.AppError
	SearchChannels(teamId, term string) ([]string, *model.AppError)
	DeleteChannel(channel *model.Channel) *model.AppError
	IndexUser(user *model.User, teamsIds, channelsIds []string) *model.AppError
	SearchUsersInChannel(teamId, channelId, term string, options *model.UserSearchOptions) ([]string, []string, *model.AppError)
	SearchUsersInTeam(teamId, term string, options *model.UserSearchOptions) ([]string, *model.AppError)
	DeleteUser(user *model.User) *model.AppError
	TestConfig(cfg *model.Config) *model.AppError
	PurgeIndexes() *model.AppError
	DataRetentionDeleteIndexes(cutoff time.Time) *model.AppError
}
