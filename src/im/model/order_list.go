package model

import (
	"encoding/json"
	"io"
	"sort"
)

type OrderList struct {
	Order  []string          `json:"order"`
	Orders map[string]*Order `json:"orders"`
}

func NewOrderList() *OrderList {
	return &OrderList{
		Order:  make([]string, 0),
		Orders: make(map[string]*Order),
	}
}

func (o *OrderList) ToSlice() []*Order {
	var order []*Order
	for _, id := range o.Order {
		order = append(order, o.Orders[id])
	}
	return order
}

func (o *OrderList) WithRewrittenImageURLs(f func(string) string) *OrderList {
	copy := *o
	copy.Orders = make(map[string]*Order)
	/*for id, transaction := range o.Orders {
		copy.Orders[id] = transaction.WithRewrittenImageURLs(f)
	}*/
	return &copy
}

func (o *OrderList) StripActionIntegrations() {
	order := o.Orders
	o.Orders = make(map[string]*Order)
	for id, transaction := range order {
		pcopy := *transaction
		o.Orders[id] = &pcopy
	}
}

func (o *OrderList) ToJson() string {
	copy := *o
	copy.StripActionIntegrations()
	b, err := json.Marshal(&copy)
	if err != nil {
		return ""
	} else {
		return string(b)
	}
}

func (o *OrderList) MakeNonNil() {
	if o.Order == nil {
		o.Order = make([]string, 0)
	}

	if o.Orders == nil {
		o.Orders = make(map[string]*Order)
	}

	for _, v := range o.Orders {
		v.MakeNonNil()
	}
}

func (o *OrderList) AddOrder(id string) {

	if o.Order == nil {
		o.Order = make([]string, 0, 128)
	}

	o.Order = append(o.Order, id)
}

func (o *OrderList) AddItem(transaction *Order) {

	if o.Orders == nil {
		o.Orders = make(map[string]*Order)
	}

	o.Orders[transaction.Id] = transaction
}

func (o *OrderList) Extend(other *OrderList) {
	for _, transactionId := range other.Order {
		if _, ok := o.Orders[transactionId]; !ok {
			o.AddItem(other.Orders[transactionId])
			o.AddOrder(transactionId)
		}
	}
}

func (o *OrderList) SortByCreateAt() {
	sort.Slice(o.Order, func(i, j int) bool {
		return o.Orders[o.Order[i]].CreateAt > o.Orders[o.Order[j]].CreateAt
	})
}

func (o *OrderList) Etag() string {

	id := "0"
	var t int64 = 0

	for _, v := range o.Orders {
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

func OrderListFromJson(data io.Reader) *OrderList {
	var o *OrderList
	json.NewDecoder(data).Decode(&o)
	return o
}
