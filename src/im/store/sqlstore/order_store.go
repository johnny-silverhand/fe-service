package sqlstore

import (
	"database/sql"
	"fmt"
	sq "github.com/Masterminds/squirrel"
	"github.com/mattermost/gorp"
	"im/model"
	"im/store"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type SqlOrderStore struct {
	SqlStore
	// ordersQuery is a starting point for all queries that return one or more Orders.
	ordersQuery sq.SelectBuilder
}

func NewSqlOrderStore(sqlStore SqlStore) store.OrderStore {
	s := &SqlOrderStore{
		SqlStore: sqlStore,
	}

	s.ordersQuery = s.getQueryBuilder().
		Select("O.*").
		From("Orders O")

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

func generateOrderStatusQuery(query sq.SelectBuilder, terms []string, include bool, isPostgreSQL bool) sq.SelectBuilder {
	var sep string
	var op string
	if include {
		sep = " OR "
		op = "="
	} else {
		sep = " AND "
		op = "!="
	}
	searchFields := []string{}
	termArgs := []interface{}{}
	for _, term := range terms {
		if isPostgreSQL {
			searchFields = append(searchFields, fmt.Sprintf("lower(%s) %s ? ", "O.Status", op))
		} else {
			searchFields = append(searchFields, fmt.Sprintf("%s %s ? ", "O.Status", op))
		}
		termArgs = append(termArgs, fmt.Sprintf("%s", strings.TrimLeft(term, "@")))
	}
	query = query.Where(fmt.Sprintf("(%s)", strings.Join(searchFields, sep)), termArgs...)

	return query
}

func (s SqlOrderStore) GetAllOrders(options model.OrderGetOptions) store.StoreChannel {
	isPostgreSQL := s.DriverName() == model.DATABASE_DRIVER_POSTGRES
	return store.Do(func(result *store.StoreResult) {
		var total string
		var OrderStats *model.OrdersStats
		endOfDay := model.GetEndOfDayMillis(time.Now(), 0)
		query := s.ordersQuery.
			Join("Users U ON O.UserId = U.Id").
			Join("(SELECT SUBSTRING(Props, 14, 26) AS OrderId FROM Posts WHERE Type = ?) P ON O.Id = P.OrderId", model.POST_WITH_METADATA).
			Where("O.DeleteAt = 0").
			Where("U.AppId = ?", options.AppId).
			OrderBy("O.CreateAt DESC").
			Offset(uint64(options.Page * options.PerPage)).
			Limit(uint64(options.PerPage))

		r := <-s.Count(model.OrderCountOptions{AppId: options.AppId})
		if r.Err != nil {
			result.Err = r.Err
			return
		}
		OrderStats = r.Data.(*model.OrdersStats)
		statuses := model.ORDER_STATUS_DECLINED + " " + model.ORDER_STATUS_SHIPPED
		if options.Status == model.ORDER_STADY_CLOSED {
			query = generateOrderStatusQuery(query, strings.Fields(statuses), true, isPostgreSQL)
			total = strconv.FormatInt(OrderStats.ClosedCount, 10)
		} else {
			if options.Status == model.ORDER_STADY_DEFERRED {
				query = generateOrderStatusQuery(query, strings.Fields(statuses), false, isPostgreSQL).
					Where("O.DeliveryAt > ?", endOfDay)
				total = strconv.FormatInt(OrderStats.DeferredCount, 10)
			} else {
				query = generateOrderStatusQuery(query, strings.Fields(statuses), false, isPostgreSQL).
					Where("O.DeliveryAt <= ?", endOfDay)
				total = strconv.FormatInt(OrderStats.CurrentCount, 10)
			}
		}

		queryString, args, err := query.ToSql()

		if err != nil {
			result.Err = model.NewAppError("SqlOrderStore.GetAllOrders", "store.sql_order.get_orders.app_error", nil, err.Error(), http.StatusInternalServerError)
			return
		}
		var orders []*model.Order
		if _, err := s.GetMaster().Select(&orders, queryString, args...); err != nil {
			result.Err = model.NewAppError("SqlOrderStore.GetAllOrders", "store.sql_order.get_root_orders.app_error", nil, err.Error(), http.StatusInternalServerError)
			return
		}

		list := model.NewOrderList()
		list.Total = total
		for _, p := range orders {
			list.AddItem(p)
			list.AddOrder(p.Id)
		}
		list.MakeNonNil()

		var descending bool
		if len(options.Sort) > 0 {
			if options.Sort[0:1] == "-" {
				descending = true
				options.Sort = options.Sort[1:]
			} else {
				descending = false
			}
		} else {
			descending = true
		}

		list.SortBy(options.Sort, descending)

		result.Data = list
	})
}

func (s SqlOrderStore) GetAllOrdersSince(time int64, options model.OrderGetOptions) store.StoreChannel {
	return store.Do(func(result *store.StoreResult) {

		/*list := model.NewOrderList()
		var latestDeliveryAt int64 = 0
		for _, p := range orders {
			list.AddItem(p)
			if p.DeliveryAt > time {
				list.AddOrder(p.Id)
			}
			if latestDeliveryAt < p.DeliveryAt {
				latestDeliveryAt = p.DeliveryAt
			}
		}
		list.MakeNonNil()*/

	})
}

func (s SqlOrderStore) GetAllOrdersBefore(orderId string, options model.OrderGetOptions) store.StoreChannel {
	return s.getAllOrdersAround(orderId, true, options)
}

func (s SqlOrderStore) GetAllOrdersAfter(orderId string, options model.OrderGetOptions) store.StoreChannel {
	return s.getAllOrdersAround(orderId, false, options)
}

func (s SqlOrderStore) getAllOrdersAround(orderId string, before bool, options model.OrderGetOptions) store.StoreChannel {
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

		query := s.ordersQuery.
			Join("Users U ON O.UserId = U.Id").
			Join("(SELECT SUBSTRING(Props, 14, 26) AS OrderId FROM Posts WHERE Type = ?) P ON O.Id = P.OrderId", model.POST_WITH_METADATA).
			Where("O.CreateAt "+direction+" (SELECT O.CreateAt FROM Orders O WHERE O.Id = ?)", orderId).
			Where("O.DeleteAt = 0").
			Where("U.AppId = ?", options.AppId).
			OrderBy("O.CreateAt " + sort).
			OrderBy("O.DeliveryAt " + sort).
			Offset(uint64(options.Page * options.PerPage)).
			Limit(uint64(options.PerPage))

		queryString, args, err := query.ToSql()
		if err != nil {
			result.Err = model.NewAppError("SqlOrderStore.getAllOrdersAround", "store.sql_order.get_orders_around.get.app_error", nil, err.Error(), http.StatusInternalServerError)
			return
		}
		var orders []*model.Order
		if _, err := s.GetReplica().Select(&orders, queryString, args...); err != nil {
			result.Err = model.NewAppError("SqlOrderStore.getAllOrdersAround", "store.sql_order.get_orders_around.get.app_error", nil, err.Error(), http.StatusInternalServerError)
			return
		}
		var count struct {
			Total string
		}

		var re = regexp.MustCompile(`SELECT O.\*`)
		queryString = re.ReplaceAllString(queryString, `SELECT COUNT(*) AS Total`)
		re = regexp.MustCompile(`LIMIT .* OFFSET .*`)
		queryString = re.ReplaceAllString(queryString, ``)
		if err := s.GetMaster().SelectOne(&count, queryString, args...); err != nil {
			result.Err = model.NewAppError("SqlOrderStore.GetAllOrders", "store.sql_order.get_count_orders.app_error", nil, err.Error(), http.StatusInternalServerError)
			return
		}
		list := model.NewOrderList()
		list.Total = count.Total

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

func (s SqlOrderStore) GetByUserId(options model.OrderGetOptions) store.StoreChannel {
	return store.Do(func(result *store.StoreResult) {
		sort := model.GetOrder(options.Sort)

		query := s.getQueryBuilder().
			Select("o.*").
			From("Orders o").
			Join("Users u ON o.UserId = u.Id").
			Where("o.DeleteAt = 0").
			Where("u.Id = ? AND u.AppId = ?", options.UserId, options.AppId).
			Offset(uint64(options.Page * options.PerPage)).
			Limit(uint64(options.PerPage))

		if sort.Validate() {
			query = query.OrderBy(sort.Column + " " + sort.Type)
		} else {
			query = query.OrderBy("o.CreateAt DESC")
		}

		queryString, args, err := query.ToSql()

		if err != nil {
			result.Err = model.NewAppError("SqlOrderStore.GetAllPage", "store.sql_orders.get_all_page.app_error", nil, err.Error(), http.StatusInternalServerError)
			return
		}

		var orders []*model.Order
		if _, err := s.GetMaster().Select(&orders, queryString, args...); err != nil {
			result.Err = model.NewAppError("SqlOrderStore.GetAllPage", "store.sql_orders.get_all_page.app_error", nil, err.Error(), http.StatusInternalServerError)
			return
		}

		list := model.NewOrderList()
		list.MakeNonNil()
		for _, p := range orders {
			list.AddItem(p)
			list.AddOrder(p.Id)
		}
		list.Total = strconv.Itoa(len(orders))

		result.Data = list
		/*var orders []*model.Order

		query := `SELECT * FROM Orders WHERE UserId = :UserId LIMIT :Limit OFFSET :Offset`

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
		}*/
	})
}

func (s SqlOrderStore) SetOrderPayed(orderId string) store.StoreChannel {
	return store.Do(func(result *store.StoreResult) {

		ts := model.GetMillis()

		_, err := s.GetMaster().Exec("UPDATE Orders SET Payed = :Payed, UpdateAt =:UpdateAt, PayedAt = :PayedAt, Status = :Status WHERE Id = :Id ", map[string]interface{}{"Payed": true, "UpdateAt": ts, "Id": orderId, "PayedAt": ts, "Status": model.ORDER_STATUS_AWAITING_FULFILLMENT})
		if err != nil {
			result.Err = model.NewAppError("SqlOfficeStore.Publish", "store.sql_offices.publish.app_error", nil, err.Error(), http.StatusInternalServerError)

		}

	})
}

func (s SqlOrderStore) SetOrderCancel(orderId string) store.StoreChannel {
	return store.Do(func(result *store.StoreResult) {

		ts := model.GetMillis()

		_, err := s.GetMaster().Exec("UPDATE Orders SET Canceled = :Canceled, UpdateAt =:UpdateAt, CanceledAt = :CanceledAt, Status = :Status WHERE Id = :Id ", map[string]interface{}{"Canceled": true, "UpdateAt": ts, "Id": orderId, "CanceledAt": ts, "Status": model.ORDER_STATUS_DECLINED})
		if err != nil {
			result.Err = model.NewAppError("SqlOrderStore.CancelOrder", "store.sql_order.cancel_order.app_error", nil, err.Error(), http.StatusInternalServerError)

		}

	})
}

func (s SqlOrderStore) Count(options model.OrderCountOptions) store.StoreChannel {
	isPostgreSQL := s.DriverName() == model.DATABASE_DRIVER_POSTGRES
	return store.Do(func(result *store.StoreResult) {
		var query sq.SelectBuilder
		Totals := new(model.OrdersStats)
		endOfDay := model.GetEndOfDayMillis(time.Now(), 0)
		query = sq.Select("COUNT(*)").From("Orders O").
			Join("Users U ON O.UserId = U.Id").
			Join("(SELECT SUBSTRING(Props, 14, 26) AS OrderId FROM Posts WHERE Type = ?) P ON O.Id = P.OrderId", model.POST_WITH_METADATA).
			Where("O.DeleteAt = 0").
			Where("O.DeliveryAt <= ? AND U.AppId = ?", endOfDay, options.AppId)

		query = generateOrderStatusQuery(query, strings.Fields(model.ORDER_STATUS_DECLINED+" "+model.ORDER_STATUS_SHIPPED), false, isPostgreSQL)
		queryString, args, err := query.ToSql()
		if err != nil {
			result.Err = model.NewAppError("SqlOrderStore.CountCurrent", "store.sql_order.app_error", nil, err.Error(), http.StatusInternalServerError)
			return
		}
		if count, err := s.GetReplica().SelectInt(queryString, args...); err != nil {
			result.Err = model.NewAppError("SqlOrderStore.CountCurrent", "store.sql_order.get_total_count.app_error", nil, err.Error(), http.StatusInternalServerError)
			return
		} else {
			Totals.CurrentCount = count
		}

		query = sq.Select("COUNT(*)").From("Orders O").
			Join("Users U ON O.UserId = U.Id").
			Join("(SELECT SUBSTRING(Props, 14, 26) AS OrderId FROM Posts WHERE Type = ?) P ON O.Id = P.OrderId ", model.POST_WITH_METADATA).
			Where("O.DeleteAt = 0").
			Where("O.DeliveryAt > ? AND U.AppId = ?", endOfDay, options.AppId)
		query = generateOrderStatusQuery(query, strings.Fields(model.ORDER_STATUS_DECLINED+" "+model.ORDER_STATUS_SHIPPED), false, isPostgreSQL)
		queryString, args, err = query.ToSql()
		if err != nil {
			result.Err = model.NewAppError("SqlOrderStore.CountDeferred", "store.sql_order.app_error", nil, err.Error(), http.StatusInternalServerError)
			return
		}
		if count, err := s.GetReplica().SelectInt(queryString, args...); err != nil {
			result.Err = model.NewAppError("SqlOrderStore.CountDeferred", "store.sql_order.get_total_count.app_error", nil, err.Error(), http.StatusInternalServerError)
			return
		} else {
			Totals.DeferredCount = count
		}

		query = sq.Select("COUNT(*)").From("Orders O").
			Join("Users U ON O.UserId = U.Id").
			Join("(SELECT SUBSTRING(Props, 14, 26) AS OrderId FROM Posts WHERE Type = ?) P ON O.Id = P.OrderId", model.POST_WITH_METADATA).
			Where("O.DeleteAt = 0").
			Where("U.AppId = ?", options.AppId)

		query = generateOrderStatusQuery(query, strings.Fields(model.ORDER_STATUS_DECLINED+" "+model.ORDER_STATUS_SHIPPED), true, isPostgreSQL)
		queryString, args, err = query.ToSql()
		if err != nil {
			result.Err = model.NewAppError("SqlOrderStore.CountClosed", "store.sql_order.app_error", nil, err.Error(), http.StatusInternalServerError)
			return
		}
		if count, err := s.GetReplica().SelectInt(queryString, args...); err != nil {
			result.Err = model.NewAppError("SqlOrderStore.CountClosed", "store.sql_order.get_total_count.app_error", nil, err.Error(), http.StatusInternalServerError)
			return
		} else {
			Totals.ClosedCount = count
		}

		Totals.TotalCount = Totals.CurrentCount + Totals.DeferredCount + Totals.ClosedCount
		result.Data = Totals
	})
}

func (s SqlOrderStore) GetMetricsForOrders(appId string, beginAt int64, expireAt int64) store.StoreChannel {
	return store.Do(func(result *store.StoreResult) {
		query := s.getQueryBuilder().
			Select("sum(o.Price) AS TotalPrice, sum(o.DiscountValue) AS TotalDiscount, avg(o.Price) AS AvgPrice, sum(Canceled) AS TotalReturn").
			From("Orders o").
			Join("Users u ON u.Id = o.UserId").
			Where("u.AppId = ? AND u.Roles = ?", appId, model.CHANNEL_USER_ROLE_ID)
			//Where("o.Payed = ?", true)
		queryString, args, err := query.ToSql()
		if err != nil {
			result.Err = model.NewAppError("SqlOrderStore.GetMetricsForOrders", "store.sql_order.app_error", nil, err.Error(), http.StatusInternalServerError)
			return
		}
		var metrics *model.MetricsForOrders
		if err := s.GetReplica().SelectOne(&metrics, queryString, args...); err != nil {
			result.Err = model.NewAppError("SqlOrderStore.GetMetricsForOrders", "store.sql_order.app_error", nil, err.Error(), http.StatusInternalServerError)
			return
		}

		query = s.getQueryBuilder().
			Select("FROM_UNIXTIME(o.CreateAt / 1000, '%d.%m.%Y') AS Date, count(*) AS Count").
			From("Orders o").
			Join("Users u ON o.UserId = u.Id").
			Where("u.AppId = ? AND u.Roles = ?", appId, model.CHANNEL_USER_ROLE_ID).
			Where("o.CreateAt >= ? AND o.CreateAt <= ?", beginAt, expireAt).
			GroupBy("Date")
		queryString, args, err = query.ToSql()
		if err != nil {
			result.Err = model.NewAppError("SqlOrderStore.GetMetricsForOrders", "store.sql_order.app_error", nil, err.Error(), http.StatusInternalServerError)
			return
		}
		if _, err := s.GetReplica().Select(&metrics.OrdersByDay, queryString, args...); err != nil {
			result.Err = model.NewAppError("SqlOrderStore.GetMetricsForOrders", "store.sql_order.app_error", nil, err.Error(), http.StatusInternalServerError)
			return
		}
		metrics.Total = metrics.TotalPrice + metrics.TotalDiscount

		result.Data = metrics
	})
}
