CREATE TABLE `ai_proxy_providers`
(
    `id`          CHAR(36)     NOT NULL COMMENT 'primary key',
    `created_at`  DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `updated_at`  DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    `deleted_at`  DATETIME     NOT NULL DEFAULT '1970-01-01 00:00:00' COMMENT '删除时间, 1970-01-01 00:00:00 表示未删除',

    `name`        VARCHAR(128) NOT NULL COMMENT 'provider name',
    `instance_id` CHAR(512)    NOT NULL COMMENT '',
    `host`        VARCHAR(128) NOT NULL COMMENT '',
    `scheme`      VARCHAR(16)  NOT NULL COMMENT '',
    `description` VARCHAR(512) NOT NULL COMMENT '',
    `aes_key`     CHAR(16)     NOT NULL COMMENT '',
    `api_key`     VARCHAR(512) NOT NULL COMMENT '',
    `metadata`    LONGTEXT     NOT NULL COMMENT '',

    PRIMARY KEY (`id`),
    INDEX `idx_instance_id` (`instance_id`)
) ENGINE = InnoDB
  DEFAULT CHARSET = utf8mb4
    COMMENT 'ai 供应商服务实例';
