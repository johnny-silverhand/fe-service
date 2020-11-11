package sqlstore

import (
	"database/sql"
	"im/model"
	"im/store"
	"net/http"
)

type SqlApplicationStore struct {
	SqlStore
}

func NewSqlApplicationStore(sqlStore SqlStore) store.ApplicationStore {
	s := &SqlApplicationStore{sqlStore}

	for _, db := range sqlStore.GetAllConns() {
		table := db.AddTableWithName(model.Application{}, "Applications").SetKeys(false, "Id")

		table.ColMap("Id").SetMaxSize(26)

	}

	return s
}

func (s SqlApplicationStore) CreateIndexesIfNotExists() {
	s.CreateIndexIfNotExists("idx_applications_update_at", "Applications", "UpdateAt")
	s.CreateIndexIfNotExists("idx_applications_create_at", "Applications", "CreateAt")
	s.CreateIndexIfNotExists("idx_applications_delete_at", "Applications", "DeleteAt")
}

func (s SqlApplicationStore) Activate(appId string) store.StoreChannel {
	return store.Do(func(result *store.StoreResult) {

		_, err := s.GetMaster().Exec("UPDATE Applications SET Active = :Active, UpdateAt =:UpdateAt WHERE Id = :Id ",
			map[string]interface{}{
				"Active":   true,
				"UpdateAt": model.GetMillis(),
				"Id":       appId})
		if err != nil {
			result.Err = model.NewAppError("SqlApplicationStore.Publish", "store.sql_applications.publish.app_error", nil, err.Error(), http.StatusInternalServerError)

		}

	})
}
func (s SqlApplicationStore) Deactivate(appId string) store.StoreChannel {
	return store.Do(func(result *store.StoreResult) {

		_, err := s.GetMaster().Exec("UPDATE Applications SET Active = :Active, UpdateAt =:UpdateAt WHERE Id = :Id ",
			map[string]interface{}{
				"Active":   false,
				"UpdateAt": model.GetMillis(),
				"Id":       appId})
		if err != nil {
			result.Err = model.NewAppError("SqlApplicationStore.Disable", "store.sql_applications.publish.app_error", nil, err.Error(), http.StatusInternalServerError)

		}

	})
}

func (s *SqlApplicationStore) Save(application *model.Application) store.StoreChannel {
	return store.Do(func(result *store.StoreResult) {

		if len(application.Id) > 0 {
			result.Err = model.NewAppError("SqlApplicationStore.Save", "store.sql_application.save.existing.app_error", nil, "id="+application.Id, http.StatusBadRequest)
			return
		}

		application.PreSave()

		if result.Err = application.IsValid(); result.Err != nil {
			return
		}

		if err := s.GetMaster().Insert(application); err != nil {
			result.Err = model.NewAppError("SqlApplicationStore.Save", "store.sql_application.save.app_error", nil, "id="+application.Id+", "+err.Error(), http.StatusInternalServerError)
		} else {
			result.Data = application
		}
	})
}

func (s *SqlApplicationStore) Update(newApplication *model.Application) store.StoreChannel {
	return store.Do(func(result *store.StoreResult) {
		newApplication.UpdateAt = model.GetMillis()
		newApplication.PreCommit()

		if _, err := s.GetMaster().Update(newApplication); err != nil {
			result.Err = model.NewAppError("SqlApplicationStore.Update", "store.sql_application.update.app_error", nil, "id="+newApplication.Id+", "+err.Error(), http.StatusInternalServerError)
		} else {
			result.Data = newApplication
		}
	})
}

func (s *SqlApplicationStore) Get(id string) store.StoreChannel {
	return store.Do(func(result *store.StoreResult) {
		var application *model.Application
		if err := s.GetReplica().SelectOne(&application,
			`SELECT *
					FROM Applications
					WHERE Id = :Id  AND DeleteAt = 0`, map[string]interface{}{"Id": id}); err != nil {
			if err == sql.ErrNoRows {
				result.Err = model.NewAppError("SqlApplicationStore.Get", "store.sql_applications.get.app_error", nil, err.Error(), http.StatusNotFound)
			} else {
				result.Err = model.NewAppError("SqlApplicationStore.Get", "store.sql_applications.get.app_error", nil, err.Error(), http.StatusInternalServerError)
			}
		} else {
			result.Data = application
		}
	})
}

