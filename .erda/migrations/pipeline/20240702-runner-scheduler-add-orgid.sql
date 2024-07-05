ALTER TABLE `dice_runner_tasks`
    ADD COLUMN `org_id` bigint(20) NOT NULL DEFAULT '0' COMMENT 'org id' AFTER `job_id`;
# add index
CREATE INDEX `idx_org_id` ON `dice_runner_tasks` (`org_id`);
