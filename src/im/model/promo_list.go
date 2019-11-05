package model

import (
	"encoding/json"
	"io"
	"sort"
)

type PromoList struct {
	Order []string         `json:"order"`
	Promos map[string]*Promo `json:"promos"`
}

func NewPromoList() *PromoList {
	return &PromoList{
		Order: make([]string, 0),
		Promos: make(map[string]*Promo),
	}
}

func (o *PromoList) ToSlice() []*Promo {
	var promos []*Promo
	for _, id := range o.Order {
		promos = append(promos, o.Promos[id])
	}
	return promos
}

func (o *PromoList) WithRewrittenImageURLs(f func(string) string) *PromoList {
	copy := *o
	copy.Promos = make(map[string]*Promo)
	/*for id, promo := range o.Promos {
		copy.Promos[id] = promo.WithRewrittenImageURLs(f)
	}*/
	return &copy
}

func (o *PromoList) StripActionIntegrations() {
	promos := o.Promos
	o.Promos = make(map[string]*Promo)
	for id, promo := range promos {
		pcopy := *promo
		o.Promos[id] = &pcopy
	}
}

func (o *PromoList) ToJson() string {
	copy := *o
	copy.StripActionIntegrations()
	b, err := json.Marshal(&copy)
	if err != nil {
		return ""
	} else {
		return string(b)
	}
}

func (o *PromoList) MakeNonNil() {
	if o.Order == nil {
		o.Order = make([]string, 0)
	}

	if o.Promos == nil {
		o.Promos = make(map[string]*Promo)
	}

	for _, v := range o.Promos {
		v.MakeNonNil()
	}
}

func (o *PromoList) AddOrder(id string) {

	if o.Order == nil {
		o.Order = make([]string, 0, 128)
	}

	o.Order = append(o.Order, id)
}

func (o *PromoList) AddPromo(promo *Promo) {

	if o.Promos == nil {
		o.Promos = make(map[string]*Promo)
	}

	o.Promos[promo.Id] = promo
}

func (o *PromoList) Extend(other *PromoList) {
	for _, promoId := range other.Order {
		if _, ok := o.Promos[promoId]; !ok {
			o.AddPromo(other.Promos[promoId])
			o.AddOrder(promoId)
		}
	}
}

func (o *PromoList) SortByCreateAt() {
	sort.Slice(o.Order, func(i, j int) bool {
		return o.Promos[o.Order[i]].CreateAt > o.Promos[o.Order[j]].CreateAt
	})
}

func (o *PromoList) Etag() string {

	id := "0"
	var t int64 = 0

	for _, v := range o.Promos {
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


func PromoListFromJson(data io.Reader) *PromoList {
	var o *PromoList
	json.NewDecoder(data).Decode(&o)
	return o
}
