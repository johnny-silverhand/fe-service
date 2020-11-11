package sqlstore

import (
	"database/sql"
	"im/model"
	"im/store"
	"net/http"
)

type SqlBasketStore struct {
	SqlStore
}

func NewSqlBasketStore(sqlStore SqlStore) store.BasketStore {
	s := &SqlBasketStore{sqlStore}

	for _, db := range sqlStore.GetAllConns() {
		table := db.AddTableWithName(model.Basket{}, "Baskets").SetKeys(false, "Id")

		table.ColMap("Id").SetMaxSize(26)

	}

	return s
}

func (s SqlBasketStore) CreateIndexesIfNotExists() {
	s.CreateIndexIfNotExists("idx_baskets_update_at", "Baskets", "UpdateAt")
	s.CreateIndexIfNotExists("idx_baskets_create_at", "Baskets", "CreateAt")
	s.CreateIndexIfNotExists("idx_baskets_delete_at", "Baskets", "DeleteAt")
}

func (s *SqlBasketStore) Save(basket *model.Basket) store.StoreChannel {
	return store.Do(func(result *store.StoreResult) {
		if len(basket.Id) > 0 {
			result.Err = model.NewAppError("SqlBasketStore.Save", "store.sql_basket.save.existing.app_error", nil, "id="+basket.Id, http.StatusBadRequest)
			return
		}

		basket.PreSave()

		if result.Err = basket.IsValid(); result.Err != nil {
			return
		}

		if err := s.GetMaster().Insert(basket); err != nil {
			result.Err = model.NewAppError("SqlBasketStore.Save", "store.sql_basket.save.app_error", nil, "id="+basket.Id+", "+err.Error(), http.StatusInternalServerError)
		} else {
			result.Data = basket
		}
	})
}

func (s *SqlBasketStore) Update(newBasket *model.Basket) store.StoreChannel {
	return store.Do(func(result *store.StoreResult) {
		newBasket.UpdateAt = model.GetMillis()
		newBasket.PreCommit()

		if _, err := s.GetMaster().Update(newBasket); err != nil {
			result.Err = model.NewAppError("SqlBasketStore.Update", "store.sql_basket.update.app_error", nil, "id="+newBasket.Id+", "+err.Error(), http.StatusInternalServerError)
		} else {
			result.Data = newBasket
		}
	})
}

func (s *SqlBasketStore) GetByUserId(userId string) store.StoreChannel {
	return store.Do(func(result *store.StoreResult) {
		var basket []*model.Basket
		if _, err := s.GetReplica().Select(&basket,
			`SELECT *
					FROM Baskets
					WHERE OrderId = '' AND UserId = :UserId AND DeleteAt = 0`, map[string]interface{}{"UserId": userId}); err != nil {
			if err == sql.ErrNoRows {
				result.Err = model.NewAppError("SqlBasketStore.Get", "store.sql_baskets.get.app_error", nil, err.Error(), http.StatusNotFound)
			} else {
				result.Err = model.NewAppError("SqlBasketStore.Get", "store.sql_baskets.get.app_error", nil, err.Error(), http.StatusInternalServerError)
			}
		} else {
			result.Data = basket
		}
	})
}

func (s *SqlBasketStore) GetByOrderId(orderId string) store.StoreChannel {
	return store.Do(func(result *store.StoreResult) {
		var basket []*model.Basket
		if _, err := s.GetReplica().Select(&basket,
			`SELECT *
					FROM Baskets
					WHERE OrderId = :OrderId AND DeleteAt = 0`, map[string]interface{}{"OrderId": orderId}); err != nil {
			if err == sql.ErrNoRows {
				result.Err = model.NewAppError("SqlBasketStore.Get", "store.sql_baskets.get.app_error", nil, err.Error(), http.StatusNotFound)
			} else {
				result.Err = model.NewAppError("SqlBasketStore.Get", "store.sql_baskets.get.app_error", nil, err.Error(), http.StatusInternalServerError)
			}
		} else {
			result.Data = basket
		}
	})
}

func (s *SqlBasketStore) Overwrite(basket *model.Basket) store.StoreChannel {
	return store.Do(func(result *store.StoreResult) {
		basket.UpdateAt = model.GetMillis()

		if result.Err = basket.IsValid(); result.Err != nil {
			return
		}

		if _, err := s.GetMaster().Update(basket); err != nil {
			result.Err = model.NewAppError("SqlBasketStore.Overwrite", "store.sql_basket.overwrite.app_error", nil, "id="+basket.Id+", "+err.Error(), http.StatusInternalServerError)
		} else {
			result.Data = basket
		}
	})
}

func (s *SqlBasketStore) Delete(basketId string, time int64, deleteByID string) store.StoreChannel {
	return store.Do(func(result *store.StoreResult) {

		appErr := func(errMsg string) *model.AppError {
			return model.NewAppError("SqlBasketStore.Delete", "store.sql_basket.delete.app_error", nil, "id="+basketId+", err="+errMsg, http.StatusInternalServerError)
		}

		var basket model.Basket
		err := s.GetReplica().SelectOne(&basket, "SELECT * FROM Baskets WHERE Id = :Id AND DeleteAt = 0", map[string]interface{}{"Id": basketId})
		if err != nil {
			result.Err = appErr(err.Error())
		}

		_, err = s.GetMaster().Exec("UPDATE Baskets SET DeleteAt = :DeleteAt, UpdateAt = :UpdateAt WHERE Id = :Id", map[string]interface{}{"DeleteAt": time, "UpdateAt": time, "Id": basketId})
		if err != nil {
			result.Err = appErr(err.Error())
		}
	})
}
