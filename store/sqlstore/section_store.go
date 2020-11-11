package sqlstore

import (
	"im/model"
	"im/store"
	"net/http"
)

type SqlSectionStore struct {
	SqlStore
}

func NewSqlSectionStore(sqlStore SqlStore) store.SectionStore {
	s := &SqlSectionStore{sqlStore}

	for _, db := range sqlStore.GetAllConns() {
		table := db.AddTableWithName(model.Section{}, "Sections").SetKeys(false, "Id")
		table.ColMap("Id").SetMaxSize(26)
		table.ColMap("Name").SetMaxSize(250)
		table.ColMap("ParentId").SetMaxSize(26)
	}

	return s
}

func (s SqlSectionStore) CreateIndexesIfNotExists() {
	s.CreateIndexIfNotExists("idx_sections_id", "Sections", "Id")
	s.CreateIndexIfNotExists("idx_sections_parent_id", "Sections", "ParentId")
}

func (s SqlSectionStore) GetAll(parentId string) store.StoreChannel {
	return store.Do(func(result *store.StoreResult) {
		sections := &model.SectionList{}
		if _, err := s.GetReplica().Select(sections, `
				SELECT *
				FROM Sections`, map[string]interface{}{"Id": parentId}); err != nil {
			result.Err = model.NewAppError("SqlSectionStore.GetChildren", "store.sql_section.get_children.app_error", nil, err.Error(), http.StatusInternalServerError)
		} else {
			result.Data = sections
		}
	})
}

func (s SqlSectionStore) Insert(section *model.Section) store.StoreChannel {
	return store.Do(func(result *store.StoreResult) {

		var parent *model.Section
		if err := s.GetReplica().SelectOne(&parent, "SELECT * FROM Sections WHERE Id = :Id", map[string]interface{}{"Id": section.ParentId}); err == nil {
			if _, err := s.GetMaster().Exec("UPDATE Sections SET Lft= CASE WHEN Lft > :Lft THEN Lft+2 ELSE Lft END, Rgt= CASE WHEN Rgt > :Rgt THEN Rgt+2 ELSE Rgt END", map[string]interface{}{"Lft": parent.Rgt, "Rgt": parent.Rgt - 1}); err == nil {
				//if _, err := s.GetMaster().Exec("UPDATE Sections SET Rgt = Lft + 2 WHERE Rgt >= :Rgt AND Lft < :Lft", map[string]interface{}{"Lft": parent.Rgt,  "Rgt": parent.Rgt -1}); err == nil {

				section.Depth = parent.Depth + 1
				section.Lft = parent.Rgt
				section.Rgt = parent.Rgt + 1

				if err := s.GetMaster().Insert(section); err != nil {
					result.Err = model.NewAppError("SqlSectionStore.SaveOrUpdate", "store.sql_status.save.app_error", nil, err.Error(), http.StatusInternalServerError)
				} else {
					result.Data = section
				}
				//}
			}
		}

	})
}
