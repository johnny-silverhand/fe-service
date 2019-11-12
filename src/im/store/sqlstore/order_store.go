package sqlstore

import (
	"database/sql"
	"github.com/mattermost/gorp"
	"im/model"
	"im/store"
	"net/http"
)

type SqlOrderStore struct {
	SqlStore
}

func NewSqlOrderStore(sqlStore SqlStore) store.OrderStore {
	s := &SqlOrderStore{sqlStore}

	for _, db := range sqlStore.GetAllConns() {
		table := db.AddTableWithName(model.Order{}, "Orders").SetKeys(false, "Id")

		table.ColMap("Id").SetMaxSize(26)

	}

	return s
}

func (s SqlOrderStore) CreateIndexesIfNotExists() {
	s.CreateIndexIfNotExists("idx_orders_update_at", "Orders", "UpdateAt")
	s.CreateIndexIfNotExists("idx_orders_create_at", "Orders", "CreateAt")
	s.CreateIndexIfNotExists("idx_orders_delete_at", "Orders", "DeleteAt")
}

func (s SqlOrderStore) Cancel(orderId string) store.StoreChannel {
	return store.Do(func(result *store.StoreResult) {

		_, err := s.GetMaster().Exec("UPDATE Orders SET Canceled = :Canceled, UpdateAt =:UpdateAt WHERE Id = :Id ", map[string]interface{}{"Canceled": true, "UpdateAt": model.GetMillis(), "Id": orderId})
		if err != nil {
			result.Err = model.NewAppError("SqlOrderStore.Cancel", "store.sql_orders.cancel.app_error", nil, err.Error(), http.StatusInternalServerError)

		}

	})
}

func (s *SqlOrderStore) Save(order *model.Order) store.StoreChannel {
	return store.Do(func(result *store.StoreResult) {
		if len(order.Id) > 0 {
			result.Err = model.NewAppError("SqlOrderStore.Save", "store.sql_order.save.existing.app_error", nil, "id="+order.Id, http.StatusBadRequest)
			return
		}

		order.PreSave()

		if result.Err = order.IsValid(); result.Err != nil {
			return
		}

		if err := s.GetMaster().Insert(order); err != nil {
			result.Err = model.NewAppError("SqlOrderStore.Save", "store.sql_order.save.app_error", nil, "id="+order.Id+", "+err.Error(), http.StatusInternalServerError)
		} else {
			result.Data = order
		}
	})
}

func (s *SqlOrderStore) Update(newOrder *model.Order) store.StoreChannel {
	return store.Do(func(result *store.StoreResult) {
		newOrder.UpdateAt = model.GetMillis()
		newOrder.PreCommit()

		if _, err := s.GetMaster().Update(newOrder); err != nil {
			result.Err = model.NewAppError("SqlOrderStore.Update", "store.sql_order.update.app_error", nil, "id="+newOrder.Id+", "+err.Error(), http.StatusInternalServerError)
		} else {
			result.Data = newOrder
		}
	})
}

func (s *SqlOrderStore) Get(id string) store.StoreChannel {
	return store.Do(func(result *store.StoreResult) {
		var order *model.Order
		if err := s.GetReplica().SelectOne(&order,
			`SELECT *
					FROM Orders
					WHERE Id = :Id  AND DeleteAt = 0`, map[string]interface{}{"Id": id}); err != nil {
			if err == sql.ErrNoRows {
				result.Err = model.NewAppError("SqlOrderStore.Get", "store.sql_orders.get.app_error", nil, err.Error(), http.StatusNotFound)
			} else {
				result.Err = model.NewAppError("SqlOrderStore.Get", "store.sql_orders.get.app_error", nil, err.Error(), http.StatusInternalServerError)
			}
		} else {
			result.Data = order
		}
	})
}

