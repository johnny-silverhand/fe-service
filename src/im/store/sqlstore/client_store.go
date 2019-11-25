package sqlstore

import (
	"database/sql"
	"im/model"
	"im/store"
	"net/http"
)

type SqlClientStore struct {
	SqlStore
}

func NewSqlClientStore(sqlStore SqlStore) store.ClientStore {
	s := &SqlClientStore{sqlStore}

	for _, db := range sqlStore.GetAllConns() {
		table := db.AddTableWithName(model.Client{}, "Clients").SetKeys(false, "Id")

		table.ColMap("Id").SetMaxSize(26)
		table.ColMap("Name").SetMaxSize(255)
		table.ColMap("Preview").SetMaxSize(255)
		table.ColMap("Description").SetMaxSize(2000)
		table.ColMap("Phone").SetMaxSize(2000)

	}

	return s
}

func (s SqlClientStore) CreateIndexesIfNotExists() {
	s.CreateIndexIfNotExists("idx_clients_update_at", "Clients", "UpdateAt")
	s.CreateIndexIfNotExists("idx_clients_create_at", "Clients", "CreateAt")
	s.CreateIndexIfNotExists("idx_clients_delete_at", "Clients", "DeleteAt")
}

func (s SqlClientStore) Activate(clientId string) store.StoreChannel {
	return store.Do(func(result *store.StoreResult) {

		_, err := s.GetMaster().Exec("UPDATE Clients SET Active = :Active, UpdateAt =:UpdateAt WHERE Id = :Id ", map[string]interface{}{"Active": true, "UpdateAt": model.GetMillis(), "Id": clientId})
		if err != nil {
			result.Err = model.NewAppError("SqlClientStore.Publish", "store.sql_clients.publish.app_error", nil, err.Error(), http.StatusInternalServerError)

		}

	})
}
func (s SqlClientStore) Deactivate(clientId string) store.StoreChannel {
	return store.Do(func(result *store.StoreResult) {

		_, err := s.GetMaster().Exec("UPDATE Clients SET Active = :Active, UpdateAt =:UpdateAt WHERE Id = :Id ", map[string]interface{}{"Active": false, "UpdateAt": model.GetMillis(), "Id": clientId})
		if err != nil {
			result.Err = model.NewAppError("SqlClientStore.Disable", "store.sql_clients.publish.app_error", nil, err.Error(), http.StatusInternalServerError)

		}

	})
}

func (s *SqlClientStore) Save(client *model.Client) store.StoreChannel {
	return store.Do(func(result *store.StoreResult) {
		if len(client.Id) > 0 {
			result.Err = model.NewAppError("SqlClientStore.Save", "store.sql_client.save.existing.app_error", nil, "id="+client.Id, http.StatusBadRequest)
			return
		}

		client.PreSave()

		if result.Err = client.IsValid(); result.Err != nil {
			return
		}

		if err := s.GetMaster().Insert(client); err != nil {
			result.Err = model.NewAppError("SqlClientStore.Save", "store.sql_client.save.app_error", nil, "id="+client.Id+", "+err.Error(), http.StatusInternalServerError)
		} else {
			result.Data = client
		}
	})
}

func (s *SqlClientStore) Update(newClient *model.Client) store.StoreChannel {
	return store.Do(func(result *store.StoreResult) {
		newClient.UpdateAt = model.GetMillis()
		newClient.PreCommit()

		if _, err := s.GetMaster().Update(newClient); err != nil {
			result.Err = model.NewAppError("SqlClientStore.Update", "store.sql_client.update.app_error", nil, "id="+newClient.Id+", "+err.Error(), http.StatusInternalServerError)
		} else {
			result.Data = newClient
		}
	})
}

func (s *SqlClientStore) Get(id string) store.StoreChannel {
	return store.Do(func(result *store.StoreResult) {
		var client *model.Client
		if err := s.GetReplica().SelectOne(&client,
			`SELECT *
					FROM Clients
					WHERE Id = :Id  AND DeleteAt = 0`, map[string]interface{}{"Id": id}); err != nil {
			if err == sql.ErrNoRows {
				result.Err = model.NewAppError("SqlClientStore.Get", "store.sql_clients.get.app_error", nil, err.Error(), http.StatusNotFound)
			} else {
				result.Err = model.NewAppError("SqlClientStore.Get", "store.sql_clients.get.app_error", nil, err.Error(), http.StatusInternalServerError)
			}
		} else {
			result.Data = client
		}
	})
}

func (s SqlClientStore) GetAllPage(offset int, limit int, order model.ColumnOrder) store.StoreChannel {
	return store.Do(func(result *store.StoreResult) {
		var clients []*model.Client

		query := `SELECT *
                  FROM Clients`
		//ORDER BY ` + order.Column + ` `

		/*if order.Column == "price" { // cuz price is string
			query += `+ 0 ` // hack for sorting string as integer
		}*/

		query += order.Type + ` LIMIT :Limit OFFSET :Offset `

		if _, err := s.GetReplica().Select(&clients, query, map[string]interface{}{"Limit": limit, "Offset": offset}); err != nil {
			result.Err = model.NewAppError("SqlClientStore.GetAllPage", "store.sql_clients.get_all_page.app_error",
				nil, err.Error(),
				http.StatusInternalServerError)
		} else {

			list := model.NewClientList()

			for _, p := range clients {
				list.AddClient(p)
				list.AddOrder(p.Id)
			}

			list.MakeNonNil()

			result.Data = list
		}
	})
}

