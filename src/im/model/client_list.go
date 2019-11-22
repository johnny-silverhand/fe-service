package model

import (
	"encoding/json"
	"io"
	"sort"
)

type ClientList struct {
	Order []string         `json:"order"`
	Clients map[string]*Client `json:"clients"`
}

func NewClientList() *ClientList {
	return &ClientList{
		Order: make([]string, 0),
		Clients: make(map[string]*Client),
	}
}

func (o *ClientList) ToSlice() []*Client {
	var clients []*Client
	for _, id := range o.Order {
		clients = append(clients, o.Clients[id])
	}
	return clients
}

func (o *ClientList) WithRewrittenImageURLs(f func(string) string) *ClientList {
	copy := *o
	copy.Clients = make(map[string]*Client)
	/*for id, client := range o.Clients {
		copy.Clients[id] = client.WithRewrittenImageURLs(f)
	}*/
	return &copy
}

func (o *ClientList) StripActionIntegrations() {
	clients := o.Clients
	o.Clients = make(map[string]*Client)
	for id, client := range clients {
		pcopy := *client
		o.Clients[id] = &pcopy
	}
}

func (o *ClientList) ToJson() string {
	copy := *o
	copy.StripActionIntegrations()
	b, err := json.Marshal(&copy)
	if err != nil {
		return ""
	} else {
		return string(b)
	}
}

func (o *ClientList) MakeNonNil() {
	if o.Order == nil {
		o.Order = make([]string, 0)
	}

	if o.Clients == nil {
		o.Clients = make(map[string]*Client)
	}

	for _, v := range o.Clients {
		v.MakeNonNil()
	}
}

func (o *ClientList) AddOrder(id string) {

	if o.Order == nil {
		o.Order = make([]string, 0, 128)
	}

	o.Order = append(o.Order, id)
}

func (o *ClientList) AddClient(client *Client) {

	if o.Clients == nil {
		o.Clients = make(map[string]*Client)
	}

	o.Clients[client.Id] = client
}

func (o *ClientList) Extend(other *ClientList) {
	for _, clientId := range other.Order {
		if _, ok := o.Clients[clientId]; !ok {
			o.AddClient(other.Clients[clientId])
			o.AddOrder(clientId)
		}
	}
}

func (o *ClientList) SortByCreateAt() {
	sort.Slice(o.Order, func(i, j int) bool {
		return o.Clients[o.Order[i]].CreateAt > o.Clients[o.Order[j]].CreateAt
	})
}

func (o *ClientList) Etag() string {

	id := "0"
	var t int64 = 0

	for _, v := range o.Clients {
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


func ClientListFromJson(data io.Reader) *ClientList {
	var o *ClientList
	json.NewDecoder(data).Decode(&o)
	return o
}