func (s SqlOrderStore) GetAllPage(offset int, limit int, order model.ColumnOrder) store.StoreChannel {
	return store.Do(func(result *store.StoreResult) {
		var orders []*model.Order

		query := `SELECT *
                  FROM Orders`
		//ORDER BY ` + order.Column + ` `

		/*if order.Column == "price" { // cuz price is string
			query += `+ 0 ` // hack for sorting string as integer
		}*/

		query += order.Type + ` LIMIT :Limit OFFSET :Offset `

		if _, err := s.GetReplica().Select(&orders, query, map[string]interface{}{"Limit": limit, "Offset": offset}); err != nil {
			result.Err = model.NewAppError("SqlOrderStore.GetAllPage", "store.sql_orders.get_all_page.app_error",
				nil, err.Error(),
				http.StatusInternalServerError)
		} else {

			list := model.NewOrderList()

			for _, p := range orders {
				list.AddItem(p)
				list.AddOrder(p.Id)
			}

			list.MakeNonNil()

			result.Data = list
		}
	})
}

func (s *SqlOrderStore) Overwrite(order *model.Order) store.StoreChannel {
	return store.Do(func(result *store.StoreResult) {
		order.UpdateAt = model.GetMillis()

		if result.Err = order.IsValid(); result.Err != nil {
			return
		}

		if _, err := s.GetMaster().Update(order); err != nil {
			result.Err = model.NewAppError("SqlOrderStore.Overwrite", "store.sql_order.overwrite.app_error", nil, "id="+order.Id+", "+err.Error(), http.StatusInternalServerError)
		} else {
			result.Data = order
		}
	})
}

func (s *SqlOrderStore) Delete(orderId string, time int64, deleteByID string) store.StoreChannel {
	return store.Do(func(result *store.StoreResult) {

		appErr := func(errMsg string) *model.AppError {
			return model.NewAppError("SqlOrderStore.Delete", "store.sql_order.delete.app_error", nil, "id="+orderId+", err="+errMsg, http.StatusInternalServerError)
		}

		var order model.Order
		err := s.GetReplica().SelectOne(&order, "SELECT * FROM Orders WHERE Id = :Id AND DeleteAt = 0", map[string]interface{}{"Id": orderId})
		if err != nil {
			result.Err = appErr(err.Error())
		}

		_, err = s.GetMaster().Exec("UPDATE Orders SET DeleteAt = :DeleteAt, UpdateAt = :UpdateAt WHERE Id = :Id", map[string]interface{}{"DeleteAt": time, "UpdateAt": time, "Id": orderId})
		if err != nil {
			result.Err = appErr(err.Error())
		}
	})
}

func (s SqlOrderStore) GetAllOrders(offset int, limit int, allowFromCache bool) store.StoreChannel {
	return store.Do(func(result *store.StoreResult) {
		if limit > 1000 {
			result.Err = model.NewAppError("SqlOrderStore.GetAllOrders", "store.sql_order.get_orders.app_error", nil, "", http.StatusBadRequest)
			return
		}

		var orders []*model.Order
		_, err := s.GetReplica().Select(&orders, "SELECT * FROM Orders WHERE "+
			" DeleteAt = 0 "+
			" ORDER BY CreateAt DESC LIMIT :Limit OFFSET :Offset", map[string]interface{}{"Offset": offset, "Limit": limit})

		if err != nil {
			result.Err = model.NewAppError("SqlOrderStore.GetAllOrders", "store.sql_order.get_root_orders.app_error", nil, err.Error(), http.StatusInternalServerError)
		}

		if err == nil {

			list := model.NewOrderList()

			for _, p := range orders {
				list.AddItem(p)
				list.AddOrder(p.Id)
			}

			list.MakeNonNil()

			result.Data = list
		}
	})
}

func (s SqlOrderStore) GetAllOrdersSince(time int64, allowFromCache bool) store.StoreChannel {
	return store.Do(func(result *store.StoreResult) {

		var orders []*model.Order
		_, err := s.GetReplica().Select(&orders,
			`SELECT * FROM Orders WHERE UpdateAt > :Time  ORDER BY UpdateAt`,
			map[string]interface{}{"Time": time})

		if err != nil {
			result.Err = model.NewAppError("SqlOrderStore.GetAllOrdersSince", "store.sql_order.get_orders_since.app_error", nil, err.Error(), http.StatusInternalServerError)
		} else {

			list := model.NewOrderList()
			var latestUpdate int64 = 0

			for _, p := range orders {
				list.AddItem(p)
				if p.UpdateAt > time {
					list.AddOrder(p.Id)
				}
				if latestUpdate < p.UpdateAt {
					latestUpdate = p.UpdateAt
				}
			}

			//lastOrderTimeCache.AddWithExpiresInSecs(channelId, latestUpdate, LAST_MESSAGE_TIME_CACHE_SEC)

			result.Data = list
		}
	})
}

