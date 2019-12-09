package model

import (
	"encoding/json"
	"io"
	"net/http"
)

type Application struct {
	Id string `json:"id"`

	Name        string `json:"name"`
	Preview     string `json:"preview"`
	Description string `json:"description"`

	PaymentDetails string `json:"payment_details"`

	Phone       string `json:"phone"`
	BuildNumber string `json:"build_number"`

	Active bool `json:"active"`

	CreateAt int64 `json:"create_at"`
	UpdateAt int64 `json:"update_at"`
	DeleteAt int64 `json:"delete_at"`

	Email string `json:"email"`
}

type ApplicationPatch struct {
	Name           *string `json:"name"`
	Preview        *string `json:"preview"`
	Description    *string `json:"description"`
	PaymentDetails *string `json:"payment_details"`
	Phone          *string `json:"phone"`
	Active         *bool   `json:"active"`
}

func (p *Application) Patch(patch *ApplicationPatch) {

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
	if patch.Phone != nil {
		p.Phone = *patch.Phone
	}
	if patch.PaymentDetails != nil {
		p.PaymentDetails = *patch.Phone
	}
}

func (application *Application) ToJson() string {
	b, _ := json.Marshal(application)
	return string(b)
}

func ApplicationFromJson(data io.Reader) *Application {
	var application *Application
	json.NewDecoder(data).Decode(&application)
	return application
}
func (o *Application) Clone() *Application {
	copy := *o
	return &copy
}
func (o *Application) PreSave() {
	if o.Id == "" {
		o.Id = NewId()
	}

	mills := GetMillis()

	if o.CreateAt == 0 {
		o.CreateAt = mills
	}

	o.UpdateAt = mills

	o.PreCommit()
}

func (o *Application) PreCommit() {

}

func (o *Application) MakeNonNil() {

}

func (o *Application) IsValid() *AppError {

	if len(o.Id) != 26 {
		return NewAppError("Application.IsValid", "model.application.is_valid.id.app_error", nil, "", http.StatusBadRequest)
	}

	if o.CreateAt == 0 {
		return NewAppError("Application.IsValid", "model.application.is_valid.create_at.app_error", nil, "id="+o.Id, http.StatusBadRequest)
	}

	if o.UpdateAt == 0 {
		return NewAppError("Application.IsValid", "model.application.is_valid.update_at.app_error", nil, "id="+o.Id, http.StatusBadRequest)
	}

	return nil
}
