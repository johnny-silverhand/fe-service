package model

import (
	"encoding/json"
	"io"
	"net/http"
)

const (
	TRANSACTION_TYPE_BONUS = "bonus"
)

type Transaction struct {
	Id          string  `json:"id"`
	AppId       string  `json:"app_id"`
	UserId      string  `json:"user_id"`
	OrderId     string  `json:"order_id"`
	Description string  `json:"description"`
	Value       float64 `json:"value"`
	Active      bool    `json:"active"`
	CreateAt    int64   `json:"create_at"`
	UpdateAt    int64   `json:"update_at"`
	DeleteAt    int64   `json:"delete_at"`

	Type string `json:"type"`

	Code  string `json:"code,omitempty" db:"-"`
	Token string `json:"token,omitempty" db:"-"`
}

type TransactionPatch struct {
	Name        *string `json:"name"`
	Preview     *string `json:"preview"`
	Description *string `json:"description"`

	Latitude  string `json:"lat"`
	Longitude string `json:"long"`
}

func (p *Transaction) Patch(patch *TransactionPatch) {

	if patch.Description != nil {
		p.Description = *patch.Description
	}

}

func (transaction *Transaction) ToJson() string {
	b, _ := json.Marshal(transaction)
	return string(b)
}

func TransactionFromJson(data io.Reader) *Transaction {
	var transaction *Transaction
	json.NewDecoder(data).Decode(&transaction)
	return transaction
}
func (o *Transaction) Clone() *Transaction {
	copy := *o
	return &copy
}
func (o *Transaction) PreSave() {
	if o.Id == "" {
		o.Id = NewId()
	}

	if o.CreateAt == 0 {
		o.CreateAt = GetMillis()
	}

	o.UpdateAt = o.CreateAt
	o.PreCommit()
}

func (o *Transaction) PreCommit() {

}

func (o *Transaction) MakeNonNil() {

}

func (o *Transaction) IsValid() *AppError {

	if len(o.Id) != 26 {
		return NewAppError("Transaction.IsValid", "model.transaction.is_valid.id.app_error", nil, "", http.StatusBadRequest)
	}

	if o.CreateAt == 0 {
		return NewAppError("Transaction.IsValid", "model.transaction.is_valid.create_at.app_error", nil, "id="+o.Id, http.StatusBadRequest)
	}

	if o.UpdateAt == 0 {
		return NewAppError("Transaction.IsValid", "model.transaction.is_valid.update_at.app_error", nil, "id="+o.Id, http.StatusBadRequest)
	}

	return nil
}
