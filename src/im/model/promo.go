package model

import (
	"encoding/json"
	"io"
	"net/http"
)

const (
	PROMO_STATUS_DRAFT      = "draft"
	PROMO_STATUS_MODERATION = "moderation"
	PROMO_STATUS_ACCEPTED   = "accepted"
	PROMO_STATUS_REJECTED   = "rejected"
)

type Promo struct {
	Id    string `json:"id"`
	AppId string `json:"app_id"`

	Name        string `json:"name"`
	Preview     string `json:"preview"`
	Description string `json:"description"`

	Status    string `json:"status"`
	Active    bool   `json:"active"`
	CreateAt  int64  `json:"create_at"`
	UpdateAt  int64  `json:"update_at"`
	DeleteAt  int64  `json:"delete_at"`
	ProductId string `json:"product_id"`

	Push bool `json:"push"`

	FileIds StringArray `json:"file_ids,omitempty"`
	Media   []*FileInfo `db:"-" json:"media,omitempty"`

	BeginAt  int64 `json:"begin_at"`
	ExpireAt int64 `json:"expire_at"`
}

type PromoPatch struct {
	Name        *string `json:"name"`
	Preview     *string `json:"preview"`
	Description *string `json:"description"`
	ProductId   *string `json:"product_id"`
	ImageId     string  `json:"image_id,omitempty"`
	BeginAt     *int64  `json:"begin_at"`
	ExpireAt    *int64  `json:"expire_at"`
}

type PromoStatus struct {
	PromoId string `json:"promo_id"`
	Status  string `json:"status"`
	Active  bool   `json:"active"`
}

func PromoStatusFromJson(data io.Reader) *PromoStatus {
	var o *PromoStatus
	json.NewDecoder(data).Decode(&o)
	return o
}

func (o *PromoStatus) ToJson() string {
	b, _ := json.Marshal(o)
	return string(b)
}

func (p *Promo) Patch(patch *PromoPatch) {

	if patch.Name != nil {
		p.Name = *patch.Name
	}
	if patch.Preview != nil {
		p.Preview = *patch.Preview
	}
	if patch.Description != nil {
		p.Description = *patch.Description
	}
	if patch.ProductId != nil {
		p.Description = *patch.ProductId
	}
}

func (promo *Promo) ToJson() string {
	b, _ := json.Marshal(promo)
	return string(b)
}

func PromoFromJson(data io.Reader) *Promo {
	var promo *Promo
	json.NewDecoder(data).Decode(&promo)
	return promo
}
func (o *Promo) Clone() *Promo {
	copy := *o
	return &copy
}
func (o *Promo) PreSave() {
	if o.Id == "" {
		o.Id = NewId()
	}

	if o.CreateAt == 0 {
		o.CreateAt = GetMillis()
	}

	o.UpdateAt = o.CreateAt
	o.PreCommit()
}

func (o *Promo) PreCommit() {

}

func (o *Promo) MakeNonNil() {

}

func (o *Promo) IsValid() *AppError {

	if len(o.Id) != 26 {
		return NewAppError("Promo.IsValid", "model.promo.is_valid.id.app_error", nil, "", http.StatusBadRequest)
	}

	if o.CreateAt == 0 {
		return NewAppError("Promo.IsValid", "model.promo.is_valid.create_at.app_error", nil, "id="+o.Id, http.StatusBadRequest)
	}

	if o.UpdateAt == 0 {
		return NewAppError("Promo.IsValid", "model.promo.is_valid.update_at.app_error", nil, "id="+o.Id, http.StatusBadRequest)
	}

	return nil
}
