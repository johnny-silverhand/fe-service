package model

import "encoding/json"

type ProductsDiscount struct {
	Total  int64         `json:"total"`
	Limits []interface{} `json:"limits"`
}

func (o *ProductsDiscount) ToJson() string {
	b, err := json.Marshal(&o)
	if err != nil {
		return ""
	} else {
		return string(b)
	}
}
