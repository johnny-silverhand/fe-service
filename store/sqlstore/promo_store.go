package sqlstore

import (
	"database/sql"
	"im/model"
	"im/store"
	"net/http"
)

type SqlPromoStore struct {
	SqlStore
}

func NewSqlPromoStore(sqlStore SqlStore) store.PromoStore {
	s := &SqlPromoStore{sqlStore}

	for _, db := range sqlStore.GetAllConns() {
		table := db.AddTableWithName(model.Promo{}, "Promos").SetKeys(false, "Id")

		table.ColMap("Id").SetMaxSize(26)

		table.ColMap("Name").SetMaxSize(255)
		table.ColMap("Preview").SetMaxSize(255)
		table.ColMap("Description").SetMaxSize(2000)

	}

	return s
}

func (s SqlPromoStore) CreateIndexesIfNotExists() {
	s.CreateIndexIfNotExists("idx_promos_update_at", "Promos", "UpdateAt")
	s.CreateIndexIfNotExists("idx_promos_create_at", "Promos", "CreateAt")
	s.CreateIndexIfNotExists("idx_promos_delete_at", "Promos", "DeleteAt")
}

func (s SqlPromoStore) Activate(promoId string) store.StoreChannel {
	return store.Do(func(result *store.StoreResult) {

		_, err := s.GetMaster().Exec("UPDATE Promos SET Active = :Active, UpdateAt =:UpdateAt WHERE Id = :Id ", map[string]interface{}{"Active": true, "UpdateAt": model.GetMillis(), "Id": promoId})
		if err != nil {
			result.Err = model.NewAppError("SqlPromoStore.Publish", "store.sql_promos.publish.app_error", nil, err.Error(), http.StatusInternalServerError)

		}

	})
}
func (s SqlPromoStore) Deactivate(promoId string) store.StoreChannel {
	return store.Do(func(result *store.StoreResult) {

		_, err := s.GetMaster().Exec("UPDATE Promos SET Active = :Active, UpdateAt =:UpdateAt WHERE Id = :Id ", map[string]interface{}{"Active": false, "UpdateAt": model.GetMillis(), "Id": promoId})
		if err != nil {
			result.Err = model.NewAppError("SqlPromoStore.Disable", "store.sql_promos.publish.app_error", nil, err.Error(), http.StatusInternalServerError)

		}

	})
}

func (s *SqlPromoStore) Save(promo *model.Promo) store.StoreChannel {
	return store.Do(func(result *store.StoreResult) {
		if len(promo.Id) > 0 {
			result.Err = model.NewAppError("SqlPromoStore.Save", "store.sql_promo.save.existing.app_error", nil, "id="+promo.Id, http.StatusBadRequest)
			return
		}

		promo.PreSave()

		if result.Err = promo.IsValid(); result.Err != nil {
			return
		}

		if err := s.GetMaster().Insert(promo); err != nil {
			result.Err = model.NewAppError("SqlPromoStore.Save", "store.sql_promo.save.app_error", nil, "id="+promo.Id+", "+err.Error(), http.StatusInternalServerError)
		} else {
			result.Data = promo
		}
	})
}

func (s *SqlPromoStore) Update(newPromo *model.Promo) store.StoreChannel {
	return store.Do(func(result *store.StoreResult) {
		newPromo.UpdateAt = model.GetMillis()
		newPromo.PreCommit()

		if _, err := s.GetMaster().Update(newPromo); err != nil {
			result.Err = model.NewAppError("SqlPromoStore.Update", "store.sql_promo.update.app_error", nil, "id="+newPromo.Id+", "+err.Error(), http.StatusInternalServerError)
		} else {
			result.Data = newPromo
		}
	})
}

func (s *SqlPromoStore) Get(id string) store.StoreChannel {
	return store.Do(func(result *store.StoreResult) {
		var promo *model.Promo
		if err := s.GetReplica().SelectOne(&promo,
			`SELECT *
					FROM Promos
					WHERE Id = :Id  AND DeleteAt = 0`, map[string]interface{}{"Id": id}); err != nil {
			if err == sql.ErrNoRows {
				result.Err = model.NewAppError("SqlPromoStore.Get", "store.sql_promos.get.app_error", nil, err.Error(), http.StatusNotFound)
			} else {
				result.Err = model.NewAppError("SqlPromoStore.Get", "store.sql_promos.get.app_error", nil, err.Error(), http.StatusInternalServerError)
			}
		} else {
			result.Data = promo
		}
	})
}

