ALTER TABLE `chart_meta` 
MODIFY COLUMN `name` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL COMMENT '名称' AFTER `id`,
MODIFY COLUMN `title` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL COMMENT '图表标题' AFTER `name`,
MODIFY COLUMN `metricsName` varchar(127) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL COMMENT '指标名称' AFTER `title`,
MODIFY COLUMN `fields` varchar(4096) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL COMMENT '字段配置' AFTER `metricsName`,
MODIFY COLUMN `parameters` varchar(4096) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL COMMENT '查询参数' AFTER `fields`,
MODIFY COLUMN `type` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL COMMENT '图表类型分组' AFTER `parameters`,
MODIFY COLUMN `order` int(11) NOT NULL COMMENT '顺序' AFTER `type`,
MODIFY COLUMN `unit` varchar(16) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL COMMENT '原始数据的单位' AFTER `order`;

ALTER TABLE `chart_meta` COMMENT = '监控图表元数据配置';

ALTER TABLE `sp_account` 
MODIFY COLUMN `id` int(11) NOT NULL AUTO_INCREMENT COMMENT '主键' FIRST,
MODIFY COLUMN `auth_id` int(11) NULL DEFAULT NULL COMMENT '认证ID' AFTER `id`,
MODIFY COLUMN `username` varchar(255) CHARACTER SET utf8 COLLATE utf8_general_ci NULL DEFAULT NULL COMMENT '用户名' AFTER `auth_id`,
MODIFY COLUMN `password` varchar(255) CHARACTER SET utf8 COLLATE utf8_general_ci NULL DEFAULT NULL COMMENT '密码' AFTER `username`;

ALTER TABLE `sp_authentication` 
MODIFY COLUMN `id` int(11) NOT NULL AUTO_INCREMENT COMMENT '主键' FIRST,
MODIFY COLUMN `project_id` int(11) UNSIGNED ZEROFILL NULL DEFAULT NULL COMMENT '项目ID' AFTER `id`,
MODIFY COLUMN `name` varchar(255) CHARACTER SET utf8 COLLATE utf8_general_ci NULL DEFAULT NULL COMMENT '显示名称' AFTER `project_id`,
MODIFY COLUMN `extra` varchar(2048) CHARACTER SET utf8 COLLATE utf8_general_ci NOT NULL DEFAULT '' COMMENT '额外数据' AFTER `name`;

ALTER TABLE `sp_chart_meta` 
MODIFY COLUMN `name` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL COMMENT '名称' AFTER `id`,
MODIFY COLUMN `title` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL COMMENT '图表标题' AFTER `name`,
MODIFY COLUMN `metricsName` varchar(127) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL COMMENT '指标名称' AFTER `title`,
MODIFY COLUMN `fields` varchar(4096) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL COMMENT '字段配置' AFTER `metricsName`,
MODIFY COLUMN `parameters` varchar(4096) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL COMMENT '查询参数' AFTER `fields`,
MODIFY COLUMN `type` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL COMMENT '图表类型分组' AFTER `parameters`,
MODIFY COLUMN `order` int(11) NOT NULL COMMENT '顺序' AFTER `type`,
MODIFY COLUMN `unit` varchar(16) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL COMMENT '原始数据的单位' AFTER `order`;

