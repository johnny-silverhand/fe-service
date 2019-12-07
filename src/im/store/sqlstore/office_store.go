package sqlstore

import (
	"database/sql"
	"im/model"
	"im/store"
	"net/http"
)

type SqlOfficeStore struct {
	SqlStore
}

func NewSqlOfficeStore(sqlStore SqlStore) store.OfficeStore {
	s := &SqlOfficeStore{sqlStore}

	for _, db := range sqlStore.GetAllConns() {
		table := db.AddTableWithName(model.Office{}, "Offices").SetKeys(false, "Id")

		table.ColMap("Id").SetMaxSize(26)
		table.ColMap("Name").SetMaxSize(255)
		table.ColMap("Preview").SetMaxSize(255)
		table.ColMap("Description").SetMaxSize(2000)

	}

	return s
}

func (s SqlOfficeStore) CreateIndexesIfNotExists() {
	s.CreateIndexIfNotExists("idx_offices_update_at", "Offices", "UpdateAt")
	s.CreateIndexIfNotExists("idx_offices_create_at", "Offices", "CreateAt")
	s.CreateIndexIfNotExists("idx_offices_delete_at", "Offices", "DeleteAt")
}

func (s SqlOfficeStore) Activate(officeId string) store.StoreChannel {
	return store.Do(func(result *store.StoreResult) {

		_, err := s.GetMaster().Exec("UPDATE Offices SET Active = :Active, UpdateAt =:UpdateAt WHERE Id = :Id ", map[string]interface{}{"Active": true, "UpdateAt": model.GetMillis(), "Id": officeId})
		if err != nil {
			result.Err = model.NewAppError("SqlOfficeStore.Publish", "store.sql_offices.publish.app_error", nil, err.Error(), http.StatusInternalServerError)

		}

	})
}
func (s SqlOfficeStore) Deactivate(officeId string) store.StoreChannel {
	return store.Do(func(result *store.StoreResult) {

		_, err := s.GetMaster().Exec("UPDATE Offices SET Active = :Active, UpdateAt =:UpdateAt WHERE Id = :Id ", map[string]interface{}{"Active": false, "UpdateAt": model.GetMillis(), "Id": officeId})
		if err != nil {
			result.Err = model.NewAppError("SqlOfficeStore.Disable", "store.sql_offices.publish.app_error", nil, err.Error(), http.StatusInternalServerError)

		}

	})
}

func (s *SqlOfficeStore) Save(office *model.Office) store.StoreChannel {
	return store.Do(func(result *store.StoreResult) {
		if len(office.Id) > 0 {
			result.Err = model.NewAppError("SqlOfficeStore.Save", "store.sql_office.save.existing.app_error", nil, "id="+office.Id, http.StatusBadRequest)
			return
		}

		office.PreSave()

		if result.Err = office.IsValid(); result.Err != nil {
			return
		}

		if err := s.GetMaster().Insert(office); err != nil {
			result.Err = model.NewAppError("SqlOfficeStore.Save", "store.sql_office.save.app_error", nil, "id="+office.Id+", "+err.Error(), http.StatusInternalServerError)
		} else {
			result.Data = office
		}
	})
}

func (s *SqlOfficeStore) Update(newOffice *model.Office) store.StoreChannel {
	return store.Do(func(result *store.StoreResult) {
		newOffice.UpdateAt = model.GetMillis()
		newOffice.PreCommit()

		if _, err := s.GetMaster().Update(newOffice); err != nil {
			result.Err = model.NewAppError("SqlOfficeStore.Update", "store.sql_office.update.app_error", nil, "id="+newOffice.Id+", "+err.Error(), http.StatusInternalServerError)
		} else {
			result.Data = newOffice
		}
	})
}

