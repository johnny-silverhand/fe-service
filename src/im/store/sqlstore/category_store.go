package sqlstore

import (
	"bytes"
	"database/sql"
	"errors"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"im/model"
	"im/store"
	"net/http"
	"strconv"
)

var categorySQL CategorySQL

func init() {
	categorySQL.tblName = "categories"
}

type CategorySQL struct {
	tblName   string
	Id        string
	clientId  string
	ParentId  string
	createdAt *int64
	updatedAt *int64

	//deletedAt *int64
}

func (t *CategorySQL) SelectSQL() string {
	return "SELECT `Id`, `Name`, `ParentId`, `Depth`, `Lft`, `Rgt` FROM " +
		t.tblName + " WHERE `ClientId`=\"" + t.clientId + "\" AND "
}

func (t *CategorySQL) SelectSQL2() string {
	return "SELECT `Id`, `Name`, `ParentId`, `Depth`, `Lft`, `Rgt` FROM " +
		t.tblName + " WHERE "
}

func (t *CategorySQL) SelectChildrenSQL() string {
	return "SELECT `Children`.`Id`, `Children`.`Name`, `Children`.`ParentId`, `Children`.`Depth`, `Children`.`Lft`, `Children`.`Rgt` FROM " +
		t.tblName + " AS `Children`, " + t.tblName + " AS `Parent` WHERE `Children`.`ClientId`=" + t.clientId +
		" AND `Parent`.`ClientId`=\"" + t.clientId + "\" AND "
}
func (t *CategorySQL) SelectParentsSQL() string {
	return "SELECT `Parent`.`Id`, `Parent`.`Name`, `Parent`.`ParentId`, `Parent`.`Depth`, `Parent`.`Lft`, `Parent`.`Rgt` FROM " +
		t.tblName + " AS `Children`, " + t.tblName + " AS `Parent` WHERE `Children`.`ClientId`=\"" + t.clientId +
		"\" AND `Parent`.`ClientId`=\"" + t.clientId + "\" AND "
}
func (t *CategorySQL) MoveOnAddSQL() string {
	return "UPDATE " + t.tblName + " SET `Lft`=CASE WHEN `Lft`>? THEN `Lft`+2 ELSE `Lft` END, `Rgt`=CASE WHEN `Rgt`>? " +
		"THEN `Rgt`+2 ELSE `Rgt` END WHERE `ClientId`=\"" + t.clientId + "\""
}
func (t *CategorySQL) MoveOnDeleteSQL() string {
	return "UPDATE " + t.tblName + " SET `Lft`=CASE WHEN `Lft`>? THEN `Lft`-? ELSE `Lft` END, `Rgt`=CASE WHEN `Rgt`>? " +
		"THEN `Rgt`-? ELSE `Rgt` END WHERE `ClientId`=\"" + t.clientId + "\""
}
func (t *CategorySQL) MoveOnLevelUpSQL() string {
	return "UPDATE " + t.tblName + " SET `Lft`=`Lft`-1, `Rgt`=`Rgt`-1, `Depth`=`Depth`-1 WHERE `ClientId`=" + t.clientId +
		" AND `Lft` BETWEEN ? AND ?"
}
func (t *CategorySQL) UpdateParentIdSQL() string {
	return "UPDATE " + t.tblName + " AS `Children`, " + t.tblName + " AS `Parent` SET `Children`.`ParentId`=`Parent`.`ParentId` " +
		"WHERE `Children`.`ClientId`=\"" + t.clientId + "\" AND `Parent`.`ClientId`=\"" + t.clientId +
		"\" AND `Children`.`ParentId`=`Parent`.`Id` AND `Children`.`Lft` BETWEEN ? AND ?"
}
func (t *CategorySQL) InsertSQL() string {
	return "INSERT INTO " + t.tblName + "(`Id` ,`Name`, `ParentId`, `Depth`, `Lft`, `Rgt`, `ClientId`, `CreateAt`, `UpdateAt`) " +
		"VALUES(?,?,?,?,?,?,?," + strconv.FormatInt(*t.createdAt, 10) + "," + strconv.FormatInt(*t.updatedAt, 10) + ")"
}

