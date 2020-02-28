package sqlstore

import (
	"database/sql"
	"im/model"
	"im/store"
	"net/http"
)

type SqlExtraStore struct {
	SqlStore
}

func NewSqlExtraStore(sqlStore SqlStore) store.ExtraStore {
	s := &SqlExtraStore{sqlStore}

	for _, db := range sqlStore.GetAllConns() {
		table := db.AddTableWithName(model.Extra{}, "Extras").SetKeys(false, "Id")

		table.ColMap("Id").SetMaxSize(26)

	}

	return s
}

func (s SqlExtraStore) CreateIndexesIfNotExists() {
	s.CreateIndexIfNotExists("idx_extras_update_at", "Extras", "UpdateAt")
	s.CreateIndexIfNotExists("idx_extras_create_at", "Extras", "CreateAt")
	s.CreateIndexIfNotExists("idx_extras_delete_at", "Extras", "DeleteAt")
}

func (s *SqlExtraStore) Save(extra *model.Extra) store.StoreChannel {
	return store.Do(func(result *store.StoreResult) {
		if len(extra.Id) > 0 {
			result.Err = model.NewAppError("SqlExtraStore.Save", "store.sql_extra.save.existing.app_error", nil, "id="+extra.Id, http.StatusBadRequest)
			return
		}

		extra.PreSave()

		if result.Err = extra.IsValid(); result.Err != nil {
			return
		}

		if err := s.GetMaster().Insert(extra); err != nil {
			result.Err = model.NewAppError("SqlExtraStore.Save", "store.sql_extra.save.app_error", nil, "id="+extra.Id+", "+err.Error(), http.StatusInternalServerError)
		} else {
			result.Data = extra
		}
	})
}

func (s *SqlExtraStore) Update(newExtra *model.Extra) store.StoreChannel {
	return store.Do(func(result *store.StoreResult) {
		newExtra.UpdateAt = model.GetMillis()
		newExtra.PreCommit()

		if _, err := s.GetMaster().Update(newExtra); err != nil {
			result.Err = model.NewAppError("SqlExtraStore.Update", "store.sql_extra.update.app_error", nil, "id="+newExtra.Id+", "+err.Error(), http.StatusInternalServerError)
		} else {
			result.Data = newExtra
		}
	})
}

func (s *SqlExtraStore) Get(id string) store.StoreChannel {
	return store.Do(func(result *store.StoreResult) {
		var extra *model.Extra
		if err := s.GetReplica().SelectOne(&extra,
			`SELECT *
					FROM Extras
					WHERE Id = :Id  AND DeleteAt = 0`, map[string]interface{}{"Id": id}); err != nil {
			if err == sql.ErrNoRows {
				result.Err = model.NewAppError("SqlExtraStore.Get", "store.sql_extras.get.app_error", nil, err.Error(), http.StatusNotFound)
			} else {
				result.Err = model.NewAppError("SqlExtraStore.Get", "store.sql_extras.get.app_error", nil, err.Error(), http.StatusInternalServerError)
			}
		} else {
			result.Data = extra
		}
	})
}

func (s SqlExtraStore) GetAllPage(offset int, limit int, order model.ColumnOrder) store.StoreChannel {
	return store.Do(func(result *store.StoreResult) {
		var extras []*model.Extra

		query := `SELECT *
                  FROM Extras`
		//ORDER BY ` + order.Column + ` `

		/*if order.Column == "price" { // cuz price is string
			query += `+ 0 ` // hack for sorting string as integer
		}*/

		query += order.Type + ` LIMIT :Limit OFFSET :Offset `

		if _, err := s.GetReplica().Select(&extras, query, map[string]interface{}{"Limit": limit, "Offset": offset}); err != nil {
			result.Err = model.NewAppError("SqlExtraStore.GetAllPage", "store.sql_extras.get_all_page.app_error",
				nil, err.Error(),
				http.StatusInternalServerError)
		} else {

			list := model.NewExtraList()

			for _, p := range extras {
				list.AddExtra(p)
				list.AddOrder(p.Id)
			}

			list.MakeNonNil()

			result.Data = list
		}
	})
}

