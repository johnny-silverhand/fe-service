drop procedure if exists `r_tree_traversal`;

-- --------------------------------------------------------------------------------
-- Routine DDL
-- Note: comments before and after the routine body will not be stored by the server
-- --------------------------------------------------------------------------------
DELIMITER $$

CREATE PROCEDURE `r_tree_traversal`(

    IN CRUD VARCHAR(10),

    -- Category columns --
    IN node_id VARCHAR(26),
    in parent_id VARCHAR(26),
    in node_name VARCHAR(26),
    in app_id VARCHAR(26),
    in created_at BIGINT,
    in updated_at BIGINT

)
BEGIN

    DECLARE NewLft, NewRgt, Width, HasLeafs, OldLft, OldRgt, ParentRgt, SubtreeSize, NewDepth INTEGER;
    DECLARE Superior, SuperiorParent Varchar(26);

    CASE CRUD

        WHEN 'insert' THEN
            -- INSERT OPERATION FOR CATEGORIES TABLE
            SELECT Rgt,`Depth` INTO NewLft, NewDepth FROM Categories c WHERE Id = parent_id;

            IF (NewLft >= 1) THEN
                UPDATE Categories SET Rgt = Rgt + 2 WHERE Rgt >= NewLft;
                UPDATE Categories SET Lft = Lft + 2 WHERE Lft > NewLft;
                INSERT INTO
                    Categories (Id, AppId, Name, ParentId, CreateAt, UpdateAt, Lft, Rgt, `Depth`)
                VALUES (node_id, app_id, node_name, parent_id, created_at, updated_at, NewLft, (NewLft + 1), NewDepth + 1);
            ELSE
                UPDATE Categories SET Rgt = Rgt + 2;
                UPDATE Categories SET Lft = Lft + 2;
                insert into	Categories (Id, AppId, Name, ParentId, CreateAt, UpdateAt, Lft, Rgt, `Depth`)
                VALUES (node_id, app_id, node_name, parent_id, created_at, updated_at, 1, 2, 1);

            END IF;

        WHEN 'delete' THEN
            -- DELETE OPERATION FOR CATEGORIES TABLE2
            SELECT Lft, Rgt, (Lft - Rgt), (Rgt - Lft + 1), ParentId
            INTO NewLft, NewRgt, HasLeafs, Width, SuperiorParent
            FROM Categories WHERE Id = node_id;

            DELETE FROM Categories WHERE id = node_id;

            IF (HasLeafs = 1) THEN
                DELETE FROM Categories WHERE Lft BETWEEN NewLft AND NewRgt;
                UPDATE Categories SET Rgt = Rgt - Width WHERE Rgt > NewRgt;
                UPDATE Categories SET Lft = Lft - Width WHERE Lft > NewRgt;
            ELSE
                DELETE FROM Categories WHERE Lft = NewLft;
                UPDATE Categories SET Rgt = Rgt - 1, Lft = Lft - 1, ParentId = SuperiorParent
                WHERE Lft BETWEEN NewLft AND NewRgt;
                UPDATE Categories SET Rgt = Rgt - 2 WHERE Rgt > NewRgt;
                UPDATE Categories SET Lft = Lft - 2 WHERE Lft > NewRgt;
            END IF;

        WHEN 'move' THEN

            IF (node_id != parent_id) THEN
                CREATE TEMPORARY TABLE IF NOT EXISTS CategoriesTemporary like Categories;
                -- put subtree into temporary table
                INSERT INTO CategoriesTemporary (Id, Lft, Rgt, ParentId, Depth, Name, AppId, CreateAt, UpdateAt)
                SELECT t1.Id,
                       (t1.Lft - (SELECT MIN(Lft) FROM Categories WHERE Id = node_id)) AS Lft,
                       (t1.Rgt - (SELECT MIN(Lft) FROM Categories WHERE Id = node_id)) AS Rgt,
                       t1.ParentId,
                       t1.Depth,
                       t1.Name,
                       t1.AppId,
                       t1.CreateAt,
                       t1.UpdateAt
                FROM Categories AS t1, Categories AS t2
                WHERE t1.Lft BETWEEN t2.Lft AND t2.Rgt
                  AND t2.Id = node_id;

                DELETE FROM Categories WHERE Id IN (SELECT Id FROM CategoriesTemporary);

                SELECT Rgt INTO ParentRgt FROM Categories WHERE Id = parent_id;
                SET SubtreeSize = (SELECT (MAX(Rgt) + 1) FROM CategoriesTemporary);

                -- make a gap in the tree
                UPDATE Categories
                SET Lft = (CASE WHEN Lft > ParentRgt THEN Lft + SubtreeSize ELSE Lft END),
                    Rgt = (CASE WHEN Rgt >= ParentRgt THEN Rgt + SubtreeSize ELSE Rgt END)
                WHERE Rgt >= ParentRgt;

                INSERT INTO Categories (Id, Lft, Rgt, ParentId,Depth,Name, AppId, CreateAt, UpdateAt)
                SELECT Id, Lft + ParentRgt, Rgt + ParentRgt, ParentId, Depth,Name,AppId, CreateAt, UpdateAt
                FROM CategoriesTemporary;

                UPDATE Categories
                SET Lft = (SELECT COUNT(*) FROM vw_lftrgt AS v WHERE v.Lft <= Categories.Lft),
                    Rgt = (SELECT COUNT(*) FROM vw_lftrgt AS v WHERE v.Lft <= Categories.Rgt);

                DELETE FROM CategoriesTemporary;
                UPDATE Categories
                SET ParentId = parent_id
                WHERE Id = node_id;
            END IF;

        WHEN 'order' THEN

            SELECT Lft, Rgt, (Rgt - Lft + 1), ParentId INTO OldLft, OldRgt, Width, Superior
            FROM Categories WHERE Id = node_id;

            -- is parent
            SELECT ParentId INTO SuperiorParent FROM Categories WHERE Id = parent_id;

            IF (Superior = SuperiorParent) THEN
                SELECT (Rgt + 1) INTO NewLft FROM Categories WHERE Id = parent_id;
            ELSE
                SELECT (Lft + 1) INTO NewLft FROM Categories WHERE Id = parent_id;
            END IF;

            IF (NewLft != OldLft) THEN
                CREATE TEMPORARY TABLE IF NOT EXISTS CategoriesTemporary LIKE Categories;

                INSERT INTO CategoriesTemporary (Id, Lft, Rgt, ParentId, AppId, Depth, Name, CreateAt, UpdateAt)
                SELECT t1.Id,
                       (t1.Lft - (SELECT MIN(Lft) FROM Categories WHERE Id = node_id)) AS Lft,
                       (t1.Rgt - (SELECT MIN(Lft) FROM Categories WHERE Id = node_id)) AS Rgt,
                       t1.ParentId,
                       t1.AppId,
                       t1.Depth,
                       t1.Name,
                       t1.CreateAt,
                       t1.UpdateAt
                FROM Categories AS t1, Categories AS t2
                WHERE t1.Lft BETWEEN t2.Lft AND t2.Rgt AND t2.Id = node_id;


                DELETE FROM Categories WHERE Id IN (SELECT Id FROM CategoriesTemporary);

                IF(NewLft < OldLft) THEN -- move up
                    UPDATE Categories SET Lft = Lft + Width WHERE Lft >= NewLft AND Lft < OldLft;
                    UPDATE Categories SET Rgt = Rgt + Width WHERE Rgt > NewLft AND Rgt < OldRgt;
                    UPDATE CategoriesTemporary SET Lft = NewLft + Lft, Rgt = NewLft + Rgt;
                END IF;

                IF(NewLft > OldLft) THEN -- move down
                    UPDATE Categories SET Lft = Lft - Width WHERE Lft > OldLft AND Lft < NewLft;
                    UPDATE Categories SET Rgt = Rgt - Width WHERE Rgt > OldRgt AND Rgt < NewLft;
                    UPDATE CategoriesTemporary SET Lft = (NewLft - Width) + Lft, Rgt = (NewLft - Width) + Rgt;
                END IF;

                INSERT INTO Categories (Id, Lft, Rgt, ParentId, AppId, Name, Depth, CreateAt, UpdateAt )
                SELECT Id, Lft, Rgt, ParentId, AppId, Name, Depth, CreateAt, UpdateAt
                FROM CategoriesTemporary;

                DELETE FROM CategoriesTemporary;
            END IF;
        END CASE;

end
