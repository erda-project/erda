CREATE TABLE erda_license
(
    `id`          BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '许可证ID',
    `entity_type` ENUM('PLATFORM', 'ORG') NOT NULL COMMENT '实体类型',
    `entity_id`   BIGINT NOT NULL DEFAULT 0 COMMENT '实体ID',
    `enc_license` TEXT COMMENT '加密许可证信息',
    `created_at`  TIMESTAMP       DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `updated_at`  TIMESTAMP       DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    PRIMARY KEY (`id`),
    UNIQUE KEY entity_entity_id (entity_type, entity_id)
) COMMENT '许可证信息表';
