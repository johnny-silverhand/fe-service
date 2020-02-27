package model

import "encoding/json"

type MetricsForRegister struct {
	TotalClients           int                             `json:"total_clients"`
	ClientsWithOrders      int                             `json:"clients_with_orders"`
	ClientsPaidWithBonuses int                             `json:"clients_paid_with_bonuses"`
	ClientsChargeBonuses   int                             `json:"clients_charge_bonuses"`
	ClientsAppNotInstalled int                             `json:"clients_app_not_installed"`
	ClientsBonuses         int                             `json:"clients_bonuses"`
	ClientsDiscardBonuses  int                             `json:"clients_discard_bonuses"`
	RegisterClientsByDay   []*AdditionalMetricsForRegister `json:"register_clients_by_day,omitempty"`
}

func (o *MetricsForRegister) ToJson() string {
	b, _ := json.Marshal(o)
	return string(b)
}

type AdditionalMetricsForRegister struct {
	Date  string `json:"date"`
	Count int    `json:"count"`
}

type MetricsForOrders struct {
	Total         float64                         `json:"total"`
	TotalDiscount float64                         `json:"total_discount"`
	TotalPrice    float64                         `json:"total_price"`
	AvgPrice      float64                         `json:"avg_price"`
	TotalReturn   float64                         `json:"total_return"`
	OrdersByDay   []*AdditionalMetricsForRegister `json:"orders_by_day,omitempty"`
}

func (o *MetricsForOrders) ToJson() string {
	b, _ := json.Marshal(o)
	return string(b)
}

type AdditionalMetricsForOrders struct {
	Date  string `json:"date"`
	Count int    `json:"count"`
}
