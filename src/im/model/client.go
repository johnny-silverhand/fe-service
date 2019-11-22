package model

import (
	"encoding/json"
	"io"
	"net/http"
)

type Client struct {
	Id          string `json:"id"`

	Name        string `json:"name"`
	Preview     string `json:"preview"`
	Description string `json:"description"`

	Active      bool   `json:"active"`

	CreateAt    int64  `json:"create_at"`
	UpdateAt    int64  `json:"update_at"`
	DeleteAt    int64  `json:"delete_at"`

}

type ClientPatch struct {
	Name        *string `json:"name"`
	Preview     *string `json:"preview"`
	Description *string `json:"description"`
	Active      *bool   `json:"active"`

}

func (p *Client) Patch(patch *ClientPatch) {

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

func (client *Client) ToJson() string {
	b, _ := json.Marshal(client)
	return string(b)
}

func ClientFromJson(data io.Reader) *Client {
	var client *Client
	json.NewDecoder(data).Decode(&client)
	return client
}
func (o *Client) Clone() *Client {
	copy := *o
	return &copy
}
func (o *Client) PreSave() {
	if o.Id == "" {
		o.Id = NewId()
	}

	if o.CreateAt == 0 {
		o.CreateAt = GetMillis()
	}

	o.UpdateAt = o.CreateAt
	o.PreCommit()
}

func (o *Client) PreCommit() {

}

func (o *Client) MakeNonNil() {

}

func (o *Client) IsValid() *AppError {

	if len(o.Id) != 26 {
		return NewAppError("Client.IsValid", "model.client.is_valid.id.app_error", nil, "", http.StatusBadRequest)
	}

	if o.CreateAt == 0 {
		return NewAppError("Client.IsValid", "model.client.is_valid.create_at.app_error", nil, "id="+o.Id, http.StatusBadRequest)
	}

	if o.UpdateAt == 0 {
		return NewAppError("Client.IsValid", "model.client.is_valid.update_at.app_error", nil, "id="+o.Id, http.StatusBadRequest)
	}

	return nil
}