func (s SqlPromoStore) GetAllPage(offset int, limit int, order model.ColumnOrder) store.StoreChannel {
	return store.Do(func(result *store.StoreResult) {
		var promos []*model.Promo

		query := `SELECT *
                  FROM Promos`
		//ORDER BY ` + order.Column + ` `

		/*if order.Column == "price" { // cuz price is string
			query += `+ 0 ` // hack for sorting string as integer
		}*/

		query += order.Type + ` LIMIT :Limit OFFSET :Offset `

		if _, err := s.GetReplica().Select(&promos, query, map[string]interface{}{"Limit": limit, "Offset": offset}); err != nil {
			result.Err = model.NewAppError("SqlPromoStore.GetAllPage", "store.sql_promos.get_all_page.app_error",
				nil, err.Error(),
				http.StatusInternalServerError)
		} else {

			list := model.NewPromoList()

			for _, p := range promos {
				list.AddPromo(p)
				list.AddOrder(p.Id)
			}

			list.MakeNonNil()

			result.Data = list
		}
	})
}

func (s SqlPromoStore) GetAllPageByApp(offset int, limit int, order model.ColumnOrder, appId string) store.StoreChannel {
	return store.Do(func(result *store.StoreResult) {
		var promos []*model.Promo

		query := `SELECT *
                  FROM Promos WHERE AppId = :AppId `

		query += order.Type + ` LIMIT :Limit OFFSET :Offset `

		if _, err := s.GetReplica().Select(&promos, query, map[string]interface{}{"Limit": limit, "Offset": offset, "AppId": appId}); err != nil {
			result.Err = model.NewAppError("SqlPromoStore.GetAllPageByApp", "store.sql_promos.get_all_page_by_app.app_error",
				nil, err.Error(),
				http.StatusInternalServerError)
		} else {

			list := model.NewPromoList()

			for _, p := range promos {
				list.AddPromo(p)
				list.AddOrder(p.Id)
			}

			list.MakeNonNil()

			result.Data = list
		}
	})
}

func (s *SqlPromoStore) Overwrite(promo *model.Promo) store.StoreChannel {
	return store.Do(func(result *store.StoreResult) {
		promo.UpdateAt = model.GetMillis()

		if result.Err = promo.IsValid(); result.Err != nil {
			return
		}

		if _, err := s.GetMaster().Update(promo); err != nil {
			result.Err = model.NewAppError("SqlPromoStore.Overwrite", "store.sql_promo.overwrite.app_error", nil, "id="+promo.Id+", "+err.Error(), http.StatusInternalServerError)
		} else {
			result.Data = promo
		}
	})
}

func (s *SqlPromoStore) Delete(promoId string, time int64, deleteByID string) store.StoreChannel {
	return store.Do(func(result *store.StoreResult) {

		appErr := func(errMsg string) *model.AppError {
			return model.NewAppError("SqlPromoStore.Delete", "store.sql_promo.delete.app_error", nil, "id="+promoId+", err="+errMsg, http.StatusInternalServerError)
		}

		var promo model.Promo
		err := s.GetReplica().SelectOne(&promo, "SELECT * FROM Promos WHERE Id = :Id AND DeleteAt = 0", map[string]interface{}{"Id": promoId})
		if err != nil {
			result.Err = appErr(err.Error())
		}

		_, err = s.GetMaster().Exec("UPDATE Promos SET DeleteAt = :DeleteAt, UpdateAt = :UpdateAt WHERE Id = :Id", map[string]interface{}{"DeleteAt": time, "UpdateAt": time, "Id": promoId})
		if err != nil {
			result.Err = appErr(err.Error())
		}
	})
}

