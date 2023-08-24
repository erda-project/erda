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

    PRIMARY KEY (`id`),
    INDEX `idx_access_key_id` (`access_key_id`)
) ENGINE = InnoDB
  DEFAULT CHARSET = utf8mb4 COMMENT 'ai-proxy 客户端';

CREATE TABLE `ai_proxy_model_provider`
(
    `id`         CHAR(36)     NOT NULL COMMENT 'primary key',
    `created_at` DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `updated_at` DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    `deleted_at` DATETIME     NOT NULL DEFAULT '1970-01-01 00:00:00' COMMENT '删除时间, 1970-01-01 00:00:00 表示未删除',

    `name`       VARCHAR(191) NOT NULL COMMENT '供应商名称',
    `desc`       VARCHAR(1024)         DEFAULT NULL COMMENT '供应商描述',
    `api_key`    varchar(128) NOT NULL COMMENT '供应商级别的 api-key，例如 openai 的 sk，可以使用该供应商下的所有模型',
    `homepage`   VARCHAR(1024)         DEFAULT NULL COMMENT '供应商主页',
    `doc_site`   VARCHAR(1024)         DEFAULT NULL COMMENT '供应商文档',
    `metadata`   MEDIUMTEXT   NOT NULL COMMENT '供应商元数据',

    PRIMARY KEY (`id`),
    INDEX `idx_api_key` (`api_key`)
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
    `type`        VARCHAR(32)  NOT NULL COMMENT '模型类型，例如 text-generation / image / video / embedding / text-moderation / text+visual(多模态) 等',
    `provider_id` CHAR(36)     NOT NULL COMMENT '模型供应商 id',
    `api_key`     varchar(128) NOT NULL COMMENT '模型级别的 api-key，优先级比 provider 级别更高',
    `metadata`    MEDIUMTEXT   NOT NULL COMMENT '模型元数据',

    PRIMARY KEY (`id`),
    INDEX `idx_api_key` (`api_key`),
    UNIQUE INDEX `idx_name_under_provider` (`name`, `provider_id`),
    INDEX `idx_type` (`type`)
) ENGINE = InnoDB
  DEFAULT CHARSET = utf8mb4 COMMENT 'ai-proxy 模型';

CREATE TABLE `ai_proxy_client_model_relation`
(
    `id`         CHAR(36)    NOT NULL COMMENT 'primary key',
    `created_at` DATETIME    NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `updated_at` DATETIME    NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    `deleted_at` DATETIME    NOT NULL DEFAULT '1970-01-01 00:00:00' COMMENT '删除时间, 1970-01-01 00:00:00 表示未删除',

    `client_id`  CHAR(36)    NOT NULL COMMENT '客户端 id',
    `model_id`   CHAR(36)    NOT NULL COMMENT '模型 id',
    `model_type` VARCHAR(32) NOT NULL COMMENT '模型类型',
    `is_default` TINYINT(1)  NOT NULL DEFAULT 0 COMMENT '每个 client 可以为每个 model_type 设置一个默认的 model',

    PRIMARY KEY (`id`),
    UNIQUE INDEX `idx_clientid_modelid` (`client_id`, `model_id`), # 一个模型在一个客户端下只能被关联一次
    INDEX `idx_clientid_modelid_modeltype_isdefault` (`client_id`, `model_id`, `model_type`, `is_default`)
) ENGINE = InnoDB
  DEFAULT CHARSET = utf8mb4 COMMENT 'ai-proxy 客户端模型关联表';

CREATE TABLE `ai_proxy_client_prompt`
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
    INDEX `idx_clientid` (`client_id`)
) ENGINE = InnoDB
  DEFAULT CHARSET = utf8mb4 COMMENT 'ai-proxy 提示词';

CREATE TABLE `ai_proxy_session`
(
    `id`             CHAR(36)     NOT NULL COMMENT 'primary key',
    `created_at`     DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `updated_at`     DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    `deleted_at`     DATETIME     NOT NULL DEFAULT '1970-01-01 00:00:00' COMMENT '删除时间, 1970-01-01 00:00:00 表示未删除',

    `client_id`      CHAR(36)     NOT NULL COMMENT '会话所属的客户端 id',
    `model_id`       CHAR(36)     NOT NULL COMMENT '会话用到的模型 id',
    `scene`          VARCHAR(128) NOT NULL COMMENT '会话场景，用于区分功能，例如：chat, 测试用例, API 测试 等',
    `user_id`        VARCHAR(128) NOT NULL COMMENT '客户端传入的自定义 user_id，客户端用来区分用户',

    `name`           VARCHAR(128) NOT NULL DEFAULT '' COMMENT '会话名称。可以为空，由 AI 总结生成',
    `topic`          TEXT         NOT NULL COMMENT '会话主题',
    `num_of_ctx_msg` INT          NOT NULL DEFAULT 0 COMMENT '上下文消息个数，0 表示不使用上下文，1 表示使用上一条消息作为上下文，2 表示使用上两条消息作为上下文，以此类推；一问一答为 2 条消息',
    `source`         VARCHAR(128) NOT NULL COMMENT '接入应用: dingtalk, vscode-plugin, jetbrains-plugin ...',
    `is_archived`    BOOLEAN      NOT NULL DEFAULT false COMMENT '是否归档',
    `reset_at`       DATETIME     NOT NULL DEFAULT '1970-01-01 00:00:00' COMMENT '删除时间, 1970-01-01 00:00:00 表示未删除',
    `model`          VARCHAR(128) NOT NULL COMMENT '调用的模型名称: gpt-3.5-turbo, gpt-4-8k, ...',
    `temperature`    DECIMAL      NOT NULL DEFAULT 0.7 COMMENT 'Higher values will make the output more random, while lower values will make it more focused and deterministic',

    PRIMARY KEY (`id`),
    INDEX `idx_user_id` (`user_id`),
    INDEX `idx_name` (`name`)
) ENGINE = InnoDB
  DEFAULT CHARSET = utf8mb4
    COMMENT 'AI 会话管理表';
