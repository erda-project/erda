CREATE TABLE `ai_proxy_client`
(
    `id`            CHAR(36)     NOT NULL COMMENT 'primary key',
    `created_at`    DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `updated_at`    DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    `deleted_at`    DATETIME     NOT NULL DEFAULT '1970-01-01 00:00:00' COMMENT '删除时间, 1970-01-01 00:00:00 表示未删除',

    `name`          VARCHAR(191) NOT NULL COMMENT '客户端名称',
    `desc`          VARCHAR(1024)         DEFAULT NULL COMMENT '客户端描述',
    `access_key_id` CHAR(32)     NOT NULL COMMENT '客户端 AK',
    `secret_key_id` CHAR(32)     NOT NULL COMMENT '客户端 SK',
    `metadata`      MEDIUMTEXT   NOT NULL COMMENT '客户端元数据',

    PRIMARY KEY (`id`),
    INDEX `idx_access_key_id` (`access_key_id`),
    UNIQUE INDEX `unique_name` (`name`, `deleted_at`)
) ENGINE = InnoDB
  DEFAULT CHARSET = utf8mb4 COMMENT 'ai-proxy 客户端';

CREATE TABLE `ai_proxy_model_provider`
(
    `id`         CHAR(36)     NOT NULL COMMENT 'primary key',
    `created_at` DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `updated_at` DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    `deleted_at` DATETIME     NOT NULL DEFAULT '1970-01-01 00:00:00' COMMENT '删除时间, 1970-01-01 00:00:00 表示未删除',

    `name`       VARCHAR(191) NOT NULL COMMENT '供应商名称，可以区分多账号或多地区，例如 azure-us-esat / azure-japan-east 等',
    `desc`       VARCHAR(1024)         DEFAULT NULL COMMENT '供应商描述',
    `type`       VARCHAR(191) NOT NULL COMMENT '供应商类型，例如 openai / azure 等',
    `api_key`    varchar(191) NOT NULL COMMENT '供应商级别的 api-key，例如 openai 的 sk，可以使用该供应商下的所有模型',
    `metadata`   MEDIUMTEXT   NOT NULL COMMENT '供应商元数据',

    PRIMARY KEY (`id`),
    INDEX `idx_api_key` (`api_key`),
    UNIQUE INDEX `unique_name` (`name`, `deleted_at`)
) ENGINE = InnoDB
  DEFAULT CHARSET = utf8mb4 COMMENT 'ai-proxy 模型供应商';

CREATE TABLE `ai_proxy_model`
(
    `id`          CHAR(36)     NOT NULL COMMENT 'primary key',
    `created_at`  DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `updated_at`  DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    `deleted_at`  DATETIME     NOT NULL DEFAULT '1970-01-01 00:00:00' COMMENT '删除时间, 1970-01-01 00:00:00 表示未删除',

    `name`        VARCHAR(191) NOT NULL COMMENT '模型名称, provider 下唯一',
    `desc`        VARCHAR(1024)         DEFAULT NULL COMMENT '供应商描述',
    `type`        VARCHAR(32)  NOT NULL COMMENT '模型类型，例如 text-generation / image / audio / embedding / text-moderation / text+visual(多模态) 等',
    `provider_id` CHAR(36)     NOT NULL COMMENT '模型供应商 id',
    `api_key`     varchar(191) NOT NULL COMMENT '模型级别的 api-key，优先级比 provider 级别更高',
    `metadata`    MEDIUMTEXT   NOT NULL COMMENT '模型元数据',

    PRIMARY KEY (`id`),
    INDEX `idx_api_key` (`api_key`),
    UNIQUE INDEX `unique_name_under_provider` (`name`, `provider_id`, `deleted_at`),
    INDEX `idx_type` (`type`)
) ENGINE = InnoDB
  DEFAULT CHARSET = utf8mb4 COMMENT 'ai-proxy 模型';