ALTER TABLE `sp_chart_profile` 
MODIFY COLUMN `created_at` datetime(0) NOT NULL COMMENT '创建时间' AFTER `id`,
MODIFY COLUMN `updated_at` datetime(0) NULL DEFAULT NULL COMMENT '更新时间' AFTER `created_at`,
MODIFY COLUMN `unique_id` varchar(125) CHARACTER SET utf8 COLLATE utf8_general_ci NULL DEFAULT NULL COMMENT '唯一ID' AFTER `updated_at`,
MODIFY COLUMN `layout` text CHARACTER SET utf8 COLLATE utf8_general_ci NULL COMMENT '布局信息' AFTER `unique_id`,
MODIFY COLUMN `drawer_info_map` text CHARACTER SET utf8 COLLATE utf8_general_ci NULL COMMENT '数据' AFTER `layout`,
MODIFY COLUMN `url_config` text CHARACTER SET utf8 COLLATE utf8_general_ci NULL COMMENT 'url配置' AFTER `drawer_info_map`,
MODIFY COLUMN `name` varchar(125) CHARACTER SET utf8 COLLATE utf8_general_ci NOT NULL COMMENT '名称' AFTER `url_config`,
MODIFY COLUMN `category` varchar(125) CHARACTER SET utf8 COLLATE utf8_general_ci NOT NULL COMMENT '类目' AFTER `name`,
MODIFY COLUMN `cluster_name` varchar(125) CHARACTER SET utf8 COLLATE utf8_general_ci NULL DEFAULT '' COMMENT '集群名称' AFTER `category`,
MODIFY COLUMN `organization` varchar(125) CHARACTER SET utf8 COLLATE utf8_general_ci NULL DEFAULT '' COMMENT '组织' AFTER `cluster_name`,
MODIFY COLUMN `editable` tinyint(1) UNSIGNED NULL DEFAULT 0 COMMENT '是否可编辑' AFTER `organization`,
MODIFY COLUMN `deletable` tinyint(1) NULL DEFAULT 0 COMMENT '是否可删除' AFTER `editable`,
MODIFY COLUMN `cluster_level` tinyint(1) NULL DEFAULT 0 COMMENT '集群级别' AFTER `deletable`,
MODIFY COLUMN `cluster_type` varchar(20) CHARACTER SET utf8 COLLATE utf8_general_ci NOT NULL DEFAULT '' COMMENT '集群类型' AFTER `cluster_level`;

ALTER TABLE `sp_dashboard_block` 
MODIFY COLUMN `domain` varchar(255) CHARACTER SET utf8 COLLATE utf8_general_ci NULL DEFAULT NULL COMMENT '域名称' AFTER `desc`,
MODIFY COLUMN `scope` varchar(255) CHARACTER SET utf8 COLLATE utf8_general_ci NULL DEFAULT NULL COMMENT '范围' AFTER `domain`,
MODIFY COLUMN `scope_id` varchar(255) CHARACTER SET utf8 COLLATE utf8_general_ci NULL DEFAULT NULL COMMENT '范围ID' AFTER `scope`,
MODIFY COLUMN `version` varchar(32) CHARACTER SET utf8 COLLATE utf8_general_ci NOT NULL COMMENT '版本' AFTER `updated_at`;

ALTER TABLE `sp_dashboard_block_system` 
MODIFY COLUMN `domain` varchar(255) CHARACTER SET utf8 COLLATE utf8_general_ci NULL DEFAULT NULL COMMENT '域名称' AFTER `desc`,
MODIFY COLUMN `scope` varchar(255) CHARACTER SET utf8 COLLATE utf8_general_ci NULL DEFAULT NULL COMMENT '范围' AFTER `domain`,
MODIFY COLUMN `scope_id` varchar(255) CHARACTER SET utf8 COLLATE utf8_general_ci NULL DEFAULT NULL COMMENT '范围ID' AFTER `scope`,
MODIFY COLUMN `version` varchar(32) CHARACTER SET utf8 COLLATE utf8_general_ci NOT NULL COMMENT '版本' AFTER `updated_at`;

ALTER TABLE `sp_history` 
MODIFY COLUMN `id` int(11) NOT NULL AUTO_INCREMENT COMMENT '主键' FIRST,
MODIFY COLUMN `metric_id` int(11) NULL DEFAULT NULL COMMENT '指标ID' AFTER `id`,
MODIFY COLUMN `status_id` int(11) NULL DEFAULT NULL COMMENT '状态ID' AFTER `metric_id`,
MODIFY COLUMN `latency` bigint(64) NOT NULL COMMENT '延迟' AFTER `status_id`,
MODIFY COLUMN `count` int(11) NOT NULL COMMENT '请求次数' AFTER `latency`,
MODIFY COLUMN `code` int(11) UNSIGNED ZEROFILL NOT NULL COMMENT '状态码' AFTER `count`,
MODIFY COLUMN `created_at` timestamp(0) NOT NULL DEFAULT CURRENT_TIMESTAMP(0) ON UPDATE CURRENT_TIMESTAMP(0) COMMENT '创建时间' AFTER `code`;

