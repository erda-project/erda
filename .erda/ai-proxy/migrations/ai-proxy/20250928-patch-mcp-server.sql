ALTER TABLE `ai_proxy_mcp_server` DROP INDEX `name`;

ALTER TABLE `ai_proxy_mcp_server`
    ADD COLUMN `deleted_at` DATETIME NOT NULL DEFAULT '1970-01-01 00:00:00' COMMENT '删除时间, 1970-01-01 00:00:00 表示未删除',
    ADD COLUMN `scope_type` VARCHAR(64) NOT NULL DEFAULT 'org' COMMENT '作用域类型',
    ADD COLUMN `scope_id`   VARCHAR(64) NOT NULL DEFAULT '0' COMMENT '作用域 ID',
    ADD UNIQUE KEY uniq_name_scope_version (`name`, `scope_type`, `scope_id`, `version`, `deleted_at`),
    ADD INDEX idx_scope (`scope_type`, `scope_id`);
