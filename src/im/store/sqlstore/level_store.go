package sqlstore

import (
	"database/sql"
	"im/model"
	"im/store"
	"net/http"
)

type SqlLevelStore struct {
	SqlStore
}

func NewSqlLevelStore(sqlStore SqlStore) store.LevelStore {
	s := &SqlLevelStore{sqlStore}

	for _, db := range sqlStore.GetAllConns() {
		table := db.AddTableWithName(model.Level{}, "Levels").SetKeys(false, "Id")

		table.ColMap("Id").SetMaxSize(26)
		table.ColMap("Name").SetMaxSize(255)

	}

	return s
}

func (s SqlLevelStore) CreateIndexesIfNotExists() {
	s.CreateIndexIfNotExists("idx_levels_update_at", "Levels", "UpdateAt")
	s.CreateIndexIfNotExists("idx_levels_create_at", "Levels", "CreateAt")
	s.CreateIndexIfNotExists("idx_levels_delete_at", "Levels", "DeleteAt")
}

func (s SqlLevelStore) Activate(levelId string) store.StoreChannel {
	return store.Do(func(result *store.StoreResult) {

		_, err := s.GetMaster().Exec("UPDATE Levels SET Active = :Active, UpdateAt =:UpdateAt WHERE Id = :Id ", map[string]interface{}{"Active": true, "UpdateAt": model.GetMillis(), "Id": levelId})
		if err != nil {
			result.Err = model.NewAppError("SqlLevelStore.Publish", "store.sql_levels.publish.app_error", nil, err.Error(), http.StatusInternalServerError)

		}

	})
}
func (s SqlLevelStore) Deactivate(levelId string) store.StoreChannel {
	return store.Do(func(result *store.StoreResult) {

		_, err := s.GetMaster().Exec("UPDATE Levels SET Active = :Active, UpdateAt =:UpdateAt WHERE Id = :Id ", map[string]interface{}{"Active": false, "UpdateAt": model.GetMillis(), "Id": levelId})
		if err != nil {
			result.Err = model.NewAppError("SqlLevelStore.Disable", "store.sql_levels.publish.app_error", nil, err.Error(), http.StatusInternalServerError)

		}

	})
}

func (s *SqlLevelStore) Save(level *model.Level) store.StoreChannel {
	return store.Do(func(result *store.StoreResult) {
		if len(level.Id) > 0 {
			result.Err = model.NewAppError("SqlLevelStore.Save", "store.sql_level.save.existing.app_error", nil, "id="+level.Id, http.StatusBadRequest)
			return
		}

		level.PreSave()

		if result.Err = level.IsValid(); result.Err != nil {
			return
		}

		if err := s.GetMaster().Insert(level); err != nil {
			result.Err = model.NewAppError("SqlLevelStore.Save", "store.sql_level.save.app_error", nil, "id="+level.Id+", "+err.Error(), http.StatusInternalServerError)
		} else {
			result.Data = level
		}
	})
}

func (s *SqlLevelStore) Update(newLevel *model.Level) store.StoreChannel {
	return store.Do(func(result *store.StoreResult) {
		newLevel.UpdateAt = model.GetMillis()
		newLevel.PreCommit()

		if _, err := s.GetMaster().Update(newLevel); err != nil {
			result.Err = model.NewAppError("SqlLevelStore.Update", "store.sql_level.update.app_error", nil, "id="+newLevel.Id+", "+err.Error(), http.StatusInternalServerError)
		} else {
			result.Data = newLevel
		}
	})
}

func (s *SqlLevelStore) Get(id string) store.StoreChannel {
	return store.Do(func(result *store.StoreResult) {
		var level *model.Level
		if err := s.GetReplica().SelectOne(&level,
			`SELECT *
					FROM Levels
					WHERE Id = :Id  AND DeleteAt = 0`, map[string]interface{}{"Id": id}); err != nil {
			if err == sql.ErrNoRows {
				result.Err = model.NewAppError("SqlLevelStore.Get", "store.sql_levels.get.app_error", nil, err.Error(), http.StatusNotFound)
			} else {
				result.Err = model.NewAppError("SqlLevelStore.Get", "store.sql_levels.get.app_error", nil, err.Error(), http.StatusInternalServerError)
			}
		} else {
			result.Data = level
		}
	})
}

