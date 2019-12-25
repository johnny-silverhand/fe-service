package sqlstore

import (
	"database/sql"
	"fmt"
	sq "github.com/Masterminds/squirrel"
	"im/mlog"
	"im/model"
	"im/store"
	"net/http"
	"strings"
)

var (
	PRODUCT_SEARCH_TYPE_NAMES = []string{"Name"}
)

type SqlProductStore struct {
	SqlStore

	productsQuery sq.SelectBuilder
}

func NewSqlProductStore(sqlStore SqlStore) store.ProductStore {
	s := &SqlProductStore{
		SqlStore: sqlStore,
	}

	s.productsQuery = s.getQueryBuilder().
		Select("p.*").
		From("Products p")

	for _, db := range sqlStore.GetAllConns() {
		table := db.AddTableWithName(model.Product{}, "Products").SetKeys(false, "Id")

		table.ColMap("Id").SetMaxSize(26)
		table.ColMap("AppId").SetMaxSize(26)
		table.ColMap("Name").SetMaxSize(255)
		table.ColMap("Preview").SetMaxSize(2000)
		table.ColMap("Description").SetMaxSize(32)

		table.ColMap("CategoryId").SetMaxSize(26)
		table.ColMap("FileIds").SetMaxSize(150)

	}

	return s
}

func (s SqlProductStore) CreateIndexesIfNotExists() {
	s.CreateIndexIfNotExists("idx_products_clientid", "Products", "AppId")
	s.CreateIndexIfNotExists("idx_products_update_at", "Products", "UpdateAt")
	s.CreateIndexIfNotExists("idx_products_create_at", "Products", "CreateAt")
	s.CreateIndexIfNotExists("idx_products_delete_at", "Products", "DeleteAt")
}

func (s SqlProductStore) GetExtras(product *model.Product) store.StoreChannel {
	return store.Do(func(result *store.StoreResult) {
		var products []*model.Product
		var extraIds []int
		var query string = "SELECT `ProductExtraId` FROM `products_extras` WHERE `ProductId` = :ProductId"

		if _, err := s.GetMaster().Select(&extraIds, query, map[string]interface{}{"ProductId": product.Id}); err != nil {
			result.Err = model.NewAppError("SqlProductStore.GetExtras", "store.sql_products.select_extras.app_error", nil, err.Error(), http.StatusInternalServerError)
			return
		} else {
			var inQueryList []string
			queryArgs := make(map[string]interface{})
			for i, extraId := range extraIds {
				inQueryList = append(inQueryList, fmt.Sprintf(":pid%v", i))
				queryArgs[fmt.Sprintf("pid%v", i)] = extraId
			}
			inQuery := strings.Join(inQueryList, ", ")

			if _, err = s.GetMaster().Select(&products, "SELECT * FROM `products` WHERE `Id` IN ("+inQuery+")", queryArgs); err != nil {
				result.Err = model.NewAppError("SqlProductStore.GetExtras", "store.sql_products.get_extras.app_error", nil, err.Error(), http.StatusInternalServerError)
				return
			} else {
				result.Data = products
				return
			}
		}
	})

	/*var inQueryList []string
	  queryArgs := make(map[string]interface{})
	  for i, productId := range productsIds {
	      inQueryList = append(inQueryList, fmt.Sprintf(":pid%v", i))
	      queryArgs[fmt.Sprintf("pid%v", i)] = productId
	  }
	  inQuery := strings.Join(inQueryList, ", ")

	  return store.Do(func(result *store.StoreResult) {

	      var extraIds []int
	      var categoriesQuery string = "SELECT `CategoryId` FROM `products_extras` WHERE `ProductId` IN ("+ inQuery +") GROUP BY `CategoryId`"

	      if _, err := s.GetMaster().Select(&extraIds, categoriesQuery, queryArgs); err != nil {
	          result.Err = model.NewAppError("SqlProductStore.GetExtraProduct", "store.sql_products.get_extra_products.app_error", nil, err.Error(), http.StatusInternalServerError)
	      } else {
	          result.Data = extraIds
	      }
	  })*/
}

