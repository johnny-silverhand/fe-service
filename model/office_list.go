package model

import (
	"encoding/json"
	"io"
	"sort"
)

type OfficeList struct {
	Order   []string           `json:"order"`
	Offices map[string]*Office `json:"offices"`
}

func NewOfficeList() *OfficeList {
	return &OfficeList{
		Order:   make([]string, 0),
		Offices: make(map[string]*Office),
	}
}

func (o *OfficeList) ToSlice() []*Office {
	var offices []*Office
	for _, id := range o.Order {
		offices = append(offices, o.Offices[id])
	}
	return offices
}

func (o *OfficeList) WithRewrittenImageURLs(f func(string) string) *OfficeList {
	copy := *o
	copy.Offices = make(map[string]*Office)
	/*for id, office := range o.Offices {
		copy.Offices[id] = office.WithRewrittenImageURLs(f)
	}*/
	return &copy
}

func (o *OfficeList) StripActionIntegrations() {
	offices := o.Offices
	o.Offices = make(map[string]*Office)
	for id, office := range offices {
		pcopy := *office
		o.Offices[id] = &pcopy
	}
}

func (o *OfficeList) ToJson() string {
	copy := *o
	copy.StripActionIntegrations()
	b, err := json.Marshal(&copy)
	if err != nil {
		return ""
	} else {
		return string(b)
	}
}

func (o *OfficeList) MakeNonNil() {
	if o.Order == nil {
		o.Order = make([]string, 0)
	}

	if o.Offices == nil {
		o.Offices = make(map[string]*Office)
	}

	for _, v := range o.Offices {
		v.MakeNonNil()
	}
}

func (o *OfficeList) AddOrder(id string) {

	if o.Order == nil {
		o.Order = make([]string, 0, 128)
	}

	o.Order = append(o.Order, id)
}

func (o *OfficeList) AddOffice(office *Office) {

	if o.Offices == nil {
		o.Offices = make(map[string]*Office)
	}

	o.Offices[office.Id] = office
}

func (o *OfficeList) Extend(other *OfficeList) {
	for _, officeId := range other.Order {
		if _, ok := o.Offices[officeId]; !ok {
			o.AddOffice(other.Offices[officeId])
			o.AddOrder(officeId)
		}
	}
}

func (o *OfficeList) SortByCreateAt() {
	sort.Slice(o.Order, func(i, j int) bool {
		return o.Offices[o.Order[i]].CreateAt > o.Offices[o.Order[j]].CreateAt
	})
}

func (o *OfficeList) Etag() string {

	id := "0"
	var t int64 = 0

	for _, v := range o.Offices {
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

func OfficeListFromJson(data io.Reader) *OfficeList {
	var o *OfficeList
	json.NewDecoder(data).Decode(&o)
	return o
}
