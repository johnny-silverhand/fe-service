package model

import (
	"encoding/json"
	"io"
)

type OrdersStats struct {
	TotalCount    int64 `json:"total_count"`
	CurrentCount  int64 `json:"current_count"`
	DeferredCount int64 `json:"deferred_count"`
	ClosedCount   int64 `json:"closed_count"`
}

func (o *OrdersStats) ToJson() string {
	b, _ := json.Marshal(o)
	return string(b)
}

func OrdersStatsFromJson(data io.Reader) *OrdersStats {
	var o *OrdersStats
	json.NewDecoder(data).Decode(&o)
	return o
}
