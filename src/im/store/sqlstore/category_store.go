package sqlstore

import (
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	"im/model"
	"im/store"
	"net/http"
	"strconv"
)

type SqlCategoryStore struct {
	SqlStore
}

func NewSqlCategoryStore(sqlStore SqlStore) store.CategoryStore {
	cs := &SqlCategoryStore{SqlStore: sqlStore}

	for _, db := range sqlStore.GetAllConns() {
		table := db.AddTableWithName(model.Category{}, "Categories").SetKeys(false, "Id")

		table.ColMap("Id").SetMaxSize(26)
		table.ColMap("AppId").SetMaxSize(32)
		table.ColMap("Name").SetMaxSize(32)
		table.ColMap("ParentId").SetMaxSize(26)
		table.ColMap("CreateAt").SetMaxSize(26)
		table.ColMap("UpdateAt").SetMaxSize(26)
		//table.ColMap("DeleteAt").SetMaxSize(26)

	}

	return cs
}

func (s SqlCategoryStore) CreateIndexesIfNotExists() {
	s.CreateIndexIfNotExists("idx_categories_app_id", "Categories", "AppId")
	s.CreateIndexIfNotExists("idx_categories_update_at", "Categories", "UpdateAt")
	s.CreateIndexIfNotExists("idx_categories_create_at", "Categories", "CreateAt")
}

func (s SqlCategoryStore) Update(category *model.Category) store.StoreChannel {
	return store.Do(func(result *store.StoreResult) {
		category.UpdateAt = model.GetMillis()
		category.PreCommit()

		if _, err := s.GetMaster().Update(category); err != nil {
			result.Err = model.NewAppError("SqlCategoryStore.Update", "store.sql_post.update.app_error",
				nil, "id="+category.Id+", "+err.Error(), http.StatusInternalServerError)
		} else {
			result.Data = category
		}
	})
}

func (s SqlCategoryStore) Get(id string) store.StoreChannel {
	return store.Do(func(result *store.StoreResult) {
		var query = `select c.* from categories c 
					 where Id = :Id`

		var category *model.Category
		if err := s.GetReplica().SelectOne(&category,
			query, map[string]interface{}{"Id": id}); err != nil {
			if err == sql.ErrNoRows {
				result.Err = model.NewAppError("SqlCategoryStore.Get",
					"store.sql_category.get.app_error", nil, err.Error(), http.StatusNotFound)
			} else {
				result.Err = model.NewAppError("SqlCategoryStore.Get", "store.sql_category.get.app_error",
					nil, err.Error(), http.StatusInternalServerError)
			}
		} else {
			result.Data = category
		}
	})
}

func (s SqlCategoryStore) GetAllPage(offset int, limit int) store.StoreChannel {

	return store.Do(func(result *store.StoreResult) {
		var query = `select c.*, childs.cnt as CntChild 
					from categories c 
					 left join (
						select ParentId, count(Id) cnt
						from categories
						where ParentId is not null
						group by ParentId ) childs on c.Id = childs.ParentId
					 order by UpdateAt Desc, Id Desc
					 limit :Limit
					 offset :Offset`

		var categories []*model.Category
		if _, err := s.GetReplica().Select(&categories,
			query, map[string]interface{}{"Limit": limit, "Offset": offset}); err != nil {
			result.Err = model.NewAppError("SqlCategoryStore.GetAllPage", "store.sql_category.get_all_page.app_error", nil, err.Error(), http.StatusInternalServerError)
		} else {
			result.Data = categories
		}
	})
}

func (s SqlCategoryStore) GetAllByApp(appId string) store.StoreChannel {
	return store.Do(func(result *store.StoreResult) {
		var query = `select c.*, childs.cnt as CntChild
					 from categories c 
					 left join (
						select ParentId, count(Id) cnt
						from categories
						where ParentId is not null
						group by ParentId ) childs on c.Id = childs.ParentId
					 where AppId = :AppId
					 order by UpdateAt Desc, Id Desc`
		var category *model.Category
		if err := s.GetReplica().SelectOne(&category,
			query, map[string]interface{}{"AppId": appId}); err != nil {
			result.Err = model.NewAppError("SqlCategoryStore.GetAllByAppId", "store.sql_category.get_all_by_app_id.app_error", nil, err.Error(), http.StatusNotFound)
		} else {
			result.Data = category
		}
	})
}

