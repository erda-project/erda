CREATE TABLE `pipeline_definitions` (
  `id` varchar(36) NOT NULL COMMENT '自增id',
  `created_at` datetime NOT NULL DEFAULT '1970-01-01 00:00:00' COMMENT 'CREATED AT',
  `updated_at` datetime NOT NULL DEFAULT '1970-01-01 00:00:00' ON UPDATE CURRENT_TIMESTAMP COMMENT 'UPDATED AT',
  `name` varchar(36) NOT NULL DEFAULT '' COMMENT '定义名称',
  `cost_time` bigint(20) NOT NULL DEFAULT '0' COMMENT '上次执行的耗时',
  `creator` varchar(36) NOT NULL DEFAULT '' COMMENT '定义创建者',
  `executor` varchar(36) NOT NULL DEFAULT '' COMMENT '最后一次执行者',
  `started_at` datetime NOT NULL DEFAULT '1970-01-01 00:00:00' COMMENT '最后一次流水线开始时间',
  `ended_at` datetime NOT NULL DEFAULT '1970-01-01 00:00:00' COMMENT '最后一次流水线结束时间',
  `soft_deleted_at` bigint(20) NOT NULL DEFAULT '0' COMMENT '软删除',
  `pipeline_source_id` varchar(36) NOT NULL DEFAULT '' COMMENT 'pipeline 来源id',
  `pipeline_definition_extra_id` varchar(36) NOT NULL DEFAULT '' COMMENT '详情 id',
  `category` varchar(20) NOT NULL DEFAULT '' COMMENT '类型',
  `status` varchar(20) NOT NULL COMMENT '最后一次流水线执行状态',
  `pipeline_id` bigint(20) DEFAULT NULL COMMENT '最后一次执行的流水线id',
  PRIMARY KEY (`id`),
  KEY `pipeline_source_id_index` (`pipeline_source_id`),
  KEY `pipeline_id_index` (`pipeline_id`),
  KEY `pipeline_definition_extra_id_index` (`pipeline_definition_extra_id`),
  KEY `name_index` (`name`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='pipeline definition table';


CREATE TABLE `pipeline_definition_extras` (
  `id` varchar(36) NOT NULL DEFAULT '' COMMENT '主键',
  `extra` mediumtext NOT NULL COMMENT '详细信息',
  `created_at` datetime NOT NULL DEFAULT '1970-01-01 00:00:00' COMMENT '创建时间',
  `updated_at` datetime NOT NULL DEFAULT '1970-01-01 00:00:00' ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='pipeline definition extra table';

CREATE TABLE `pipeline_sources` (
  `id` varchar(36) NOT NULL DEFAULT '' COMMENT '主键',
  `source_type` varchar(20) NOT NULL COMMENT '来源类型',
  `soft_deleted_at` bigint(20) NOT NULL DEFAULT '0' COMMENT '软删除',
  `created_at` datetime NOT NULL DEFAULT '1970-01-01 00:00:00' COMMENT '创建时间',
  `updated_at` datetime NOT NULL DEFAULT '1970-01-01 00:00:00' ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  `version_lock` bigint(20) NOT NULL DEFAULT '1' COMMENT '乐观锁',
  `pipeline_yml` mediumtext NOT NULL COMMENT 'yml 文件内容',
  `remote` varchar(200) NOT NULL DEFAULT '' COMMENT '源地址',
  `ref` varchar(200) NOT NULL DEFAULT '' COMMENT '源地址位置',
  `path` varchar(200) NOT NULL DEFAULT '' COMMENT '源地址位置目录位置',
  `name` varchar(200) NOT NULL DEFAULT '' COMMENT '流水线名称',
  PRIMARY KEY (`id`),
  KEY `source_type_index` (`source_type`),
  KEY `remote_index` (`remote`),
  KEY `name_index` (`name`),
  KEY `ref_index` (`ref`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='pipeline definition source table';


ALTER TABLE `pipeline_bases` ADD COLUMN `pipeline_definition_id` varchar(36) DEFAULT '' COMMENT '流水线定义id';
ALTER TABLE `pipeline_crons` ADD COLUMN `pipeline_definition_id` varchar(36) DEFAULT '' COMMENT '流水线定义id';
