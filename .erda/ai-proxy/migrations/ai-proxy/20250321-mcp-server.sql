CREATE TABLE `ai_proxy_mcp_server`
(
    `id`           CHAR(36)     NOT NULL COMMENT 'primary key',
    `name`         VARCHAR(64)  NOT NULL COMMENT 'MCP Server 名称',
    `version`      VARCHAR(64)  NOT NULL COMMENT 'MCP Server 版本',
    `description`  TEXT COMMENT '描述信息',
    `endpoint`     VARCHAR(191) NOT NULL COMMENT 'MCP Server URL',
    `config`       TEXT         NOT NULL COMMENT '配置信息，JSON 结构',
    `is_published` BOOLEAN      NOT NULL DEFAULT FALSE COMMENT '发布状态',
    `created_at`   DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `updated_at`   DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    `deleted_at`   DATETIME              DEFAULT NULL COMMENT '删除时间',
    UNIQUE (`name`, `version`),
    PRIMARY KEY (`id`),
    INDEX          `idx_name_version` (`name`, `version`),
    INDEX          `idx_is_published` (`is_published`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='MCP 服务列表';