func (s SqlLevelStore) GetAllPage(offset int, limit int, order model.ColumnOrder) store.StoreChannel {
	return store.Do(func(result *store.StoreResult) {
		var levels []*model.Level

		query := `SELECT *
                  FROM Levels`
		//ORDER BY ` + order.Column + ` `

		/*if order.Column == "price" { // cuz price is string
			query += `+ 0 ` // hack for sorting string as integer
		}*/

		query += order.Type + ` LIMIT :Limit OFFSET :Offset `

		if _, err := s.GetReplica().Select(&levels, query, map[string]interface{}{"Limit": limit, "Offset": offset}); err != nil {
			result.Err = model.NewAppError("SqlLevelStore.GetAllPage", "store.sql_levels.get_all_page.app_error",
				nil, err.Error(),
				http.StatusInternalServerError)
		} else {

			list := model.NewLevelList()

			for _, p := range levels {
				list.AddLevel(p)
				list.AddOrder(p.Id)
			}

			list.MakeNonNil()

			result.Data = list
		}
	})
}

func (s *SqlLevelStore) Overwrite(level *model.Level) store.StoreChannel {
	return store.Do(func(result *store.StoreResult) {
		level.UpdateAt = model.GetMillis()

		if result.Err = level.IsValid(); result.Err != nil {
			return
		}

		if _, err := s.GetMaster().Update(level); err != nil {
			result.Err = model.NewAppError("SqlLevelStore.Overwrite", "store.sql_level.overwrite.app_error", nil, "id="+level.Id+", "+err.Error(), http.StatusInternalServerError)
		} else {
			result.Data = level
		}
	})
}

func (s *SqlLevelStore) Delete(levelId string, time int64, deleteByID string) store.StoreChannel {
	return store.Do(func(result *store.StoreResult) {

		appErr := func(errMsg string) *model.AppError {
			return model.NewAppError("SqlLevelStore.Delete", "store.sql_level.delete.app_error", nil, "id="+levelId+", err="+errMsg, http.StatusInternalServerError)
		}

		var level model.Level
		err := s.GetReplica().SelectOne(&level, "SELECT * FROM Levels WHERE Id = :Id AND DeleteAt = 0", map[string]interface{}{"Id": levelId})
		if err != nil {
			result.Err = appErr(err.Error())
		}

		_, err = s.GetMaster().Exec("UPDATE Levels SET DeleteAt = :DeleteAt, UpdateAt = :UpdateAt WHERE Id = :Id", map[string]interface{}{"DeleteAt": time, "UpdateAt": time, "Id": levelId})
		if err != nil {
			result.Err = appErr(err.Error())
		}
	})
}

func (s *SqlLevelStore) DeleteApplicationLevels(appId string) store.StoreChannel {
	return store.Do(func(result *store.StoreResult) {

		appErr := func(errMsg string) *model.AppError {
			return model.NewAppError("SqlLevelStore.Delete", "store.sql_level.delete.app_error", nil, "id="+appId+", err="+errMsg, http.StatusInternalServerError)
		}

		/*var level model.Level
		err := s.GetReplica().SelectOne(&level, "SELECT * FROM Levels WHERE Id = :Id AND DeleteAt = 0", map[string]interface{}{"Id": levelId})
		if err != nil {
			result.Err = appErr(err.Error())
		}*/
		time := model.GetMillis()
		_, err := s.GetMaster().Exec("UPDATE Levels SET DeleteAt = :DeleteAt, UpdateAt = :UpdateAt WHERE AppId = :AppId", map[string]interface{}{"DeleteAt": time, "UpdateAt": time, "AppId": appId})
		if err != nil {
			result.Err = appErr(err.Error())
		}
	})
}