func (s SqlApplicationStore) GetAllPage(offset int, limit int, order model.ColumnOrder) store.StoreChannel {
	return store.Do(func(result *store.StoreResult) {
		var applications []*model.Application

		query := `SELECT *
                  FROM Applications`

		query += order.Type + ` LIMIT :Limit OFFSET :Offset `

		if _, err := s.GetReplica().Select(&applications, query, map[string]interface{}{"Limit": limit, "Offset": offset}); err != nil {
			result.Err = model.NewAppError("SqlApplicationStore.GetAllPage", "store.sql_applications.get_all_page.app_error",
				nil, err.Error(),
				http.StatusInternalServerError)
		} else {

			list := model.NewApplicationList()

			for _, p := range applications {
				list.AddApplication(p)
				list.AddOrder(p.Id)
			}

			list.MakeNonNil()

			result.Data = list
		}
	})
}

func (s *SqlApplicationStore) Overwrite(application *model.Application) store.StoreChannel {
	return store.Do(func(result *store.StoreResult) {
		application.UpdateAt = model.GetMillis()

		if result.Err = application.IsValid(); result.Err != nil {
			return
		}

		if _, err := s.GetMaster().Update(application); err != nil {
			result.Err = model.NewAppError("SqlApplicationStore.Overwrite", "store.sql_application.overwrite.app_error", nil, "id="+application.Id+", "+err.Error(), http.StatusInternalServerError)
		} else {
			result.Data = application
		}
	})
}

func (s *SqlApplicationStore) Delete(appId string, time int64, deleteByID string) store.StoreChannel {
	return store.Do(func(result *store.StoreResult) {

		appErr := func(errMsg string) *model.AppError {
			return model.NewAppError("SqlApplicationStore.Delete", "store.sql_application.delete.app_error", nil, "id="+appId+", err="+errMsg, http.StatusInternalServerError)
		}

		var application model.Application
		err := s.GetReplica().SelectOne(&application, "SELECT * FROM Applications WHERE Id = :Id AND DeleteAt = 0", map[string]interface{}{"Id": appId})
		if err != nil {
			result.Err = appErr(err.Error())
		}

		_, err = s.GetMaster().Exec("UPDATE Applications SET DeleteAt = :DeleteAt, UpdateAt = :UpdateAt WHERE Id = :Id", map[string]interface{}{"DeleteAt": time, "UpdateAt": time, "Id": appId})
		if err != nil {
			result.Err = appErr(err.Error())
		}
	})
}

func (s SqlApplicationStore) GetAllApplications(offset int, limit int, allowFromCache bool) store.StoreChannel {
	return store.Do(func(result *store.StoreResult) {
		if limit > 1000 {
			result.Err = model.NewAppError("SqlApplicationStore.GetAllApplications", "store.sql_application.get_applications.app_error", nil, "", http.StatusBadRequest)
			return
		}

		var applications []*model.Application
		_, err := s.GetReplica().Select(&applications, "SELECT * FROM Applications WHERE "+
			" DeleteAt = 0 "+
			" ORDER BY CreateAt DESC LIMIT :Limit OFFSET :Offset", map[string]interface{}{"Offset": offset, "Limit": limit})

		if err != nil {
			result.Err = model.NewAppError("SqlApplicationStore.GetAllApplications", "store.sql_application.get_root_applications.app_error", nil, err.Error(), http.StatusInternalServerError)
		}

		if err == nil {

			list := model.NewApplicationList()

			for _, p := range applications {
				list.AddApplication(p)
				list.AddOrder(p.Id)
			}

			list.MakeNonNil()

			result.Data = list
		}
	})
}

