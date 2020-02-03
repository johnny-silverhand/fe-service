DROP VIEW  IF EXISTS`vw_lftrgt`;
CREATE VIEW `vw_lftrgt` AS select `Categories`.`Lft` AS `Lft` from `Categories` union select `Categories`.`Rgt` AS `Rgt` from `Categories`;