func (s SqlLevelStore) GetAllLevels(offset int, limit int, allowFromCache bool, appId *string) store.StoreChannel {
	return store.Do(func(result *store.StoreResult) {
		if limit > 1000 {
			result.Err = model.NewAppError("SqlLevelStore.GetAllLevels", "store.sql_level.get_levels.app_error", nil, "", http.StatusBadRequest)
			return
		}

		var appQuery string
		if appId != nil {
			appQuery = " AND AppId = :AppId "
		} else {
			appQuery = ""
		}

		var levels []*model.Level
		_, err := s.GetReplica().Select(&levels, "SELECT * FROM Levels WHERE "+
			" DeleteAt = 0 "+appQuery+
			" ORDER BY Lvl ASC LIMIT :Limit OFFSET :Offset", map[string]interface{}{"Offset": offset, "Limit": limit, "AppId": appId})

		if err != nil {
			result.Err = model.NewAppError("SqlLevelStore.GetAllLevels", "store.sql_level.get_root_levels.app_error", nil, err.Error(), http.StatusInternalServerError)
		}

		if err == nil {

			list := model.NewLevelList()

			for _, p := range levels {
				list.AddLevel(p)
				list.AddOrder(p.Id)
			}

			list.MakeNonNil()

			result.Data = list
		}
	})
}

func (s SqlLevelStore) GetAllLevelsSince(time int64, allowFromCache bool, appId *string) store.StoreChannel {
	return store.Do(func(result *store.StoreResult) {

		var levels []*model.Level
		_, err := s.GetReplica().Select(&levels,
			`SELECT * FROM Levels WHERE UpdateAt > :Time  ORDER BY Lvl ASC`,
			map[string]interface{}{"Time": time})

		if err != nil {
			result.Err = model.NewAppError("SqlLevelStore.GetAllLevelsSince", "store.sql_level.get_levels_since.app_error", nil, err.Error(), http.StatusInternalServerError)
		} else {

			list := model.NewLevelList()
			var latestUpdate int64 = 0

			for _, p := range levels {
				list.AddLevel(p)
				if p.UpdateAt > time {
					list.AddOrder(p.Id)
				}
				if latestUpdate < p.UpdateAt {
					latestUpdate = p.UpdateAt
				}
			}

			//lastLevelTimeCache.AddWithExpiresInSecs(channelId, latestUpdate, LAST_MESSAGE_TIME_CACHE_SEC)

			result.Data = list
		}
	})
}

func (s SqlLevelStore) GetAllLevelsBefore(levelId string, numLevels int, offset int, appId *string) store.StoreChannel {
	return s.getAllLevelsAround(levelId, numLevels, offset, true, appId)
}

func (s SqlLevelStore) GetAllLevelsAfter(levelId string, numLevels int, offset int, appId *string) store.StoreChannel {
	return s.getAllLevelsAround(levelId, numLevels, offset, false, appId)
}

func (s SqlLevelStore) getAllLevelsAround(levelId string, numLevels int, offset int, before bool, appId *string) store.StoreChannel {
	return store.Do(func(result *store.StoreResult) {
		var direction string
		//var sort string
		if before {
			direction = "<"
			//sort = "DESC"
		} else {
			direction = ">"
			//sort = "ASC"
		}

		var appQuery string
		if appId != nil {
			appQuery = " AND AppId = :AppId "
		} else {
			appQuery = ""
		}

		var levels []*model.Level

		_, err := s.GetReplica().Select(&levels,
			`SELECT * FROM Levels WHERE (CreateAt `+direction+` (SELECT CreateAt FROM Levels WHERE Id = :LevelId)) `+appQuery+`
			ORDER BY Lvl ASC
			OFFSET :Offset LIMIT :NumLevels`,
			map[string]interface{}{"LevelId": levelId, "NumLevels": numLevels, "Offset": offset, "AppId": appId})

		if err != nil {
			result.Err = model.NewAppError("SqlLevelStore.getAllLevelsAround", "store.sql_level.get_levels_around.get.app_error", nil, err.Error(), http.StatusInternalServerError)
		} else {

			list := model.NewLevelList()

			// We need to flip the order if we selected backwards
			if before {
				for _, p := range levels {
					list.AddLevel(p)
					list.AddOrder(p.Id)
				}
			} else {
				l := len(levels)
				for i := range levels {
					list.AddLevel(levels[l-i-1])
					list.AddOrder(levels[l-i-1].Id)
				}
			}

			result.Data = list
		}
	})
}