func (s SqlPromoStore) GetAllPromos(offset int, limit int, options *model.PromoGetOptions) store.StoreChannel {
	return store.Do(func(result *store.StoreResult) {

		query := s.getQueryBuilder().
			Select("P.*").
			From("Promos P").
			Where("P.DeleteAt = ?", 0).
			Offset(uint64(options.Page * options.PerPage)).Limit(uint64(options.PerPage))

		/*if options == nil {
			result.Err = model.NewAppError("SqlPromoStore.GetAllPromos", "store.sql_promo.get_promos.app_error", nil, "", http.StatusBadRequest)
			return
		}*/

		/*if limit > 1000 {
			result.Err = model.NewAppError("SqlPromoStore.GetAllPromos", "store.sql_promo.get_promos.app_error", nil, "", http.StatusBadRequest)
			return
		}*/

		if len(options.AppId) > 0 {
			query = query.Where("P.AppId = ? ", options.AppId)
		}

		if len(options.Status) > 0 {
			query = query.Where("P.Status = ? ", options.Status)
		}

		if options.Mobile {
			millis := model.GetMillis()
			query = query.Where("P.Active = ? AND P.BeginAt <= ? AND P.ExpireAt >= ? ", true, millis, millis)
		}

		/*var whereClause string
		queryArgs := make(map[string]interface{})

		if options.AppId != "" {
			whereClause = whereClause + " p.AppId = :AppId AND "
		}

		if options.Status != "" {
			whereClause = whereClause + " p.Status = :Status AND "
		}

		if options.Active != nil {
			whereClause = whereClause + " p.Active = :Active AND "
		}

		if options.Mobile {
			whereClause = whereClause + " p.BeginAt >= :BeginAt AND p.ExpireAt <= :ExpireAt AND "
			queryArgs["BeginAt"] = model.GetMillis()
			queryArgs["ExpireAt"] = model.GetMillis()
		}

		query := "SELECT * FROM Promos p " +
			" WHERE " + whereClause +
			" DeleteAt = 0 " +
			" ORDER BY CreateAt DESC " +
			"LIMIT :Limit OFFSET :Offset"

		queryArgs["Limit"] = limit
		queryArgs["Offset"] = offset
		queryArgs["AppId"] = options.AppId
		queryArgs["Status"] = options.Status
		queryArgs["Active"] = options.Active*/

		queryString, args, err := query.ToSql()
		if err != nil {
			result.Err = model.NewAppError("SqlUserStore.GetAllProfiles", "store.sql_user.app_error", nil, err.Error(), http.StatusInternalServerError)
			return
		}

		var promos []*model.Promo
		if _, err := s.GetReplica().Select(&promos, queryString, args...); err != nil {
			result.Err = model.NewAppError("SqlPromoStore.GetAllPromos", "store.sql_promo.get_root_promos.app_error", nil, err.Error(), http.StatusInternalServerError)
			return
		}

		/*var promos []*model.Promo
		_, err := s.GetReplica().Select(&promos, query, queryArgs)

		if err != nil {
			result.Err = model.NewAppError("SqlPromoStore.GetAllPromos", "store.sql_promo.get_root_promos.app_error", nil, err.Error(), http.StatusInternalServerError)
		}*/

		if err == nil {

			list := model.NewPromoList()

			for _, p := range promos {
				list.AddPromo(p)
				list.AddOrder(p.Id)
			}

			list.MakeNonNil()

			result.Data = list
		}
	})
}

func (s SqlPromoStore) GetAllPromosSince(time int64, options *model.PromoGetOptions) store.StoreChannel {
	return store.Do(func(result *store.StoreResult) {

		if options == nil {
			result.Err = model.NewAppError("SqlPromoStore.GetAllPromosSince", "store.sql_promo.get_promos_since.app_error", nil, "", http.StatusBadRequest)
			return
		}

		appQuery := ""

		if options.AppId != "" {
			appQuery = " AND AppId = :AppId "
		}

		var promos []*model.Promo
		_, err := s.GetReplica().Select(&promos,
			`SELECT * FROM Promos WHERE UpdateAt > :Time `+appQuery+` ORDER BY UpdateAt`,
			map[string]interface{}{"Time": time, "AppId": options.AppId})

		if err != nil {
			result.Err = model.NewAppError("SqlPromoStore.GetAllPromosSince", "store.sql_promo.get_promos_since.app_error", nil, err.Error(), http.StatusInternalServerError)
		} else {

			list := model.NewPromoList()
			var latestUpdate int64 = 0

			for _, p := range promos {
				list.AddPromo(p)
				if p.UpdateAt > time {
					list.AddOrder(p.Id)
				}
				if latestUpdate < p.UpdateAt {
					latestUpdate = p.UpdateAt
				}
			}

			//lastPromoTimeCache.AddWithExpiresInSecs(channelId, latestUpdate, LAST_MESSAGE_TIME_CACHE_SEC)

			result.Data = list
		}
	})
}

