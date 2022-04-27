CREATE TABLE erda_gallery_room
(
    `id`           VARCHAR(36)  NOT NULL PRIMARY KEY COMMENT 'primary key',
    `created_at`   DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `updated_at`   DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    `deleted_at`   DATETIME     NOT NULL DEFAULT '1970-01-01 00:00:00' COMMENT '删除时间',

    `key`          VARCHAR(36)  NOT NULL DEFAULT '' COMMENT 'room key 用于接口认证',
    `secret`       VARCHAR(36)  NOT NULL DEFAULT '' COMMENT 'room secret 用于接口认证',
    `algorithm`    VARCHAR(36)  NOT NULL DEFAULT 'SHA256' 'room 认证的算法',

    `name`         VARCHAR(191) NOT NULL DEFAULT '' COMMENT 'room 标识',
    `display_name` VARCHAR(255) NOT NULL DEFAULT '' COMMENT 'room 名称',
    `logo`         VARCHAR(512) NOT NULL DEFAULT '' COMMENT 'room logo 地址',
    `config`       TEXT         NOT NULL COMMENT 'room 配置信息',
    `locale`       VARCHAR(50)  NOT NULL DEFAULT '' COMMENT 'room 默认所在地',
    `is_public`    TINYINT(1)   NOT NULL DEFAULT '' COMMENT 'room 是否公开'
) DEFAULT CHARACTER SET UTF8MB4 COMMENT 'gallery room 即租户';

CREATE TABLE erda_gallery_artifacts
(
    `id`           VARCHAR(36)  NOT NULL PRIMARY KEY COMMENT 'primary key',
    `created_at`   DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `updated_at`   DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    `deleted_at`   DATETIME     NOT NULL DEFAULT '1970-01-01 00:00:00' COMMENT '删除时间',

    `room_id`      VARCHAR(36)  NOT NULL DEFAULT 0 COMMENT 'erda_gallery_room.id',
    `room_name`    VARCHAR(50)  NOT NULL DEFAULT '' COMMENT 'erda_gallery_room.name 冗余存储',

    `type`         VARCHAR(64)  NOT NULL DEFAULT '' COMMENT 'artifacts 类型, 目前都是 artifacts/project',
    `name`         VARCHAR(191) NOT NULL DEFAULT '' COMMENT 'artifacts 标识',
    `version`      VARCHAR(128) NOT NULL DEFAULT '' COMMENT 'artifacts 版本',
    `display_name` VARCHAR(255) NOT NULL DEFAULT '' COMMENT 'artifacts 名称',

    INDEX idx_room_name_type (`room_id`, `name`, `type`),
    UNIQUE uk_room_name_type_version (`room_id`, `name`, `type`, `version`)
) DEFAULT CHARACTER SET UTF8MB4 COMMENT 'artifacts 记录表';

CREATE TABLE erda_gallery_artifacts_presentation
(
    `id`                VARCHAR(36)   NOT NULL PRIMARY KEY COMMENT 'primary key',
    `created_at`        DATETIME      NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `updated_at`        DATETIME      NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    `deleted_at`        DATETIME      NOT NULL DEFAULT '1970-01-01 00:00:00' COMMENT '删除时间',

    `artifacts_id`      VARCHAR(36)   NOT NULL DEFAULT '' COMMENT 'erda_gallery_artifacts.id',

    `values`            TEXT          NOT NULL COMMENT '预留字段 map[string]string',
    `ref`               VARCHAR(512)  NOT NULL DEFAULT '' COMMENT '.ref: 展示信息时的回调, 是实现了特定接口的 http 链接, 这个接口要返回展示信息的结构',
    `name`              VARCHAR(191)  NOT NULL DEFAULT '' COMMENT '.info.name: artifact 标识',
    `display_name`      VARCHAR(255)  NOT NULL DEFAULT '' COMMENT '.info.displayName: artifact 名称',
    `type`              VARCHAR(64)   NOT NULL DEFAULT '' COMMENT '.info.type: artifact 类型',
    `version`           VARCHAR(128)  NOT NULL DEFAULT '' COMMENT '.info.version: artifact 版本号',
    `summary`           VARCHAR(255)  NOT NULL DEFAULT '' COMMENT '.info.summary: artifact 概述',
    `desc`              VARCHAR(1024) NOT NULL DEFAULT '' COMMENT '.info.desc: artifact 描述 [预留字段]',
    `contact_name`      VARCHAR(191)  NOT NULL DEFAULT '' COMMENT '.info.contact.name: 联系人名称',
    `contact_url`       VARCHAR(512)  NOT NULL DEFAULT '' COMMENT '.info.contact.url: 联系人链接',
    `contact_email`     VARCHAR(128)  NOT NULL DEFAULT '' COMMENT '.info.contact.email: 联系人 email',
    `is_open_sourced`   BOOLEAN       NOT NULL DEFAULT 0 COMMENT '.info.opensource.isOpenSourced: 是否开源',
    `opensource_url`    VARCHAR(512)  NOT NULL DEFAULT '' COMMENT '.info.opensource.url: 开源仓库地址',
    `license_name`      VARCHAR(64)   NOT NULL DEFAULT '' COMMENT '.info.opensource.license.name: 开源协议, 如 "Apache 2.0"',
    `license_url`       VARCHAR(512)  NOT NULL DEFAULT '' COMMENT '.info.opensource.license.url: 开源协议地址, 如 "https://www.apache.org/licenses/LICENSE-2.0.html"',
    `logo_url`          VARCHAR(512)  NOT NULL DEFAULT '' COMMENT 'info.logURL: logo 链接',
    `homepage_name`     VARCHAR(128)  NOT NULL DEFAULT '' COMMENT '.info.homepage.name: 主页名称',
    `homepage_url`      VARCHAR(512)  NOT NULL DEFAULT '' COMMENT '.info.homepage.url: 主页地址',
    `homepage_logo_url` VARCHAR(512)  NOT NULL DEFAULT '' COMMENT '.info.homepage.logURL: 主页 logo 链接',
    `is_downloadable`   BOOLEAN       NOT NULL DEFAULT 0 COMMENT '.download.downloadable: 是否可以下载',
    `download_url`      VARCHAR(512)  NOT NULL DEFAULT '' COMMENT '.download.url: 下载地址',
    `readme`            LONGTEXT      NOT NULL COMMENT '.readme: 主要的显示信息, 形如 [{"lang": "zh_CN", "source": "https://...", "text": "..."}]',
    `schemas`           MEDIUMTEXT    NOT NULL COMMENT '.schemas, 有效的 Openapi3 协议 SchemaRef 片段的集合, 用于展示参数, 如 {"params": [{}]}',
    `labels`            VARCHAR(512)  NOT NULL DEFAULT '' COMMENT '.labels: 标签, 形如 a=A&b=B 的键值对列表',
    `catalog`           VARCHAR(64)   NOT NULL DEFAULT '' COMMENT '.catalog: 分目'
) DEFAULT CHARACTER SET UTF8MB4 COMMENT '如何展示 artifacts 的信息';
