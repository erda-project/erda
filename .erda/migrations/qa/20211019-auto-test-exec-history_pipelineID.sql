ALTER TABLE `dice_autotest_exec_history` ADD `pipeline_id` bigint(20) NOT NULL DEFAULT 0 COMMENT '父流水线ID';
ALTER TABLE `dice_autotest_exec_history` ADD INDEX `idx_pipeline_id` (`pipeline_id`);