func (s *SqlOfficeStore) Get(id string) store.StoreChannel {
	return store.Do(func(result *store.StoreResult) {
		var office *model.Office
		if err := s.GetReplica().SelectOne(&office,
			`SELECT *
					FROM Offices
					WHERE Id = :Id  AND DeleteAt = 0`, map[string]interface{}{"Id": id}); err != nil {
			if err == sql.ErrNoRows {
				result.Err = model.NewAppError("SqlOfficeStore.Get", "store.sql_offices.get.app_error", nil, err.Error(), http.StatusNotFound)
			} else {
				result.Err = model.NewAppError("SqlOfficeStore.Get", "store.sql_offices.get.app_error", nil, err.Error(), http.StatusInternalServerError)
			}
		} else {
			result.Data = office
		}
	})
}

func (s SqlOfficeStore) GetAllPage(offset int, limit int, order model.ColumnOrder) store.StoreChannel {
	return store.Do(func(result *store.StoreResult) {
		var offices []*model.Office

		query := `SELECT *
                  FROM Offices`
		//ORDER BY ` + order.Column + ` `

		/*if order.Column == "price" { // cuz price is string
			query += `+ 0 ` // hack for sorting string as integer
		}*/

		query += order.Type + ` LIMIT :Limit OFFSET :Offset `

		if _, err := s.GetReplica().Select(&offices, query, map[string]interface{}{"Limit": limit, "Offset": offset}); err != nil {
			result.Err = model.NewAppError("SqlOfficeStore.GetAllPage", "store.sql_offices.get_all_page.app_error",
				nil, err.Error(),
				http.StatusInternalServerError)
		} else {

			list := model.NewOfficeList()

			for _, p := range offices {
				list.AddOffice(p)
				list.AddOrder(p.Id)
			}

			list.MakeNonNil()

			result.Data = list
		}
	})
}

func (s *SqlOfficeStore) Overwrite(office *model.Office) store.StoreChannel {
	return store.Do(func(result *store.StoreResult) {
		office.UpdateAt = model.GetMillis()

		if result.Err = office.IsValid(); result.Err != nil {
			return
		}

		if _, err := s.GetMaster().Update(office); err != nil {
			result.Err = model.NewAppError("SqlOfficeStore.Overwrite", "store.sql_office.overwrite.app_error", nil, "id="+office.Id+", "+err.Error(), http.StatusInternalServerError)
		} else {
			result.Data = office
		}
	})
}

func (s *SqlOfficeStore) Delete(officeId string, time int64, deleteByID string) store.StoreChannel {
	return store.Do(func(result *store.StoreResult) {

		appErr := func(errMsg string) *model.AppError {
			return model.NewAppError("SqlOfficeStore.Delete", "store.sql_office.delete.app_error", nil, "id="+officeId+", err="+errMsg, http.StatusInternalServerError)
		}

		var office model.Office
		err := s.GetReplica().SelectOne(&office, "SELECT * FROM Offices WHERE Id = :Id AND DeleteAt = 0", map[string]interface{}{"Id": officeId})
		if err != nil {
			result.Err = appErr(err.Error())
		}

		_, err = s.GetMaster().Exec("UPDATE Offices SET DeleteAt = :DeleteAt, UpdateAt = :UpdateAt WHERE Id = :Id", map[string]interface{}{"DeleteAt": time, "UpdateAt": time, "Id": officeId})
		if err != nil {
			result.Err = appErr(err.Error())
		}
	})
}

