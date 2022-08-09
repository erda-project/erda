ALTER TABLE `erda_dev_flow` DROP INDEX `idx_flow_rule_name`;
ALTER TABLE `erda_dev_flow` ADD INDEX `idx_flow_rule_name_app_id` (`flow_rule_name`, `app_id`);
