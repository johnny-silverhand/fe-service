package model

import (
	"encoding/json"
	"io"
	"sort"
)

type TransactionList struct {
	Order []string         `json:"order"`
	Transactions map[string]*Transaction `json:"transactions"`
}

func NewTransactionList() *TransactionList {
	return &TransactionList{
		Order: make([]string, 0),
		Transactions: make(map[string]*Transaction),
	}
}

func (o *TransactionList) ToSlice() []*Transaction {
	var transactions []*Transaction
	for _, id := range o.Order {
		transactions = append(transactions, o.Transactions[id])
	}
	return transactions
}

func (o *TransactionList) WithRewrittenImageURLs(f func(string) string) *TransactionList {
	copy := *o
	copy.Transactions = make(map[string]*Transaction)
	/*for id, transaction := range o.Transactions {
		copy.Transactions[id] = transaction.WithRewrittenImageURLs(f)
	}*/
	return &copy
}

func (o *TransactionList) StripActionIntegrations() {
	transactions := o.Transactions
	o.Transactions = make(map[string]*Transaction)
	for id, transaction := range transactions {
		pcopy := *transaction
		o.Transactions[id] = &pcopy
	}
}

func (o *TransactionList) ToJson() string {
	copy := *o
	copy.StripActionIntegrations()
	b, err := json.Marshal(&copy)
	if err != nil {
		return ""
	} else {
		return string(b)
	}
}

func (o *TransactionList) MakeNonNil() {
	if o.Order == nil {
		o.Order = make([]string, 0)
	}

	if o.Transactions == nil {
		o.Transactions = make(map[string]*Transaction)
	}

	for _, v := range o.Transactions {
		v.MakeNonNil()
	}
}

func (o *TransactionList) AddOrder(id string) {

	if o.Order == nil {
		o.Order = make([]string, 0, 128)
	}

	o.Order = append(o.Order, id)
}

func (o *TransactionList) AddTransaction(transaction *Transaction) {

	if o.Transactions == nil {
		o.Transactions = make(map[string]*Transaction)
	}

	o.Transactions[transaction.Id] = transaction
}

func (o *TransactionList) Extend(other *TransactionList) {
	for _, transactionId := range other.Order {
		if _, ok := o.Transactions[transactionId]; !ok {
			o.AddTransaction(other.Transactions[transactionId])
			o.AddOrder(transactionId)
		}
	}
}

func (o *TransactionList) SortByCreateAt() {
	sort.Slice(o.Order, func(i, j int) bool {
		return o.Transactions[o.Order[i]].CreateAt > o.Transactions[o.Order[j]].CreateAt
	})
}

func (o *TransactionList) Etag() string {

	id := "0"
	var t int64 = 0

	for _, v := range o.Transactions {
		if v.UpdateAt > t {
			t = v.UpdateAt
			id = v.Id
		} else if v.UpdateAt == t && v.Id > id {
			t = v.UpdateAt
			id = v.Id
		}
	}

	orderId := ""
	if len(o.Order) > 0 {
		orderId = o.Order[0]
	}

	return Etag(orderId, id, t)
}


func TransactionListFromJson(data io.Reader) *TransactionList {
	var o *TransactionList
	json.NewDecoder(data).Decode(&o)
	return o
}
