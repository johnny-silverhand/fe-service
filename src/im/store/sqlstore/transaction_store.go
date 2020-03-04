package sqlstore

import (
	"database/sql"
	"im/model"
	"im/store"
	"net/http"
)

type SqlTransactionStore struct {
	SqlStore
}

func NewSqlTransactionStore(sqlStore SqlStore) store.TransactionStore {
	s := &SqlTransactionStore{sqlStore}

	for _, db := range sqlStore.GetAllConns() {
		table := db.AddTableWithName(model.Transaction{}, "Transactions").SetKeys(false, "Id")
		table.ColMap("Id").SetMaxSize(26)

	}

	return s
}

func (s SqlTransactionStore) CreateIndexesIfNotExists() {
	s.CreateIndexIfNotExists("idx_transactions_update_at", "Transactions", "UpdateAt")
	s.CreateIndexIfNotExists("idx_transactions_create_at", "Transactions", "CreateAt")
	s.CreateIndexIfNotExists("idx_transactions_delete_at", "Transactions", "DeleteAt")
}

func (s *SqlTransactionStore) Save(transaction *model.Transaction) store.StoreChannel {
	return store.Do(func(result *store.StoreResult) {
		if len(transaction.Id) > 0 {
			result.Err = model.NewAppError("SqlTransactionStore.Save", "store.sql_transaction.save.existing.app_error", nil, "id="+transaction.Id, http.StatusBadRequest)
			return
		}

		transaction.PreSave()

		if result.Err = transaction.IsValid(); result.Err != nil {
			return
		}

		if err := s.GetMaster().Insert(transaction); err != nil {
			result.Err = model.NewAppError("SqlTransactionStore.Save", "store.sql_transaction.save.app_error", nil, "id="+transaction.Id+", "+err.Error(), http.StatusInternalServerError)
		} else {
			result.Data = transaction
		}
	})
}

func (s *SqlTransactionStore) Update(newTransaction *model.Transaction) store.StoreChannel {
	return store.Do(func(result *store.StoreResult) {
		newTransaction.UpdateAt = model.GetMillis()
		newTransaction.PreCommit()

		if _, err := s.GetMaster().Update(newTransaction); err != nil {
			result.Err = model.NewAppError("SqlTransactionStore.Update", "store.sql_transaction.update.app_error", nil, "id="+newTransaction.Id+", "+err.Error(), http.StatusInternalServerError)
		} else {
			result.Data = newTransaction
		}
	})
}

func (s *SqlTransactionStore) Get(id string) store.StoreChannel {
	return store.Do(func(result *store.StoreResult) {
		var transaction *model.Transaction
		if err := s.GetReplica().SelectOne(&transaction,
			`SELECT *
					FROM Transactions
					WHERE Id = :Id  AND DeleteAt = 0`, map[string]interface{}{"Id": id}); err != nil {
			if err == sql.ErrNoRows {
				result.Err = model.NewAppError("SqlTransactionStore.Get", "store.sql_transactions.get.app_error", nil, err.Error(), http.StatusNotFound)
			} else {
				result.Err = model.NewAppError("SqlTransactionStore.Get", "store.sql_transactions.get.app_error", nil, err.Error(), http.StatusInternalServerError)
			}
		} else {
			result.Data = transaction
		}
	})
}

func (s SqlTransactionStore) GetAllPage(offset int, limit int, order model.ColumnOrder) store.StoreChannel {
	return store.Do(func(result *store.StoreResult) {
		var transactions []*model.Transaction

		query := `SELECT *
                  FROM Transactions`
		//ORDER BY ` + order.Column + ` `

		/*if order.Column == "price" { // cuz price is string
			query += `+ 0 ` // hack for sorting string as integer
		}*/

		query += order.Type + ` LIMIT :Limit OFFSET :Offset `

		if _, err := s.GetReplica().Select(&transactions, query, map[string]interface{}{"Limit": limit, "Offset": offset}); err != nil {
			result.Err = model.NewAppError("SqlTransactionStore.GetAllPage", "store.sql_transactions.get_all_page.app_error",
				nil, err.Error(),
				http.StatusInternalServerError)
		} else {

			list := model.NewTransactionList()

			for _, p := range transactions {
				list.AddTransaction(p)
				list.AddOrder(p.Id)
			}

			list.MakeNonNil()

			result.Data = list
		}
	})
}

func (s *SqlTransactionStore) Overwrite(transaction *model.Transaction) store.StoreChannel {
	return store.Do(func(result *store.StoreResult) {
		transaction.UpdateAt = model.GetMillis()

		if result.Err = transaction.IsValid(); result.Err != nil {
			return
		}

		if _, err := s.GetMaster().Update(transaction); err != nil {
			result.Err = model.NewAppError("SqlTransactionStore.Overwrite", "store.sql_transaction.overwrite.app_error", nil, "id="+transaction.Id+", "+err.Error(), http.StatusInternalServerError)
		} else {
			result.Data = transaction
		}
	})
}

