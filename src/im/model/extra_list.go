package model

import (
	"encoding/json"
	"io"
	"sort"
)

type ExtraList struct {
	Order []string         `json:"order"`
	Extras map[string]*Extra `json:"extras"`
}

func NewExtraList() *ExtraList {
	return &ExtraList{
		Order: make([]string, 0),
		Extras: make(map[string]*Extra),
	}
}

func (o *ExtraList) ToSlice() []*Extra {
	var extras []*Extra
	for _, id := range o.Order {
		extras = append(extras, o.Extras[id])
	}
	return extras
}

func (o *ExtraList) WithRewrittenImageURLs(f func(string) string) *ExtraList {
	copy := *o
	copy.Extras = make(map[string]*Extra)
	/*for id, extra := range o.Extras {
		copy.Extras[id] = extra.WithRewrittenImageURLs(f)
	}*/
	return &copy
}

func (o *ExtraList) StripActionIntegrations() {
	extras := o.Extras
	o.Extras = make(map[string]*Extra)
	for id, extra := range extras {
		pcopy := *extra
		o.Extras[id] = &pcopy
	}
}

func (o *ExtraList) ToJson() string {
	copy := *o
	copy.StripActionIntegrations()
	b, err := json.Marshal(&copy)
	if err != nil {
		return ""
	} else {
		return string(b)
	}
}

func (o *ExtraList) MakeNonNil() {
	if o.Order == nil {
		o.Order = make([]string, 0)
	}

	if o.Extras == nil {
		o.Extras = make(map[string]*Extra)
	}

	for _, v := range o.Extras {
		v.MakeNonNil()
	}
}

func (o *ExtraList) AddOrder(id string) {

	if o.Order == nil {
		o.Order = make([]string, 0, 128)
	}

	o.Order = append(o.Order, id)
}

func (o *ExtraList) AddExtra(extra *Extra) {

	if o.Extras == nil {
		o.Extras = make(map[string]*Extra)
	}

	o.Extras[extra.Id] = extra
}

func (o *ExtraList) Extend(other *ExtraList) {
	for _, extraId := range other.Order {
		if _, ok := o.Extras[extraId]; !ok {
			o.AddExtra(other.Extras[extraId])
			o.AddOrder(extraId)
		}
	}
}

func (o *ExtraList) SortByCreateAt() {
	sort.Slice(o.Order, func(i, j int) bool {
		return o.Extras[o.Order[i]].CreateAt > o.Extras[o.Order[j]].CreateAt
	})
}

func (o *ExtraList) Etag() string {

	id := "0"
	var t int64 = 0

	for _, v := range o.Extras {
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


func ExtraListFromJson(data io.Reader) *ExtraList {
	var o *ExtraList
	json.NewDecoder(data).Decode(&o)
	return o
}