func (s SqlProductStore) Publish(product *model.Product) store.StoreChannel {
	return store.Do(func(result *store.StoreResult) {
		if _, err := s.GetMaster().Query("UPDATE `products` SET `Status`=? WHERE `Id`=?", true, product.Id); err != nil {
			result.Err = model.NewAppError("SqlProductStore.Publish", "store.sql_products.publish.app_error", nil, err.Error(), http.StatusInternalServerError)
		}
	})
}

func (s *SqlProductStore) Save(product *model.Product) store.StoreChannel {
	return store.Do(func(result *store.StoreResult) {
		if len(product.Id) > 0 {
			result.Err = model.NewAppError("SqlPostStore.Save", "store.sql_post.save.existing.app_error", nil, "id="+product.Id, http.StatusBadRequest)
			return
		}

		product.PreSave()

		if result.Err = product.IsValid(); result.Err != nil {
			return
		}

		if err := s.GetMaster().Insert(product); err != nil {
			result.Err = model.NewAppError("SqlPostStore.Save", "store.sql_post.save.app_error", nil, "id="+product.Id+", "+err.Error(), http.StatusInternalServerError)
		} else {
			result.Data = product
		}
	})
}

func (s *SqlProductStore) Update(newProduct *model.Product) store.StoreChannel {
	return store.Do(func(result *store.StoreResult) {
		newProduct.UpdateAt = model.GetMillis()
		newProduct.PreCommit()

		if _, err := s.GetMaster().Update(newProduct); err != nil {
			result.Err = model.NewAppError("SqlProductStore.Update", "store.sql_post.update.app_error", nil, "id="+newProduct.Id+", "+err.Error(), http.StatusInternalServerError)
		} else {
			result.Data = newProduct
		}
	})
}

func (s *SqlProductStore) Get(id string) store.StoreChannel {
	return store.Do(func(result *store.StoreResult) {
		var product *model.Product
		if err := s.GetReplica().SelectOne(&product,
			`SELECT *
					FROM Products
					WHERE Id = :Id  AND DeleteAt = 0`, map[string]interface{}{"Id": id}); err != nil {
			if err == sql.ErrNoRows {
				result.Err = model.NewAppError("SqlProductStore.Get", "store.sql_products.get.app_error", nil, err.Error(), http.StatusNotFound)
			} else {
				result.Err = model.NewAppError("SqlProductStore.Get", "store.sql_products.get.app_error", nil, err.Error(), http.StatusInternalServerError)
			}
		} else {
			result.Data = product
		}
	})
}

func (s SqlProductStore) GetAllByCategoryId(categoryId string, offset int, limit int, allowFromCache bool) store.StoreChannel {
	return store.Do(func(result *store.StoreResult) {

		//add cache with allowFromCache

		var products []*model.Product
		query := `SELECT * from Products WHERE CategoryId = :CategoryId AND DeleteAt = 0`
		if _, err := s.GetReplica().Select(&products, query, map[string]interface{}{"CategoryId": categoryId}); err != nil {
			result.Err = model.NewAppError("SqlProductStore.GetAllByCategoryId", "store.sql_products.get_all_by_category_id.app_error", nil, err.Error(), http.StatusNotFound)
		} else {

			list := model.NewProductList()

			for _, p := range products {
				list.AddProduct(p)
				list.AddOrder(p.Id)
			}

			list.MakeNonNil()

			result.Data = list
		}
	})
}

