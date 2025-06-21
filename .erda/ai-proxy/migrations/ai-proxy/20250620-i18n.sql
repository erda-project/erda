CREATE TABLE `ai_proxy_i18n`
(
    `id`         CHAR(36)     NOT NULL COMMENT 'primary key',
    `created_at` DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `updated_at` DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    `deleted_at` DATETIME     NOT NULL DEFAULT '1970-01-01 00:00:00' COMMENT '删除时间, 1970-01-01 00:00:00 表示未删除',

    `category`   VARCHAR(50)  NOT NULL COMMENT '配置分类，如：publisher, ability, model_type',
    `item_key`   VARCHAR(100) NOT NULL COMMENT '项目标识，如：openai, anthropic, text_generation',
    `field_name` VARCHAR(50)  NOT NULL COMMENT '字段名称，如：name, description, logo',
    `locale`     VARCHAR(10)  NOT NULL COMMENT '语言标识，如：zh, en, *（通用配置）',
    `value`      TEXT         NOT NULL COMMENT '配置值',

    PRIMARY KEY (`id`),
    UNIQUE INDEX `unique_config_key` (`category`, `item_key`, `field_name`, `locale`, `deleted_at`),
    INDEX `idx_category` (`category`),
    INDEX `idx_item_key` (`item_key`),
    INDEX `idx_locale` (`locale`)
) ENGINE = InnoDB
  DEFAULT CHARSET = utf8mb4 COMMENT 'AI-Proxy 国际化配置表';

-- Initialize
INSERT INTO `ai_proxy_i18n` (`id`, `category`, `item_key`, `field_name`, `locale`, `value`) VALUES (UUID(), 'publisher', 'openai', 'name', 'en', 'OpenAI');
INSERT INTO `ai_proxy_i18n` (`id`, `category`, `item_key`, `field_name`, `locale`, `value`) VALUES (UUID(), 'publisher', 'openai', 'name', 'zh', 'OpenAI');
INSERT INTO `ai_proxy_i18n` (`id`, `category`, `item_key`, `field_name`, `locale`, `value`) VALUES (UUID(), 'publisher', 'openai', 'logo', '*', 'placeholder');

INSERT INTO `ai_proxy_i18n` (`id`, `category`, `item_key`, `field_name`, `locale`, `value`) VALUES (UUID(), 'publisher', 'anthropic', 'name', 'en', 'Anthropic');
INSERT INTO `ai_proxy_i18n` (`id`, `category`, `item_key`, `field_name`, `locale`, `value`) VALUES (UUID(), 'publisher', 'anthropic', 'name', 'zh', 'Anthropic');
INSERT INTO `ai_proxy_i18n` (`id`, `category`, `item_key`, `field_name`, `locale`, `value`) VALUES (UUID(), 'publisher', 'anthropic', 'logo', '*', 'placeholder');

INSERT INTO `ai_proxy_i18n` (`id`, `category`, `item_key`, `field_name`, `locale`, `value`) VALUES (UUID(), 'publisher', 'deepseek', 'name', 'en', 'DeepSeek');
INSERT INTO `ai_proxy_i18n` (`id`, `category`, `item_key`, `field_name`, `locale`, `value`) VALUES (UUID(), 'publisher', 'deepseek', 'name', 'zh', '深度求索');
INSERT INTO `ai_proxy_i18n` (`id`, `category`, `item_key`, `field_name`, `locale`, `value`) VALUES (UUID(), 'publisher', 'deepseek', 'logo', '*', 'placeholder');

INSERT INTO `ai_proxy_i18n` (`id`, `category`, `item_key`, `field_name`, `locale`, `value`) VALUES (UUID(), 'publisher', 'bytedance', 'name', 'en', 'ByteDance');
INSERT INTO `ai_proxy_i18n` (`id`, `category`, `item_key`, `field_name`, `locale`, `value`) VALUES (UUID(), 'publisher', 'bytedance', 'name', 'zh', '字节跳动');
INSERT INTO `ai_proxy_i18n` (`id`, `category`, `item_key`, `field_name`, `locale`, `value`) VALUES (UUID(), 'publisher', 'bytedance', 'logo', '*', 'placeholder');