CREATE TABLE `ai_proxy_client_model_relation`
(
    `id`         CHAR(36) NOT NULL COMMENT 'primary key',
    `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    `deleted_at` DATETIME NOT NULL DEFAULT '1970-01-01 00:00:00' COMMENT '删除时间, 1970-01-01 00:00:00 表示未删除',

    `client_id`  CHAR(36) NOT NULL COMMENT '客户端 id',
    `model_id`   CHAR(36) NOT NULL COMMENT '模型 id',

    PRIMARY KEY (`id`),
    UNIQUE INDEX `unique_clientid_modelid` (`client_id`, `model_id`, `deleted_at`) # 一个模型在一个客户端下只能被关联一次。即使实际在物理上是同一个 model，但是在 model 表里也会有多条记录（model_type 不同）。
) ENGINE = InnoDB
  DEFAULT CHARSET = utf8mb4 COMMENT 'ai-proxy 客户端模型关联表';

CREATE TABLE `ai_proxy_prompt`
(
    `id`         CHAR(36)      NOT NULL COMMENT 'primary key',
    `created_at` DATETIME      NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `updated_at` DATETIME      NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    `deleted_at` DATETIME      NOT NULL DEFAULT '1970-01-01 00:00:00' COMMENT '删除时间, 1970-01-01 00:00:00 表示未删除',

    `name`       VARCHAR(191)  NOT NULL COMMENT 'prompt 名称',
    `desc`       VARCHAR(1024) NOT NULL COMMENT 'prompt 描述',
    `client_id`  CHAR(36)               DEFAULT NULL COMMENT '无 client_id 说明是平台级别的；有 client_id 则为客户端专属',
    `messages`   LONGTEXT      NOT NULL COMMENT '数组，一组 message，格式为: [{"role": "role", "message": "content"}]',
    `metadata`   MEDIUMTEXT    NOT NULL COMMENT 'prompt 元数据',

    PRIMARY KEY (`id`),
    INDEX `idx_name` (`name`),
    UNIQUE `unique_clientid_name` (`client_id`, `name`, `deleted_at`)
) ENGINE = InnoDB
  DEFAULT CHARSET = utf8mb4 COMMENT 'ai-proxy 提示词';

CREATE TABLE `ai_proxy_session`
(
    `id`             CHAR(36)     NOT NULL COMMENT 'primary key',
    `created_at`     DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `updated_at`     DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    `deleted_at`     DATETIME     NOT NULL DEFAULT '1970-01-01 00:00:00' COMMENT '删除时间, 1970-01-01 00:00:00 表示未删除',

    `client_id`      CHAR(36)     NOT NULL COMMENT '会话所属的客户端 id',
    `prompt_id`      CHAR(36)     NOT NULL COMMENT '会话使用的 prompt id',
    `model_id`       CHAR(36)     NOT NULL COMMENT '会话用到的模型 id',
    `scene`          VARCHAR(191) NOT NULL COMMENT '会话场景，用于区分功能，例如：chat, 测试用例, API 测试 等',
    `user_id`        VARCHAR(191) NOT NULL COMMENT '客户端传入的自定义 user_id，客户端用来区分用户',

    `name`           VARCHAR(191) NOT NULL DEFAULT '' COMMENT '会话名称。可以为空，由 AI 总结生成',
    `topic`          TEXT         NOT NULL COMMENT '会话主题',
    `num_of_ctx_msg` INT          NOT NULL DEFAULT 0 COMMENT '上下文消息个数，0 表示不使用上下文，1 表示使用上一条消息作为上下文，2 表示使用上两条消息作为上下文，以此类推；一问一答为 2 条消息',
    `is_archived`    BOOLEAN      NOT NULL DEFAULT false COMMENT '是否归档',
    `reset_at`       DATETIME     NOT NULL DEFAULT '1970-01-01 00:00:00' COMMENT '删除时间, 1970-01-01 00:00:00 表示未删除',
    `temperature`    DECIMAL      NOT NULL DEFAULT 0.7 COMMENT 'Higher values will make the output more random, while lower values will make it more focused and deterministic',
    `metadata`       MEDIUMTEXT   NOT NULL COMMENT '会话元数据',

    PRIMARY KEY (`id`),
    INDEX `idx_user_id` (`user_id`),
    INDEX `idx_name` (`name`)
) ENGINE = InnoDB
  DEFAULT CHARSET = utf8mb4
    COMMENT 'AI 会话管理表';

CREATE TABLE `ai_proxy_filter_audit`
(
    `id`                   CHAR(36)     NOT NULL COMMENT 'primary key',
    `created_at`           DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `updated_at`           DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    `deleted_at`           DATETIME     NOT NULL DEFAULT '1970-01-01 00:00:00' COMMENT '删除时间, 1970-01-01 00:00:00 表示未删除',
    `request_at`           DATETIME     NULL     DEFAULT CURRENT_TIMESTAMP COMMENT '请求到达时间',
    `response_at`          DATETIME     NULL     DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '响应到达时间',

    `status`               SMALLINT     NULL COMMENT 'http response status',
    `username`             VARCHAR(128) NULL COMMENT '用户名',
    `source`               VARCHAR(128) NULL COMMENT '接入应用: dingtalk, vscode-plugin, jetbrains-plugin ...',
    `prompt`               MEDIUMTEXT   NULL COMMENT '提示语',
    `completion`           LONGTEXT     NULL COMMENT 'AI 回复多个 choices 中的一个',
    `request_body`         LONGTEXT     NULL COMMENT '请求的 Body',
    `response_body`        LONGTEXT     NULL COMMENT '响应的 Body',
    `actual_request_body`  LONGTEXT     NULL COMMENT '实际调用大模型请求的 Body',
    `actual_response_body` LONGTEXT     NULL COMMENT '实际调用大模型响应的 Body',
    `user_agent`           TEXT         NULL COMMENT 'http 客户端 User-Agent',
    `x_request_id`         VARCHAR(64)  NULL COMMENT 'http 请求中的 X-Request-Id',

    `auth_key`             CHAR(64)     NULL COMMENT 'auth key: api_key or token',
    `client_id`            char(36)     NULL COMMENT '客户端 id',
    `model_id`             char(36)     NULL COMMENT '模型 id',
    `session_id`           VARCHAR(64)  NULL COMMENT '对话标识',

    `email`                VARCHAR(64)  NULL COMMENT '用户邮箱',

    `operation_id`         VARCHAR(128) NULL COMMENT '调用的接口名称, HTTP Method + Path',

    `res_func_call_name`   VARCHAR(128) NULL COMMENT 'function_call name in response message',

    `metadata`             LONGTEXT     NULL COMMENT '客户端要审计的其他信息',
    PRIMARY KEY (`id`),
    INDEX `idx_username` (`username`),
    INDEX `idx_email` (`email`),
    INDEX `idx_auth_key` (`auth_key`),
    INDEX `idx_session_id` (`session_id`)
) ENGINE = InnoDB
  DEFAULT CHARSET = utf8mb4
    COMMENT 'AI 审计表';

CREATE TABLE `ai_proxy_client_token`
(
    `id`         CHAR(36)     NOT NULL COMMENT 'primary key',
    `created_at` DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `updated_at` DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    `deleted_at` DATETIME     NOT NULL DEFAULT '1970-01-01 00:00:00' COMMENT '删除时间, 1970-01-01 00:00:00 表示未删除',

    `client_id`  CHAR(36)     NOT NULL COMMENT '会话所属的客户端 id',
    `user_id`    VARCHAR(191) NOT NULL COMMENT '客户端传入的自定义 user_id，客户端用来区分用户',
    `token`      CHAR(34)     NOT NULL COMMENT 't_ 前缀，len: uuid(32)+2',
    `expired_at` DATETIME     NOT NULL DEFAULT '1970-01-01 00:00:00',
    `metadata`   MEDIUMTEXT   NOT NULL COMMENT 'Token 元数据，主要包含 user 额外信息，用于审计',

    PRIMARY KEY (`id`),
    INDEX `idx_token` (`token`),
    UNIQUE INDEX `unique_clientid_userid` (`client_id`, `user_id`, `deleted_at`)
) ENGINE = InnoDB
  DEFAULT CHARSET = utf8mb4
    COMMENT 'AI 客户端 Token 表';