func (s SqlProductStore) GetAllPage(offset int, limit int, options *model.ProductGetOptions) store.StoreChannel {
	return store.Do(func(result *store.StoreResult) {

		if options == nil {
			result.Err = model.NewAppError("SqlProductStore.GetAllPage", "store.sql_products.get_all_page.app_error", nil, "", http.StatusBadRequest)
			return
		}

		var officeQuery string
		var whereClause string
		queryArgs := make(map[string]interface{})

		if options.OfficeId != "" {
			officeQuery = " INNER JOIN Offices o ON o.Id = :OfficeId "
		} else {
			officeQuery = ""
		}

		if options.AppId != "" {
			//applicationQuery = " INNER JOIN Applications a ON a.Id = :AppId "
			whereClause = whereClause + " p.AppId = :AppId AND "
		}

		if options.CategoryId != "" {
			var rootCategory *model.Category
			if err := s.GetMaster().SelectOne(&rootCategory, `SELECT * FROM Categories WHERE Id = :Id`, map[string]interface{}{"Id": options.CategoryId}); err != nil {
				result.Err = model.NewAppError("SqlProductStore.GetAllPage",
					"store.sql_products.get_category.app_error", nil, err.Error(), http.StatusInternalServerError)
				return
			}

			var categories []*model.Category
			if _, err := s.GetMaster().Select(&categories, `SELECT * FROM Categories WHERE Lft >= :Lft and Rgt <= :Rgt and AppId = :AppId`,
				map[string]interface{}{
					"Lft":   rootCategory.Lft,
					"Rgt":   rootCategory.Rgt,
					"AppId": rootCategory.AppId,
				}); err != nil {
				if err == sql.ErrNoRows {
					result.Err = model.NewAppError("SqlProductStore.GetAllPage",
						"store.sql_products.get_categories.app_error", nil, err.Error(), http.StatusNotFound)
				} else {
					result.Err = model.NewAppError("SqlProductStore.GetAllPage", "store.sql_products.get_categories.app_error",
						nil, err.Error(), http.StatusInternalServerError)
				}
			}

			var inQueryList []string
			for i, category := range categories {
				inQueryList = append(inQueryList, fmt.Sprintf(":CategoryId%v", i))
				queryArgs[fmt.Sprintf("CategoryId%v", i)] = category.Id
			}
			inQuery := strings.Join(inQueryList, ", ")
			whereClause = whereClause + " p.CategoryId IN (" + inQuery + ") AND "
		}

		if options.Status != "" {
			whereClause = whereClause + " p.Status = :Status AND "
		}

		query := "SELECT p.* " +
			"FROM Products p " + officeQuery +
			"WHERE " + whereClause +
			" p.DeleteAt = 0 " +
			"LIMIT :Limit OFFSET :Offset"

		queryArgs["Limit"] = limit
		queryArgs["Offset"] = offset
		queryArgs["OfficeId"] = options.OfficeId
		queryArgs["AppId"] = options.AppId
		queryArgs["Status"] = options.Status

		var products []*model.Product

		if _, err := s.GetReplica().Select(&products, query, queryArgs); err != nil {
			result.Err = model.NewAppError("SqlProductStore.GetAllPage", "store.sql_products.get_all_page.app_error",
				nil, err.Error(),
				http.StatusInternalServerError)
		} else {

			list := model.NewProductList()

			for _, p := range products {
				list.AddProduct(p)
				list.AddOrder(p.Id)
			}

			list.MakeNonNil()

			result.Data = list
		}
	})
}

func (s SqlProductStore) GetAllPageByApp(offset int, limit int, order model.ColumnOrder, appId string) store.StoreChannel {
	return store.Do(func(result *store.StoreResult) {

		var products []*model.Product
		query := `SELECT *
                  FROM Products
                  WHERE AppId = :AppId
				  AND DeleteAt = 0
                  ORDER BY ` + order.Column + ` `

		query += order.Type + ` LIMIT :Limit OFFSET :Offset `

		if _, err := s.GetReplica().Select(&products, query, map[string]interface{}{"Limit": limit, "Offset": offset, "AppId": appId}); err != nil {
			result.Err = model.NewAppError("SqlProductStore.GetAllPageByClient", "store.sql_products.get_all_page_by_client.app_error",
				nil, err.Error(),
				http.StatusInternalServerError)
		} else {

			list := model.NewProductList()

			for _, p := range products {
				list.AddProduct(p)
				list.AddOrder(p.Id)
			}

			list.MakeNonNil()

			result.Data = list
		}
	})
}

