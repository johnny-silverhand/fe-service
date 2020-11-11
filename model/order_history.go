package model

import (
	"encoding/json"
	"io"
)

type OrderHistory struct {
	Id        string    `json:"id"`
	Status    string    `json:"status"`
	Value     int       `json:"value"`
	CreatedAt int64     `json:"created_at"`
	Positions []*Basket `json:"positions"`
}

func (order *OrderHistory) ToJson() string {
	b, _ := json.Marshal(order)
	return string(b)
}

func OrderHistoryFromJson(data io.Reader) *OrderHistory {
	var order *OrderHistory
	json.NewDecoder(data).Decode(&order)
	return order
}