func (s SqlOfficeStore) GetAllOffices(offset int, limit int, allowFromCache bool, appId *string) store.StoreChannel {
	return store.Do(func(result *store.StoreResult) {
		if limit > 1000 {
			result.Err = model.NewAppError("SqlOfficeStore.GetAllOffices", "store.sql_office.get_offices.app_error", nil, "", http.StatusBadRequest)
			return
		}

		var appQuery string
		if appId != nil {
			appQuery = " AND AppId = :AppId "
		} else {
			appQuery = ""
		}

		var offices []*model.Office
		_, err := s.GetReplica().Select(&offices, "SELECT * FROM Offices WHERE "+
			" DeleteAt = 0 "+appQuery+
			" ORDER BY CreateAt DESC LIMIT :Limit OFFSET :Offset", map[string]interface{}{"Offset": offset, "Limit": limit, "AppId": appId})

		if err != nil {
			result.Err = model.NewAppError("SqlOfficeStore.GetAllOffices", "store.sql_office.get_root_offices.app_error", nil, err.Error(), http.StatusInternalServerError)
		}

		if err == nil {

			list := model.NewOfficeList()

			for _, p := range offices {
				list.AddOffice(p)
				list.AddOrder(p.Id)
			}

			list.MakeNonNil()

			result.Data = list
		}
	})
}

func (s SqlOfficeStore) GetAllOfficesSince(time int64, allowFromCache bool, appId *string) store.StoreChannel {
	return store.Do(func(result *store.StoreResult) {

		var offices []*model.Office
		_, err := s.GetReplica().Select(&offices,
			`SELECT * FROM Offices WHERE UpdateAt > :Time  ORDER BY UpdateAt`,
			map[string]interface{}{"Time": time})

		if err != nil {
			result.Err = model.NewAppError("SqlOfficeStore.GetAllOfficesSince", "store.sql_office.get_offices_since.app_error", nil, err.Error(), http.StatusInternalServerError)
		} else {

			list := model.NewOfficeList()
			var latestUpdate int64 = 0

			for _, p := range offices {
				list.AddOffice(p)
				if p.UpdateAt > time {
					list.AddOrder(p.Id)
				}
				if latestUpdate < p.UpdateAt {
					latestUpdate = p.UpdateAt
				}
			}

			//lastOfficeTimeCache.AddWithExpiresInSecs(channelId, latestUpdate, LAST_MESSAGE_TIME_CACHE_SEC)

			result.Data = list
		}
	})
}

func (s SqlOfficeStore) GetAllOfficesBefore(officeId string, numOffices int, offset int, appId *string) store.StoreChannel {
	return s.getAllOfficesAround(officeId, numOffices, offset, true, appId)
}

func (s SqlOfficeStore) GetAllOfficesAfter(officeId string, numOffices int, offset int, appId *string) store.StoreChannel {
	return s.getAllOfficesAround(officeId, numOffices, offset, false, appId)
}

func (s SqlOfficeStore) getAllOfficesAround(officeId string, numOffices int, offset int, before bool, appId *string) store.StoreChannel {
	return store.Do(func(result *store.StoreResult) {
		var direction string
		var sort string
		if before {
			direction = "<"
			sort = "DESC"
		} else {
			direction = ">"
			sort = "ASC"
		}

		var appQuery string
		if appId != nil {
			appQuery = " AND AppId = :AppId "
		} else {
			appQuery = ""
		}

		var offices []*model.Office

		_, err := s.GetReplica().Select(&offices,
			`SELECT * FROM Offices WHERE (CreateAt `+direction+` (SELECT CreateAt FROM Offices WHERE Id = :OfficeId))`+appQuery+
				`ORDER BY CreateAt `+sort+`
			OFFSET :Offset LIMIT :NumOffices`,
			map[string]interface{}{"OfficeId": officeId, "NumOffices": numOffices, "Offset": offset, "AppId": appId})

		if err != nil {
			result.Err = model.NewAppError("SqlOfficeStore.getAllOfficesAround", "store.sql_office.get_offices_around.get.app_error", nil, err.Error(), http.StatusInternalServerError)
		} else {

			list := model.NewOfficeList()

			// We need to flip the order if we selected backwards
			if before {
				for _, p := range offices {
					list.AddOffice(p)
					list.AddOrder(p.Id)
				}
			} else {
				l := len(offices)
				for i := range offices {
					list.AddOffice(offices[l-i-1])
					list.AddOrder(offices[l-i-1].Id)
				}
			}

			result.Data = list
		}
	})
}