func (s SqlOrderStore) GetAllOrdersBefore(orderId string, numOrders int, offset int) store.StoreChannel {
	return s.getAllOrdersAround(orderId, numOrders, offset, true)
}

func (s SqlOrderStore) GetAllOrdersAfter(orderId string, numOrders int, offset int) store.StoreChannel {
	return s.getAllOrdersAround(orderId, numOrders, offset, false)
}

func (s SqlOrderStore) getAllOrdersAround(orderId string, numOrders int, offset int, before bool) store.StoreChannel {
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

		var orders []*model.Order

		_, err := s.GetReplica().Select(&orders,
			`SELECT
			    *
			FROM
			    Orders
			WHERE (CreateAt `+direction+` (SELECT CreateAt FROM Orders WHERE Id = :OrderId))
			ORDER BY CreateAt `+sort+`
			LIMIT :NumOrders OFFSET :Offset `,
			map[string]interface{}{"OrderId": orderId, "NumOrders": numOrders, "Offset": offset})

		if err != nil {
			result.Err = model.NewAppError("SqlOrderStore.getAllOrdersAround", "store.sql_order.get_orders_around.get.app_error", nil, err.Error(), http.StatusInternalServerError)
		} else {

			list := model.NewOrderList()

			// We need to flip the order if we selected backwards
			if before {
				for _, p := range orders {
					list.AddItem(p)
					list.AddOrder(p.Id)
				}
			} else {
				l := len(orders)
				for i := range orders {
					list.AddItem(orders[l-i-1])
					list.AddOrder(orders[l-i-1].Id)
				}
			}

			result.Data = list
		}
	})
}

func (s SqlOrderStore) GetFromMaster(id string) store.StoreChannel {
	return s.get(id, true, false)
}

func (s SqlOrderStore) get(id string, master bool, allowFromCache bool) store.StoreChannel {
	return store.Do(func(result *store.StoreResult) {
		var db *gorp.DbMap
		if master {
			db = s.GetMaster()
		} else {
			db = s.GetReplica()
		}

		/*		if allowFromCache {
				if cacheItem, ok := orderCache.Get(id); ok {
					result.Data = (orderItem.(*model.Order)).DeepCopy()
					return
				}
			}*/

		obj, err := db.Get(model.Order{}, id)
		if err != nil {
			result.Err = model.NewAppError("SqlOrderStore.Get", "store.sql_order.get.find.app_error", nil, "id="+id+", "+err.Error(), http.StatusInternalServerError)
			return
		}

		if obj == nil {
			result.Err = model.NewAppError("SqlOrderStore.Get", "store.sql_order.get.existing.app_error", nil, "id="+id, http.StatusNotFound)
			return
		}

		result.Data = obj.(*model.Order)
		//orderCache.AddWithExpiresInSecs(id, obj.(*model.Order), CHANNEL_CACHE_SEC)
	})
}

func (s SqlOrderStore) SaveBasket(orderId string, positions []*model.Basket) store.StoreChannel {
	return store.Do(func(result *store.StoreResult) {

		// Grab the order we are saving this basket to
		cr := <-s.GetFromMaster(orderId)
		if cr.Err != nil {
			result.Err = cr.Err
			return
		}

		order := cr.Data.(*model.Order)

		transaction, err := s.GetMaster().Begin()
		if err != nil {
			result.Err = model.NewAppError("SqlOrderStore.SaveBasket", "store.sql_channel.save_basket.open_transaction.app_error", nil, err.Error(), http.StatusInternalServerError)
			return
		}
		defer finalizeTransaction(transaction)

		for _, basket := range positions {
			*result = s.saveBasketT(transaction, basket, order)
			if result.Err != nil {
				return
			}
		}

		if err := transaction.Commit(); err != nil {
			result.Err = model.NewAppError("SqlOrderStore.SaveBasket", "store.sql_channel.save_basket.commit_transaction.app_error", nil, err.Error(), http.StatusInternalServerError)
			return
		}
	})
}