func (s SqlApplicationStore) GetAllApplicationsSince(time int64, allowFromCache bool) store.StoreChannel {
	return store.Do(func(result *store.StoreResult) {

		var applications []*model.Application
		_, err := s.GetReplica().Select(&applications,
			`SELECT * FROM Applications WHERE UpdateAt > :Time  ORDER BY UpdateAt`,
			map[string]interface{}{"Time": time})

		if err != nil {
			result.Err = model.NewAppError("SqlApplicationStore.GetAllApplicationsSince", "store.sql_application.get_applications_since.app_error", nil, err.Error(), http.StatusInternalServerError)
		} else {

			list := model.NewApplicationList()
			var latestUpdate int64 = 0

			for _, p := range applications {
				list.AddApplication(p)
				if p.UpdateAt > time {
					list.AddOrder(p.Id)
				}
				if latestUpdate < p.UpdateAt {
					latestUpdate = p.UpdateAt
				}
			}

			//lastApplicationTimeCache.AddWithExpiresInSecs(channelId, latestUpdate, LAST_MESSAGE_TIME_CACHE_SEC)

			result.Data = list
		}
	})
}

func (s SqlApplicationStore) GetAllApplicationsBefore(appId string, numApplications int, offset int) store.StoreChannel {
	return s.getAllApplicationsAround(appId, numApplications, offset, true)
}

func (s SqlApplicationStore) GetAllApplicationsAfter(appId string, numApplications int, offset int) store.StoreChannel {
	return s.getAllApplicationsAround(appId, numApplications, offset, false)
}

func (s SqlApplicationStore) getAllApplicationsAround(appId string, numApplications int, offset int, before bool) store.StoreChannel {
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

		var applications []*model.Application

		_, err := s.GetReplica().Select(&applications,
			`SELECT
			    *
			FROM
			    Applications
			WHERE (CreateAt `+direction+` (SELECT CreateAt FROM Applications WHERE Id = :ApplicationId))
			ORDER BY CreateAt `+sort+`
			OFFSET :Offset LIMIT :NumApplications`,
			map[string]interface{}{"ApplicationId": appId, "NumApplications": numApplications, "Offset": offset})

		if err != nil {
			result.Err = model.NewAppError("SqlApplicationStore.getAllApplicationsAround", "store.sql_application.get_applications_around.get.app_error", nil, err.Error(), http.StatusInternalServerError)
		} else {

			list := model.NewApplicationList()

			// We need to flip the order if we selected backwards
			if before {
				for _, p := range applications {
					list.AddApplication(p)
					list.AddOrder(p.Id)
				}
			} else {
				l := len(applications)
				for i := range applications {
					list.AddApplication(applications[l-i-1])
					list.AddOrder(applications[l-i-1].Id)
				}
			}

			result.Data = list
		}
	})
}

func (s SqlApplicationStore) GetApplications(options *model.ApplicationGetOptions) store.StoreChannel {
	return store.Do(func(result *store.StoreResult) {
		query := s.getQueryBuilder().
			Select("A.*").
			From("Applications A").
			Where("A.DeleteAt = ?", 0).
			Offset(uint64(options.Page * options.PerPage)).Limit(uint64(options.PerPage))

		if len(options.Email) > 0 {
			query = query.Where("A.Email = ? ", options.Email)
		}

		queryString, args, err := query.ToSql()
		if err != nil {
			result.Err = model.NewAppError("SqlUserStore.GetAllProfiles", "store.sql_user.app_error", nil, err.Error(), http.StatusInternalServerError)
			return
		}

		var applications []*model.Application
		if _, err := s.GetReplica().Select(&applications, queryString, args...); err != nil {
			result.Err = model.NewAppError("SqlApplicationStore.GetAllApplications", "store.sql_application.get_root_applications.app_error", nil, err.Error(), http.StatusInternalServerError)
			return
		}

		/*var applications []*model.Application
		_, err := s.GetReplica().Select(&applications, "SELECT * FROM Applications WHERE "+
			" DeleteAt = 0 "+
			" ORDER BY CreateAt DESC LIMIT :Limit OFFSET :Offset", map[string]interface{}{"Offset": offset, "Limit": limit})

		if err != nil {
			result.Err = model.NewAppError("SqlApplicationStore.GetAllApplications", "store.sql_application.get_root_applications.app_error", nil, err.Error(), http.StatusInternalServerError)
		}*/

		if err == nil {

			list := model.NewApplicationList()

			for _, p := range applications {
				list.AddApplication(p)
				list.AddOrder(p.Id)
			}

			list.MakeNonNil()

			result.Data = list
		}
	})
}
