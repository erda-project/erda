CREATE TABLE IF NOT EXISTS `ai_proxy_mcp_server_template`
(
    `id`         BIGINT(20) UNSIGNED NOT NULL AUTO_INCREMENT COMMENT 'Primary Key',
    `mcp_name`   varchar(255)        NOT NULL COMMENT 'MCP 名称',
    `version`    varchar(128)        NOT NULL COMMENT 'MCP 版本',
    `template`   TEXT COMMENT '配置模板',

    `created_at` DATETIME            NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `updated_at` DATETIME            NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    `deleted_at` DATETIME            NOT NULL DEFAULT '1970-01-01 00:00:00' COMMENT '删除时间, 1970-01-01 00:00:00 表示未删除',

    PRIMARY KEY (`id`),
    UNIQUE INDEX `unique_name_version` (`mcp_name`, `version`)
    ) ENGINE = InnoDB
    DEFAULT CHARSET = utf8mb4 COMMENT 'ai-proxy MCP 配置模板表';

INSERT INTO `ai_proxy_mcp_server_template` (`mcp_name`, `version`, `template`)
VALUES ('enterprise-tools', '1.0.0', '[]'),
       ('html-generator', '2.0.0', '[
    {
      "default": "",
      "description": "生成HTML的额外说明",
      "name": "instructions",
      "required": false,
    },
    {
      "description": "模型名称",
      "name": "model_name",
      "required": true,
    },
    {
      "description": "模型发布商",
      "name": "model_publisher",
      "required": true,
    }
  ]'),
       ('search-oil', '1.0.0', '[]'),
       ('pptx-agent', '1.0.0', '[]'),
       ('mcp-timer', '1.0.0', '[]'),
       ('mcp-server-baidu-maps', '1.11.0', '[]'),
       ('playwright', '1.0.0', '[]'),
       ('office-word-mcp-server', '1.0.0', '[]'),
       ('mcp-milvus', '1.0.0', '[]'),
       ('mcp-fetch', '1.0.0', '[]'),
       ('search-engine', '3.0.0','[]'),
       ('mcp-python-script', '1.0.0', '[]'),
       ('mcp-calculator', '1.0.0', '[]'),
       ('search-engine', '1.0.0', '[]'),
       ('mcp-ocr', '1.0.0', '[]'),
       ('mcp-email', '1.0.0','[]'),
       ('pymupdf4llm-mcp-server', '1.0.0', '[]'),
       ('12306', '1.0.0', '[]'),
       ('mcp-exchange-rate', '1.0.0', '[]');


CREATE TABLE IF NOT EXISTS `ai_proxy_mcp_server_config_instance`
(
    `id`            CHAR(36)     NOT NULL COMMENT 'Primary Key',
    `created_at`    DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `updated_at`    DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    `deleted_at`    DATETIME     NOT NULL DEFAULT '1970-01-01 00:00:00' COMMENT '删除时间, 1970-01-01 00:00:00 表示未删除',

    `instance_name` VARCHAR(255) NOT NULL DEFAULT 'default' COMMENT '默认实例名称',
    `client_id`     CHAR(36)     NOT NULL COMMENT '客户端 id',
    `config`        TEXT COMMENT 'MCP 配置',
    `mcp_name`      varchar(255) NOT NULL COMMENT 'MCP 名称',
    `version`       varchar(128) NOT NULL COMMENT 'MCP 版本',

    PRIMARY KEY (`id`),
    UNIQUE INDEX `unique_mcp_name_version_client_id` (`instance_name`, `mcp_name`, `version`, `client_id`, `deleted_at`)
    ) ENGINE = InnoDB
    DEFAULT CHARSET = utf8mb4 COMMENT 'ai-proxy MCP 配置实例表';