func (s SqlCategoryStore) GetAllByAppPage(appId string, offset int, limit int) store.StoreChannel {
	return store.Do(func(result *store.StoreResult) {
		var query = `select c.*, childs.cnt as CntChild
					 from categories c 
					 left join (
						select ParentId, count(Id) cnt
						from categories
						where ParentId is not null
						group by ParentId ) childs on c.Id = childs.ParentId
					 where AppId = :AppId
					 order by UpdateAt Desc, Id Desc
					 limit :Limit
					 offset :Offset`
		var categories []*model.Category
		if _, err := s.GetReplica().Select(&categories,
			query, map[string]interface{}{"AppId": appId, "Limit": limit, "Offset": offset}); err != nil {
			result.Err = model.NewAppError("SqlCategoryStore.GetAllByAppIdPage", "store.sql_category.get_all_by_app_id_page.app_error", nil, err.Error(), http.StatusInternalServerError)
		} else {
			result.Data = categories
		}
	})
}

func (s SqlCategoryStore) GetDescendants(category *model.Category) store.StoreChannel {
	return store.Do(func(result *store.StoreResult) {
		var descendants []*model.Category
		if _, err := s.GetReplica().Select(&descendants,
			`SELECT *
					FROM categories
					WHERE ParentId = :ParentId
					ORDER BY UpdateAt DESC, Id DESC`, map[string]interface{}{"ParentId": category.Id}); err != nil {
			result.Err = model.NewAppError("SqlCategoryStore.GetDescendants", "store.sql_category.get_descendants.app_error", nil, err.Error(), http.StatusInternalServerError)
		} else {
			result.Data = descendants
		}
	})
}

func (s SqlCategoryStore) GetCategoryPath(categoryId string) store.StoreChannel {
	return store.Do(func(result *store.StoreResult) {
		var (
			root       *model.Category
			categories []*model.Category
		)
		rootQuery := `select * from categories where id = :Id`
		if err := s.GetMaster().SelectOne(&root, rootQuery, map[string]interface{}{"Id": categoryId}); err != nil {
			result.Err = model.NewAppError("SqlCategoryStore.GetCategoryPath",
				"store.sql_category.get_category_path.app_error", nil, err.Error(), http.StatusInternalServerError)
			return
		}

		if _, err := s.GetMaster().Select(&categories, `
			select * 
			from categories 
			where lft >= :Lft and rgt <= :Rgt and appId = :AppId 
			order by lft desc`,
			map[string]interface{}{
				"Lft":   root.Lft,
				"Rgt":   root.Rgt,
				"AppId": root.AppId,
			}); err != nil {
			if err == sql.ErrNoRows {
				result.Err = model.NewAppError("SqlCategoryStore.GetCategoryPath",
					"store.sql_category.get_category_path.app_error", nil, err.Error(), http.StatusNotFound)
			} else {
				result.Err = model.NewAppError("SqlCategoryStore.GetCategoryPath", "store.sql_category.get_category_path.app_error",
					nil, err.Error(), http.StatusInternalServerError)
			}
		} else {
			result.Data = categories
		}
	})
}

func (s SqlCategoryStore) GetCategoriesByIds(categoryIds []string) store.StoreChannel {
	return store.Do(func(result *store.StoreResult) {
		var categories []*model.Category

		props := make(map[string]interface{})
		idQuery := ""
		for i, categoryId := range categoryIds {
			if len(idQuery) > 0 {
				idQuery += ", "
			}
			props["categoryId"+strconv.Itoa(i)] = categoryId
			idQuery += ":categoryId" + strconv.Itoa(i)
		}

		if _, err := s.GetMaster().Select(&categories, `
			SELECT distinct p.* 
			FROM categories AS n, categories AS p 
			WHERE n.lft BETWEEN p.lft AND p.rgt AND n.Id IN (`+idQuery+`) 
			ORDER BY p.lft`,
			props); err != nil {

			if err == sql.ErrNoRows {
				result.Err = model.NewAppError("SqlCategoryStore.GetCategoriesByIds",
					"store.sql_category.get_categories_by_ids.app_error", nil, err.Error(), http.StatusNotFound)
			} else {
				result.Err = model.NewAppError("SqlCategoryStore.GetCategoriesByIds", "store.sql_category.get_categories_by_ids.app_error",
					nil, err.Error(), http.StatusInternalServerError)
			}

		} else {
			result.Data = categories
		}
	})
}