ALTER TABLE `sp_log_deployment` 
MODIFY COLUMN `id` int(10) NOT NULL AUTO_INCREMENT COMMENT '主键' FIRST,
MODIFY COLUMN `org_id` varchar(32) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL COMMENT '企业ID' AFTER `id`,
MODIFY COLUMN `cluster_name` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL COMMENT '集群名' AFTER `org_id`,
MODIFY COLUMN `cluster_type` tinyint(4) NOT NULL COMMENT '集群类型' AFTER `cluster_name`,
MODIFY COLUMN `es_url` varchar(1024) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL COMMENT 'es的url地址' AFTER `cluster_type`,
MODIFY COLUMN `es_config` varchar(1024) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL COMMENT 'es相关的配置' AFTER `es_url`,
MODIFY COLUMN `kafka_servers` varchar(1024) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL COMMENT 'kafka服务器地址' AFTER `es_config`,
MODIFY COLUMN `kafka_config` varchar(1024) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL COMMENT 'kafka配置' AFTER `kafka_servers`,
MODIFY COLUMN `collector_url` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL DEFAULT '' COMMENT 'collector组件地址' AFTER `kafka_config`,
MODIFY COLUMN `domain` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL COMMENT '泛域名' AFTER `collector_url`,
MODIFY COLUMN `created` datetime(0) NOT NULL COMMENT '创建时间' AFTER `domain`,
MODIFY COLUMN `updated` datetime(0) NOT NULL COMMENT '更新时间' AFTER `created`;

ALTER TABLE `sp_log_instance` 
MODIFY COLUMN `id` int(11) NOT NULL AUTO_INCREMENT COMMENT '主键' FIRST,
MODIFY COLUMN `log_key` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL COMMENT '租户ID' AFTER `id`,
MODIFY COLUMN `org_id` varchar(32) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL COMMENT '企业ID' AFTER `log_key`,
MODIFY COLUMN `org_name` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL COMMENT '企业名' AFTER `org_id`,
MODIFY COLUMN `cluster_name` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL COMMENT '集群名' AFTER `org_name`,
MODIFY COLUMN `project_id` varchar(32) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL COMMENT '项目ID' AFTER `cluster_name`,
MODIFY COLUMN `project_name` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL COMMENT '项目名' AFTER `project_id`,
MODIFY COLUMN `workspace` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL COMMENT '环境' AFTER `project_name`,
MODIFY COLUMN `application_id` varchar(32) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL COMMENT '应用ID' AFTER `workspace`,
MODIFY COLUMN `application_name` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL COMMENT '应用名称' AFTER `application_id`,
MODIFY COLUMN `runtime_id` varchar(32) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL COMMENT 'runtime部署ID' AFTER `application_name`,
MODIFY COLUMN `runtime_name` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL COMMENT 'runtime部署名称' AFTER `runtime_id`,
MODIFY COLUMN `config` varchar(1023) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL COMMENT '部署配置' AFTER `runtime_name`,
MODIFY COLUMN `version` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL COMMENT '版本' AFTER `config`,
MODIFY COLUMN `plan` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL COMMENT '规格' AFTER `version`,
MODIFY COLUMN `is_delete` tinyint(1) NOT NULL COMMENT '是否删除' AFTER `plan`,
MODIFY COLUMN `created` datetime(0) NOT NULL COMMENT '创建时间' AFTER `is_delete`,
MODIFY COLUMN `updated` datetime(0) NOT NULL COMMENT '更新时间' AFTER `created`;

