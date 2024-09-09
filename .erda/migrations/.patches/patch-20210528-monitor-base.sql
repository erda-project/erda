ALTER TABLE sp_dashboard_block CONVERT TO CHARACTER SET utf8mb4;

ALTER TABLE sp_dashboard_block_system CONVERT TO CHARACTER SET utf8mb4;

ALTER TABLE sp_report_task CONVERT TO CHARACTER SET utf8mb4;

ALTER TABLE `sp_dashboard_block` MODIFY COLUMN `scope` varchar (191) DEFAULT NULL;
ALTER TABLE `sp_dashboard_block` MODIFY COLUMN `scope_id` varchar (191) DEFAULT NULL;
ALTER TABLE `sp_dashboard_block_system` MODIFY COLUMN `scope` varchar (191) DEFAULT NULL;
ALTER TABLE `sp_dashboard_block_system` MODIFY COLUMN `scope_id` varchar (191) DEFAULT NULL;
ALTER TABLE `sp_report_task` MODIFY COLUMN `scope` varchar (191) DEFAULT NULL COMMENT 'scope';
ALTER TABLE `sp_report_task` MODIFY COLUMN `scope_id` varchar (191) DEFAULT NULL COMMENT 'scope id';