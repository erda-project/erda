CREATE TABLE `erda_profile_app`
(
    `id`              bigint(20) unsigned NOT NULL AUTO_INCREMENT COMMENT 'primary key',
    `units`           varchar(50)  NOT NULL COMMENT '单位',
    `sample_type`     varchar(50)  NOT NULL COMMENT '采样类型',
    `sample_rate`     int(11) NOT NULL COMMENT '采样率',
    `aggregation_type` varchar(50)  NOT NULL COMMENT '聚合类型',
    `name`            varchar(50)  NOT NULL COMMENT '指标名称',
    `updated_at`      datetime     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    `created_at`      datetime     NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `deleted_at`      bigint(20) NOT NULL DEFAULT '0' COMMENT '删除时间',
    `is_deleted`      tinyint(1)  NOT NULL DEFAULT '0' COMMENT '是否删除',
    `org_id`          varchar(100) NOT NULL COMMENT '组织id',
    `org_name`        varchar(50)  NOT NULL COMMENT '组织名称',
    `project_id`      varchar(100) NOT NULL COMMENT '项目id',
    `project_name`    varchar(50)  NOT NULL COMMENT '项目名称',
    `app_id`          varchar(100) NOT NULL COMMENT '应用id',
    `app_name`        varchar(50)  NOT NULL COMMENT '应用名称',
    `spy_name`        varchar(50) NOT NULL COMMENT '可持续性观测名称',
    `workspace`       varchar(50)  NOT NULL COMMENT '部署环境',
    `cluster_name`    varchar(50)  NOT NULL COMMENT '集群名称',
    `service_name`    varchar(50)  NOT NULL COMMENT '服务名称',
    PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=9 DEFAULT CHARSET=utf8mb4 COMMENT ='可持续性观测应用记录表';