ALTER TABLE `sp_log_metric_config` 
MODIFY COLUMN `id` int(11) NOT NULL AUTO_INCREMENT COMMENT '主键' FIRST,
MODIFY COLUMN `org_id` int(11) NOT NULL COMMENT '企业ID' AFTER `id`,
MODIFY COLUMN `org_name` varchar(128) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL DEFAULT '' COMMENT '企业名' AFTER `org_id`,
MODIFY COLUMN `scope` varchar(32) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL COMMENT '范围' AFTER `org_name`,
MODIFY COLUMN `scope_id` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL COMMENT '范围ID' AFTER `scope`,
MODIFY COLUMN `name` varchar(128) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL COMMENT '名称' AFTER `scope_id`,
MODIFY COLUMN `metric` varchar(128) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL COMMENT '指标名' AFTER `name`,
MODIFY COLUMN `filters` text CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL COMMENT '过滤配置' AFTER `metric`,
MODIFY COLUMN `processors` text CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL COMMENT '日志处理器配置' AFTER `filters`,
MODIFY COLUMN `enable` tinyint(1) NOT NULL COMMENT '是否启用' AFTER `processors`,
MODIFY COLUMN `create_time` datetime(0) NOT NULL COMMENT '创建时间' AFTER `enable`,
MODIFY COLUMN `update_time` datetime(0) NOT NULL COMMENT '更新时间' AFTER `create_time`;

ALTER TABLE `sp_maintenance` 
MODIFY COLUMN `id` int(11) NOT NULL AUTO_INCREMENT COMMENT '主键' FIRST,
MODIFY COLUMN `project_id` int(11) NULL DEFAULT NULL COMMENT '项目ID' AFTER `id`,
MODIFY COLUMN `name` varchar(255) CHARACTER SET utf8 COLLATE utf8_general_ci NULL DEFAULT NULL COMMENT '名称' AFTER `project_id`,
MODIFY COLUMN `duration` int(11) NOT NULL DEFAULT 0 COMMENT '持续时间' AFTER `name`,
MODIFY COLUMN `start_time` timestamp(0) NOT NULL DEFAULT CURRENT_TIMESTAMP(0) ON UPDATE CURRENT_TIMESTAMP(0) COMMENT '开始时间' AFTER `duration`;

ALTER TABLE `sp_metric` 
MODIFY COLUMN `id` int(11) NOT NULL AUTO_INCREMENT COMMENT '主键' FIRST,
MODIFY COLUMN `project_id` int(11) NULL DEFAULT NULL COMMENT '项目ID' AFTER `id`,
MODIFY COLUMN `service_id` int(11) NULL DEFAULT NULL COMMENT '服务ID' AFTER `project_id`,
MODIFY COLUMN `name` varchar(255) CHARACTER SET utf8 COLLATE utf8_general_ci NOT NULL DEFAULT '' COMMENT '名称' AFTER `service_id`,
MODIFY COLUMN `url` varchar(255) CHARACTER SET utf8 COLLATE utf8_general_ci NULL DEFAULT NULL COMMENT 'url地址' AFTER `name`,
MODIFY COLUMN `mode` varchar(255) CHARACTER SET utf8 COLLATE utf8_general_ci NOT NULL DEFAULT '' COMMENT '模式' AFTER `url`,
MODIFY COLUMN `extra` varchar(1024) CHARACTER SET utf8 COLLATE utf8_general_ci NOT NULL DEFAULT '' COMMENT '额外的数据' AFTER `mode`,
MODIFY COLUMN `account_id` int(11) NOT NULL COMMENT '账户ID' AFTER `extra`,
MODIFY COLUMN `status` int(11) NULL DEFAULT 0 COMMENT '状态' AFTER `account_id`,
MODIFY COLUMN `env` varchar(36) CHARACTER SET utf8 COLLATE utf8_general_ci NOT NULL DEFAULT 'PROD' COMMENT '环境' AFTER `status`;

ALTER TABLE `sp_metric_meta` 
MODIFY COLUMN `scope` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL COMMENT '范围' AFTER `id`,
MODIFY COLUMN `scope_id` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL COMMENT '范围ID' AFTER `scope`,
MODIFY COLUMN `group` varchar(128) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL COMMENT '指标分组' AFTER `scope_id`,
MODIFY COLUMN `metric` varchar(128) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL COMMENT '指标名' AFTER `group`,
MODIFY COLUMN `name` varchar(128) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL COMMENT '显示名称' AFTER `metric`,
MODIFY COLUMN `tags` text CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL COMMENT '指标的标签' AFTER `name`,
MODIFY COLUMN `fields` text CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL COMMENT '指标的字段' AFTER `tags`,
MODIFY COLUMN `create_time` datetime(0) NOT NULL COMMENT '创建时间' AFTER `fields`,
MODIFY COLUMN `update_time` datetime(0) NOT NULL COMMENT '更新时间' AFTER `create_time`;

