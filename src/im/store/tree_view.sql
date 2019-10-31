DROP VIEW  IF EXISTS`vw_lftrgt`;
CREATE VIEW `vw_lftrgt` AS select `categories`.`Lft` AS `Lft` from `categories` union select `categories`.`Rgt` AS `Rgt` from `categories`;