func (s *SqlExtraStore) Overwrite(extra *model.Extra) store.StoreChannel {
	return store.Do(func(result *store.StoreResult) {
		extra.UpdateAt = model.GetMillis()

		if result.Err = extra.IsValid(); result.Err != nil {
			return
		}

		if _, err := s.GetMaster().Update(extra); err != nil {
			result.Err = model.NewAppError("SqlExtraStore.Overwrite", "store.sql_extra.overwrite.app_error", nil, "id="+extra.Id+", "+err.Error(), http.StatusInternalServerError)
		} else {
			result.Data = extra
		}
	})
}

func (s *SqlExtraStore) Delete(extraId string, time int64, deleteByID string) store.StoreChannel {
	return store.Do(func(result *store.StoreResult) {

		appErr := func(errMsg string) *model.AppError {
			return model.NewAppError("SqlExtraStore.Delete", "store.sql_extra.delete.app_error", nil, "id="+extraId+", err="+errMsg, http.StatusInternalServerError)
		}

		var extra model.Extra
		err := s.GetReplica().SelectOne(&extra, "SELECT * FROM Extras WHERE Id = :Id AND DeleteAt = 0", map[string]interface{}{"Id": extraId})
		if err != nil {
			result.Err = appErr(err.Error())
		}

		_, err = s.GetMaster().Exec("UPDATE Extras SET DeleteAt = :DeleteAt, UpdateAt = :UpdateAt WHERE Id = :Id", map[string]interface{}{"DeleteAt": time, "UpdateAt": time, "Id": extraId})
		if err != nil {
			result.Err = appErr(err.Error())
		}
	})
}

func (s SqlExtraStore) GetAllExtras(offset int, limit int, allowFromCache bool) store.StoreChannel {
	return store.Do(func(result *store.StoreResult) {
		if limit > 1000 {
			result.Err = model.NewAppError("SqlExtraStore.GetAllExtras", "store.sql_extra.get_extras.app_error", nil, "", http.StatusBadRequest)
			return
		}

		var extras []*model.Extra
		_, err := s.GetReplica().Select(&extras, "SELECT * FROM Extras WHERE "+
			" DeleteAt = 0 "+
			" ORDER BY CreateAt DESC LIMIT :Limit OFFSET :Offset", map[string]interface{}{"Offset": offset, "Limit": limit})

		if err != nil {
			result.Err = model.NewAppError("SqlExtraStore.GetAllExtras", "store.sql_extra.get_root_extras.app_error", nil, err.Error(), http.StatusInternalServerError)
		}

		if err == nil {

			list := model.NewExtraList()

			for _, p := range extras {
				list.AddExtra(p)
				list.AddOrder(p.Id)
			}

			list.MakeNonNil()

			result.Data = list
		}
	})
}

func (s SqlExtraStore) GetAllExtrasSince(time int64, allowFromCache bool) store.StoreChannel {
	return store.Do(func(result *store.StoreResult) {

		var extras []*model.Extra
		_, err := s.GetReplica().Select(&extras,
			`SELECT * FROM Extras WHERE UpdateAt > :Time  ORDER BY UpdateAt`,
			map[string]interface{}{"Time": time})

		if err != nil {
			result.Err = model.NewAppError("SqlExtraStore.GetAllExtrasSince", "store.sql_extra.get_extras_since.app_error", nil, err.Error(), http.StatusInternalServerError)
		} else {

			list := model.NewExtraList()
			var latestUpdate int64 = 0

			for _, p := range extras {
				list.AddExtra(p)
				if p.UpdateAt > time {
					list.AddOrder(p.Id)
				}
				if latestUpdate < p.UpdateAt {
					latestUpdate = p.UpdateAt
				}
			}

			//lastExtraTimeCache.AddWithExpiresInSecs(channelId, latestUpdate, LAST_MESSAGE_TIME_CACHE_SEC)

			result.Data = list
		}
	})
}

func (s SqlExtraStore) GetAllExtrasBefore(extraId string, numExtras int, offset int) store.StoreChannel {
	return s.getAllExtrasAround(extraId, numExtras, offset, true)
}

