package model

import (
	"encoding/json"
	"io"
	"net/http"
)

type Basket struct {
	Id            string  `json:"id"`
	OrderId       string  `json:"order_id"`
	ProductId     string  `json:"product_id"`
	Price         float64 `json:"price"`
	Currency      string  `json:"currency"`
	InsertAt      int64   `json:"insert_at"`
	UpdateAt      int64   `json:"update_at"`
	RefreshAt     int64   `json:"refresh_at"`
	Quantity      int     `json:"quantity"`
	Name          string  `json:"name"`
	DiscountPrice int64   `json:"discount_price"`
	DiscountValue int64   `json:"discount_value"`
	Sort          int     `json:"sort"`

	Product *Product `db:"-" json:"product"`

	CreateAt int64 `json:"create_at"`
	DeleteAt int64 `json:"delete_at"`

	Cashback float64 `json:"cashback"`
}

type BasketPatch struct {
	Name        *string `json:"name"`
	Preview     *string `json:"preview"`
	Description *string `json:"description"`

	Latitude  string `json:"lat"`
	Longitude string `json:"long"`
}

func (p *Basket) Patch(patch *BasketPatch) {

	if patch.Name != nil {
		p.Name = *patch.Name
	}

}

func (basket *Basket) ToJson() string {
	b, _ := json.Marshal(basket)
	return string(b)
}

func BasketFromJson(data io.Reader) *Basket {
	var basket *Basket
	json.NewDecoder(data).Decode(&basket)
	return basket
}
func (o *Basket) Clone() *Basket {
	copy := *o
	return &copy
}
func (o *Basket) PreSave() {
	if o.Id == "" {
		o.Id = NewId()
	}

	if o.CreateAt == 0 {
		o.CreateAt = GetMillis()
	}

	o.UpdateAt = o.CreateAt
	o.PreCommit()
}

func (o *Basket) PreCommit() {

}
func (o *Basket) Fil(order *Order) {

	o.OrderId = order.Id
}

func (o *Basket) MakeNonNil() {

}

func (o *Basket) IsValid() *AppError {

	if len(o.Id) != 26 {
		return NewAppError("Basket.IsValid", "model.basket.is_valid.id.app_error", nil, "", http.StatusBadRequest)
	}

	if o.CreateAt == 0 {
		return NewAppError("Basket.IsValid", "model.basket.is_valid.create_at.app_error", nil, "id="+o.Id, http.StatusBadRequest)
	}

	if o.UpdateAt == 0 {
		return NewAppError("Basket.IsValid", "model.basket.is_valid.update_at.app_error", nil, "id="+o.Id, http.StatusBadRequest)
	}

	return nil
}