ALTER TABLE `sp_metric_metas` 
MODIFY COLUMN `id` int(11) NOT NULL AUTO_INCREMENT COMMENT '主键' FIRST,
MODIFY COLUMN `metric` varchar(64) CHARACTER SET utf8 COLLATE utf8_general_ci NOT NULL COMMENT '指标名' AFTER `id`,
MODIFY COLUMN `meta_metric` varchar(64) CHARACTER SET utf8 COLLATE utf8_general_ci NOT NULL COMMENT '元数据指标名' AFTER `metric`,
MODIFY COLUMN `name` varchar(64) CHARACTER SET utf8 COLLATE utf8_general_ci NOT NULL COMMENT '显示名称' AFTER `meta_metric`,
MODIFY COLUMN `type` char(16) CHARACTER SET utf8 COLLATE utf8_general_ci NOT NULL COMMENT '类型' AFTER `name`,
MODIFY COLUMN `unit` char(16) CHARACTER SET utf8 COLLATE utf8_general_ci NOT NULL COMMENT '单位' AFTER `type`,
MODIFY COLUMN `force_tags` text CHARACTER SET utf8 COLLATE utf8_general_ci NULL COMMENT '强制需要加上的tag' AFTER `unit`,
MODIFY COLUMN `group_id` varchar(32) CHARACTER SET utf8 COLLATE utf8_general_ci NOT NULL COMMENT '分组ID' AFTER `force_tags`,
MODIFY COLUMN `created_at` datetime(0) NULL DEFAULT NULL COMMENT '创建时间' AFTER `group_id`,
MODIFY COLUMN `updated_at` datetime(0) NULL DEFAULT NULL COMMENT '更新时间' AFTER `created_at`;

ALTER TABLE `sp_monitor` 
MODIFY COLUMN `id` int(11) NOT NULL AUTO_INCREMENT COMMENT '主键' FIRST,
MODIFY COLUMN `monitor_id` varchar(55) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL DEFAULT '' COMMENT '监控租户ID' AFTER `id`,
MODIFY COLUMN `terminus_key` varchar(55) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL DEFAULT '' COMMENT '监控租户ID' AFTER `monitor_id`,
MODIFY COLUMN `terminus_key_runtime` varchar(55) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NULL DEFAULT NULL COMMENT '监控租户ID,废弃' AFTER `terminus_key`,
MODIFY COLUMN `workspace` varchar(55) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NULL DEFAULT NULL COMMENT '环境' AFTER `terminus_key_runtime`,
MODIFY COLUMN `runtime_id` varchar(55) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL DEFAULT '' COMMENT 'runtime部署ID' AFTER `workspace`,
MODIFY COLUMN `runtime_name` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NULL DEFAULT NULL COMMENT 'runtime部署名称' AFTER `runtime_id`,
MODIFY COLUMN `application_id` varchar(55) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NULL DEFAULT NULL COMMENT '应用ID' AFTER `runtime_name`,
MODIFY COLUMN `application_name` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NULL DEFAULT NULL COMMENT '应用名称' AFTER `application_id`,
MODIFY COLUMN `project_id` varchar(55) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NULL DEFAULT NULL COMMENT '项目ID' AFTER `application_name`,
MODIFY COLUMN `project_name` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NULL DEFAULT NULL COMMENT '项目名称' AFTER `project_id`,
MODIFY COLUMN `org_id` varchar(55) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NULL DEFAULT NULL COMMENT '企业ID' AFTER `project_name`,
MODIFY COLUMN `org_name` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NULL DEFAULT NULL COMMENT '企业名' AFTER `org_id`,
MODIFY COLUMN `cluster_id` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NULL DEFAULT NULL COMMENT '集群ID' AFTER `org_name`,
MODIFY COLUMN `cluster_name` varchar(125) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NULL DEFAULT NULL COMMENT '集群名' AFTER `cluster_id`,
MODIFY COLUMN `config` varchar(1023) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL DEFAULT '' COMMENT '配置信息' AFTER `cluster_name`,
MODIFY COLUMN `callback_url` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NULL DEFAULT NULL COMMENT '回调地址' AFTER `config`,
MODIFY COLUMN `version` varchar(55) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NULL DEFAULT NULL COMMENT '版本' AFTER `callback_url`,
MODIFY COLUMN `plan` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NULL DEFAULT NULL COMMENT '规格' AFTER `version`,
MODIFY COLUMN `is_delete` tinyint(1) NOT NULL DEFAULT 0 COMMENT '是否已删除' AFTER `plan`,
MODIFY COLUMN `created` datetime(0) NOT NULL COMMENT '创建时间' AFTER `is_delete`,
MODIFY COLUMN `updated` datetime(0) NOT NULL COMMENT '更新时间' AFTER `created`;

