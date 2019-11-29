package model

import (
	"encoding/json"
	"io"
	"net/http"
)

type Office struct {
	Id          string `json:"id"`
	AppId       string `json:"app_id"`
	Name        string `json:"name"`
	Preview     string `json:"preview"`
	Description string `json:"description"`
	Active      bool   `json:"active"`
	CreateAt    int64  `json:"create_at"`
	UpdateAt    int64  `json:"update_at"`
	DeleteAt    int64  `json:"delete_at"`
	Latitude    string `json:"lat"`
	Longitude   string `json:"long"`
}

type OfficePatch struct {
	Name        *string `json:"name"`
	Preview     *string `json:"preview"`
	Description *string `json:"description"`
	Active      *bool   `json:"active"`
	Latitude    string  `json:"lat"`
	Longitude   string  `json:"long"`
}

func (p *Office) Patch(patch *OfficePatch) {

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

func (office *Office) ToJson() string {
	b, _ := json.Marshal(office)
	return string(b)
}

func OfficeFromJson(data io.Reader) *Office {
	var office *Office
	json.NewDecoder(data).Decode(&office)
	return office
}
func (o *Office) Clone() *Office {
	copy := *o
	return &copy
}
func (o *Office) PreSave() {
	if o.Id == "" {
		o.Id = NewId()
	}

	if o.CreateAt == 0 {
		o.CreateAt = GetMillis()
	}

	o.UpdateAt = o.CreateAt
	o.PreCommit()
}

func (o *Office) PreCommit() {

}

func (o *Office) MakeNonNil() {

}

func (o *Office) IsValid() *AppError {

	if len(o.Id) != 26 {
		return NewAppError("Office.IsValid", "model.office.is_valid.id.app_error", nil, "", http.StatusBadRequest)
	}

	if o.CreateAt == 0 {
		return NewAppError("Office.IsValid", "model.office.is_valid.create_at.app_error", nil, "id="+o.Id, http.StatusBadRequest)
	}

	if o.UpdateAt == 0 {
		return NewAppError("Office.IsValid", "model.office.is_valid.update_at.app_error", nil, "id="+o.Id, http.StatusBadRequest)
	}

	return nil
}
