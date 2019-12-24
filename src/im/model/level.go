package model

import (
	"encoding/json"
	"io"
	"net/http"
)

type Level struct {
	Id       string  `json:"id"`
	Name     string  `json:"name"`
	Active   bool    `json:"active"`
	Lvl      int64   `json:"lvl"`
	Value    float64 `json:"value"`
	CreateAt int64   `json:"create_at"`
	UpdateAt int64   `json:"update_at"`
	DeleteAt int64   `json:"delete_at"`
	AppId    string  `json:"app_id"`

	Invited     int `db:"-" json:"invited"`
	BonusEarned int `db:"-" json:"bonus_earned"`
}

type LevelPatch struct {
	Name   *string  `json:"name"`
	Active *bool    `json:"active"`
	Lvl    *int64   `json:"lvl"`
	Value  *float64 `json:"value"`
}

func (p *Level) Patch(patch *LevelPatch) {

	if patch.Name != nil {
		p.Name = *patch.Name
	}
	if patch.Active != nil {
		p.Active = *patch.Active
	}
	if patch.Lvl != nil {
		p.Lvl = *patch.Lvl
	}
	if patch.Value != nil {
		p.Value = *patch.Value
	}
}

func (level *Level) ToJson() string {
	b, _ := json.Marshal(level)
	return string(b)
}

func LevelFromJson(data io.Reader) *Level {
	var level *Level
	json.NewDecoder(data).Decode(&level)
	return level
}

func LevelsFromJson(data io.Reader) []*Level {
	var levels []*Level
	json.NewDecoder(data).Decode(&levels)
	return levels
}

func LevelPatchFromJson(data io.Reader) *LevelPatch {
	var patch *LevelPatch
	json.NewDecoder(data).Decode(&patch)
	return patch
}

func (o *Level) Clone() *Level {
	copy := *o
	return &copy
}
func (o *Level) PreSave() {
	if o.Id == "" {
		o.Id = NewId()
	}

	if o.CreateAt == 0 {
		o.CreateAt = GetMillis()
	}

	o.UpdateAt = o.CreateAt
	o.PreCommit()
}

func (o *Level) PreCommit() {

}

func (o *Level) MakeNonNil() {

}

func (o *Level) IsValid() *AppError {

	if len(o.Id) != 26 {
		return NewAppError("Level.IsValid", "model.level.is_valid.id.app_error", nil, "", http.StatusBadRequest)
	}

	if o.CreateAt == 0 {
		return NewAppError("Level.IsValid", "model.level.is_valid.create_at.app_error", nil, "id="+o.Id, http.StatusBadRequest)
	}

	if o.UpdateAt == 0 {
		return NewAppError("Level.IsValid", "model.level.is_valid.update_at.app_error", nil, "id="+o.Id, http.StatusBadRequest)
	}

	return nil
}