func (s SqlPromoStore) GetAllPromosBefore(promoId string, numPromos int, offset int, options *model.PromoGetOptions) store.StoreChannel {
	return s.getAllPromosAround(promoId, numPromos, offset, true, options)
}

func (s SqlPromoStore) GetAllPromosAfter(promoId string, numPromos int, offset int, options *model.PromoGetOptions) store.StoreChannel {
	return s.getAllPromosAround(promoId, numPromos, offset, false, options)
}

func (s SqlPromoStore) getAllPromosAround(promoId string, numPromos int, offset int, before bool, options *model.PromoGetOptions) store.StoreChannel {
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

		if options == nil {
			result.Err = model.NewAppError("SqlPromoStore.getAllPromosAround", "store.sql_promo.get_promos_around.get.app_error", nil, "", http.StatusBadRequest)
			return
		}

		appQuery := ""

		if options.AppId != "" {
			appQuery = " AND AppId = :AppId "
		}

		var promos []*model.Promo

		_, err := s.GetReplica().Select(&promos,
			`SELECT
			    *
			FROM
			    Promos
			WHERE (CreateAt `+direction+` (SELECT CreateAt FROM Promos WHERE Id = :PromoId)) `+appQuery+`
			ORDER BY CreateAt `+sort+`
			OFFSET :Offset LIMIT :NumPromos`,
			map[string]interface{}{"PromoId": promoId, "NumPromos": numPromos, "Offset": offset, "AppId": options.AppId})

		if err != nil {
			result.Err = model.NewAppError("SqlPromoStore.getAllPromosAround", "store.sql_promo.get_promos_around.get.app_error", nil, err.Error(), http.StatusInternalServerError)
		} else {

			list := model.NewPromoList()

			// We need to flip the order if we selected backwards
			if before {
				for _, p := range promos {
					list.AddPromo(p)
					list.AddOrder(p.Id)
				}
			} else {
				l := len(promos)
				for i := range promos {
					list.AddPromo(promos[l-i-1])
					list.AddOrder(promos[l-i-1].Id)
				}
			}

			result.Data = list
		}
	})
}

func (s SqlPromoStore) GetForModeration(options *model.PromoGetOptions) store.StoreChannel {
	return store.Do(func(result *store.StoreResult) {
		if options == nil {
			result.Err = model.NewAppError("SqlPromoStore.GetForModeration", "store.sql_promo.promo_get_options.app_error", nil, "", http.StatusBadRequest)
			return
		}

		var whereClause string
		queryArgs := make(map[string]interface{})

		if options.AppId != "" {
			whereClause = whereClause + " p.AppId = :AppId AND "
		}

		//if options.Status != "" {
		whereClause = whereClause + " p.Status = :Status AND "
		//}

		/*if options.Active != nil {
			whereClause = whereClause + " p.Active = :Active AND "
		}*/

		query := "SELECT * FROM Promos p " +
			" WHERE " + whereClause +
			" DeleteAt = 0 " +
			" ORDER BY CreateAt DESC "

		queryArgs["AppId"] = options.AppId
		queryArgs["Status"] = model.PROMO_STATUS_MODERATION
		//queryArgs["Active"] = options.Active

		var promos []*model.Promo
		_, err := s.GetReplica().Select(&promos, query, queryArgs)

		if err != nil {
			result.Err = model.NewAppError("SqlPromoStore.GetForModeration", "store.sql_promo.get_for_moderation.app_error", nil, err.Error(), http.StatusInternalServerError)
		}

		if err == nil {

			list := model.NewPromoList()

			for _, p := range promos {
				list.AddPromo(p)
				list.AddOrder(p.Id)
			}

			list.MakeNonNil()

			result.Data = list
		}
	})
}
