CREATE TABLE `ai_proxy_providers`
(
    `id`          CHAR(36)      NOT NULL COMMENT 'primary key',
    `created_at`  DATETIME      NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `updated_at`  DATETIME      NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    `deleted_at`  DATETIME      NOT NULL DEFAULT '1970-01-01 00:00:00' COMMENT '删除时间, 1970-01-01 00:00:00 表示未删除',

    `name`        VARCHAR(128)  NOT NULL DEFAULT '' COMMENT 'AI 供应商名称: azure, openai, tongyi, ...',
    `instance_id` VARCHAR(512)  NOT NULL DEFAULT '' COMMENT 'AI 服务实例 id',
    `host`        VARCHAR(128)  NOT NULL DEFAULT '' COMMENT 'AI 服务域名',
    `scheme`      VARCHAR(16)   NOT NULL DEFAULT 'https' COMMENT 'AI 服务',
    `description` VARCHAR(1024) NOT NULL DEFAULT '' COMMENT 'AI 供应商描述',
    `doc_site`    VARCHAR(512)  NOT NULL DEFAULT '' COMMENT 'AI 供应商网站地址或文档地址',
    `aes_key`     CHAR(16)      NOT NULL DEFAULT '' COMMENT 'AES 对称算法种子 Key',
    `api_key`     VARCHAR(512)  NOT NULL DEFAULT '' COMMENT 'AES 对称算法加密后的 api_key',
    `metadata`    LONGTEXT      NOT NULL COMMENT 'AI 服务实例其他元信息',

    PRIMARY KEY (`id`),
    INDEX `idx_instance_id` (`instance_id`),
    INDEX `idx_name_instance_id` (`name`, `instance_id`)
) ENGINE = InnoDB
  DEFAULT CHARSET = utf8mb4
    COMMENT 'ai 供应商服务实例';
