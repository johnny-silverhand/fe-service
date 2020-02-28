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

type UserMetricsForRating struct {
	User
	OrdersCount int64   `json:"orders_count"`
	OrdersSum   float64 `json:"orders_sum"`
}

func (o *UserMetricsForRating) ToJson() string {
	b, _ := json.Marshal(o)
	return string(b)
}

type UserMetricsForRatingList struct {
	Order                []string                         `json:"order"`
	UserMetricsForRating map[string]*UserMetricsForRating `json:"users"`
}

func NewUserMetricsForRatingList() *UserMetricsForRatingList {
	return &UserMetricsForRatingList{
		Order:                make([]string, 0),
		UserMetricsForRating: make(map[string]*UserMetricsForRating),
	}
}

func (o *UserMetricsForRatingList) ToJson() string {
	copy := *o
	b, err := json.Marshal(&copy)
	if err != nil {
		return ""
	} else {
		return string(b)
	}
}

func (o *UserMetricsForRatingList) MakeNonNil() {
	if o.Order == nil {
		o.Order = make([]string, 0)
	}

	if o.UserMetricsForRating == nil {
		o.UserMetricsForRating = make(map[string]*UserMetricsForRating)
	}

	for _, v := range o.UserMetricsForRating {
		v.MakeNonNil()
	}
}

func (o *UserMetricsForRatingList) AddOrder(id string) {

	if o.Order == nil {
		o.Order = make([]string, 0, 128)
	}

	o.Order = append(o.Order, id)
}

func (o *UserMetricsForRatingList) AddItem(user *UserMetricsForRating) {

	if o.UserMetricsForRating == nil {
		o.UserMetricsForRating = make(map[string]*UserMetricsForRating)
	}

	o.UserMetricsForRating[user.Id] = user
}