/*
					*** STORED PROCEDURE CALLS ***
 https://www.we-rc.com/blog/2015/07/19/nested-set-model-practical-examples-part-i
		На всякий случай нужно проверить существует ли процедура в боевой базе. На момент
		написани этого текста SQL Скрипты с процедурой лежали в store/
*/

func (s SqlCategoryStore) Create(category *model.Category) store.StoreChannel {

	//store.Do(func(result *store.StoreResult) {
	if len(category.Id) == 0 {
		category.PreSave()
	}

	if _, err := s.GetMaster().Exec(`
			call r_tree_traversal('insert',:Id, :AppId, :ParentId,:Name, :CreateAt, :UpdateAt);`,
		map[string]interface{}{
			"Id":       category.Id,
			"AppId":    category.AppId,
			"ParentId": category.ParentId,
			"Name":     category.Name,
			"CreateAt": category.CreateAt,
			"UpdateAt": category.UpdateAt,
		}); err != nil {

		//result.Err = model.NewAppError("SqlCategoryStore.CreateCategory", "store.sql_category.create_category.app_error", nil, err.Error(), http.StatusInternalServerError)
	}
	//})

	return s.Get(category.Id)
}

func (s SqlCategoryStore) Delete(category *model.Category) store.StoreChannel {
	return store.Do(func(result *store.StoreResult) {
		if _, err := s.GetMaster().Exec(`
			call r_tree_traversal('delete', :Id, :AppId, :ParentId, :Name, :CreateAt, :UpdateAt);`,
			map[string]interface{}{
				"Id":       category.Id,
				"AppId":    category.AppId,
				"ParentId": category.ParentId,
				"Name":     category.Name,
				"CreateAt": category.CreateAt,
				"UpdateAt": category.UpdateAt,
			}); err != nil {
			result.Err = model.NewAppError("SqlCategoryStore.DeleteCategory", "store.sql_category.delete_category.app_error", nil, err.Error(), http.StatusInternalServerError)
		}
	})
}

func (s SqlCategoryStore) Move(category *model.Category) store.StoreChannel {
	return store.Do(func(result *store.StoreResult) {
		if _, err := s.GetMaster().Exec(`
			call r_tree_traversal('move',:Id, :AppId, :ParentId, :Name, :CreateAt, :UpdateAt);`,
			map[string]interface{}{
				"Id":       category.Id,
				"AppId":    category.AppId,
				"ParentId": category.ParentId,
				"Name":     category.Name,
				"CreateAt": category.CreateAt,
				"UpdateAt": category.UpdateAt,
			}); err != nil {
			result.Err = model.NewAppError("SqlCategoryStore.MoveCategory", "store.sql_category.move_category.app_error", nil, err.Error(), http.StatusInternalServerError)
		}
	})
}

func (s SqlCategoryStore) Order(category *model.Category) store.StoreChannel {
	return store.Do(func(result *store.StoreResult) {
		if _, err := s.GetMaster().Exec(`
			call r_tree_traversal('order', :Id, :AppId, :ParentId, :Name, :CreateAt, :UpdateAt);`,
			map[string]interface{}{
				"Id":       category.Id,
				"AppId":    category.AppId,
				"ParentId": category.DestinationId,
				"Name":     category.Name,
				"CreateAt": category.CreateAt,
				"UpdateAt": category.UpdateAt,
			}); err != nil {
			result.Err = model.NewAppError("SqlCategoryStore.OrderCategory", "store.sql_category.order_category.app_error", nil, err.Error(), http.StatusInternalServerError)
		}
	})
}