func (s SqlExtraStore) GetAllExtrasAfter(extraId string, numExtras int, offset int) store.StoreChannel {
	return s.getAllExtrasAround(extraId, numExtras, offset, false)
}

func (s SqlExtraStore) getAllExtrasAround(extraId string, numExtras int, offset int, before bool) store.StoreChannel {
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

		var extras []*model.Extra

		_, err := s.GetReplica().Select(&extras,
			`SELECT
			    *
			FROM
			    Extras
			WHERE (CreateAt `+direction+` (SELECT CreateAt FROM Extras WHERE Id = :ExtraId))
			ORDER BY CreateAt `+sort+`
			OFFSET :Offset LIMIT :NumExtras`,
			map[string]interface{}{"ExtraId": extraId, "NumExtras": numExtras, "Offset": offset})

		if err != nil {
			result.Err = model.NewAppError("SqlExtraStore.getAllExtrasAround", "store.sql_extra.get_extras_around.get.app_error", nil, err.Error(), http.StatusInternalServerError)
		} else {

			list := model.NewExtraList()

			// We need to flip the order if we selected backwards
			if before {
				for _, p := range extras {
					list.AddExtra(p)
					list.AddOrder(p.Id)
				}
			} else {
				l := len(extras)
				for i := range extras {
					list.AddExtra(extras[l-i-1])
					list.AddOrder(extras[l-i-1].Id)
				}
			}

			result.Data = list
		}
	})
}

func (s SqlExtraStore) GetExtraProductsByIds(productIds []string, allowFromCache bool) store.StoreChannel {
	return store.Do(func(result *store.StoreResult) {
		list := model.NewProductList()
		list.MakeNonNil()
		keys, params := StringsToQueryParams(productIds)

		query := s.getQueryBuilder().
			Select("p.*").
			From("Products p").
			LeftJoin("Extras ex ON (p.Id = ex.ProductId)").
			Where("p.DeleteAt = ? AND ex.DeleteAt = ?", 0, 0).
			Where("ex.RefId IN "+keys, params...).
			OrderBy("ex.Primary DESC")

		queryString, args, err := query.ToSql()

		if err != nil {
			//result.Err = model.NewAppError("SqlExtraStore.GetExtraProductsByIds", "store.sql_extra.get.app_error", nil, err.Error(), http.StatusInternalServerError)
			result.Data = list
			return
		}

		var products []*model.Product
		if _, err := s.GetMaster().Select(&products, queryString, args...); err != nil {
			//result.Err = model.NewAppError("SqlExtraStore.GetExtraProductsByIds", "store.sql_extra.get.app_error", nil, err.Error(), http.StatusInternalServerError)
			result.Data = list
			return
		}

		for _, p := range products {
			list.AddProduct(p)
			list.AddOrder(p.Id)
		}

		result.Data = list
	})
}

func (s SqlExtraStore) DeleteForProduct(productId string) store.StoreChannel {
	return store.Do(func(result *store.StoreResult) {
		query := s.getQueryBuilder().
			Delete("Extras").
			Where("RefId = ?", productId)
		queryString, args, err := query.ToSql()
		if err != nil {
			result.Err = model.NewAppError("SqlExtraStore.DeleteForProduct", "store.sql_extra.delete_for_product.app_error", nil, err.Error(), http.StatusInternalServerError)
			return
		}

		if _, err := s.GetMaster().Exec(queryString, args...); err != nil {
			result.Err = model.NewAppError("SqlExtraStore.DeleteForProduct",
				"store.sql_extra.delete_for_product.app_error", nil, "product_id="+productId+", err="+err.Error(), http.StatusInternalServerError)
			return
		} else {
			result.Data = productId
		}

		/*if _, err := s.GetMaster().Exec(`UPDATE Extras SET DeleteAt = :DeleteAt WHERE RefId = :ProductId`,
			map[string]interface{}{"DeleteAt": model.GetMillis(), "ProductId": productId}); err != nil {
			result.Err = model.NewAppError("SqlExtraStore.DeleteForProduct",
				"store.sql_extra.delete_for_product.app_error", nil, "product_id="+productId+", err="+err.Error(), http.StatusInternalServerError)
		} else {
			result.Data = productId
		}*/
	})
}