ALTER TABLE `sp_monitor_config` 
MODIFY COLUMN `type` varchar(64) CHARACTER SET utf8 COLLATE utf8_general_ci NOT NULL COMMENT '类型,log/metric' AFTER `org_name`,
MODIFY COLUMN `hash` varchar(64) CHARACTER SET utf8 COLLATE utf8_general_ci NOT NULL COMMENT '哈希值,hash(org_id,type,names+filters)' AFTER `key`;

ALTER TABLE `sp_monitor_config_register` 
MODIFY COLUMN `scope` varchar(32) CHARACTER SET utf8 COLLATE utf8_general_ci NOT NULL COMMENT '范围' AFTER `id`,
MODIFY COLUMN `scope_id` varchar(64) CHARACTER SET utf8 COLLATE utf8_general_ci NOT NULL DEFAULT '' COMMENT '范围ID' AFTER `scope`,
MODIFY COLUMN `namespace` varchar(255) CHARACTER SET utf8 COLLATE utf8_general_ci NOT NULL COMMENT '环境，dev/test/staging/prod/other' AFTER `scope_id`,
MODIFY COLUMN `type` varchar(64) CHARACTER SET utf8 COLLATE utf8_general_ci NOT NULL COMMENT '类型，metric/log' AFTER `namespace`,
MODIFY COLUMN `names` varchar(255) CHARACTER SET utf8 COLLATE utf8_general_ci NOT NULL COMMENT '名字' AFTER `type`,
MODIFY COLUMN `filters` varchar(4096) CHARACTER SET utf8 COLLATE utf8_general_ci NOT NULL DEFAULT '' COMMENT '过滤器配置' AFTER `names`,
MODIFY COLUMN `enable` tinyint(1) NOT NULL DEFAULT 1 COMMENT '是否启用' AFTER `filters`,
MODIFY COLUMN `update_time` datetime(0) NOT NULL COMMENT '更新时间' AFTER `enable`,
MODIFY COLUMN `desc` varchar(64) CHARACTER SET utf8 COLLATE utf8_general_ci NOT NULL DEFAULT '' COMMENT '描述信息' AFTER `update_time`,
MODIFY COLUMN `hash` varchar(64) CHARACTER SET utf8 COLLATE utf8_general_ci NOT NULL COMMENT 'hash值' AFTER `desc`;

ALTER TABLE `sp_project` 
MODIFY COLUMN `id` int(11) NOT NULL AUTO_INCREMENT COMMENT '主键' FIRST,
MODIFY COLUMN `identity` varchar(255) CHARACTER SET utf8 COLLATE utf8_general_ci NOT NULL COMMENT '唯一标示' AFTER `id`,
MODIFY COLUMN `name` varchar(255) CHARACTER SET utf8 COLLATE utf8_general_ci NULL DEFAULT NULL COMMENT '名称' AFTER `identity`,
MODIFY COLUMN `description` varchar(255) CHARACTER SET utf8 COLLATE utf8_general_ci NULL DEFAULT NULL COMMENT '描述信息' AFTER `name`,
MODIFY COLUMN `ats` varchar(255) CHARACTER SET utf8 COLLATE utf8_general_ci NOT NULL COMMENT '废弃字段' AFTER `description`,
MODIFY COLUMN `callback` varchar(255) CHARACTER SET utf8 COLLATE utf8_general_ci NULL DEFAULT NULL COMMENT '回调地址' AFTER `ats`,
MODIFY COLUMN `project_id` int(11) NOT NULL DEFAULT 0 COMMENT '项目ID' AFTER `callback`;