func (s SqlProductStore) GetAllByAppId(appId string) store.StoreChannel {
	return store.Do(func(result *store.StoreResult) {
		var products []*model.Product
		if err := s.GetReplica().SelectOne(&products,
			`SELECT *
                    FROM Products
                    WHERE AppId = :AppId AND DeleteAt = 0`, map[string]interface{}{"appId": appId}); err != nil {
			result.Err = model.NewAppError("SqlProductStore.GetAllByAppId", "store.sql_products.get_all_by_app_id.app_error", nil, err.Error(), http.StatusNotFound)
		} else {

			list := model.NewProductList()

			for _, p := range products {
				list.AddProduct(p)
				list.AddOrder(p.Id)
			}

			list.MakeNonNil()

			result.Data = list
		}
	})
}

/*func (s SqlProductStore) GetAllByAppIdPage(clientId string, offset int, limit int, order model.ColumnOrder, categoryId string) store.StoreChannel {

	return store.Do(func(result *store.StoreResult) {

		var inQueryList []string
		var prefixSub string
		var products []*model.Product
		var descendants []*model.Category
		queryArgs := make(map[string]interface{})

		if len(categoryId) > 0 {
			descendantsQuery := `SELECT Children.* FROM Categories  AS Children, Categories AS Parent WHERE Parent.Id=:ParentId AND Children.Left BETWEEN Parent.Left AND Parent.Right`

			if _, err := s.GetReplica().Select(&descendants, descendantsQuery, map[string]interface{}{"ParentId": categoryId}); err != nil {
				result.Err = model.NewAppError("SqlProductStore.GetAllByAppIdPage", "store.sql_products.get_all_by_client_id_page.app_error", nil, err.Error(), http.StatusInternalServerError)
			} else {

				for i, node := range descendants {
					inQueryList = append(inQueryList, fmt.Sprintf(":CategoryId%v", i))
					queryArgs[fmt.Sprintf("CategoryId%v", i)] = node.Id
				}

				inQuery := strings.Join(inQueryList, ", ")
				prefixSub = ` AND CategoryId IN (` + inQuery + `)`
			}
		}

		query := `SELECT *
                  FROM Products
                  WHERE DeleteAt = 0`

		if len(prefixSub) > 0 {
			query += prefixSub
		}

		query += ` ORDER BY ` + order.Column + ` `
		query += order.Type + ` LIMIT :Limit OFFSET :Offset `

		queryArgs["AppId"] = clientId
		queryArgs["Limit"] = limit
		queryArgs["Offset"] = offset

		if _, err := s.GetReplica().Select(&products, query, queryArgs); err != nil {
			result.Err = model.NewAppError("SqlProductStore.GetAllByAppIdPage", "store.sql_products.get_all_by_client_id_page.app_error", nil, err.Error(), http.StatusInternalServerError)
		} else {

			list := model.NewProductList()

			for _, p := range products {
				list.AddProduct(p)
				list.AddOrder(p.Id)
			}

			list.MakeNonNil()

			result.Data = list

		}

	})
}*/

func (s *SqlProductStore) Delete(productId string, time int64, deleteByID string) store.StoreChannel {
	return store.Do(func(result *store.StoreResult) {

		appErr := func(errMsg string) *model.AppError {
			return model.NewAppError("SqlProductStore.Delete", "store.sql_product.delete.app_error", nil, "id="+productId+", err="+errMsg, http.StatusInternalServerError)
		}

		var product model.Product
		err := s.GetReplica().SelectOne(&product, "SELECT * FROM Products WHERE Id = :Id AND DeleteAt = 0", map[string]interface{}{"Id": productId})
		if err != nil {
			result.Err = appErr(err.Error())
		}

		_, err = s.GetMaster().Exec("UPDATE Products SET DeleteAt = :DeleteAt, UpdateAt = :UpdateAt WHERE Id = :Id", map[string]interface{}{"DeleteAt": time, "UpdateAt": time, "Id": productId})
		if err != nil {
			result.Err = appErr(err.Error())
		}
	})
}