func (s *SqlClientStore) Overwrite(client *model.Client) store.StoreChannel {
	return store.Do(func(result *store.StoreResult) {
		client.UpdateAt = model.GetMillis()

		if result.Err = client.IsValid(); result.Err != nil {
			return
		}

		if _, err := s.GetMaster().Update(client); err != nil {
			result.Err = model.NewAppError("SqlClientStore.Overwrite", "store.sql_client.overwrite.app_error", nil, "id="+client.Id+", "+err.Error(), http.StatusInternalServerError)
		} else {
			result.Data = client
		}
	})
}

func (s *SqlClientStore) Delete(clientId string, time int64, deleteByID string) store.StoreChannel {
	return store.Do(func(result *store.StoreResult) {

		appErr := func(errMsg string) *model.AppError {
			return model.NewAppError("SqlClientStore.Delete", "store.sql_client.delete.app_error", nil, "id="+clientId+", err="+errMsg, http.StatusInternalServerError)
		}

		var client model.Client
		err := s.GetReplica().SelectOne(&client, "SELECT * FROM Clients WHERE Id = :Id AND DeleteAt = 0", map[string]interface{}{"Id": clientId})
		if err != nil {
			result.Err = appErr(err.Error())
		}

		_, err = s.GetMaster().Exec("UPDATE Clients SET DeleteAt = :DeleteAt, UpdateAt = :UpdateAt WHERE Id = :Id", map[string]interface{}{"DeleteAt": time, "UpdateAt": time, "Id": clientId})
		if err != nil {
			result.Err = appErr(err.Error())
		}
	})
}

func (s SqlClientStore) GetAllClients(offset int, limit int, allowFromCache bool) store.StoreChannel {
	return store.Do(func(result *store.StoreResult) {
		if limit > 1000 {
			result.Err = model.NewAppError("SqlClientStore.GetAllClients", "store.sql_client.get_clients.app_error", nil, "", http.StatusBadRequest)
			return
		}

		var clients []*model.Client
		_, err := s.GetReplica().Select(&clients, "SELECT * FROM Clients WHERE "+
			" DeleteAt = 0 "+
			" ORDER BY CreateAt DESC LIMIT :Limit OFFSET :Offset", map[string]interface{}{"Offset": offset, "Limit": limit})

		if err != nil {
			result.Err = model.NewAppError("SqlClientStore.GetAllClients", "store.sql_client.get_root_clients.app_error", nil, err.Error(), http.StatusInternalServerError)
		}

		if err == nil {

			list := model.NewClientList()

			for _, p := range clients {
				list.AddClient(p)
				list.AddOrder(p.Id)
			}

			list.MakeNonNil()

			result.Data = list
		}
	})
}

func (s SqlClientStore) GetAllClientsSince(time int64, allowFromCache bool) store.StoreChannel {
	return store.Do(func(result *store.StoreResult) {

		var clients []*model.Client
		_, err := s.GetReplica().Select(&clients,
			`SELECT * FROM Clients WHERE UpdateAt > :Time  ORDER BY UpdateAt`,
			map[string]interface{}{"Time": time})

		if err != nil {
			result.Err = model.NewAppError("SqlClientStore.GetAllClientsSince", "store.sql_client.get_clients_since.app_error", nil, err.Error(), http.StatusInternalServerError)
		} else {

			list := model.NewClientList()
			var latestUpdate int64 = 0

			for _, p := range clients {
				list.AddClient(p)
				if p.UpdateAt > time {
					list.AddOrder(p.Id)
				}
				if latestUpdate < p.UpdateAt {
					latestUpdate = p.UpdateAt
				}
			}

			//lastClientTimeCache.AddWithExpiresInSecs(channelId, latestUpdate, LAST_MESSAGE_TIME_CACHE_SEC)

			result.Data = list
		}
	})
}

func (s SqlClientStore) GetAllClientsBefore(clientId string, numClients int, offset int) store.StoreChannel {
	return s.getAllClientsAround(clientId, numClients, offset, true)
}

func (s SqlClientStore) GetAllClientsAfter(clientId string, numClients int, offset int) store.StoreChannel {
	return s.getAllClientsAround(clientId, numClients, offset, false)
}

func (s SqlClientStore) getAllClientsAround(clientId string, numClients int, offset int, before bool) store.StoreChannel {
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

		var clients []*model.Client

		_, err := s.GetReplica().Select(&clients,
			`SELECT
			    *
			FROM
			    Clients
			WHERE (CreateAt `+direction+` (SELECT CreateAt FROM Clients WHERE Id = :ClientId))
			ORDER BY CreateAt `+sort+`
			OFFSET :Offset LIMIT :NumClients`,
			map[string]interface{}{"ClientId": clientId, "NumClients": numClients, "Offset": offset})

		if err != nil {
			result.Err = model.NewAppError("SqlClientStore.getAllClientsAround", "store.sql_client.get_clients_around.get.app_error", nil, err.Error(), http.StatusInternalServerError)
		} else {

			list := model.NewClientList()

			// We need to flip the order if we selected backwards
			if before {
				for _, p := range clients {
					list.AddClient(p)
					list.AddOrder(p.Id)
				}
			} else {
				l := len(clients)
				for i := range clients {
					list.AddClient(clients[l-i-1])
					list.AddOrder(clients[l-i-1].Id)
				}
			}

			result.Data = list
		}
	})
}
