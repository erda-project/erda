CREATE TABLE `ai_proxy_mcp_server`
(
    `id`                 CHAR(36)     NOT NULL COMMENT 'primary key',
    `name`               VARCHAR(64)  NOT NULL COMMENT 'MCP Server 名称',
    `version`            VARCHAR(64)  NOT NULL COMMENT 'MCP Server 版本',
    `description`        TEXT COMMENT '描述信息',
    `endpoint`           VARCHAR(191) NOT NULL COMMENT 'MCP Server URL',
    `config`             TEXT         NOT NULL COMMENT '配置信息，JSON 结构',
    `is_published`       TINYINT(1)   NOT NULL DEFAULT 0 COMMENT '是否发布',
    `is_default_version` TINYINT(1)   DEFAULT 0 COMMENT '是否是默认版本',
    `created_at`         DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `updated_at`         DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    `deleted_at`         DATETIME              DEFAULT NULL COMMENT '删除时间',
    UNIQUE (`name`, `version`),
    PRIMARY KEY (`id`),
    INDEX                `idx_is_published` (`is_published`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='MCP 服务列表';
