package model

import (
	"encoding/json"
	"io"
	"sort"
)

type Category struct {
	Id            string      `json:"id"`
	ClientId      string      `json:"client_id"`
	Name          string      `json:"name"`
	ParentId      string      `json:"parent_id"`
	CreateAt      int64       `json:"create_at"`
	UpdateAt      int64       `json:"update_at"`
	Lft           int         `json:"lft"`
	Rgt           int         `json:"rgt"`
	Depth         int         `json:"depth"`
	CountChildren int         `db:"-" json:"count_children"`
	Children      []*Category `db:"-" json:"children"`
	Products      []*Product  `db:"-" json:"products"`
}

type CategoryPatch struct {
	Id       string `db:"Id"`
	ClientId string `db:"ClientId"`
	Name     string `db:"Name"`
	ParentId string `db:"ParentId"`
	CreateAt *int64 `db:"CreateAt"`
	UpdateAt *int64 `db:"UpdateAt"`
	DeleteAt *int64 `db:"DeleteAt"`
}

func (c *Category) NewCp(id string, name string) *CategoryPatch {
	cp := CategoryPatch{}
	cp.Id = id
	cp.Name = name
	return &cp
}

func (category *Category) SetPatch() *CategoryPatch {
	patch := CategoryPatch{}
	patch.ClientId = category.ClientId
	patch.ParentId = category.ParentId
	patch.Name = category.Name
	return &patch
}

func (category *Category) ToJson() string {
	b, _ := json.Marshal(category)
	return string(b)
}

func CategoriesAllToJson(categories []*Category) string {
	if len(categories) < 2 {
		m, _ := json.Marshal(categories)
		return string(m)
	}
	sort.Slice(categories, func(i, j int) bool {
		return categories[i].Lft < categories[j].Lft
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
		return tree[i].Lft < tree[j].Lft
	})
	outdata, err := json.Marshal(tree)
	if err != nil {
		panic(err)
	}
	return string(outdata)
}

func CategoriesToJson(categories []*Category) string {
	sort.Slice(categories, func(i, j int) bool {
		return categories[i].CreateAt < categories[j].CreateAt
	})
	slice := make(map[string]*Category)
	for i, _ := range categories {
		slice[categories[i].Id] = categories[i]
	}
	for i, category := range categories {
		if len(category.ParentId) > 0 && slice[category.ParentId] != nil {
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
		return tree[i].CreateAt < tree[j].CreateAt
	})
	outdata, err := json.Marshal(tree)
	if err != nil {
		panic(err)
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
