ALTER TABLE erda_rule CHANGE COLUMN `updator` `actor` varchar(191) NOT NULL DEFAULT '' COMMENT 'actor';

ALTER TABLE erda_rule_exec_history ADD `actor` varchar(191) NOT NULL DEFAULT '' COMMENT 'actor';