func (s *SqlProductStore) Overwrite(product *model.Product) store.StoreChannel {
	return store.Do(func(result *store.StoreResult) {
		product.UpdateAt = model.GetMillis()

		if result.Err = product.IsValid(); result.Err != nil {
			return
		}

		if _, err := s.GetMaster().Update(product); err != nil {
			result.Err = model.NewAppError("SqlProductStore.Overwrite", "store.sql_product.overwrite.app_error", nil, "id="+product.Id+", "+err.Error(), http.StatusInternalServerError)
		} else {
			result.Data = product
		}
	})
}

func (s SqlProductStore) GetProductsByIds(productIds []string, allowFromCache bool) store.StoreChannel {
	return store.Do(func(result *store.StoreResult) {
		keys, params := MapStringsToQueryParams(productIds, "Product")

		query := `SELECT * FROM Products WHERE DeleteAt = 0 AND Id IN ` + keys

		var products []*model.Product
		_, err := s.GetReplica().Select(&products, query, params)

		if err != nil {
			mlog.Error(fmt.Sprint(err))
			result.Err = model.NewAppError("SqlProductStore.GetProductsByIds", "store.sql_product.get_products_by_ids.app_error", nil, "", http.StatusInternalServerError)
		} else {
			result.Data = products
		}
	})
}

func (s SqlProductStore) GetAll() store.StoreChannel {
	return store.Do(func(result *store.StoreResult) {
		var products []*model.Product
		query := `SELECT * FROM products WHERE DeleteAt = 0`
		if _, err := s.GetReplica().Select(&products, query, map[string]interface{}{}); err != nil {
			result.Err = model.NewAppError("SqlProductStore.GetAll", "store.sql_products.get_all.app_error",
				nil, err.Error(),
				http.StatusInternalServerError)
		} else {
			list := model.NewProductList()
			for _, p := range products {
				list.AddProduct(p)
				list.AddOrder(p.Id)
			}
			list.MakeNonNil()
			result.Data = list
		}
	})
}

func (s SqlProductStore) Search(categoryId, terms string, page, perPage int) store.StoreChannel {
	return store.Do(func(result *store.StoreResult) {

		var rootCategory *model.Category
		if err := s.GetMaster().SelectOne(&rootCategory, `SELECT * FROM Categories WHERE Id = :Id`, map[string]interface{}{"Id": categoryId}); err != nil {
			result.Err = model.NewAppError("SqlProductStore.GetAllPage",
				"store.sql_products.get_category.app_error", nil, err.Error(), http.StatusInternalServerError)
			return
		}

		var categories []*model.Category
		if _, err := s.GetMaster().Select(&categories, `SELECT * FROM Categories WHERE Lft >= :Lft and Rgt <= :Rgt and AppId = :AppId`,
			map[string]interface{}{
				"Lft":   rootCategory.Lft,
				"Rgt":   rootCategory.Rgt,
				"AppId": rootCategory.AppId,
			}); err != nil {
			if err == sql.ErrNoRows {
				result.Err = model.NewAppError("SqlProductStore.GetAllPage",
					"store.sql_products.get_categories.app_error", nil, err.Error(), http.StatusNotFound)
			} else {
				result.Err = model.NewAppError("SqlProductStore.GetAllPage", "store.sql_products.get_categories.app_error",
					nil, err.Error(), http.StatusInternalServerError)
			}
		}

		var inQueryList []string
		var categoryIds []string
		queryArgs := make(map[string]interface{})
		for i, category := range categories {
			inQueryList = append(inQueryList, fmt.Sprintf(":CategoryId%v", i))
			queryArgs[fmt.Sprintf("CategoryId%v", i)] = category.Id
			categoryIds = append(categoryIds, category.Id)
		}
		//inQuery := strings.Join(inQueryList, ", ")

		query := s.productsQuery.
			//Where(fmt.Sprintf("CategoryId IN (%s)", inQuery), queryArgs).
			OrderBy("UpdateAt DESC").
			Limit(uint64(perPage))
		*result = s.performProductSearch(query, terms, categoryIds)
	})
}

