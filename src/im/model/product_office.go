// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See License.txt for license information.

package model

import (
	"encoding/json"
	"io"
	"net/http"
)

type ProductOffice struct {
	Id string `json:"id"`

	ProductId string `json:"product_id"`
	OfficeId  string `json:"office_id"`

	CreateAt int64 `json:"create_at"`
	UpdateAt int64 `json:"update_at"`
	DeleteAt int64 `json:"delete_at"`
}

func (po *ProductOffice) ToJson() string {
	b, _ := json.Marshal(po)
	return string(b)
}

func ProductOfficeFromJson(data io.Reader) *ProductOffice {
	decoder := json.NewDecoder(data)

	var po ProductOffice
	if err := decoder.Decode(&po); err != nil {
		return nil
	} else {
		return &po
	}
}

func ProductOfficesToJson(pos []*ProductOffice) string {
	b, _ := json.Marshal(pos)
	return string(b)
}

func ProductOfficesFromJson(data io.Reader) []*ProductOffice {
	decoder := json.NewDecoder(data)

	var pos []*ProductOffice
	if err := decoder.Decode(&pos); err != nil {
		return nil
	} else {
		return pos
	}
}

func (o *ProductOffice) PreSave() {
	if o.Id == "" {
		o.Id = NewId()
	}

	mills := GetMillis()

	if o.CreateAt == 0 {
		o.CreateAt = mills
	}

	o.UpdateAt = mills
}

func (o *ProductOffice) IsValid() *AppError {
	if len(o.Id) != 26 {
		return NewAppError("ProductOffice.IsValid", "model.file_info.is_valid.id.app_error", nil, "", http.StatusBadRequest)
	}

	if o.CreateAt == 0 {
		return NewAppError("ProductOffice.IsValid", "model.file_info.is_valid.create_at.app_error", nil, "id="+o.Id, http.StatusBadRequest)
	}

	if o.UpdateAt == 0 {
		return NewAppError("ProductOffice.IsValid", "model.file_info.is_valid.update_at.app_error", nil, "id="+o.Id, http.StatusBadRequest)
	}

	return nil
}

func NewProductOffice(officeId, productId string) *ProductOffice {
	po := &ProductOffice{
		OfficeId:  officeId,
		ProductId: productId,
	}

	return po
}
