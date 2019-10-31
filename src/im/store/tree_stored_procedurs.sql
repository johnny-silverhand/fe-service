use fe;
drop procedure if exists `r_tree_traversal`;

-- --------------------------------------------------------------------------------
-- Routine DDL
-- Note: comments before and after the routine body will not be stored by the server
-- --------------------------------------------------------------------------------
DELIMITER $$

CREATE PROCEDURE `r_tree_traversal`(

    IN Crud VARCHAR(10),

    -- Category columns --
    IN node_id VARCHAR(26),
    in client_id VARCHAR(26),
    in parent_id VARCHAR(26),
    in node_name VARCHAR(26),
    in created_at BIGINT,
    in updated_at BIGINT

)
BEGIN

    DECLARE new_lft, new_rgt, width, has_leafs, old_lft, old_rgt, parent_rgt, subtree_size,new_depth INTEGER;
    DECLARE superior, superior_parent Varchar(26);

    CASE Crud

        WHEN 'insert' THEN
            -- INSERT OPERATION FOR CATEGORIES TABLE
            SELECT Rgt,`Depth` INTO new_lft, new_depth FROM categories c WHERE Id = parent_id;

            IF (new_lft >= 1) THEN
                UPDATE categories SET Rgt = Rgt + 2 WHERE Rgt >= new_lft;
                UPDATE categories SET Lft = Lft + 2 WHERE Lft > new_lft;
                INSERT INTO
                    categories (Id,ClientId, Name, ParentId, CreateAt, UpdateAt, Lft, Rgt, `Depth`)
                VALUES (node_id,client_id,node_name,parent_id,created_at,updated_at, new_lft, (new_lft + 1),new_depth + 1);
            ELSE
                UPDATE categories SET Rgt = Rgt + 2;
                UPDATE categories SET Lft = Lft + 2;
                insert into	categories (Id,ClientId, Name, ParentId, CreateAt, UpdateAt, Lft, Rgt, `Depth`)
                VALUES (node_id,client_id,node_name,parent_id,created_at,updated_at, 1,2,1);

            END IF;

        WHEN 'delete' THEN
            -- DELETE OPERATION FOR CATEGORIES TABLE2
            SELECT Lft, Rgt, (Lft - Rgt), (Rgt - Lft + 1), ParentId
            INTO new_lft, new_rgt, has_leafs, width, superior_parent
            FROM categories WHERE Id = node_id;

            DELETE FROM categories WHERE id = node_id;

            IF (has_leafs = 1) THEN
                DELETE FROM categories WHERE Lft BETWEEN new_lft AND new_rgt;
                UPDATE categories SET Rgt = Rgt - width WHERE Rgt > new_rgt;
                UPDATE categories SET Lft = lft - width WHERE Lft > new_rgt;
            ELSE
                DELETE FROM categories WHERE lft = new_lft;
                UPDATE categories SET Rgt = Rgt - 1, Lft = Lft - 1, ParentId = superior_parent
                WHERE Lft BETWEEN new_lft AND new_rgt;
                UPDATE categories SET Rgt = Rgt - 2 WHERE Rgt > new_rgt;
                UPDATE categories SET Lft = Lft - 2 WHERE Lft > new_rgt;
            END IF;

        WHEN 'move' THEN

            IF (node_id != parent_id) THEN
                CREATE TEMPORARY TABLE IF NOT EXISTS categories_temp like categories;
                -- put subtree into temporary table
                INSERT INTO categories_temp (Id, Lft, Rgt, ParentId, Depth, Name, ClientId, CreateAt,UpdateAt)
                SELECT t1.Id,
                       (t1.Lft - (SELECT MIN(lft) FROM categories WHERE Id = node_id)) AS Lft,
                       (t1.Rgt - (SELECT MIN(lft) FROM categories WHERE Id = node_id)) AS Rgt,
                       t1.ParentId,
                       t1.Depth,
                       t1.Name,
                       t1.ClientId,
                       t1.CreateAt,
                       t1.UpdateAt
                FROM categories AS t1, categories AS t2
                WHERE t1.Lft BETWEEN t2.Lft AND t2.Rgt
                  AND t2.Id = node_id;

                DELETE FROM categories WHERE Id IN (SELECT Id FROM categories_temp);

                SELECT Rgt INTO parent_rgt FROM categories WHERE Id = parent_id;
                SET subtree_size = (SELECT (MAX(Rgt) + 1) FROM categories_temp);

                -- make a gap in the tree
                UPDATE categories
                SET Lft = (CASE WHEN Lft > parent_rgt THEN Lft + subtree_size ELSE Lft END),
                    Rgt = (CASE WHEN Rgt >= parent_rgt THEN Rgt + subtree_size ELSE Rgt END)
                WHERE Rgt >= parent_rgt;

                INSERT INTO categories (Id, Lft, Rgt, ParentId,Depth,Name, ClientId, CreateAt, UpdateAt)
                SELECT Id, Lft + parent_rgt, Rgt + parent_rgt, ParentId, Depth,Name,ClientId, CreateAt, UpdateAt
                FROM categories_temp;

                UPDATE categories
                SET Lft = (SELECT COUNT(*) FROM vw_lftrgt AS v WHERE v.Lft <= categories.Lft),
                    Rgt = (SELECT COUNT(*) FROM vw_lftrgt AS v WHERE v.Lft <= categories.Rgt);

                DELETE FROM categories_temp;
                UPDATE categories
                SET ParentId = parent_id
                WHERE Id = node_id;
            END IF;

        WHEN 'order' THEN

            SELECT Lft, Rgt, (Rgt - Lft + 1), ParentId INTO old_lft, old_rgt, width, superior
            FROM categories WHERE Id = node_id;

            -- is parent
            SELECT ParentId INTO superior_parent FROM categories WHERE Id = parent_id;

            IF (superior = superior_parent) THEN
                SELECT (Rgt + 1) INTO new_lft FROM categories WHERE Id = parent_id;
            ELSE
                SELECT (Lft + 1) INTO new_lft FROM categories WHERE Id = parent_id;
            END IF;

            IF (new_lft != old_lft) THEN
                CREATE TEMPORARY TABLE IF NOT EXISTS categories_temp LIKE categories;

                INSERT INTO categories_temp (Id, Lft, Rgt, ParentId, ClientId, Depth, Name, CreateAt, UpdateAt)
                    SELECT t1.Id,
                (t1.Lft - (SELECT MIN(Lft) FROM categories WHERE Id = node_id)) AS Lft,
                (t1.Rgt - (SELECT MIN(Lft) FROM categories WHERE Id = node_id)) AS Rgt,
                    t1.ParentId,
                    t1.ClientId,
                    t1.Depth,
                    t1.Name,
                    t1.CreateAt,
                    t1.UpdateAt
                FROM categories AS t1, categories AS t2
                WHERE t1.Lft BETWEEN t2.Lft AND t2.Rgt AND t2.Id = node_id;


                DELETE FROM categories WHERE Id IN (SELECT Id FROM categories_temp);

                IF(new_lft < old_lft) THEN -- move up
                    UPDATE categories SET Lft = Lft + width WHERE Lft >= new_lft AND lft < old_lft;
                    UPDATE categories SET Rgt = Rgt + width WHERE Rgt > new_lft AND Rgt < old_rgt;
                    UPDATE categories_temp SET Lft = new_lft + Lft, Rgt = new_lft + Rgt;
                END IF;

                IF(new_lft > old_lft) THEN -- move down
                    UPDATE categories SET Lft = Lft - width WHERE Lft > old_lft AND Lft < new_lft;
                    UPDATE categories SET Rgt = Rgt - width WHERE Rgt > old_rgt AND Rgt < new_lft;
                    UPDATE categories_temp SET Lft = (new_lft - width) + Lft, Rgt = (new_lft - width) + Rgt;
                END IF;

                INSERT INTO categories (Id, lft, rgt, ParentId, ClientId, Name, Depth, CreateAt, UpdateAt )
                SELECT Id, lft, rgt, ParentId, ClientId, Name, Depth, CreateAt, UpdateAt
                FROM categories_temp;

                DELETE FROM categories_temp;
            END IF;
        END CASE;

end