func (s SqlProductStore) performProductSearch(query sq.SelectBuilder, terms string, categoryIds []string) store.StoreResult {
	result := store.StoreResult{}

	// These chars must be removed from the like query.
	for _, c := range ignoreLikeSearchChar {
		terms = strings.Replace(terms, c, "", -1)
	}

	// These chars must be escaped in the like query.
	for _, c := range escapeLikeSearchChar {
		terms = strings.Replace(terms, c, "*"+c, -1)
	}

	searchType := PRODUCT_SEARCH_TYPE_NAMES

	isPostgreSQL := s.DriverName() == model.DATABASE_DRIVER_POSTGRES

	if strings.TrimSpace(terms) != "" {
		query = generateProductSearchQuery(query, strings.Fields(terms), searchType, isPostgreSQL, categoryIds)
	}

	queryString, args, err := query.ToSql()
	if err != nil {
		result.Err = model.NewAppError("SqlUserStore.Search", "store.sql_product.app_error", nil, err.Error(), http.StatusInternalServerError)
		return result
	}

	var products []*model.Product
	if _, err := s.GetReplica().Select(&products, queryString, args...); err != nil {
		result.Err = model.NewAppError("SqlUserStore.Search", "store.sql_product.search.app_error", nil,
			fmt.Sprintf("terms=%v, search_type=%v, %v", terms, searchType, err.Error()), http.StatusInternalServerError)
	} else {

		list := model.NewProductList()

		for _, p := range products {

			list.AddProduct(p)
			list.AddOrder(p.Id)
		}

		list.MakeNonNil()

		result.Data = list

	}

	return result
}

func generateProductSearchQuery(query sq.SelectBuilder, terms []string, fields []string, isPostgreSQL bool, categoryIds []string) sq.SelectBuilder {
	for _, term := range terms {
		searchFields := []string{}
		termArgs := []interface{}{}
		for _, field := range fields {
			if isPostgreSQL {
				searchFields = append(searchFields, fmt.Sprintf("lower(%s) LIKE lower(?) escape '*' ", field))
			} else {
				searchFields = append(searchFields, fmt.Sprintf("%s LIKE ? escape '*' ", field))
			}
			termArgs = append(termArgs, fmt.Sprintf("%%%s%%", term))
		}
		query = query.Where(fmt.Sprintf("(%s)", strings.Join(searchFields, " OR ")), termArgs...)
	}

	/*for _, categoryId := range categoryIds {
		searchFields := []string{}
		termArgs := []interface{}{}

			if isPostgreSQL {
				searchFields = append(searchFields, fmt.Sprintf("lower(%s) LIKE lower(?) escape '*' ", "CategoryId"))
			} else {
				searchFields = append(searchFields, fmt.Sprintf("%s LIKE ? escape '*' ", "CategoryId"))
			}
			termArgs = append(termArgs, fmt.Sprintf("%s%%", strings.TrimLeft(categoryId, "@")))

	}*/

	var inQueryList []string
	queryArgs := []interface{}{}
	//queryArgs := make(map[string]interface{})
	for _, categoryId := range categoryIds {
		/*inQueryList = append(inQueryList, fmt.Sprintf(":CategoryId%v", i))
		queryArgs[fmt.Sprintf("CategoryId%v", i)] = category.Id
		categoryIds = append(categoryIds, category.Id)*/
		inQueryList = append(inQueryList, "?")
		queryArgs = append(queryArgs, fmt.Sprintf("%s", strings.TrimLeft(categoryId, "@")))
	}
	inQuery := strings.Join(inQueryList, ", ")

	query = query.Where(fmt.Sprintf("CategoryId IN (%s)", inQuery), queryArgs...)

	return query
}
