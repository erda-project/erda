CREATE TABLE `ai_proxy_token_usage`
(
    `id`              BIGINT(20) UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '自增主键',
    `created_at`      DATETIME(3)         NOT NULL DEFAULT CURRENT_TIMESTAMP(3) COMMENT '创建时间',
    `updated_at`      DATETIME(3)         NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3) COMMENT '更新时间',

    `call_id`         VARCHAR(64)                  DEFAULT NULL COMMENT '调用 ID',
    `x_request_id`    VARCHAR(64)                  DEFAULT NULL COMMENT '请求 ID',
    `client_id`       CHAR(36)                     DEFAULT NULL COMMENT '客户端 ID',
    `client_token_id` CHAR(36)                     DEFAULT NULL COMMENT '客户端 Token ID',
    `provider_id`     CHAR(36)                     DEFAULT NULL COMMENT '模型供应商 ID',
    `model_id`        CHAR(36)                     DEFAULT NULL COMMENT '模型 ID',
    `input_tokens`    BIGINT(20) UNSIGNED          DEFAULT NULL COMMENT '输入 token 数',
    `output_tokens`   BIGINT(20) UNSIGNED          DEFAULT NULL COMMENT '输出 token 数',
    `total_tokens`    BIGINT(20) UNSIGNED          DEFAULT NULL COMMENT '总 token 数',
    `is_estimated`    TINYINT(1)          NOT NULL DEFAULT 0 COMMENT '是否为估算值',
    `metadata`        MEDIUMTEXT                   DEFAULT NULL COMMENT '元数据',
    `usage_details`   TEXT                         DEFAULT NULL COMMENT '用量详情',

    PRIMARY KEY (`id`),
    INDEX `idx_call_id` (`call_id`),
    INDEX `idx_client_id` (`client_id`),
    INDEX `idx_client_token_id` (`client_token_id`),
    INDEX `idx_provider_id` (`provider_id`),
    INDEX `idx_model_id` (`model_id`),
    INDEX `idx_created_at` (`created_at`)
) ENGINE = InnoDB
  DEFAULT CHARSET = utf8mb4 COMMENT 'ai-proxy token 用量记录';
