CREATE TABLE `filter_audit`
(
    `id`                    char(36)      NOT NULL COMMENT 'primary key',
    `created_at`            DATETIME      NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `updated_at`            DATETIME      NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    `deleted_at`            DATETIME      NOT NULL DEFAULT '1970-01-01 00:00:00' COMMENT '删除时间',

    `session_id`            varchar(64)   NOT NULL COMMENT '对话标识',
    `chat_type`             varchar(32)   NOT NULL COMMENT '对话类型',
    `chat_title`            varchar(64)   NOT NULL COMMENT '',
    `chat_id`               varchar(64)   NOT NULL COMMENT '',
    `source`                varchar(128)  NOT NULL COMMENT '接入应用: dingtalk, vscode-plugin, jetbrains-plugin ...',
    `user_info`             varchar(1024) NOT NULL COMMENT '用户标识',
    `provider`              varchar(128)  NOT NULL COMMENT 'AI 能力提供商: openai, baidu, alibaba',
    `model`                 varchar(128)  NOT NULL COMMENT '调用的模型名称: gpt-3.5-turbo, gpt-4-8k, ...',
    `operation_id`          varchar(128)  NOT NULL COMMENT '调用的接口名称: CreateCompletion',
    `prompt`                text          NOT NULL COMMENT '提示语',
    `completion`            text          NOT NULL COMMENT 'AI 回复多个 choices 中的一个',

    `request_at`            DATETIME      NOT NULL DEFAULT '1970-01-01 00:00:00' COMMENT '请求到达时间',
    `response_at`           DATETIME      NOT NULL DEFAULT '1970-01-01 00:00:00' COMMENT '响应到达时间',
    `request_content_type`  varchar(32)   NOT NULL COMMENT '请求使用的 Content-Type',
    `request_body`          text          NOT NULL COMMENT '请求的 Body',
    `response_content_type` varchar(32)   NOT NULL COMMENT '响应使用的 Content-Type',
    `response_body`         text          NOT NULL COMMENT '响应的 Body',
    `user_agent`            varchar(128)  NOT NULL COMMENT 'http 客户端 User-Agent',
    `server`                varchar(32)   NOT NULL COMMENT 'response server',
    `status`                varchar(32)   NOT NULL COMMENT 'http response status',
    `status_code`           int           NOT NULL COMMENT 'http response status code',
    PRIMARY KEY (`id`)
) ENGINE = InnoDB
  DEFAULT CHARSET = utf8mb4
  COLLATE = utf8mb4_general_ci;