func (s SqlOrderStore) saveBasketT(transaction *gorp.Transaction, basket *model.Basket, order *model.Order) store.StoreResult {
	result := store.StoreResult{}

	basket.PreSave()

	if result.Err = basket.IsValid(); result.Err != nil {
		return result
	}

	if err := transaction.Insert(basket); err != nil {

		result.Err = model.NewAppError("SqlOrderStore.SaveBasket", "store.sql_channel.save_basket.save.app_error", nil, err.Error(), http.StatusInternalServerError)
		return result
	}

	result.Data = basket
	return result
}

func (s *SqlOrderStore) SaveWithBasket(order *model.Order) store.StoreChannel {
	return store.Do(func(result *store.StoreResult) {
		if len(order.Id) > 0 {
			result.Err = model.NewAppError("SqlOrderStore.Save", "store.sql_order.save.existing.app_error", nil, "id="+order.Id, http.StatusBadRequest)
			return
		}

		order.PreSave()

		//var bresult store.StoreResult

		if result.Err = order.IsValid(); result.Err != nil {
			return
		}

		transaction, err := s.GetMaster().Begin()
		if err != nil {
			result.Err = model.NewAppError("SqlOrderStore.SaveBasket", "store.sql_channel.save_basket.open_transaction.app_error", nil, err.Error(), http.StatusInternalServerError)
			return
		}
		defer finalizeTransaction(transaction)

		var basket []*model.Basket

		if err := transaction.Insert(order); err != nil {
			result.Err = model.NewAppError("SqlOrderStore.Save", "store.sql_order.save.app_error", nil, "id="+order.Id+", "+err.Error(), http.StatusInternalServerError)
		} else {

			for _, ps := range order.Positions {

				ps.Fil(order)

				bresult := s.saveBasketT(transaction, ps, order)

				if bresult.Err != nil {
					transaction.Rollback()
					result.Err = bresult.Err
					return
				}

				basket = append(basket, bresult.Data.(*model.Basket))
			}

			if err := transaction.Commit(); err != nil {
				result.Err = model.NewAppError("SqlOrderStore.SaveBasket", "store.sql_channel.save_basket.commit_transaction.app_error", nil, err.Error(), http.StatusInternalServerError)
				return
			}

			order.Positions = basket

			result.Data = order

		}
	})
}

func (s SqlOrderStore) GetByUserId(userId string, offset int, limit int, order model.ColumnOrder) store.StoreChannel {
	return store.Do(func(result *store.StoreResult) {
		var orders []*model.Order

		query := `SELECT *
                  FROM Orders
WHERE UserId = :UserId `
		//ORDER BY ` + order.Column + ` `

		/*if order.Column == "price" { // cuz price is string
			query += `+ 0 ` // hack for sorting string as integer
		}*/

		query += /*order.Type + */ ` LIMIT :Limit OFFSET :Offset `

		if _, err := s.GetReplica().Select(&orders, query, map[string]interface{}{"UserId": userId, "Limit": limit, "Offset": offset}); err != nil {
			result.Err = model.NewAppError("SqlOrderStore.GetAllPage", "store.sql_orders.get_all_page.app_error",
				nil, err.Error(),
				http.StatusInternalServerError)
		} else {

			list := model.NewOrderList()

			for _, p := range orders {
				list.AddItem(p)
				list.AddOrder(p.Id)
			}

			list.MakeNonNil()

			result.Data = list
		}
	})
}

func (s SqlOrderStore) SetOrderPayed(orderId string) store.StoreChannel {
	return store.Do(func(result *store.StoreResult) {

		ts := model.GetMillis()

		_, err := s.GetMaster().Exec("UPDATE Orders SET Payed = :Payed, UpdateAt =:UpdateAt, PayedAt = :PayedAt WHERE Id = :Id ", map[string]interface{}{"Payed": true, "UpdateAt": ts, "Id": orderId, "PayedAt": ts})
		if err != nil {
			result.Err = model.NewAppError("SqlOfficeStore.Publish", "store.sql_offices.publish.app_error", nil, err.Error(), http.StatusInternalServerError)

		}

	})
}
