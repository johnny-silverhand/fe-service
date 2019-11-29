package model

import (
	"encoding/json"
	"io"
	"sort"
)

type Category struct {
	Id            string       `json:"id"`
	AppId         string       `json:"app_id"`
	Name          string       `json:"name"`
	ParentId      string       `json:"parent_id"`
	CreateAt      int64        `json:"create_at"`
	UpdateAt      int64        `json:"update_at"`
	Lft           int          `json:"lft"`
	Rgt           int          `json:"rgt"`
	Depth         int          `json:"depth"`
	CountChildren int          `db:"-" json:"count_children"`
	Children      []*Category  `db:"-" json:"children"`
	DestinationId string       `db:"-" json:"destination_id,omitempty"`
	ProductList   *ProductList `db:"-" json:"products,omitempty"`
}

type CategoryPatch struct {
	Id       string `db:"id"`
	AppId    string `db:"app_id"`
	Name     string `db:"name"`
	ParentId string `db:"parent_id"`
	CreateAt *int64 `db:"create_at"`
	UpdateAt *int64 `db:"update_at"`
	DeleteAt *int64 `db:"delete_at"`
}

func (c *Category) NewCp(id string, name string) *CategoryPatch {
	cp := CategoryPatch{}
	cp.Id = id
	cp.Name = name
	return &cp
}

func (category *Category) SetPatch() *CategoryPatch {
	patch := CategoryPatch{}
	patch.AppId = category.AppId
	patch.ParentId = category.ParentId
	patch.Name = category.Name
	return &patch
}

func (category *Category) ToJson() string {
	b, _ := json.Marshal(category)
	return string(b)
}

func CategoriesToJson(categories []*Category) string {
	if len(categories) < 2 {
		m, _ := json.Marshal(categories)
		return string(m)
	}
	sort.Slice(categories, func(i, j int) bool {
		return categories[i].Lft > categories[j].Lft
	})
	slice := make(map[string]*Category)
	for i, _ := range categories {
		slice[categories[i].Id] = categories[i]
	}
	for i, category := range categories {
		if len(category.ParentId) > 0 && slice[category.ParentId] != nil {
			slice[category.ParentId].CountChildren += 1
			slice[category.ParentId].Children = append(slice[category.ParentId].Children, categories[i])
		}
	}
	tree := []*Category{}
	for _, category := range slice {
		if len(category.ParentId) == 0 {
			tree = append(tree, category)
		}
	}
	sort.Slice(tree, func(i, j int) bool {
		return tree[i].Lft > tree[j].Lft
	})
	outdata, err := json.Marshal(tree)
	if err != nil {
		panic(err) // dont use panic
	}
	return string(outdata)
}

func (o *Category) PreSave() {
	if o.Id == "" {
		o.Id = NewId()
	}

	if o.CreateAt == 0 {
		o.CreateAt = GetMillis()
	}

	o.UpdateAt = o.CreateAt
	o.PreCommit()
}

func (o *Category) PreCommit() {

}

func CategoryFromJson(data io.Reader) *Category {
	var category *Category
	json.NewDecoder(data).Decode(&category)
	return category
}
