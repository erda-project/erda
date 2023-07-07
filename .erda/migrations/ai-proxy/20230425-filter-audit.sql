CREATE TABLE `ai_proxy_filter_audit`
(
    `id`                    CHAR(36)     NOT NULL COMMENT 'primary key',
    `created_at`            DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `updated_at`            DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    `deleted_at`            DATETIME     NOT NULL DEFAULT '1970-01-01 00:00:00' COMMENT '删除时间, 1970-01-01 00:00:00 表示未删除',

    `app_key_sha256`        CHAR(64)     NOT NULL DEFAULT '' COMMENT '请求使用的 app_key sha256 哈希值',

    `username`              VARCHAR(128) NOT NULL COMMENT '用户名称, source=dingtalk时, 为钉钉用户名称',
    `phone_number`          VARCHAR(32)  NOT NULL COMMENT '用户手机号码, source=dingtalk时, 为钉钉账号注册手机号',
    `job_number`            VARCHAR(32)  NOT NULL COMMENT '用户工号, source=dingtalk时, 为用户在其组织内的工号',
    `email`                 VARCHAR(64)  NOT NULL COMMENT '用户邮箱',
    `dingtalk_staff_id`     VARCHAR(64)  NOT NULL COMMENT '用户钉钉号',

    `session_id`            VARCHAR(64)  NOT NULL COMMENT '对话标识',
    `chat_type`             VARCHAR(32)  NOT NULL COMMENT '对话类型',
    `chat_title`            VARCHAR(128) NOT NULL COMMENT 'source=dingtalk时, 私聊时为 private, 群聊时为群名称',
    `chat_id`               VARCHAR(64)  NOT NULL COMMENT '钉钉聊天 id',
    `source`                VARCHAR(128) NOT NULL COMMENT '接入应用: dingtalk, vscode-plugin, jetbrains-plugin ...',
    `provider`              VARCHAR(128) NOT NULL COMMENT 'AI 能力提供商: openai, azure...',
    `model`                 VARCHAR(128) NOT NULL COMMENT '调用的模型名称: gpt-3.5-turbo, gpt-4-8k, ...',
    `operation_id`          VARCHAR(128) NOT NULL COMMENT '调用的接口名称, HTTP Method + Path',
    `prompt`                MEDIUMTEXT   NOT NULL COMMENT '提示语',
    `completion`            LONGTEXT     NOT NULL COMMENT 'AI 回复多个 choices 中的一个',

    `request_at`            DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '请求到达时间',
    `response_at`           DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '响应到达时间',
    `request_content_type`  VARCHAR(32)  NOT NULL COMMENT '请求使用的 Content-Type',
    `request_body`          LONGTEXT     NOT NULL COMMENT '请求的 Body',
    `response_content_type` VARCHAR(32)  NOT NULL COMMENT '响应使用的 Content-Type',
    `response_body`         LONGTEXT     NOT NULL COMMENT '响应的 Body',
    `user_agent`            VARCHAR(128) NOT NULL COMMENT 'http 客户端 User-Agent',
    `server`                VARCHAR(32)  NOT NULL COMMENT 'response server',
    `status`                VARCHAR(32)  NOT NULL COMMENT 'http response status',
    `status_code`           INT          NOT NULL COMMENT 'http response status code',
    PRIMARY KEY (`id`),
    INDEX `idx_job_number` (`job_number`),
    INDEX `idx_dingtalk_staff_id` (`dingtalk_staff_id`)
) ENGINE = InnoDB
  DEFAULT CHARSET = utf8mb4
    COMMENT 'AI 插件之审计';

CREATE TABLE `ai_proxy_sessions`
(
    `id`             CHAR(36)     NOT NULL COMMENT 'primary key',
    `created_at`     DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `updated_at`     DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    `deleted_at`     DATETIME     NOT NULL DEFAULT '1970-01-01 00:00:00' COMMENT '删除时间, 1970-01-01 00:00:00 表示未删除',

    `user_id`        VARCHAR(128) NOT NULL DEFAULT '' COMMENT '用户id',
    `name`           VARCHAR(128) NOT NULL DEFAULT '' COMMENT '会话名称',
    `topic`          TEXT         NOT NULL COMMENT '会话主题',
    `context_length` INT          NOT NULL DEFAULT 0 COMMENT '上下文长度',
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
