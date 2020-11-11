package model

import (
	"encoding/json"
	"io"
	"sort"
)

type LevelList struct {
	Order  []string          `json:"order"`
	Levels map[string]*Level `json:"levels"`
}

func NewLevelList() *LevelList {
	return &LevelList{
		Order:  make([]string, 0),
		Levels: make(map[string]*Level),
	}
}

func (o *LevelList) ToSlice() []*Level {
	var levels []*Level
	for _, id := range o.Order {
		levels = append(levels, o.Levels[id])
	}
	return levels
}

func (o *LevelList) WithRewrittenImageURLs(f func(string) string) *LevelList {
	copy := *o
	copy.Levels = make(map[string]*Level)
	/*for id, level := range o.Levels {
		copy.Levels[id] = level.WithRewrittenImageURLs(f)
	}*/
	return &copy
}

func (o *LevelList) StripActionIntegrations() {
	levels := o.Levels
	o.Levels = make(map[string]*Level)
	for id, level := range levels {
		pcopy := *level
		o.Levels[id] = &pcopy
	}
}

func (o *LevelList) ToJson() string {
	copy := *o
	copy.StripActionIntegrations()
	b, err := json.Marshal(&copy)
	if err != nil {
		return ""
	} else {
		return string(b)
	}
}

func (o *LevelList) MakeNonNil() {
	if o.Order == nil {
		o.Order = make([]string, 0)
	}

	if o.Levels == nil {
		o.Levels = make(map[string]*Level)
	}

	for _, v := range o.Levels {
		v.MakeNonNil()
	}
}

func (o *LevelList) AddOrder(id string) {

	if o.Order == nil {
		o.Order = make([]string, 0, 128)
	}

	o.Order = append(o.Order, id)
}

func (o *LevelList) AddLevel(level *Level) {

	if o.Levels == nil {
		o.Levels = make(map[string]*Level)
	}

	o.Levels[level.Id] = level
}

func (o *LevelList) Extend(other *LevelList) {
	for _, levelId := range other.Order {
		if _, ok := o.Levels[levelId]; !ok {
			o.AddLevel(other.Levels[levelId])
			o.AddOrder(levelId)
		}
	}
}

func (o *LevelList) SortByCreateAt() {
	sort.Slice(o.Order, func(i, j int) bool {
		return o.Levels[o.Order[i]].CreateAt > o.Levels[o.Order[j]].CreateAt
	})
}

func (o *LevelList) SortByLvl() {
	sort.Slice(o.Order, func(i, j int) bool {
		return o.Levels[o.Order[i]].Lvl < o.Levels[o.Order[j]].Lvl
	})
}

func (o *LevelList) Etag() string {

	id := "0"
	var t int64 = 0

	for _, v := range o.Levels {
		if v.UpdateAt > t {
			t = v.UpdateAt
			id = v.Id
		} else if v.UpdateAt == t && v.Id > id {
			t = v.UpdateAt
			id = v.Id
		}
	}

	orderId := ""
	if len(o.Order) > 0 {
		orderId = o.Order[0]
	}

	return Etag(orderId, id, t)
}

func LevelListFromJson(data io.Reader) *LevelList {
	var o *LevelList
	json.NewDecoder(data).Decode(&o)
	return o
}
