ALTER TABLE `pipeline_definition` ADD INDEX `idx_category` (`category`);
ALTER TABLE `pipeline_definition` ADD INDEX `idx_executor` (`executor`);
ALTER TABLE `pipeline_definition` ADD INDEX `idx_creator` (`creator`);
ALTER TABLE `pipeline_definition` ADD INDEX `idx_status` (`status`);