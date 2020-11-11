package model

import (
	"encoding/json"
	"io"
	"net/http"
)

const SBERBANK_AQUIRING_TYPE = "sberbank"
const ALFABANK_AQUIRING_TYPE = "alfabank"

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

	Settings    string `json:"settings"`
	MaxDiscount int    `json:"max_discount"`

	AqType     string `json:"aq_type"`
	AqUsername string `json:"aq_username"`
	AqPassword string `json:"aq_password"`

	Cash     bool    `json:"cash"`
	Cashback float64 `json:"cashback"`

	HasModeration bool  `json:"has_moderation"`
	BlockedAt     int64 `json:"blocked_at"`

	ModerationCount int    `db:"-" json:"moderation_count,omitempty"`
	RegBonus        int    `json:"reg_bonus"`
	ContactDetails  string `json:"contact_details"`

	Password string `db:"-" json:"password"`

	SmsLogin  string `json:"sms_login"`
	SmsApiKey string `json:"sms_api_key"`
}

type ApplicationPatch struct {
	Name           *string  `json:"name"`
	Preview        *string  `json:"preview"`
	Description    *string  `json:"description"`
	PaymentDetails *string  `json:"payment_details"`
	Phone          *string  `json:"phone"`
	Active         *bool    `json:"active"`
	Settings       *string  `json:"settings"`
	Email          *string  `json:"email"`
	MaxDiscount    *int     `json:"max_discount"`
	AqType         *string  `json:"aq_type"`
	AqUsername     *string  `json:"aq_username"`
	AqPassword     *string  `json:"aq_password"`
	Cash           *bool    `json:"cash"`
	Cashback       *float64 `json:"cashback"`
	HasModeration  *bool    `json:"has_moderation"`
	BlockedAt      *int64   `json:"blocked_at"`
	RegBonus       *int     `json:"reg_bonus"`
	ContactDetails *string  `json:"contact_details"`
	Password       *string  `json:"password"`
	SmsLogin       *string  `json:"sms_login"`
	SmsApiKey      *string  `json:"sms_api_key"`
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
		p.PaymentDetails = *patch.PaymentDetails
	}
	if patch.Settings != nil {
		p.Settings = *patch.Settings
	}
	if patch.Email != nil {
		p.Email = *patch.Email
	}
	if patch.MaxDiscount != nil {
		p.MaxDiscount = *patch.MaxDiscount
	}
	if patch.AqType != nil {
		p.AqType = *patch.AqType
	}
	if patch.AqUsername != nil {
		p.AqUsername = *patch.AqUsername
	}
	if patch.AqPassword != nil {
		p.AqPassword = *patch.AqPassword
	}
	if patch.Cash != nil {
		p.Cash = *patch.Cash
	}
	if patch.Cashback != nil {
		p.Cashback = *patch.Cashback
	}
	if patch.HasModeration != nil {
		p.HasModeration = *patch.HasModeration
	}
	if patch.BlockedAt != nil {
		p.BlockedAt = *patch.BlockedAt
	}
	if patch.RegBonus != nil {
		p.RegBonus = *patch.RegBonus
	}
	if patch.ContactDetails != nil {
		p.ContactDetails = *patch.ContactDetails
	}
	if patch.Password != nil {
		p.Password = *patch.Password
	}
	if patch.SmsLogin != nil {
		p.SmsLogin = *patch.SmsLogin
	}
	if patch.SmsApiKey != nil {
		p.SmsApiKey = *patch.SmsApiKey
	}
}

func (application *Application) ToJson() string {
	b, _ := json.Marshal(application)
	return string(b)
}

func ApplicationPatchFromJson(data io.Reader) *ApplicationPatch {
	var patch *ApplicationPatch
	json.NewDecoder(data).Decode(&patch)
	return patch
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
