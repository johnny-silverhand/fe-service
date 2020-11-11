package model

import (
	"encoding/json"
	"io"
	"sort"
)

type ProductList struct {
	Order    []string            `json:"order"`
	Products map[string]*Product `json:"products"`
}

func NewProductList() *ProductList {
	return &ProductList{
		Order:    make([]string, 0),
		Products: make(map[string]*Product),
	}
}

func (o *ProductList) ToSlice() []*Product {
	var products []*Product
	for _, id := range o.Order {
		products = append(products, o.Products[id])
	}
	return products
}

func (o *ProductList) WithRewrittenImageURLs(f func(string) string) *ProductList {
	copy := *o
	copy.Products = make(map[string]*Product)
	/*for id, product := range o.Products {
		copy.Products[id] = product.WithRewrittenImageURLs(f)
	}*/
	return &copy
}

func (o *ProductList) StripActionIntegrations() {
	products := o.Products
	o.Products = make(map[string]*Product)
	for id, product := range products {
		pcopy := *product
		o.Products[id] = &pcopy
	}
}

func (o *ProductList) ToJson() string {
	copy := *o
	copy.StripActionIntegrations()
	b, err := json.Marshal(&copy)
	if err != nil {
		return ""
	} else {
		return string(b)
	}
}

func (o *ProductList) MakeNonNil() {
	if o.Order == nil {
		o.Order = make([]string, 0)
	}

	if o.Products == nil {
		o.Products = make(map[string]*Product)
	}

	for _, v := range o.Products {
		v.MakeNonNil()
	}
}

func (o *ProductList) AddOrder(id string) {

	if o.Order == nil {
		o.Order = make([]string, 0, 128)
	}

	o.Order = append(o.Order, id)
}

func (o *ProductList) AddProduct(product *Product) {

	if o.Products == nil {
		o.Products = make(map[string]*Product)
	}

	o.Products[product.Id] = product
}

func (o *ProductList) Extend(other *ProductList) {
	for _, productId := range other.Order {
		if _, ok := o.Products[productId]; !ok {
			o.AddProduct(other.Products[productId])
			o.AddOrder(productId)
		}
	}
}

func (o *ProductList) SortByCreateAt() {
	sort.Slice(o.Order, func(i, j int) bool {
		return o.Products[o.Order[i]].CreateAt > o.Products[o.Order[j]].CreateAt
	})
}

func (o *ProductList) Etag() string {

	id := "0"
	var t int64 = 0

	for _, v := range o.Products {
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

func ProductListFromJson(data io.Reader) *ProductList {
	var o *ProductList
	json.NewDecoder(data).Decode(&o)
	return o
}
