ALTER TABLE `sp_alert_event` ADD INDEX `idx_scope_scope_id` (`scope`, `scope_id`);
ALTER TABLE `sp_alert_event` ADD INDEX `idx_scope` (`scope`);
ALTER TABLE `sp_alert_event` ADD INDEX `idx_scope_id` (`scope_id`);