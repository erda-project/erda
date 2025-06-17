CREATE TABLE `ai_proxy_mcp_server`
(
    `id`                 CHAR(36)     NOT NULL COMMENT 'Primary key',
    `name`               VARCHAR(64)  NOT NULL COMMENT 'MCP Server 名称',
    `version`            VARCHAR(64)  NOT NULL COMMENT 'MCP Server 版本',
    `description`        TEXT COMMENT '描述信息',
    `instruction`        TEXT COMMENT 'Agent Instruction',
    `endpoint`           VARCHAR(191) NOT NULL COMMENT 'MCP Server URL',
    `transport_type`     VARCHAR(64)  NOT NULL DEFAULT 'sse' COMMENT '传输协议类型',
    `config`             TEXT         NOT NULL COMMENT 'MCP Tool 配置信息, JSON 结构',
    `server_config`      TEXT NULL COMMENT 'Server 配置信息',
    `is_published`       TINYINT(1)   NOT NULL DEFAULT 0 COMMENT '是否发布',
    `is_default_version` TINYINT(1)   DEFAULT 0 COMMENT '是否默认版本',
    `created_at`         DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `updated_at`         DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    UNIQUE (`name`, `version`),
    PRIMARY KEY (`id`),
    INDEX                `idx_is_published` (`is_published`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='MCP 服务列表';