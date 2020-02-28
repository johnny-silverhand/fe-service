package model

import (
	"encoding/json"
	"io"
	"net/http"
)

type Extra struct {
	Id        string `json:"id"`
	ProductId string `json:"product_id"`
	RefId     string `json:"ref_id"`
	CreateAt  int64  `json:"create_at"`
	UpdateAt  int64  `json:"update_at"`
	DeleteAt  int64  `json:"delete_at"`
	Required  bool   `json:"required"`
}

func NewExtra(productId string, RefId string, required bool) *Extra {
	return &Extra{
		ProductId: productId,
		RefId:     RefId,
		Required:  required,
	}
}

type ExtraBasket struct {
	Products []string `json:"products"`
}

func (extra *ExtraBasket) ToJson() string {
	b, _ := json.Marshal(extra)
	return string(b)
}

func ExtraBasketFromJson(data io.Reader) *ExtraBasket {
	var product *ExtraBasket
	json.NewDecoder(data).Decode(&product)
	return product
}

func (product *Extra) ToJson() string {
	b, _ := json.Marshal(product)
	return string(b)
}

func ExtraFromJson(data io.Reader) *Extra {
	var product *Extra
	json.NewDecoder(data).Decode(&product)
	return product
}

func (o *Extra) Clone() *Extra {
	copy := *o
	return &copy
}
func (o *Extra) PreSave() {
	if o.Id == "" {
		o.Id = NewId()
	}

	if o.CreateAt == 0 {
		o.CreateAt = GetMillis()
	}

	o.UpdateAt = o.CreateAt
	o.PreCommit()
}

func (o *Extra) PreCommit() {

}

func (o *Extra) MakeNonNil() {

}

func (o *Extra) IsValid() *AppError {

	if len(o.Id) != 26 {
		return NewAppError("Extra.IsValid", "model.product.is_valid.id.app_error", nil, "", http.StatusBadRequest)
	}

	if o.CreateAt == 0 {
		return NewAppError("Extra.IsValid", "model.product.is_valid.create_at.app_error", nil, "id="+o.Id, http.StatusBadRequest)
	}

	if o.UpdateAt == 0 {
		return NewAppError("Extra.IsValid", "model.product.is_valid.update_at.app_error", nil, "id="+o.Id, http.StatusBadRequest)
	}

	return nil
}
