package model

import (
	"encoding/json"
	"io"
	"sort"
)

type Category struct {
	Id       		string      		`json:"id"`
	ClientId 		string      		`json:"client_id"`
	Name     		string      		`json:"name"`
	ParentId 		string     			`json:"parent_id"`
	CreateAt 		int64       		`json:"create_at"`
	UpdateAt 		int64       		`json:"update_at"`
	DeleteAt 		int64       		`json:"delete_at"`
	Lft      		int         		`json:"lft"`
	Rgt      		int         		`json:"rgt"`
	Depth    		int         		`json:"depth"`
	Children 		[]*Category 		`db:"-" json:"count_children"`
	CountChildren 	int        			`db:"-" json:"count_children"`
	//Products  []*Products `json:"Products"`
}

type CategoryPatch struct {
	Id        string    `db:"Id, primarykey, autoincrement"`
	ClientId  string    `db:"ClientId"`
	Name      string `db:"Name"`
	ParentId  string   `db:"ParentId"`
	CreatedAt *int64 `db:"CreatedAt"`
	UpdatedAt *int64 `db:"UpdatedAt"`
	DeletedAt *int64 `db:"DeletedAt"`
}

func (c *Category) NewCp(id int, name string) *CategoryPatch {

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

func CategoriesToJson(categories []*Category) string {
	sort.Slice(categories, func(i, j int) bool {
		return categories[i].Depth > categories[j].Depth
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
		return tree[i].Depth > tree[j].Depth
	})
	outdata, err := json.Marshal(tree)
	if err != nil {
		panic(err)
	}
	return string(outdata)
}

func CategoryFromJson(data io.Reader) *Category {
	var category *Category
	json.NewDecoder(data).Decode(&category)
	return category
}