func (t *CategorySQL) DeleteSQL() string {
	return "DELETE FROM " + t.tblName + " WHERE `ClientId`='" + t.clientId + "' AND "
}

// Node detail with path from root to node
type Node struct {
	ID          string
	Name        string
	ParentID    string
	Depth       int
	Path        []int
	PathName    []string
	NumChildren int
	Lft         *int
	Rgt         *int
}

type SqlCategoryStore struct {
	SqlStore
}

func NewSqlCategoryStore(sqlStore SqlStore) store.CategoryStore {
	cs := &SqlCategoryStore{SqlStore: sqlStore}

	for _, db := range sqlStore.GetAllConns() {
		table := db.AddTableWithName(model.Category{}, "Categories").SetKeys(false, "Id")

		table.ColMap("Id").SetMaxSize(26)
		table.ColMap("ClientId").SetMaxSize(32)
		table.ColMap("Name").SetMaxSize(32)
		table.ColMap("ParentId").SetMaxSize(26)
		table.ColMap("CreateAt").SetMaxSize(26)
		table.ColMap("UpdateAt").SetMaxSize(26)
		//table.ColMap("DeleteAt").SetMaxSize(26)

	}

	return cs
}

func (s SqlCategoryStore) CreateIndexesIfNotExists() {
	s.CreateIndexIfNotExists("idx_categories_client_id", "Categories", "ClientId")
	s.CreateIndexIfNotExists("idx_categories_update_at", "Categories", "UpdateAt")
	s.CreateIndexIfNotExists("idx_categories_create_at", "Categories", "CreateAt")
	//s.CreateIndexIfNotExists("idx_categories_delete_at", "Categories", "DeleteAt")
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

func (s SqlCategoryStore) GetAllByClientId(clientId string) store.StoreChannel {
	return store.Do(func(result *store.StoreResult) {
		var query = `select c.*, childs.cnt as CntChild
					 from categories c 
					 left join (
						select ParentId, count(Id) cnt
						from categories
						where ParentId is not null
						group by ParentId ) childs on c.Id = childs.ParentId
					 where ClientId = :ClientId
					 order by UpdateAt Desc, Id Desc`
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
		var query = `select c.*, childs.cnt as CntChild
					 from categories c 
					 left join (
						select ParentId, count(Id) cnt
						from categories
						where ParentId is not null
						group by ParentId ) childs on c.Id = childs.ParentId
					 where ClientId = :ClientId
					 order by UpdateAt Desc, Id Desc
					 limit :Limit
					 offset :Offset`
		var categories []*model.Category
		if _, err := s.GetReplica().Select(&categories,
			query, map[string]interface{}{"ClientId": clientId, "Limit": limit, "Offset": offset}); err != nil {
			result.Err = model.NewAppError("SqlCategoryStore.GetAllByClientIdPage", "store.sql_category.get_all_by_client_id_page.app_error", nil, err.Error(), http.StatusInternalServerError)
		} else {
			result.Data = categories
		}
	})
}

func (s SqlCategoryStore) Delete(category *model.Category) store.StoreChannel {

	return store.Do(func(result *store.StoreResult) {
		db := s.GetMaster().Db
		categorySQL.clientId = category.ClientId
		categorySQL.RemoveNodeAndDescendants(db, category.Id)
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

func (t *CategorySQL) GetNodeDetail(db *sql.DB, id string) (*Node, error) {
	var sql bytes.Buffer
	sql.WriteString(t.SelectParentsSQL())
	sql.WriteString("`Children`.`Id`=? AND `Children`.`Lft` BETWEEN `Parent`.`Lft` AND `Parent`.`Rgt` ORDER BY `Lft` ASC")

	rows, err := query(db, sql.String(), id)

	if err != nil {
		return nil, err
	}
	if len(rows) < 1 {
		return nil, nil
	}

	path := make([]int, 0, len(rows))
	pathName := make([]string, 0, len(rows))
	for _, r := range rows {
		path = append(path, atoi(r["Id"]))
		pathName = append(pathName, r["Name"])
	}

	r := rows[len(rows)-1]

	left := atoi(r["Lft"])
	rigth := atoi(r["Rgt"])

	node := &Node{
		ID:          r["Id"],
		Name:        r["Name"],
		ParentID:    r["ParentId"],
		Depth:       atoi(r["Depth"]),
		Path:        path,
		PathName:    pathName,
		NumChildren: (atoi(r["Rgt"]) - atoi(r["Lft"]) - 1) / 2,
		Lft:         &left,
		Rgt:         &rigth,
	}

	return node, nil
}

func (t *CategorySQL) AddRootNode(db *sql.DB, name string) (*string, error) {
	// move all other nodes to right, if exits
	var sql bytes.Buffer

	sql.WriteString(t.MoveOnAddSQL())

	/*
		return "UPDATE Categories SET `Lft`= CASE WHEN `Lft`> 0 THEN `Lft`+2 ELSE `Lft` END, `Rgt`=CASE WHEN `Rgt`>? " +
			"THEN `Rgt`+2 ELSE `Rgt` END WHERE `ClientId`= ClientId"
	*/
	_, err := db.Exec(sql.String(), 0, 0)
	if err != nil {
		return nil, err
	}
	sql.Reset()

	// insert root
	sql.WriteString(t.InsertSQL())
	args := []interface{}{t.Id, name, t.ParentId, 1, 1, 2, t.clientId} // parentID is nil

	result, err := db.Exec(sql.String(), args...)
	if err != nil {
		return nil, nil
	}
	affected, _ := result.RowsAffected()
	if affected < 1 {
		return nil, errors.New("nested: inserting root affected none")
	}
	return &t.Id, err
}

// GetChildren returns all immediate children of node
func (t *CategorySQL) GetChildren(db *sql.DB, id string) ([]Node, error) {
	var sql bytes.Buffer
	sql.WriteString(t.SelectSQL())
	sql.WriteString("`ParentId`=?")

	rows, err := query(db, sql.String(), id)
	if err != nil {
		return nil, err
	}

	children := make([]Node, 0, len(rows))
	for _, r := range rows {
		children = append(children, Node{
			ID:          r["Id"],
			Name:        r["Name"],
			ParentID:    r["ParentId"],
			Depth:       atoi(r["Depth"]),
			NumChildren: (atoi(r["Rgt"]) - atoi(r["Lft"]) - 1) / 2,
		})
	}
	return children, nil
}

// GetDescendants returns sub tree of node
func (t *CategorySQL) GetDescendants(db *sql.DB, id int) ([]Node, error) {
	var sql bytes.Buffer
	sql.WriteString(t.SelectChildrenSQL())
	sql.WriteString("`Parent`.`Id`=? AND `Children`.`Lft` BETWEEN `Parent`.`Lft` AND `Parent`.`Rgt`")

	rows, err := query(db, sql.String(), id)
	if err != nil {
		return nil, err
	}

	descendants := make([]Node, 0, len(rows))
	for _, r := range rows {
		descendants = append(descendants, Node{
			ID:          r["Id"],
			Name:        r["Name"],
			ParentID:    r["ParentId"],
			Depth:       atoi(r["Depth"]),
			NumChildren: (atoi(r["Rgt"]) - atoi(r["Lft"]) - 1) / 2,
		})
	}
	return descendants, nil
}

// GetNodesByDepth returns all nodes of certain depth
func (t *CategorySQL) GetNodesByDepth(db *sql.DB, depth int) ([]Node, error) {
	sql := bytes.NewBufferString(t.SelectSQL())
	sql.WriteString("`Depth`=?")

	rows, err := query(db, sql.String(), depth)
	if err != nil {
		return nil, err
	}

	nodes := make([]Node, 0, len(rows))
	for _, r := range rows {
		nodes = append(nodes, Node{
			ID:          r["Id"],
			Name:        r["Name"],
			ParentID:    r["ParentId"],
			Depth:       atoi(r["Depth"]),
			NumChildren: (atoi(r["Rgt"]) - atoi(r["Lft"]) - 1) / 2,
		})
	}
	return nodes, nil
}

// AddNodeByParent adds a new node with certain parent, new node will be the last child of the parent.
func (t *CategorySQL) AddNodeByParent(db *sql.DB, name string, parentID string) (*string, error) {
	// query parent
	var sql bytes.Buffer
	sql.WriteString(t.SelectSQL())
	sql.WriteString("`Id`=?")

	rows, err := query(db, sql.String(), parentID)
	if err != nil {
		return nil, err
	}
	if len(rows) < 1 {
		return nil, errors.New("nested: adding node with parent does not exist")
	}
	parentRight := atoi(rows[0]["Rgt"])
	parentDepth := atoi(rows[0]["Depth"])
	sql.Reset()

	// moves nodes on the right to right by 2,
	sql.WriteString(t.MoveOnAddSQL())

	_, err = db.Exec(sql.String(), parentRight, parentRight-1) //  move right index of parent to right by 2
	if err != nil {
		return nil, err
	}
	sql.Reset()

	// insert new node
	sql.WriteString(t.InsertSQL())
	args := []interface{}{t.Id, name, parentID, parentDepth + 1, parentRight, parentRight + 1, t.clientId}

	r, err := db.Exec(sql.String(), args...)
	if err != nil {
		return nil, err
	}
	row, _ := r.RowsAffected()
	if row != 1 {
		return nil, errors.New("nested: inserting affected none")
	}
	return &t.Id, err
}

// RemoveNodeAndDescendants removes node and all its descendants -- it removes the whole subtree.
func (t *CategorySQL) RemoveNodeAndDescendants(db *sql.DB, id string) error {
	// query deleting node
	var sql bytes.Buffer
	sql.WriteString(t.SelectSQL())
	sql.WriteString("`Id`=?")

	rows, err := query(db, sql.String(), id)
	if err != nil {
		return err
	}
	if len(rows) < 1 {
		return errors.New("nested: deleting node does not exist")
	}

	left := atoi(rows[0]["Lft"])
	right := atoi(rows[0]["Rgt"])
	width := right - left + 1
	sql.Reset()

	// delete node and all its descendants
	sql.WriteString(t.DeleteSQL())
	sql.WriteString("`Lft` BETWEEN ? AND ?")

	result, err := db.Exec(sql.String(), left, right)
	if err != nil {
		return err
	}
	affected, _ := result.RowsAffected()
	if affected < 1 {
		return errors.New("nested: deleting node affected none")
	}
	sql.Reset()

	// move all node on the right to left
	sql.WriteString(t.MoveOnDeleteSQL())

	_, err = db.Exec(sql.String(), right, width, right, width)
	if err != nil {
		return err
	}
	return nil
}

// RemoveOneNode removes one node and move all its descentants 1 level up -- it removes the certain node from the tree only.
func (t *CategorySQL) RemoveOneNode(db *sql.DB, id string) error {
	// query deleting node
	var sql bytes.Buffer
	sql.WriteString(t.SelectSQL())
	sql.WriteString("`Id`=?")

	rows, err := query(db, sql.String(), id)
	if err != nil {
		return err
	}
	if len(rows) < 1 {
		return errors.New("nested: deleting node does not exist")
	}
	sql.Reset()

	left := atoi(rows[0]["Lft"])
	right := atoi(rows[0]["Rgt"])

	// update pid of its descendants
	sql.WriteString(t.UpdateParentIdSQL())

	_, err = db.Exec(sql.String(), left, right)
	if err != nil {
		return err
	}
	sql.Reset()

	// delete node
	sql.WriteString(t.DeleteSQL())
	sql.WriteString("`Id`=?")

	r, err := db.Exec(sql.String(), id)
	if err != nil {
		return err
	}
	affected, _ := r.RowsAffected()
	if affected < 1 {
		return errors.New("nested: deleting node affected none")
	}
	sql.Reset()

	// move all its descentants left and up 1 step
	sql.WriteString(t.MoveOnLevelUpSQL())

	_, err = db.Exec(sql.String(), left, right) // could affect none
	if err != nil {
		return err
	}
	sql.Reset()

	// move all other nodes on the rgt to left by 2 steps
	sql.WriteString(t.MoveOnDeleteSQL())

	_, err = db.Exec(sql.String(), right, 2, right, 2) // could affect none
	if err != nil {
		return err
	}

	return nil
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
			where lft >= :Lft and rgt <= :Rgt and clientId = :ClientId 
			order by lft desc`,
			map[string]interface{}{
				"Lft":      root.Lft,
				"Rgt":      root.Rgt,
				"ClientId": root.ClientId,
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
		var (
			categories []*model.Category
		)
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

//TODO: Ниже дохера копипасты, каюсь. Поправим после показа. (с) Автор комита

func (s SqlCategoryStore) CreateCategoryBySp(category *model.Category) store.StoreChannel {

	if len(category.Id) == 0 {
		category.PreSave()
	}

	//call r_tree_traversal(:Crud, :Id, :clientId, :parentId, :name, :createAt, :updateAt)
	_, err := s.GetMaster().Exec(`
			call r_tree_traversal('insert',:Id, :ClientID, :ParentId,:Name,:CreateAt,:UpdateAt);`,
		map[string]interface{}{
			"Id":       category.Id,
			"ClientID": category.ClientId,
			"ParentId": category.ParentId,
			"Name":     category.Name,
			"CreateAt": category.CreateAt,
			"UpdateAt": category.UpdateAt,
		})
	if err != nil {
		fmt.Print("error")
	}

	return s.Get(category.Id)
}

func (s SqlCategoryStore) DeleteCategoryBySp(category *model.Category) store.StoreChannel {
	return store.Do(func(result *store.StoreResult) {
		//call r_tree_traversal(:Crud, :Id, :clientId, :parentId, :name, :createAt, :updateAt)
		_, err := s.GetMaster().Exec(`
			call r_tree_traversal('delete',:Id, :ClientID, :ParentId,:Name,:CreateAt,:UpdateAt);`,
			map[string]interface{}{
				"Id":       category.Id,
				"ClientID": category.ClientId,
				"ParentId": category.ParentId,
				"Name":     category.Name,
				"CreateAt": category.CreateAt,
				"UpdateAt": category.UpdateAt,
			})
		if err != nil {
			fmt.Print("error")
		}
	})
}

func (s SqlCategoryStore) MoveCategoryBySp(category *model.Category) store.StoreChannel {
	return store.Do(func(result *store.StoreResult) {
		_, err := s.GetMaster().Exec(`
			call r_tree_traversal('move',:Id, :ClientID, :ParentId,:Name,:CreateAt,:UpdateAt);`,
			map[string]interface{}{
				"Id":       category.Id,
				"ClientID": category.ClientId,
				"ParentId": category.ParentId,
				"Name":     category.Name,
				"CreateAt": category.CreateAt,
				"UpdateAt": category.UpdateAt,
			})
		if err != nil {
			fmt.Print("error")
		}
	})
}

func (s SqlCategoryStore) OrderCategoryBySp(category *model.Category) store.StoreChannel {
	return store.Do(func(result *store.StoreResult) {
		_, err := s.GetMaster().Exec(`
			call r_tree_traversal('order',:Id, :ClientID, :ParentId,:Name,:CreateAt,:UpdateAt);`,
			map[string]interface{}{
				"Id":       category.Id,
				"ClientID": category.ClientId,
				"ParentId": category.DestinationId,
				"Name":     category.Name,
				"CreateAt": category.CreateAt,
				"UpdateAt": category.UpdateAt,
			})
		if err != nil {
			fmt.Print("error")
		}
	})
}
