package model

import (
	"encoding/json"
	"io"
	"net/http"
)

type PaySystem struct {
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

type PaySystemPatch struct {
	Name        *string `json:"name"`
	Preview     *string `json:"preview"`
	Description *string `json:"description"`
	Active      *bool   `json:"active"`
	Latitude    string  `json:"lat"`
	Longitude   string  `json:"long"`
}

func (p *PaySystem) Patch(patch *PaySystemPatch) {

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

func (system *PaySystem) ToJson() string {
	b, _ := json.Marshal(system)
	return string(b)
}

func PaySystemFromJson(data io.Reader) *PaySystem {
	var system *PaySystem
	json.NewDecoder(data).Decode(&system)
	return system
}
func (o *PaySystem) Clone() *PaySystem {
	copy := *o
	return &copy
}
func (o *PaySystem) PreSave() {
	if o.Id == "" {
		o.Id = NewId()
	}

	if o.CreateAt == 0 {
		o.CreateAt = GetMillis()
	}

	o.UpdateAt = o.CreateAt
	o.PreCommit()
}

func (o *PaySystem) PreCommit() {

}

func (o *PaySystem) MakeNonNil() {

}

func (o *PaySystem) IsValid() *AppError {

	if len(o.Id) != 26 {
		return NewAppError("PaySystem.IsValid", "model.system.is_valid.id.app_error", nil, "", http.StatusBadRequest)
	}

	if o.CreateAt == 0 {
		return NewAppError("PaySystem.IsValid", "model.system.is_valid.create_at.app_error", nil, "id="+o.Id, http.StatusBadRequest)
	}

	if o.UpdateAt == 0 {
		return NewAppError("PaySystem.IsValid", "model.system.is_valid.update_at.app_error", nil, "id="+o.Id, http.StatusBadRequest)
	}

	return nil
}
