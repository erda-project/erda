create table erda_marketplace_gallery_artifacts
(
    `id`           VARCHAR(36)  NOT NULL PRIMARY KEY COMMENT 'primary key',
    `created_at`   DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `updated_at`   DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    `deleted_at`   DATETIME     NOT NULL DEFAULT '1970-01-01 00:00:00' COMMENT '删除时间',

    `org_id`       BIGINT       NOT NULL DEFAULT 0 COMMENT '组织 id',
    `org_name`     VARCHAR(50)  NOT NULL DEFAULT '' COMMENT '组织名称',
    `creator_id`   VARCHAR(255) NOT NULL DEFAULT '' COMMENT '创建人 user id',
    `updater_id`   VARCHAR(255) NOT NULL DEFAULT '' COMMENT '更新人 user id',

    `release_id`   VARCHAR(64)  NOT NULL DEFAULT '' COMMENT '制品 release id',
    `name`         VARCHAR(191) NOT NULL DEFAULT '' COMMENT 'gallery 名称',
    `display_name` VARCHAR(191) NOT NULL DEFAULT '' COMMENT 'gallery 显示名称',
    `version`      VARCHAR(64)  NOT NULL DEFAULT '' COMMENT 'gallery 版本',
    `type`         VARCHAR(64)  NOT NULL DEFAULT '' COMMENT 'gallery 类型, 目前都是 artifacts/project',
    `spec`         MEDIUMTEXT   NOT NULL COMMENT '预留字段',
    `changelog`    TEXT         NOT NULL COMMENT '项目制品 changelog',
    `is_default`   TINYINT(1)   NOT NULL DEFAULT 0 COMMENT '是否为默认制品版本',
    INDEX idx_release_id (`release_id`),
    INDEX idx_org_name_version (`org_id`, `name`, `version`),
    UNIQUE uk_org_name (`org_id`, `name`, `type`),
    UNIQUE uk_org_name_version (`org_id`, `name`, `type`, `version`)
) DEFAULT CHARACTER SET UTF8MB4 COMMENT '项目制品发布记录';