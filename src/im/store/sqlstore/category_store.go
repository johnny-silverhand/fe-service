package sqlstore

import (
	"database/sql"
	"im/model"
	"im/store"

	_ "github.com/go-sql-driver/mysql"
	"net/http"

	"time"
)

func init() {

}

type SqlCategoryStore struct {
	SqlStore
}

func NewSqlCategoryStore(sqlStore SqlStore) store.CategoryStore {
	cs := &SqlCategoryStore{SqlStore: sqlStore}

	for _, db := range sqlStore.GetAllConns() {
		table := db.AddTableWithName(model.Category{}, "Categories").SetKeys(false, "Id")

		table.ColMap("Id").SetMaxSize(26)
		table.ColMap("ClientId").SetMaxSize(26)
		table.ColMap("Name").SetMaxSize(255)
		table.ColMap("ParentId").SetMaxSize(26)
	}

	return cs
}

func (s SqlCategoryStore) CreateIndexesIfNotExists() {
	s.CreateIndexIfNotExists("idx_categories_client_id", "Categories", "ClientId")
	s.CreateIndexIfNotExists("idx_categories_update_at", "Categories", "UpdateAt")
	s.CreateIndexIfNotExists("idx_categories_create_at", "Categories", "CreateAt")
	s.CreateIndexIfNotExists("idx_categories_delete_at", "Categories", "DeleteAt")
}

func (s SqlCategoryStore) Save(category *model.Category) store.StoreChannel {
	return store.Do(func(result *store.StoreResult) {

		if len(category.Id) > 0 {
			result.Err = model.NewAppError("SqlPostStore.Save", "store.sql_post.save.existing.app_error", nil, "id="+category.Id, http.StatusBadRequest)
			return
		}

		category.PreSave()

		var parentRgt int = 0
		var parentDepth int = 0

		if len(category.ParentId) > 0 {
			var parent *model.Category
			if err := s.GetReplica().SelectOne(&parent, `SELECT 
						*
					FROM Categories 
					WHERE Id = :ParentId AND DeleteAt = 0`, map[string]interface{}{"ParentId": category.ParentId}); err != nil {
				result.Err = model.NewAppError("SqlCategoryStore.Save", "store.sql_category.get_all_page.app_error", nil, err.Error(), http.StatusInternalServerError)
			} else {
				parentRgt = parent.Rgt
				parentDepth = parent.Depth
			}

		}
		_, err := s.GetMaster().Exec(`UPDATE Categories SET 
Lft  = CASE WHEN Lft > :ParentRgt THEN Lft + 2 ELSE Lft END,
Rgt = CASE WHEN Rgt > :ParentRgt2 THEN Rgt + 2 ELSE Rgt END 
WHERE ClientId = :ClientId`, map[string]interface{}{"ParentRgt": parentRgt, "ParentRgt2": parentRgt - 1, "ClientId": category.ClientId})
		if err != nil {
			result.Err = model.NewAppError("SqlCategoryStore.Save", "store.sql_category.get_all_page.app_error", nil, err.Error(), http.StatusInternalServerError)
		} else {

			if len(category.ParentId) > 0 {
				category.Depth = parentDepth + 1
				category.Lft = parentRgt
				category.Rgt = parentRgt + 1
			} else {
				category.Depth = 1
				category.Lft = 1
				category.Rgt = 2
			}

			if err := s.GetMaster().Insert(category); err != nil {
				result.Err = model.NewAppError("SqlPostStore.Save", "store.sql_post.save.app_error", nil, "id="+category.Id+", "+err.Error(), http.StatusInternalServerError)
			} else {
				result.Data = category
			}
		}

	})
}

func (s *SqlCategoryStore) Update(newCategory *model.Category) store.StoreChannel {
	return store.Do(func(result *store.StoreResult) {
		newCategory.UpdateAt = model.GetMillis()
		newCategory.PreCommit()




		if _, err := s.GetMaster().Update(newCategory); err != nil {
			result.Err = model.NewAppError("SqlCategoryStore.Update", "store.sql_post.update.app_error", nil, "id="+newCategory.Id+", "+err.Error(), http.StatusInternalServerError)
		} else {
			result.Data = newCategory
		}
	})
}

func (s SqlCategoryStore) Get(id string) store.StoreChannel {
	return store.Do(func(result *store.StoreResult) {

		var category *model.Category

		var query = `SELECT Parent.*, IFNULL(Children.CountChildren, 0) AS CountChildren
FROM Categories AS Parent
       LEFT JOIN (
                 SELECT ParentId, COUNT(Id) AS CountChildren
                 FROM Categories
                 WHERE ParentId IS NOT NULL
                 GROUP BY ParentId ) AS Children ON Parent.Id = Children.ParentId
				WHERE Id = :Id
				ORDER BY Parent.Depth ASC
`

		if err := s.GetReplica().SelectOne(&category,
			query, map[string]interface{}{"Id": id}); err != nil {
			if err == sql.ErrNoRows {
				result.Err = model.NewAppError("SqlCategoryStore.Get", "store.sql_category.get.app_error", nil, err.Error(), http.StatusNotFound)
			} else {
				result.Err = model.NewAppError("SqlCategoryStore.Get", "store.sql_category.get.app_error", nil, err.Error(), http.StatusInternalServerError)
			}
		} else {
			result.Data = category
		}
	})
}