func (s *SqlTransactionStore) Delete(transactionId string, time int64, deleteByID string) store.StoreChannel {
	return store.Do(func(result *store.StoreResult) {

		appErr := func(errMsg string) *model.AppError {
			return model.NewAppError("SqlTransactionStore.Delete", "store.sql_transaction.delete.app_error", nil, "id="+transactionId+", err="+errMsg, http.StatusInternalServerError)
		}

		var transaction model.Transaction
		err := s.GetReplica().SelectOne(&transaction, "SELECT * FROM Transactions WHERE Id = :Id AND DeleteAt = 0", map[string]interface{}{"Id": transactionId})
		if err != nil {
			result.Err = appErr(err.Error())
		}

		_, err = s.GetMaster().Exec("UPDATE Transactions SET DeleteAt = :DeleteAt, UpdateAt = :UpdateAt WHERE Id = :Id", map[string]interface{}{"DeleteAt": time, "UpdateAt": time, "Id": transactionId})
		if err != nil {
			result.Err = appErr(err.Error())
		}
	})
}

func (s SqlTransactionStore) GetAllTransactions(offset int, limit int, allowFromCache bool) store.StoreChannel {
	return store.Do(func(result *store.StoreResult) {
		if limit > 1000 {
			result.Err = model.NewAppError("SqlTransactionStore.GetAllTransactions", "store.sql_transaction.get_transactions.app_error", nil, "", http.StatusBadRequest)
			return
		}

		var transactions []*model.Transaction
		_, err := s.GetReplica().Select(&transactions, "SELECT * FROM Transactions WHERE "+
			" DeleteAt = 0 "+
			" ORDER BY CreateAt DESC LIMIT :Limit OFFSET :Offset", map[string]interface{}{"Offset": offset, "Limit": limit})

		if err != nil {
			result.Err = model.NewAppError("SqlTransactionStore.GetAllTransactions", "store.sql_transaction.get_root_transactions.app_error", nil, err.Error(), http.StatusInternalServerError)
		}

		if err == nil {

			list := model.NewTransactionList()

			for _, p := range transactions {
				list.AddTransaction(p)
				list.AddOrder(p.Id)
			}

			list.MakeNonNil()

			result.Data = list
		}
	})
}

func (s SqlTransactionStore) GetAllTransactionsSince(time int64, allowFromCache bool) store.StoreChannel {
	return store.Do(func(result *store.StoreResult) {

		var transactions []*model.Transaction
		_, err := s.GetReplica().Select(&transactions,
			`SELECT * FROM Transactions WHERE UpdateAt > :Time  ORDER BY UpdateAt`,
			map[string]interface{}{"Time": time})

		if err != nil {
			result.Err = model.NewAppError("SqlTransactionStore.GetAllTransactionsSince", "store.sql_transaction.get_transactions_since.app_error", nil, err.Error(), http.StatusInternalServerError)
		} else {

			list := model.NewTransactionList()
			var latestUpdate int64 = 0

			for _, p := range transactions {
				list.AddTransaction(p)
				if p.UpdateAt > time {
					list.AddOrder(p.Id)
				}
				if latestUpdate < p.UpdateAt {
					latestUpdate = p.UpdateAt
				}
			}

			//lastTransactionTimeCache.AddWithExpiresInSecs(channelId, latestUpdate, LAST_MESSAGE_TIME_CACHE_SEC)

			result.Data = list
		}
	})
}

func (s SqlTransactionStore) GetAllTransactionsBefore(transactionId string, numTransactions int, offset int) store.StoreChannel {
	return s.getAllTransactionsAround(transactionId, numTransactions, offset, true)
}

func (s SqlTransactionStore) GetAllTransactionsAfter(transactionId string, numTransactions int, offset int) store.StoreChannel {
	return s.getAllTransactionsAround(transactionId, numTransactions, offset, false)
}

func (s SqlTransactionStore) getAllTransactionsAround(transactionId string, numTransactions int, offset int, before bool) store.StoreChannel {
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

		var transactions []*model.Transaction

		_, err := s.GetReplica().Select(&transactions,
			`SELECT
			    *
			FROM
			    Transactions
			WHERE (CreateAt `+direction+` (SELECT CreateAt FROM Transactions WHERE Id = :TransactionId))
			ORDER BY CreateAt `+sort+`
			OFFSET :Offset LIMIT :NumTransactions`,
			map[string]interface{}{"TransactionId": transactionId, "NumTransactions": numTransactions, "Offset": offset})

		if err != nil {
			result.Err = model.NewAppError("SqlTransactionStore.getAllTransactionsAround", "store.sql_transaction.get_transactions_around.get.app_error", nil, err.Error(), http.StatusInternalServerError)
		} else {

			list := model.NewTransactionList()

			// We need to flip the order if we selected backwards
			if before {
				for _, p := range transactions {
					list.AddTransaction(p)
					list.AddOrder(p.Id)
				}
			} else {
				l := len(transactions)
				for i := range transactions {
					list.AddTransaction(transactions[l-i-1])
					list.AddOrder(transactions[l-i-1].Id)
				}
			}

			result.Data = list
		}
	})
}