ALTER TABLE `sp_report_history` 
MODIFY COLUMN `scope` varchar(255) CHARACTER SET utf8 COLLATE utf8_general_ci NULL DEFAULT NULL COMMENT '范围' AFTER `id`,
MODIFY COLUMN `scope_id` varchar(255) CHARACTER SET utf8 COLLATE utf8_general_ci NULL DEFAULT NULL COMMENT '范围ID' AFTER `scope`,
MODIFY COLUMN `task_id` bigint(20) UNSIGNED NOT NULL COMMENT '任务ID' AFTER `scope_id`,
MODIFY COLUMN `dashboard_id` varchar(255) CHARACTER SET utf8 COLLATE utf8_general_ci NOT NULL COMMENT '大盘ID' AFTER `task_id`;

ALTER TABLE `sp_report_settings` 
MODIFY COLUMN `id` int(11) NOT NULL AUTO_INCREMENT COMMENT '主键' FIRST,
MODIFY COLUMN `project_id` int(11) NOT NULL COMMENT '项目ID' AFTER `id`,
MODIFY COLUMN `project_name` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL COMMENT '项目名' AFTER `project_id`,
MODIFY COLUMN `workspace` varchar(16) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL COMMENT '环境' AFTER `project_name`,
MODIFY COLUMN `created` datetime(0) NOT NULL COMMENT '创建时间' AFTER `workspace`,
MODIFY COLUMN `weekly_report_enable` tinyint(1) NOT NULL COMMENT '是否启用周报' AFTER `created`,
MODIFY COLUMN `daily_report_enable` tinyint(1) NOT NULL COMMENT '是否启用日报' AFTER `weekly_report_enable`,
MODIFY COLUMN `weekly_report_config` text CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL COMMENT '周报配置' AFTER `daily_report_enable`,
MODIFY COLUMN `daily_report_config` text CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL COMMENT '日报配置' AFTER `weekly_report_config`;

ALTER TABLE `sp_report_task` 
MODIFY COLUMN `scope` varchar(255) CHARACTER SET utf8 COLLATE utf8_general_ci NULL DEFAULT NULL COMMENT '范围' AFTER `name`,
MODIFY COLUMN `scope_id` varchar(255) CHARACTER SET utf8 COLLATE utf8_general_ci NULL DEFAULT NULL COMMENT '范围ID' AFTER `scope`,
MODIFY COLUMN `dashboard_id` varchar(255) CHARACTER SET utf8 COLLATE utf8_general_ci NOT NULL COMMENT '大盘ID' AFTER `type`,
MODIFY COLUMN `pipeline_cron_id` bigint(20) NOT NULL COMMENT '任务ID' AFTER `notify_target`;

ALTER TABLE `sp_reports` 
MODIFY COLUMN `id` int(11) NOT NULL AUTO_INCREMENT COMMENT '主键' FIRST,
MODIFY COLUMN `key` varchar(32) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL COMMENT '报告标示' AFTER `id`,
MODIFY COLUMN `start` datetime(0) NOT NULL COMMENT '开始时间' AFTER `key`,
MODIFY COLUMN `end` datetime(0) NOT NULL COMMENT '结束时间' AFTER `start`,
MODIFY COLUMN `project_id` int(11) NOT NULL COMMENT '项目ID' AFTER `end`,
MODIFY COLUMN `project_name` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL COMMENT '项目名称' AFTER `project_id`,
MODIFY COLUMN `workspace` varchar(16) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL COMMENT '环境' AFTER `project_name`,
MODIFY COLUMN `created` datetime(0) NOT NULL COMMENT '创建时间' AFTER `workspace`,
MODIFY COLUMN `version` varchar(16) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL COMMENT '版本' AFTER `created`,
MODIFY COLUMN `data` text CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL COMMENT '报告数据' AFTER `version`,
MODIFY COLUMN `type` tinyint(4) NOT NULL COMMENT '类型' AFTER `data`,
MODIFY COLUMN `terminus_key` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL DEFAULT '' COMMENT '租户ID' AFTER `type`;