func (s SqlCategoryStore) GetAllPage(offset int, limit int) store.StoreChannel {
	return store.Do(func(result *store.StoreResult) {
		var query = `SELECT Parent.*, IFNULL(Children.CountChildren, 0) AS CountChildren
FROM Categories AS Parent
       LEFT JOIN (
                 SELECT ParentId, COUNT(Id) AS CountChildren
                 FROM Categories
                 WHERE ParentId IS NOT NULL
                 GROUP BY ParentId ) AS Children ON Parent.Id = Children.ParentId
ORDER BY Parent.Depth ASC
					 LIMIT :Limit
					 OFFSET :Offset`
		var categories []*model.Category
		if _, err := s.GetReplica().Select(&categories,
			query, map[string]interface{}{"Limit": limit, "Offset": offset}); err != nil {
			result.Err = model.NewAppError("SqlCategoryStore.GetAllPage", "store.sql_category.get_all_page.app_error", nil, err.Error(), http.StatusInternalServerError)
		} else {
			result.Data = categories
		}
	})
}

func (s SqlCategoryStore) GetAllByClientId(clientId string) store.StoreChannel {
	return store.Do(func(result *store.StoreResult) {
		var query = `SELECT Parent.*, IFNULL(Children.CountChildren, 0) AS CountChildren
FROM Categories AS Parent
       LEFT JOIN (
                 SELECT ParentId, COUNT(Id) AS CountChildren
                 FROM Categories
                 WHERE ParentId IS NOT NULL
                 GROUP BY ParentId ) AS Children ON Parent.Id = Children.ParentId

					 WHERE Parent.ClientId = :ClientId
					 ORDER BY Parent.Depth ASC`
		var category *model.Category
		if err := s.GetReplica().SelectOne(&category,
			query, map[string]interface{}{"ClientId": clientId}); err != nil {
			result.Err = model.NewAppError("SqlCategoryStore.GetAllByClientId", "store.sql_category.get_all_by_client_id.app_error", nil, err.Error(), http.StatusNotFound)
		} else {
			result.Data = category
		}
	})
}

func (s SqlCategoryStore) GetAllByClientIdPage(clientId string, offset int, limit int) store.StoreChannel {
	return store.Do(func(result *store.StoreResult) {
		var query = `SELECT Parent.*, IFNULL(Children.CountChildren, 0) AS CountChildren
FROM Categories AS Parent
       LEFT JOIN (
                 SELECT ParentId, COUNT(Id) AS CountChildren
                 FROM Categories
                 WHERE ParentId IS NOT NULL
                 GROUP BY ParentId ) AS Children ON Parent.Id = Children.ParentId
					 WHERE Parent.ClientId = :ClientId
					 where ClientId = :ClientId
				ORDER BY Parent.Depth ASC
					 LIMIT :Limit
					 OFFSET :Offset`
		var categories []*model.Category
		if _, err := s.GetReplica().Select(&categories,
			query, map[string]interface{}{"ClientId": clientId, "Limit": limit, "Offset": offset}); err != nil {
			result.Err = model.NewAppError("SqlCategoryStore.GetAllByClientIdPage", "store.sql_category.get_all_by_client_id_page.app_error", nil, err.Error(), http.StatusInternalServerError)
		} else {
			result.Data = categories
		}
	})
}

func (s SqlCategoryStore) Delete(categoryId string) store.StoreChannel {
	return store.Do(func(result *store.StoreResult) {
		if _, err := s.GetMaster().Exec(`UPDATE Categories
												SET DeleteAt = :DeleteAt
	 		  					   				WHERE Id = :Id`, map[string]interface{}{"Id": categoryId, "DeleteAt": time.Now().Unix()}); err != nil {
			result.Err = model.NewAppError("SqlCategoryStore.Delete", "store.sql_category.delete.app_error", nil, err.Error(), http.StatusInternalServerError)
		} else {
			result.Data = map[string]int{"status": http.StatusOK}
		}
	})
}

func (s SqlCategoryStore) GetDescendants(category *model.Category) store.StoreChannel {
	return store.Do(func(result *store.StoreResult) {
		var descendants []*model.Category
		if _, err := s.GetReplica().Select(&descendants,
			`SELECT *
					FROM Categories
					WHERE ParentId = :ParentId
					ORDER BY Depth ASC`, map[string]interface{}{"ParentId": category.Id}); err != nil {
			result.Err = model.NewAppError("SqlCategoryStore.GetDescendants", "store.sql_category.get_descendants.app_error", nil, err.Error(), http.StatusInternalServerError)
		} else {
			result.Data = descendants
		}
	})
}
func (s SqlCategoryStore) GetWithChildren(categoryId string) store.StoreChannel {
	return store.Do(func(result *store.StoreResult) {
		var childrens []*model.Category
		if _, err := s.GetReplica().Select(&childrens,
			`SELECT Child.Id, Child.Name, Child.ParentId, Child.Depth, Child.Lft, Child.Rgt FROM Categories AS Child, Categories AS Parent WHERE Parent.Id=:ParentId AND Child.Lft BETWEEN Parent.Lft AND Parent.Rgt ORDER BY Child.Depth ASC`, map[string]interface{}{"ParentId": categoryId}); err != nil {
			result.Err = model.NewAppError("SqlCategoryStore.GetDescendants", "store.sql_category.get_descendants.app_error", nil, err.Error(), http.StatusInternalServerError)
		} else {
			result.Data = childrens
		}
	})
}
