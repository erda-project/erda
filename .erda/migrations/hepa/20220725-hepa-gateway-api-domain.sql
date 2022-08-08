ALTER TABLE `tb_gateway_upstream_api`
    ADD COLUMN `domains` VARCHAR(1024) NOT NULL DEFAULT '' COMMENT 'gateway api 的域名列表, 逗号隔开';

ALTER TABLE `tb_gateway_api`
    ADD COLUMN `domains` VARCHAR(1024) NOT NULL DEFAULT '' COMMENT 'gateway api 的域名列表, 逗号隔开';

CREATE TABLE `tb_gateway_hub_info`
(
    `id`                VARCHAR(36)  NOT NULL PRIMARY KEY COMMENT 'primary key',
    `created_at`        DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `updated_at`        DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    `deleted_at`        DATETIME     NOT NULL DEFAULT '1970-01-01 00:00:00' COMMENT '删除时间',
    `is_deleted`        CHAR(1)      NOT NULL DEFAULT 'N' COMMENT '软删除符号, 兼容 hepa 习惯',
    `create_time`       DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间, 兼容 hepa 习惯',
    `update_time`       DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间, 兼容 hepa 习惯',

    `org_id`            BIGINT(20)   NOT NULL DEFAULT 0 COMMENT 'org id',
    `org_name`          VARCHAR(50)  NOT NULL DEFAULT '' COMMENT 'org name',
    `creator_id`        VARCHAR(255) NOT NULL DEFAULT '' COMMENT 'creator user id',
    `updater_id`        VARCHAR(255) NOT NULL DEFAULT '' COMMENT 'updater user id',

    `dice_env`          VARCHAR(10)  NOT NULL DEFAULT '' COMMENT '环境',
    `dice_cluster_name` VARCHAR(191) NOT NULL DEFAULT '' COMMENT '集群名称',
    `bind_domain`       VARCHAR(225) NOT NULL DEFAULT '' COMMENT '绑定的域名',
    `description`       VARCHAR(255) NOT NULL DEFAULT '' COMMENT '描述'
) DEFAULT CHARACTER SET UTF8MB4 COMMENT 'opus 表';