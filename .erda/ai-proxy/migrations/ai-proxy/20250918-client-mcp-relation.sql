CREATE TABLE `ai_proxy_client_mcp_relation`
(
    `id`         CHAR(36) NOT NULL COMMENT 'primary key',
    `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    `deleted_at` DATETIME NOT NULL DEFAULT '1970-01-01 00:00:00' COMMENT '删除时间, 1970-01-01 00:00:00 表示未删除',

    `client_id`  CHAR(36) NOT NULL COMMENT '客户端 id',
    `scope_type` CHAR(36) NOT NULL COMMENT 'MCP 作用域，一般为org',
    `scope_id`   CHAR(36) NOT NULL COMMENT 'MCP 作用域ID，一般为orgId',

    PRIMARY KEY (`id`),
    UNIQUE INDEX `unique_clientid_modelid` (`client_id`, `scope_type`, `scope_id`, `deleted_at`)
) ENGINE = InnoDB
  DEFAULT CHARSET = utf8mb4 COMMENT 'ai-proxy 客户端 MCP 关联表';
