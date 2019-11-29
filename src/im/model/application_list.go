package model

import (
	"encoding/json"
	"io"
	"sort"
)

type ApplicationList struct {
	Order        []string                `json:"order"`
	Applications map[string]*Application `json:"applications"`
}

func NewApplicationList() *ApplicationList {
	return &ApplicationList{
		Order:        make([]string, 0),
		Applications: make(map[string]*Application),
	}
}

func (o *ApplicationList) ToSlice() []*Application {
	var applications []*Application
	for _, id := range o.Order {
		applications = append(applications, o.Applications[id])
	}
	return applications
}

func (o *ApplicationList) WithRewrittenImageURLs(f func(string) string) *ApplicationList {
	copy := *o
	copy.Applications = make(map[string]*Application)
	/*for id, application := range o.Applications {
		copy.Applications[id] = application.WithRewrittenImageURLs(f)
	}*/
	return &copy
}

func (o *ApplicationList) StripActionIntegrations() {
	applications := o.Applications
	o.Applications = make(map[string]*Application)
	for id, application := range applications {
		pcopy := *application
		o.Applications[id] = &pcopy
	}
}

func (o *ApplicationList) ToJson() string {
	copy := *o
	copy.StripActionIntegrations()
	b, err := json.Marshal(&copy)
	if err != nil {
		return ""
	} else {
		return string(b)
	}
}

func (o *ApplicationList) MakeNonNil() {
	if o.Order == nil {
		o.Order = make([]string, 0)
	}

	if o.Applications == nil {
		o.Applications = make(map[string]*Application)
	}

	for _, v := range o.Applications {
		v.MakeNonNil()
	}
}

func (o *ApplicationList) AddOrder(id string) {

	if o.Order == nil {
		o.Order = make([]string, 0, 128)
	}

	o.Order = append(o.Order, id)
}

func (o *ApplicationList) AddApplication(application *Application) {

	if o.Applications == nil {
		o.Applications = make(map[string]*Application)
	}

	o.Applications[application.Id] = application
}

func (o *ApplicationList) Extend(other *ApplicationList) {
	for _, appId := range other.Order {
		if _, ok := o.Applications[appId]; !ok {
			o.AddApplication(other.Applications[appId])
			o.AddOrder(appId)
		}
	}
}

func (o *ApplicationList) SortByCreateAt() {
	sort.Slice(o.Order, func(i, j int) bool {
		return o.Applications[o.Order[i]].CreateAt > o.Applications[o.Order[j]].CreateAt
	})
}

func (o *ApplicationList) Etag() string {

	id := "0"
	var t int64 = 0

	for _, v := range o.Applications {
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

func ApplicationListFromJson(data io.Reader) *ApplicationList {
	var o *ApplicationList
	json.NewDecoder(data).Decode(&o)
	return o
}