func (s SqlTransactionStore) GetByUserId(options model.TransactionGetOptions) store.StoreChannel {
	return store.Do(func(result *store.StoreResult) {
		sort := model.GetOrder(options.Sort)

		query := s.getQueryBuilder().
			Select("t.*").
			From("Transactions t").
			Join("Users u ON t.UserId = u.Id").
			Where("t.DeleteAt = 0").
			Where("u.Id = ? AND u.AppId = ?", options.UserId, options.AppId).
			Offset(uint64(options.Page * options.PerPage)).
			Limit(uint64(options.PerPage))

		if sort.Validate() {
			query = query.OrderBy(sort.Column + " " + sort.Type)
		} else {
			query = query.OrderBy("t.CreateAt DESC")
		}

		queryString, args, err := query.ToSql()

		if err != nil {
			result.Err = model.NewAppError("SqlTransactionStore.GetAllPage", "store.sql_transaction.app_error", nil, err.Error(), http.StatusInternalServerError)
			return
		}

		var transactions []*model.Transaction
		if _, err := s.GetMaster().Select(&transactions, queryString, args...); err != nil {
			result.Err = model.NewAppError("SqlTransactionStore.GetAllPage", "store.sql_transaction.app_error", nil, err.Error(), http.StatusInternalServerError)
			return
		}

		list := model.NewTransactionList()
		list.MakeNonNil()
		for _, p := range transactions {
			list.AddTransaction(p)
			list.AddOrder(p.Id)
		}

		result.Data = list
	})
}

func (s SqlTransactionStore) GetBonusTransactionsForUser(orderUserId string, userId string) store.StoreChannel {
	return store.Do(func(result *store.StoreResult) {
		var transactions []*model.Transaction
		query := "SELECT t.* " +
			"FROM Transactions t " +
			"JOIN Orders o ON t.OrderId = o.Id " +
			"WHERE o.UserId = :OrderUserId AND t.Type = :Type AND t.UserId = :UserId"
		if _, err := s.GetReplica().Select(&transactions, query,
			map[string]interface{}{"OrderUserId": orderUserId, "UserId": userId, "Type": model.TRANSACTION_TYPE_BONUS}); err != nil {
			result.Err = model.NewAppError("SqlTransactionStore.GetBonusTransactionsForUser", "store.sql_transaction.get_bonus_transaction_for_user.app_error",
				nil, err.Error(),
				http.StatusInternalServerError)
		} else {

			list := model.NewTransactionList()

			for _, p := range transactions {
				list.AddTransaction(p)
				list.AddOrder(p.Id)
			}

			list.MakeNonNil()

			result.Data = list
		}

	})
}

func (s SqlTransactionStore) GetMetricsForSpy(options model.UserGetOptions, beginAt int64, expireAt int64) store.StoreChannel {
	return store.Do(func(result *store.StoreResult) {

		query := s.getQueryBuilder().
			Select("o.Id AS OperatorId, o.Email AS OperatorEmail, u.Id AS UserId, u.Email AS UserEmail, FROM_UNIXTIME(t.CreateAt / 1000, '%d.%m.%Y') AS Date, SUM(CASE WHEN t.Value > 0 THEN t.Value ELSE 0 END) Charge, SUM(CASE WHEN t.Value < 0 THEN t.Value ELSE 0 END) Discard").
			From("Users u").
			LeftJoin("Transactions t ON t.UserId = u.Id").
			Join("Users o ON t.CreatedBy = o.Id").
			Where("t.CreateAt BETWEEN ? AND ?", beginAt, expireAt).
			Where("u.AppId", options.AppId).
			GroupBy("u.Id, o.Id, Date").
			OrderBy("Date DESC").
			Offset(uint64(options.Page * options.PerPage)).
			Limit(uint64(options.PerPage))

		queryString, args, err := query.ToSql()

		if err != nil {
			result.Err = model.NewAppError("SqlTransactionStore.GetMetricsForSpy", "store.sql_transaction.app_error", nil, err.Error(), http.StatusInternalServerError)
			return
		}

		var metrics []*model.UserMetricsForSpy
		if _, err := s.GetMaster().Select(&metrics, queryString, args...); err != nil {
			result.Err = model.NewAppError("SqlTransactionStore.GetMetricsForSpy", "store.sql_transaction.app_error", nil, err.Error(), http.StatusInternalServerError)
			return
		}

		/*list := model.NewUserMetricsForSpyList()
		list.MakeNonNil()
		for _, p := range metrics {
			list.AddItem(p)
			list.AddOrder(p.OperatorId)
		}*/

		result.Data = metrics
	})
}
