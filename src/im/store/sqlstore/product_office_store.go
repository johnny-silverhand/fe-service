package sqlstore

import (
	"database/sql"
	"net/http"

	"im/model"
	"im/store"
)

type SqlProductOfficeStore struct {
	SqlStore
}

func NewSqlProductOfficeStore(sqlStore SqlStore) store.ProductOfficeStore {
	s := &SqlProductOfficeStore{
		SqlStore: sqlStore,
	}

	for _, db := range sqlStore.GetAllConns() {
		table := db.AddTableWithName(model.ProductOffice{}, "ProductOffice").SetKeys(false, "Id")
		table.ColMap("Id").SetMaxSize(26)
	}

	return s
}

func (s SqlProductOfficeStore) CreateIndexesIfNotExists() {
	s.CreateIndexIfNotExists("idx_product_office_update_at", "ProductOffice", "UpdateAt")
	s.CreateIndexIfNotExists("idx_product_office_create_at", "ProductOffice", "CreateAt")
	s.CreateIndexIfNotExists("idx_product_office_delete_at", "ProductOffice", "DeleteAt")
}

func (s SqlProductOfficeStore) Save(po *model.ProductOffice) store.StoreChannel {
	return store.Do(func(result *store.StoreResult) {
		po.PreSave()
		if result.Err = po.IsValid(); result.Err != nil {
			return
		}

		if err := s.GetMaster().Insert(po); err != nil {
			result.Err = model.NewAppError("SqlProductOfficeStore.Save", "store.sql_product_office.save.app_error", nil, err.Error(), http.StatusInternalServerError)
		} else {
			result.Data = po
		}
	})
}

func (s SqlProductOfficeStore) Get(id string) store.StoreChannel {
	return store.Do(func(result *store.StoreResult) {
		po := &model.ProductOffice{}

		if err := s.GetReplica().SelectOne(po,
			`SELECT * FROM ProductOffice WHERE Id = :Id AND DeleteAt = 0`, map[string]interface{}{"Id": id}); err != nil {
			if err == sql.ErrNoRows {
				result.Err = model.NewAppError("SqlProductOfficeStore.Get", "store.sql_product_office.get.app_error", nil, "id="+id+", "+err.Error(), http.StatusNotFound)
			} else {
				result.Err = model.NewAppError("SqlProductOfficeStore.Get", "store.sql_product_office.get.app_error", nil, "id="+id+", "+err.Error(), http.StatusInternalServerError)
			}
		} else {
			result.Data = po
		}
	})
}

func (s SqlProductOfficeStore) GetForProduct(productId string, readFromMaster bool, allowFromCache bool) store.StoreChannel {
	return store.Do(func(result *store.StoreResult) {

		var infos []*model.Office

		dbmap := s.GetReplica()

		if readFromMaster {
			dbmap = s.GetMaster()
		}

		if _, err := dbmap.Select(&infos, `
			SELECT * FROM Offices WHERE Id IN ( 
				SELECT OfficeId 
				FROM ProductOffice 
				WHERE ProductId = :ProductId AND DeleteAt = 0
			)`,
			map[string]interface{}{"ProductId": productId}); err != nil {
			result.Err = model.NewAppError("SqlProductOfficeStore.GetForProduct",
				"store.sql_product_office.get_for_product.app_error", nil, "product_id="+productId+", "+err.Error(), http.StatusInternalServerError)
		} else {

			result.Data = infos
		}
	})
}

func (s SqlProductOfficeStore) AttachToPost(productOfficeId, postId, creatorId string) store.StoreChannel {
	return store.Do(func(result *store.StoreResult) {
		sqlResult, err := s.GetMaster().Exec(
			`UPDATE
					ProductOffice
				SET
					MetadataId = :MetadataId,
					MetadataType = :MetadataType 
				WHERE
					Id = :Id`, map[string]interface{}{"MetadataId": postId, "MetadataType": model.METADATA_TYPE_POST, "Id": productOfficeId, "CreatorId": creatorId})
		if err != nil {
			result.Err = model.NewAppError("SqlProductOfficeStore.AttachToPost",
				"store.sql_product_office.attach_to_post.app_error", nil, "post_id="+postId+", product_office_id="+productOfficeId+", err="+err.Error(), http.StatusInternalServerError)
			return
		}

		count, err := sqlResult.RowsAffected()
		if err != nil {
			// RowsAffected should never fail with the MySQL or Postgres drivers
			result.Err = model.NewAppError("SqlProductOfficeStore.AttachToPost",
				"store.sql_product_office.attach_to_post.app_error", nil, "post_id="+postId+", product_office_id="+productOfficeId+", err="+err.Error(), http.StatusInternalServerError)
		} else if count == 0 {
			// Could not attach the productOffice to the post
			result.Err = model.NewAppError("SqlProductOfficeStore.AttachToPost",
				"store.sql_product_office.attach_to_post.app_error", nil, "post_id="+postId+", product_office_id="+productOfficeId, http.StatusBadRequest)
		}
	})
}

func (s SqlProductOfficeStore) AttachTo(productOfficeId, metadataId, metadataType string) store.StoreChannel {
	return store.Do(func(result *store.StoreResult) {
		sqlResult, err := s.GetMaster().Exec(`UPDATE ProductOffice SET MetadataId = :MetadataId WHERE Id = :Id AND MetadataId = ''`,
			map[string]interface{}{"MetadataId": metadataId, "MetadataType": metadataType, "Id": productOfficeId})
		if err != nil {
			result.Err = model.NewAppError("SqlProductOfficeStore.AttachTo",
				"store.sql_product_office.attach_to_post.app_error", nil, "post_id="+metadataId+", product_office_id="+productOfficeId+", err="+err.Error(), http.StatusInternalServerError)
			return
		}

		count, err := sqlResult.RowsAffected()
		if err != nil {
			// RowsAffected should never fail with the MySQL or Postgres drivers
			result.Err = model.NewAppError("SqlProductOfficeStore.AttachTo",
				"store.sql_product_office.attach_to_post.app_error", nil, "post_id="+metadataId+", product_office_id="+productOfficeId+", err="+err.Error(), http.StatusInternalServerError)
		} else if count == 0 {
			// Could not attach the productOffice to the post
			result.Err = model.NewAppError("SqlProductOfficeStore.AttachTo",
				"store.sql_product_office.attach_to_post.app_error", nil, "post_id="+metadataId+", product_office_id="+productOfficeId, http.StatusBadRequest)
		}
	})
}

func (s SqlProductOfficeStore) DeleteForProduct(productId string) store.StoreChannel {
	return store.Do(func(result *store.StoreResult) {
		if _, err := s.GetMaster().Exec(`UPDATE ProductOffice SET DeleteAt = :DeleteAt WHERE ProductId = :ProductId`,
			map[string]interface{}{"DeleteAt": model.GetMillis(), "ProductId": productId}); err != nil {
			result.Err = model.NewAppError("SqlFileInfoStore.DeleteForPost",
				"store.sql_file_info.delete_for_post.app_error", nil, "product_id="+productId+", err="+err.Error(), http.StatusInternalServerError)
		} else {
			result.Data = productId
		}
	})
}
