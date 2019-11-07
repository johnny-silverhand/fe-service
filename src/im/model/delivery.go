package model

import (
	"encoding/json"
	"io"
	"net/http"
)

type Delivery struct {
	Id          string `json:"id"`
	Name        string `json:"name"`
	Preview     string `json:"preview"`
	Description string `json:"description"`
	Active      bool   `json:"active"`
	Code        string `json:"code"`
	CreateAt    int64  `json:"create_at"`
	UpdateAt    int64  `json:"update_at"`
	DeleteAt    int64  `json:"delete_at"`

}

type DeliveryPatch struct {
	Name        *string `json:"name"`
	Preview     *string `json:"preview"`
	Description *string `json:"description"`
	Active      *bool   `json:"active"`
	Latitude    string  `json:"lat"`
	Longitude   string  `json:"long"`
}

func (p *Delivery) Patch(patch *DeliveryPatch) {

	if patch.Name != nil {
		p.Name = *patch.Name
	}
	if patch.Preview != nil {
		p.Preview = *patch.Preview
	}
	if patch.Description != nil {
		p.Description = *patch.Description
	}
	if patch.Active != nil {
		p.Active = *patch.Active
	}
}

func (delivery *Delivery) ToJson() string {
	b, _ := json.Marshal(delivery)
	return string(b)
}

func DeliveryFromJson(data io.Reader) *Delivery {
	var delivery *Delivery
	json.NewDecoder(data).Decode(&delivery)
	return delivery
}

func (o *Delivery) Clone() *Delivery {
	copy := *o
	return &copy
}
func (o *Delivery) PreSave() {
	if o.Id == "" {
		o.Id = NewId()
	}

	if o.CreateAt == 0 {
		o.CreateAt = GetMillis()
	}

	o.UpdateAt = o.CreateAt
	o.PreCommit()
}

func (o *Delivery) PreCommit() {

}

func (o *Delivery) MakeNonNil() {

}

func (o *Delivery) IsValid() *AppError {

	if len(o.Id) != 26 {
		return NewAppError("Delivery.IsValid", "model.delivery.is_valid.id.app_error", nil, "", http.StatusBadRequest)
	}

	if o.CreateAt == 0 {
		return NewAppError("Delivery.IsValid", "model.delivery.is_valid.create_at.app_error", nil, "id="+o.Id, http.StatusBadRequest)
	}

	if o.UpdateAt == 0 {
		return NewAppError("Delivery.IsValid", "model.delivery.is_valid.update_at.app_error", nil, "id="+o.Id, http.StatusBadRequest)
	}

	return nil
}
