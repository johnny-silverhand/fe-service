package model

import (
	"encoding/json"
	"io"
	"net/http"
	"unicode/utf8"
)

const (
	PRODUCT_STATUS_DRAFT      = "draft"
	PRODUCT_STATUS_MODERATION = "moderation"
	PRODUCT_STATUS_ACCEPTED   = "accepted"
)

type Product struct {
	Id            string  `json:"id"`
	ClientId      string  `json:"client_id"`
	Name          string  `json:"name"`
	Preview       string  `json:"preview"`
	Description   string  `json:"description"`
	Price         float64 `json:"price,string"`
	Currency      string  `json:"currency"`
	Measure       string  `json:"measure"`
	DiscountLimit float64 `json:"discount_limit,string"`
	Cashback      float64 `json:"cashback,string"`
	Status        string  `json:"status"`
	Active        bool    `json:"active"`
	CreateAt      int64   `json:"create_at"`
	UpdateAt      int64   `json:"update_at"`
	DeleteAt      int64   `json:"delete_at"`
	CategoryId    string  `json:"category_id"`
	//Category      *Category `json:"category"`
	FileIds  StringArray `json:"file_ids,omitempty"`
	Category *Category   `json:"category,omitempty" db:"-"`
	Media    []*FileInfo `db:"-" json:"media,omitempty"`
}

type ProductPatch struct {
	Name        *string      `json:"name"`
	Preview     *string      `json:"preview"`
	Description *string      `json:"description"`
	CategoryId  *string      `json:"category_id"`
	FileIds     *StringArray `json:"file_ids"`
}

type ProductStatus struct {
	ProductId string `json:"product_id"`
	Status    string `json:"status"`
	Activate  bool   `json:"activate"`
}

func ProductStatusFromJson(data io.Reader) *ProductStatus {
	var o *ProductStatus
	json.NewDecoder(data).Decode(&o)
	return o
}

func (o *ProductStatus) ToJson() string {
	b, _ := json.Marshal(o)
	return string(b)
}

type ProductSearch struct {
	Terms          *string `json:"terms"`
	TimeZoneOffset *int    `json:"time_zone_offset"`
	Page           *int    `json:"page"`
	PerPage        *int    `json:"per_page"`
	CategoryId     *string `json:"category_id"`
}

func ProductSearchFromJson(data io.Reader) *ProductSearch {
	decoder := json.NewDecoder(data)
	var searchParam ProductSearch
	err := decoder.Decode(&searchParam)
	if err != nil {
		return nil
	}
	return &searchParam
}

func (p *Product) Patch(patch *ProductPatch) {

	if patch.Name != nil {
		p.Name = *patch.Name
	}
	if patch.Preview != nil {
		p.Preview = *patch.Preview
	}
	if patch.Description != nil {
		p.Description = *patch.Description
	}
	if patch.CategoryId != nil {
		p.Description = *patch.CategoryId
	}
	if patch.FileIds != nil {
		p.FileIds = *patch.FileIds
	}

}

func (product *Product) ToJson() string {
	b, _ := json.Marshal(product)
	return string(b)
}

func ProductFromJson(data io.Reader) *Product {
	var product *Product
	json.NewDecoder(data).Decode(&product)
	return product
}
func (o *Product) Clone() *Product {
	copy := *o
	return &copy
}
func (o *Product) PreSave() {
	if o.Id == "" {
		o.Id = NewId()
	}

	if o.CreateAt == 0 {
		o.CreateAt = GetMillis()
	}

	o.UpdateAt = o.CreateAt
	o.PreCommit()
}

func (o *Product) PreCommit() {

	if o.FileIds == nil {
		o.FileIds = []string{}
	}

	o.FileIds = RemoveDuplicateStrings(o.FileIds)
}

func (o *Product) MakeNonNil() {

}

func (o *Product) IsValid() *AppError {

	if len(o.Id) != 26 {
		return NewAppError("Product.IsValid", "model.product.is_valid.id.app_error", nil, "", http.StatusBadRequest)
	}

	if o.CreateAt == 0 {
		return NewAppError("Product.IsValid", "model.product.is_valid.create_at.app_error", nil, "id="+o.Id, http.StatusBadRequest)
	}

	if o.UpdateAt == 0 {
		return NewAppError("Product.IsValid", "model.product.is_valid.update_at.app_error", nil, "id="+o.Id, http.StatusBadRequest)
	}

	if len(o.CategoryId) == 26 && len(o.CategoryId) == 0 {
		return NewAppError("Product.IsValid", "model.product.is_valid.root_parent.app_error", nil, "", http.StatusBadRequest)
	}

	if utf8.RuneCountInString(ArrayToJson(o.FileIds)) > POST_FILEIDS_MAX_RUNES {
		return NewAppError("Product.IsValid", "model.product.is_valid.file_ids.app_error", nil, "id="+o.Id, http.StatusBadRequest)
	}

	return nil
}
