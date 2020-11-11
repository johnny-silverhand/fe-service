package model

import (
	"encoding/json"
	"io"
)

const CHANNEL_SEARCH_RESULT_DEFAULT_LIMIT = 50

type ChannelSearchResult struct {
	Post
	Channel *Channel `json:"channel"`
}

type ChannelSearchResultList struct {
	Order   []string                        `json:"order"`
	Matches map[string]*ChannelSearchResult `json:"posts"`
}

func NewChannelSearchResultList() *ChannelSearchResultList {
	return &ChannelSearchResultList{
		Order:   make([]string, 0),
		Matches: make(map[string]*ChannelSearchResult),
	}
}

func MakeChannelSearchResultList() *ChannelSearchResultList {
	return &ChannelSearchResultList{
		Order:   make([]string, 0),
		Matches: make(map[string]*ChannelSearchResult),
	}
}

// ToJson convert a Channel to a json string
func (c *ChannelSearchResultList) ToJson() string {
	b, _ := json.Marshal(c)
	return string(b)
}

// ChannelSearchFromJson will decode the input and return a Channel
func ChannelSearchResultFromJson(data io.Reader) *ChannelSearchResultList {
	var cs *ChannelSearchResultList
	json.NewDecoder(data).Decode(&cs)
	return cs
}

func (o *ChannelSearchResultList) AddOrder(id string) {

	if o.Order == nil {
		o.Order = make([]string, 0, 128)
	}

	o.Order = append(o.Order, id)
}

func (o *ChannelSearchResultList) AddMatch(post *Post, channel *Channel) {

	if o.Matches == nil {
		o.Matches = make(map[string]*ChannelSearchResult)
	}

	o.Matches[post.Id] = &ChannelSearchResult{
		Post:    *post,
		Channel: channel,
	}

}
