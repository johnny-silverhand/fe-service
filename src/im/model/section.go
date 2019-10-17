package model

import (
	"encoding/json"
	"io"
)

type Section struct {
	Id       string     `json:"id"`
	Name     string     `json:"name"`
	ParentId string     `json:"parent_id"`
	Depth    int64      `json:"depth"`
	Lft      int64      `json:"lft"`
	Rgt      int64      `json:"rgt"`
	Children []*Section `json:"children" db:"-"` //
}

func (o *Section) ToJson() string {
	b, _ := json.Marshal(o)
	return string(b)
}

func (o *Section) ToClusterJson() string {
	b, _ := json.Marshal(o)
	return string(b)
}

func SectionFromJson(data io.Reader) *Section {
	var o *Section
	json.NewDecoder(data).Decode(&o)
	return o
}

func SectionListToJson(u []*Section) string {

	b, _ := json.Marshal(u)

	return string(b)
}

func SectionListFromJson(data io.Reader) []*Section {
	var sections []*Section
	json.NewDecoder(data).Decode(&sections)
	return sections
}

type SectionList []*Section


func (o SectionList) GenerateTree() SectionList {
	query := SectionList{}
	for _, n := range o {
		if n.Depth == 0 {
			query = append(query, n)
		}
	}
	return o.runQueryRecursive(query)
}

func (sections SectionList) runQueryRecursive(query SectionList) (result SectionList) {
	for _, q := range query {
		result = append(result, &Section{
			Id:       q.Id,
			Name:     q.Name,
			Lft:      q.Lft,
			Rgt:      q.Rgt,
			Children: sections.getChildrens(q),
		})
	}
	return result
}

func (sections SectionList) getChildrens(parent *Section) SectionList {
	query := SectionList{}
	for _, n := range sections {
		if n.Depth == parent.Depth+1 && n.Lft > parent.Lft && n.Rgt < parent.Rgt {
			query = append(query, n)
		}
	}
	return sections.runQueryRecursive(query)
}

func (o *SectionList) ToJson() string {
	if b, err := json.Marshal(o); err != nil {
		return "[]"
	} else {
		return string(b)
	}
}