INSERT INTO `ai_proxy_i18n` (`id`, `category`, `item_key`, `field_name`, `locale`, `value`) VALUES (UUID(), 'publisher', 'qwen', 'name', 'en', 'Qwen');
INSERT INTO `ai_proxy_i18n` (`id`, `category`, `item_key`, `field_name`, `locale`, `value`) VALUES (UUID(), 'publisher', 'qwen', 'name', 'zh', '通义千问');
INSERT INTO `ai_proxy_i18n` (`id`, `category`, `item_key`, `field_name`, `locale`, `value`) VALUES (UUID(), 'publisher', 'qwen', 'logo', '*', 'placeholder');

INSERT INTO `ai_proxy_i18n` (`id`, `category`, `item_key`, `field_name`, `locale`, `value`) VALUES (UUID(), 'publisher', 'meta-llama', 'name', 'en', 'Meta LLaMA');
INSERT INTO `ai_proxy_i18n` (`id`, `category`, `item_key`, `field_name`, `locale`, `value`) VALUES (UUID(), 'publisher', 'meta-llama', 'name', 'zh', 'Meta LLaMA');
INSERT INTO `ai_proxy_i18n` (`id`, `category`, `item_key`, `field_name`, `locale`, `value`) VALUES (UUID(), 'publisher', 'meta-llama', 'logo', '*', 'placeholder');

INSERT INTO `ai_proxy_i18n` (`id`, `category`, `item_key`, `field_name`, `locale`, `value`) VALUES (UUID(), 'ability', 'tool', 'name', 'en', 'Tool');
INSERT INTO `ai_proxy_i18n` (`id`, `category`, `item_key`, `field_name`, `locale`, `value`) VALUES (UUID(), 'ability', 'tool', 'name', 'zh', '工具调用');
INSERT INTO `ai_proxy_i18n` (`id`, `category`, `item_key`, `field_name`, `locale`, `value`) VALUES (UUID(), 'ability', 'reasoning', 'name', 'en', 'Reasoning');
INSERT INTO `ai_proxy_i18n` (`id`, `category`, `item_key`, `field_name`, `locale`, `value`) VALUES (UUID(), 'ability', 'reasoning', 'name', 'zh', '推理');
INSERT INTO `ai_proxy_i18n` (`id`, `category`, `item_key`, `field_name`, `locale`, `value`) VALUES (UUID(), 'ability', 'image_recognition', 'name', 'en', 'Image Recognition');
INSERT INTO `ai_proxy_i18n` (`id`, `category`, `item_key`, `field_name`, `locale`, `value`) VALUES (UUID(), 'ability', 'image_recognition', 'name', 'zh', '视觉理解 (图片)');
INSERT INTO `ai_proxy_i18n` (`id`, `category`, `item_key`, `field_name`, `locale`, `value`) VALUES (UUID(), 'ability', 'video_recognition', 'name', 'en', 'Video Recognition');
INSERT INTO `ai_proxy_i18n` (`id`, `category`, `item_key`, `field_name`, `locale`, `value`) VALUES (UUID(), 'ability', 'video_recognition', 'name', 'zh', '视觉理解 (视频)');
INSERT INTO `ai_proxy_i18n` (`id`, `category`, `item_key`, `field_name`, `locale`, `value`) VALUES (UUID(), 'ability', 'responses_api', 'name', 'en', 'Responses API');
INSERT INTO `ai_proxy_i18n` (`id`, `category`, `item_key`, `field_name`, `locale`, `value`) VALUES (UUID(), 'ability', 'responses_api', 'name', 'zh', 'Responses API');
