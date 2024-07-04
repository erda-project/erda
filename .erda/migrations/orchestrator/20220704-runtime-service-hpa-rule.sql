/*
 * Copyright (c) 2021 Terminus, Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *      http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

CREATE TABLE `erda_v2_runtime_hpa`
(
    `id`                  varchar(36)  NOT NULL COMMENT '规则 ID',
    `created_at`          datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `updated_at`          datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    `rule_name`           varchar(255) NOT NULL COMMENT '规则名称',
    `rule_namespace`      varchar(255) NOT NULL COMMENT '规则部署所在的命名空间',
    `org_id`              bigint(20) unsigned NOT NULL COMMENT '组织ID',
    `org_name`            varchar(50) NOT NULL DEFAULT '' COMMENT '组织名称',
    `org_display_name`    varchar(64) NOT NULL DEFAULT '' COMMENT '组织显示名称',
    `project_id`          bigint(20) unsigned NOT NULL COMMENT '项目ID',
    `project_name`        varchar(64) NOT NULL DEFAULT '' COMMENT '项目名称',
    `proj_display_name`   varchar(64) NOT NULL DEFAULT '' COMMENT '项目显示名称',
    `application_id`      bigint(20) unsigned NOT NULL COMMENT '应用ID',
    `application_name`    varchar(64) NOT NULL DEFAULT '' COMMENT '应用名称',
    `app_display_name`    varchar(64) NOT NULL DEFAULT '' COMMENT '应用显示名称',
    `runtime_id`          bigint(20) unsigned NOT NULL COMMENT 'Runtime ID',
    `runtime_name`        varchar(255) NOT NULL DEFAULT '' COMMENT 'Runtime 名称',
    `workspace`           varchar(16) NOT NULL DEFAULT '' COMMENT '部署环境',
    `cluster_name`        varchar(255) NOT NULL DEFAULT '' COMMENT '集群名称',
    `user_id`             varchar(255) NOT NULL COMMENT '用户Id',
    `user_name`           varchar(255) NOT NULL DEFAULT ''  COMMENT '用户名 (唯一)',
    `nick_name`           varchar(128) NOT NULL DEFAULT ''  COMMENT '用户昵称',
    `service_name`        varchar(255) NOT NULL DEFAULT '' COMMENT 'HPA 规则关联的服务名称',
    `rules`               text NOT NULL COMMENT 'scaledObject json 缓存',
    `applied`             varchar(1)   NOT NULL DEFAULT 'N' COMMENT '规则是否已使用',
    `soft_deleted_at`     bigint(20) NOT NULL DEFAULT 0 COMMENT '删除时间',
    PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='Runtime HPA信息';

/*
CREATE TABLE `erda_v2_runtime_hpa_events`
(
    `id`                  varchar(36)  NOT NULL COMMENT '规则 ID',
    `created_at`          datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `updated_at`          datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    `org_id`              bigint(20) unsigned NOT NULL COMMENT '组织ID',
    `org_name`            varchar(50) NOT NULL DEFAULT '' COMMENT '组织名称',
    `runtime_id`          bigint(20) unsigned NOT NULL COMMENT 'Runtime ID',
    `service_name`        varchar(255) NOT NULL DEFAULT '' COMMENT 'HPA 规则关联的服务名称',
    `event`               text NOT NULL COMMENT 'hpa event 摘要 json 缓存',
    `soft_deleted_at`     bigint(20) NOT NULL DEFAULT 0 COMMENT '删除时间',
    PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='Runtime HPA Event 信息';
*/