ALTER TABLE `sp_service` 
MODIFY COLUMN `id` int(11) NOT NULL AUTO_INCREMENT COMMENT '主键' FIRST,
MODIFY COLUMN `name` varchar(255) CHARACTER SET utf8 COLLATE utf8_general_ci NULL DEFAULT NULL COMMENT '名称' AFTER `id`,
MODIFY COLUMN `project_id` int(11) UNSIGNED ZEROFILL NOT NULL COMMENT '项目ID' AFTER `name`;

ALTER TABLE `sp_status` 
MODIFY COLUMN `id` int(11) NOT NULL AUTO_INCREMENT COMMENT '主键' FIRST,
MODIFY COLUMN `name` varchar(255) CHARACTER SET utf8 COLLATE utf8_general_ci NULL DEFAULT NULL COMMENT '名称' AFTER `id`,
MODIFY COLUMN `color` varchar(255) CHARACTER SET utf8 COLLATE utf8_general_ci NULL DEFAULT NULL COMMENT '颜色' AFTER `name`,
MODIFY COLUMN `level` int(10) UNSIGNED ZEROFILL NOT NULL COMMENT '级别' AFTER `color`;

ALTER TABLE `sp_trace_request_history` 
MODIFY COLUMN `request_id` varchar(128) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL COMMENT '主键' FIRST,
MODIFY COLUMN `terminus_key` varchar(55) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL DEFAULT '' COMMENT '租户ID' AFTER `request_id`,
MODIFY COLUMN `url` varchar(1024) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL COMMENT '请求地址' AFTER `terminus_key`,
MODIFY COLUMN `query_string` text CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NULL COMMENT '查询参数' AFTER `url`,
MODIFY COLUMN `header` text CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NULL COMMENT '请求头' AFTER `query_string`,
MODIFY COLUMN `body` text CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NULL COMMENT '请求体' AFTER `header`,
MODIFY COLUMN `method` varchar(16) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL COMMENT '请求方法' AFTER `body`,
MODIFY COLUMN `status` int(2) NOT NULL DEFAULT 0 COMMENT '状态' AFTER `method`,
MODIFY COLUMN `response_status` int(11) NOT NULL DEFAULT 200 COMMENT '响应状态码' AFTER `status`,
MODIFY COLUMN `response_body` text CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NULL COMMENT '响应体' AFTER `response_status`,
MODIFY COLUMN `create_time` timestamp(0) NOT NULL DEFAULT CURRENT_TIMESTAMP(0) COMMENT '创建时间' AFTER `response_body`,
MODIFY COLUMN `update_time` timestamp(0) NOT NULL DEFAULT CURRENT_TIMESTAMP(0) ON UPDATE CURRENT_TIMESTAMP(0) COMMENT '更新时间' AFTER `create_time`;

ALTER TABLE `sp_user` 
MODIFY COLUMN `id` int(11) NOT NULL AUTO_INCREMENT COMMENT '主键' FIRST,
MODIFY COLUMN `username` varchar(255) CHARACTER SET utf8 COLLATE utf8_general_ci NULL DEFAULT NULL COMMENT '用户名' AFTER `id`,
MODIFY COLUMN `salt` varchar(255) CHARACTER SET utf8 COLLATE utf8_general_ci NULL DEFAULT NULL COMMENT '加密盐' AFTER `username`,
MODIFY COLUMN `password` varchar(255) CHARACTER SET utf8 COLLATE utf8_general_ci NULL DEFAULT NULL COMMENT '密码' AFTER `salt`,
MODIFY COLUMN `created_at` timestamp(0) NULL DEFAULT NULL ON UPDATE CURRENT_TIMESTAMP(0) COMMENT '创建时间' AFTER `password`;