-- MySQL dump 10.14  Distrib 5.5.68-MariaDB, for Linux (x86_64)
--
-- Host: rm-uf6zw0kj412p3u91h.mysql.rds.aliyuncs.com    Database: dice
-- ------------------------------------------------------
-- Server version	5.6.16-log

/*!40101 SET @OLD_CHARACTER_SET_CLIENT=@@CHARACTER_SET_CLIENT */;
/*!40101 SET @OLD_CHARACTER_SET_RESULTS=@@CHARACTER_SET_RESULTS */;
/*!40101 SET @OLD_COLLATION_CONNECTION=@@COLLATION_CONNECTION */;
/*!40101 SET NAMES utf8 */;
/*!40103 SET @OLD_TIME_ZONE=@@TIME_ZONE */;
/*!40103 SET TIME_ZONE='+00:00' */;
/*!40014 SET @OLD_UNIQUE_CHECKS=@@UNIQUE_CHECKS, UNIQUE_CHECKS=0 */;
/*!40014 SET @OLD_FOREIGN_KEY_CHECKS=@@FOREIGN_KEY_CHECKS, FOREIGN_KEY_CHECKS=0 */;
/*!40101 SET @OLD_SQL_MODE=@@SQL_MODE, SQL_MODE='NO_AUTO_VALUE_ON_ZERO' */;
/*!40111 SET @OLD_SQL_NOTES=@@SQL_NOTES, SQL_NOTES=0 */;

--
-- Table structure for table `ai_environment`
--

DROP TABLE IF EXISTS `ai_environment`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `ai_environment` (
  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT,
  `name` varchar(255) DEFAULT NULL,
  `description` varchar(255) DEFAULT NULL,
  `owner_name` varchar(255) DEFAULT NULL,
  `owner_id` bigint(20) unsigned DEFAULT NULL,
  `organization_id` bigint(20) unsigned DEFAULT NULL,
  `organization_name` varchar(255) DEFAULT NULL,
  `workspace` varchar(255) DEFAULT NULL,
  `created_at` datetime DEFAULT NULL,
  `updated_at` datetime DEFAULT NULL,
  `requires` text,
  `labels` text,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=102 DEFAULT CHARSET=utf8mb4 COMMENT='AI 依赖集合配置信息';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `ai_mod`
--

DROP TABLE IF EXISTS `ai_mod`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `ai_mod` (
  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT,
  `type` varchar(255) DEFAULT NULL,
  `name` varchar(255) DEFAULT NULL,
  `version` varchar(255) DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=10 DEFAULT CHARSET=utf8mb4 COMMENT='AI 依赖配置信息';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `ai_notebook`
--

DROP TABLE IF EXISTS `ai_notebook`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `ai_notebook` (
  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT,
  `name` varchar(255) DEFAULT NULL,
  `description` varchar(255) DEFAULT NULL,
  `owner_name` varchar(255) DEFAULT NULL,
  `owner_id` bigint(20) unsigned DEFAULT NULL,
  `organization_id` bigint(20) unsigned DEFAULT NULL,
  `organization_name` varchar(255) DEFAULT NULL,
  `workspace` varchar(255) DEFAULT NULL,
  `created_at` datetime DEFAULT NULL,
  `updated_at` datetime DEFAULT NULL,
  `cluster_name` varchar(255) DEFAULT NULL,
  `project_name` varchar(255) DEFAULT NULL,
  `application_name` varchar(255) DEFAULT NULL,
  `envs` text,
  `image` varchar(255) DEFAULT NULL,
  `requirement_env_id` bigint(20) unsigned DEFAULT NULL,
  `data_source_id` bigint(20) unsigned DEFAULT NULL,
  `generic_domain` varchar(255) DEFAULT NULL,
  `cluster_domain` varchar(255) DEFAULT NULL,
  `resource_cpu` double DEFAULT NULL,
  `resource_memory` int(11) DEFAULT NULL,
  `status_started_at` datetime DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=1631153073 DEFAULT CHARSET=utf8mb4 COMMENT='AI Jupyter IDE 配置信息';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `chart_meta`
--

DROP TABLE IF EXISTS `chart_meta`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `chart_meta` (
  `id` int(10) NOT NULL AUTO_INCREMENT COMMENT '主键',
  `name` varchar(64) NOT NULL,
  `title` varchar(64) NOT NULL,
  `metricsName` varchar(127) NOT NULL,
  `fields` varchar(4096) NOT NULL,
  `parameters` varchar(4096) NOT NULL,
  `type` varchar(64) NOT NULL,
  `order` int(11) NOT NULL,
  `unit` varchar(16) NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `name_unique` (`name`)
) ENGINE=InnoDB AUTO_INCREMENT=308 DEFAULT CHARSET=utf8mb4 BLOCK_FORMAT=ENCRYPTED;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `ci_v3_build_artifacts`
--

DROP TABLE IF EXISTS `ci_v3_build_artifacts`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `ci_v3_build_artifacts` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT COMMENT '主键',
  `created_at` datetime DEFAULT NULL COMMENT '创建时间',
  `updated_at` datetime DEFAULT NULL COMMENT '更新时间',
  `sha_256` varchar(128) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '构建产物的 SHA256',
  `identity_text` text COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '构建产物用于计算 SHA256 的唯一标识内容',
  `type` varchar(128) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '构建产物类型',
  `content` text COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '构建产物的内容',
  `cluster_name` varchar(255) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT '集群名',
  `pipeline_id` bigint(20) NOT NULL COMMENT '关联的流水线 ID',
  PRIMARY KEY (`id`),
  UNIQUE KEY `sha_256` (`sha_256`)
) ENGINE=InnoDB AUTO_INCREMENT=22954 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci BLOCK_FORMAT=ENCRYPTED COMMENT='buildpack action 使用的构建产物表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `ci_v3_build_caches`
--

DROP TABLE IF EXISTS `ci_v3_build_caches`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `ci_v3_build_caches` (
  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT COMMENT '主键',
  `name` varchar(200) DEFAULT NULL COMMENT '缓存名',
  `cluster_name` varchar(200) DEFAULT NULL COMMENT '集群名',
  `last_pull_at` datetime DEFAULT NULL COMMENT '缓存最近一次被拉取的时间',
  `created_at` datetime DEFAULT NULL COMMENT '创建时间',
  `updated_at` datetime DEFAULT NULL COMMENT '更新时间',
  `deleted_at` datetime DEFAULT NULL COMMENT '删除时间',
  PRIMARY KEY (`id`),
  KEY `idx_name` (`name`),
  KEY `idx_cluster_name` (`cluster_name`)
) ENGINE=InnoDB AUTO_INCREMENT=5681 DEFAULT CHARSET=utf8 BLOCK_FORMAT=ENCRYPTED COMMENT='buildpack action 使用的构建缓存';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `cloud_resource_routing`
--

DROP TABLE IF EXISTS `cloud_resource_routing`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `cloud_resource_routing` (
  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT,
  `created_at` timestamp NULL DEFAULT NULL,
  `updated_at` timestamp NULL DEFAULT NULL,
  `resource_id` varchar(128) DEFAULT NULL COMMENT '云资源id',
  `resource_name` varchar(64) DEFAULT NULL COMMENT '云资源名称',
  `resource_type` varchar(32) DEFAULT NULL COMMENT '云资源类型',
  `vendor` varchar(32) DEFAULT NULL COMMENT '云服务提供商',
  `org_id` varchar(64) DEFAULT NULL COMMENT 'org id',
  `cluster_name` varchar(64) DEFAULT NULL COMMENT '集群名',
  `project_id` varchar(64) DEFAULT NULL COMMENT '引用云资源的项目id',
  `addon_id` varchar(64) DEFAULT NULL COMMENT '引用云资源的addon_id',
  `status` varchar(16) DEFAULT NULL COMMENT '引用状态',
  `record_id` bigint(20) unsigned DEFAULT NULL,
  `detail` text,
  PRIMARY KEY (`id`),
  KEY `idx_cloud_resource_routing_resource_id` (`resource_id`),
  KEY `idx_cloud_resource_routing_project_id` (`project_id`),
  KEY `idx_cloud_resource_routing_record_id` (`record_id`)
) ENGINE=InnoDB AUTO_INCREMENT=21 DEFAULT CHARSET=utf8mb4 COMMENT='云addon关联的云资源信息';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `cm_containers`
--

DROP TABLE IF EXISTS `cm_containers`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `cm_containers` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT,
  `created_at` timestamp NULL DEFAULT NULL,
  `updated_at` timestamp NULL DEFAULT NULL,
  `container_id` varchar(64) DEFAULT NULL,
  `deleted` tinyint(1) DEFAULT NULL,
  `started_at` varchar(255) DEFAULT NULL,
  `finished_at` varchar(255) DEFAULT NULL,
  `exit_code` int(11) DEFAULT NULL,
  `privileged` tinyint(1) DEFAULT NULL,
  `cluster` varchar(255) DEFAULT NULL,
  `host_private_ip_addr` varchar(255) DEFAULT NULL,
  `ip_address` varchar(255) DEFAULT NULL,
  `image` varchar(255) DEFAULT NULL,
  `cpu` double DEFAULT NULL,
  `memory` bigint(20) DEFAULT NULL,
  `disk` bigint(20) DEFAULT NULL,
  `dice_org` varchar(255) DEFAULT NULL,
  `dice_project` varchar(40) DEFAULT NULL,
  `dice_application` varchar(255) DEFAULT NULL,
  `dice_runtime` varchar(40) DEFAULT NULL,
  `dice_service` varchar(255) DEFAULT NULL,
  `edas_app_id` varchar(64) DEFAULT NULL,
  `edas_app_name` varchar(128) DEFAULT NULL,
  `edas_group_id` varchar(64) DEFAULT NULL,
  `dice_project_name` varchar(255) DEFAULT NULL,
  `dice_application_name` varchar(255) DEFAULT NULL,
  `dice_runtime_name` varchar(255) DEFAULT NULL,
  `dice_component` varchar(255) DEFAULT NULL,
  `dice_addon` varchar(255) DEFAULT NULL,
  `dice_addon_name` varchar(255) DEFAULT NULL,
  `dice_workspace` varchar(255) DEFAULT NULL,
  `dice_shared_level` varchar(255) DEFAULT NULL,
  `status` varchar(255) DEFAULT NULL,
  `time_stamp` bigint(20) DEFAULT NULL,
  `task_id` varchar(180) DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_project_id` (`dice_project`),
  KEY `idx_runtime_id` (`dice_runtime`),
  KEY `idx_edas_app_id` (`edas_app_id`),
  KEY `task_id` (`task_id`),
  KEY `container_id` (`container_id`)
) ENGINE=InnoDB AUTO_INCREMENT=6664356 DEFAULT CHARSET=utf8mb4 BLOCK_FORMAT=ENCRYPTED COMMENT='容器实例元数据';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `cm_deployments`
--

DROP TABLE IF EXISTS `cm_deployments`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `cm_deployments` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT,
  `created_at` timestamp NULL DEFAULT NULL,
  `updated_at` timestamp NULL DEFAULT NULL,
  `org_id` bigint(20) unsigned DEFAULT NULL,
  `project_id` bigint(20) unsigned DEFAULT NULL,
  `application_id` bigint(20) unsigned DEFAULT NULL,
  `pipeline_id` bigint(20) unsigned DEFAULT NULL,
  `task_id` bigint(20) unsigned DEFAULT NULL,
  `queue_time_sec` bigint(20) DEFAULT NULL,
  `cost_time_sec` bigint(20) DEFAULT NULL,
  `project_name` varchar(255) DEFAULT NULL,
  `application_name` varchar(255) DEFAULT NULL,
  `task_name` varchar(255) DEFAULT NULL,
  `status` varchar(255) DEFAULT NULL,
  `env` varchar(255) DEFAULT NULL,
  `cluster_name` varchar(255) DEFAULT NULL,
  `user_id` varchar(255) DEFAULT NULL,
  `runtime_id` varchar(255) DEFAULT NULL,
  `release_id` varchar(255) DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `org_id` (`org_id`),
  KEY `idx_task_id` (`task_id`)
) ENGINE=InnoDB AUTO_INCREMENT=30083 DEFAULT CHARSET=utf8mb4 BLOCK_FORMAT=ENCRYPTED COMMENT='部署的服务信息';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `cm_hosts`
--

DROP TABLE IF EXISTS `cm_hosts`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `cm_hosts` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT,
  `created_at` timestamp NULL DEFAULT NULL,
  `updated_at` timestamp NULL DEFAULT NULL,
  `name` varchar(255) DEFAULT NULL,
  `org_name` varchar(100) DEFAULT NULL,
  `cluster` varchar(100) DEFAULT NULL,
  `private_addr` varchar(255) DEFAULT NULL,
  `cpus` double DEFAULT NULL,
  `cpu_usage` double DEFAULT NULL,
  `memory` bigint(20) DEFAULT NULL,
  `memory_usage` bigint(20) DEFAULT NULL,
  `disk` bigint(20) DEFAULT NULL,
  `disk_usage` bigint(20) DEFAULT NULL,
  `load5` double DEFAULT NULL,
  `labels` varchar(255) DEFAULT NULL,
  `os` varchar(255) DEFAULT NULL,
  `kernel_version` varchar(255) DEFAULT NULL,
  `system_time` varchar(255) DEFAULT NULL,
  `birthday` bigint(20) DEFAULT NULL,
  `time_stamp` bigint(20) DEFAULT NULL,
  `deleted` tinyint(1) DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `org_name` (`org_name`),
  KEY `cluster` (`cluster`)
) ENGINE=InnoDB AUTO_INCREMENT=181 DEFAULT CHARSET=utf8mb4 BLOCK_FORMAT=ENCRYPTED COMMENT='主机元数据';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `cm_jobs`
--

DROP TABLE IF EXISTS `cm_jobs`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `cm_jobs` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT,
  `created_at` timestamp NULL DEFAULT NULL,
  `updated_at` timestamp NULL DEFAULT NULL,
  `org_id` bigint(20) unsigned DEFAULT NULL,
  `project_id` bigint(20) unsigned DEFAULT NULL,
  `application_id` bigint(20) unsigned DEFAULT NULL,
  `pipeline_id` bigint(20) unsigned DEFAULT NULL,
  `task_id` bigint(20) unsigned DEFAULT NULL,
  `queue_time_sec` bigint(20) DEFAULT NULL,
  `cost_time_sec` bigint(20) DEFAULT NULL,
  `project_name` varchar(255) DEFAULT NULL,
  `application_name` varchar(255) DEFAULT NULL,
  `task_name` varchar(255) DEFAULT NULL,
  `status` varchar(255) DEFAULT NULL,
  `env` varchar(255) DEFAULT NULL,
  `cluster_name` varchar(255) DEFAULT NULL,
  `task_type` varchar(255) DEFAULT NULL,
  `user_id` varchar(255) DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `org_id` (`org_id`),
  KEY `idx_task_id` (`task_id`)
) ENGINE=InnoDB AUTO_INCREMENT=4352308 DEFAULT CHARSET=utf8mb4 BLOCK_FORMAT=ENCRYPTED COMMENT='运行的job信息';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `co_clusters`
--

DROP TABLE IF EXISTS `co_clusters`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `co_clusters` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `org_id` int(11) NOT NULL,
  `name` varchar(41) NOT NULL,
  `display_name` varchar(255) NOT NULL DEFAULT '',
  `type` enum('dcos','edas','k8s','localdocker','swarm') NOT NULL,
  `cloud_vendor` varchar(255) NOT NULL DEFAULT '',
  `logo` text NOT NULL,
  `description` text NOT NULL,
  `wildcard_domain` varchar(255) NOT NULL DEFAULT '',
  `config` text,
  `urls` text,
  `settings` text,
  `scheduler` text,
  `opsconfig` text COMMENT 'OPS配置',
  `resource` text,
  `sys` text,
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `co_clusters_name` (`name`)
) ENGINE=InnoDB AUTO_INCREMENT=4 DEFAULT CHARSET=utf8mb4 BLOCK_FORMAT=ENCRYPTED COMMENT='集群详细配置信息';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `co_clusters_bak`
--

DROP TABLE IF EXISTS `co_clusters_bak`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `co_clusters_bak` (
  `id` int(11) NOT NULL DEFAULT '0',
  `org_id` int(11) NOT NULL,
  `name` varchar(41) NOT NULL,
  `display_name` varchar(255) NOT NULL DEFAULT '',
  `type` enum('dcos','edas','k8s','localdocker','swarm') NOT NULL,
  `cloud_vendor` varchar(255) NOT NULL DEFAULT '',
  `logo` text NOT NULL,
  `description` text NOT NULL,
  `wildcard_domain` varchar(255) NOT NULL DEFAULT '',
  `config` text,
  `urls` text,
  `settings` text,
  `scheduler` text,
  `opsconfig` text COMMENT 'OPS配置',
  `resource` text,
  `sys` text,
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `dice_api_access`
--

DROP TABLE IF EXISTS `dice_api_access`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `dice_api_access` (
  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT COMMENT 'primary key',
  `asset_id` varchar(191) DEFAULT NULL COMMENT 'asset id',
  `asset_name` varchar(191) DEFAULT NULL COMMENT 'asset name',
  `org_id` bigint(20) DEFAULT NULL COMMENT 'organization id',
  `swagger_version` varchar(16) DEFAULT NULL COMMENT 'swagger version',
  `major` int(11) DEFAULT NULL COMMENT 'version major number',
  `minor` int(11) DEFAULT NULL COMMENT 'version minor number',
  `project_id` bigint(20) DEFAULT NULL COMMENT 'project id',
  `app_id` bigint(20) DEFAULT NULL COMMENT 'application id',
  `workspace` varchar(32) DEFAULT NULL COMMENT 'DEV, TEST, STAGING, PROD',
  `endpoint_id` varchar(32) DEFAULT NULL COMMENT 'gateway endpoint id',
  `authentication` varchar(32) DEFAULT NULL COMMENT 'api-key, parameter-sign, auth2',
  `authorization` varchar(32) DEFAULT NULL COMMENT 'auto, manual',
  `addon_instance_id` varchar(128) DEFAULT NULL COMMENT 'addon instance id',
  `bind_domain` varchar(256) DEFAULT NULL COMMENT 'bind domains',
  `creator_id` varchar(191) DEFAULT NULL COMMENT 'creator user id',
  `updater_id` varchar(191) DEFAULT NULL COMMENT 'updater user id',
  `created_at` datetime DEFAULT NULL COMMENT 'created datetime',
  `updated_at` datetime DEFAULT NULL COMMENT 'last updated datetime',
  `project_name` varchar(191) DEFAULT NULL COMMENT 'project name',
  `app_name` varchar(191) DEFAULT NULL COMMENT 'app name',
  `default_sla_id` bigint(20) DEFAULT NULL COMMENT 'default SLA id',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=30 DEFAULT CHARSET=utf8mb4 COMMENT='API 集市资源访问管理表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `dice_api_asset_version_instances`
--

DROP TABLE IF EXISTS `dice_api_asset_version_instances`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `dice_api_asset_version_instances` (
  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT COMMENT 'primary key',
  `name` varchar(191) DEFAULT NULL COMMENT '实例名',
  `asset_id` varchar(191) DEFAULT NULL COMMENT 'API 集市资源 id',
  `version_id` bigint(20) DEFAULT NULL COMMENT 'dice_api_asset_versions primary key',
  `type` varchar(32) DEFAULT NULL COMMENT '实例类型',
  `runtime_id` bigint(20) DEFAULT NULL COMMENT 'runtime id',
  `service_name` varchar(191) DEFAULT NULL COMMENT '服务名称',
  `endpoint_id` varchar(191) DEFAULT NULL COMMENT '流量入口 endpoint id',
  `url` varchar(1024) DEFAULT NULL COMMENT '实例 url',
  `creator_id` varchar(191) DEFAULT NULL COMMENT '创建者 user id',
  `updater_id` varchar(191) DEFAULT NULL COMMENT '更新者 user id',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  `swagger_version` varchar(16) DEFAULT NULL COMMENT 'swagger version',
  `major` int(11) DEFAULT NULL COMMENT 'major',
  `minor` int(11) DEFAULT NULL COMMENT 'minor',
  `project_id` bigint(20) DEFAULT NULL COMMENT 'project id',
  `app_id` bigint(20) DEFAULT NULL COMMENT 'application id',
  `org_id` bigint(20) DEFAULT NULL COMMENT 'organization id',
  `workspace` varchar(16) DEFAULT NULL COMMENT 'env',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=50 DEFAULT CHARSET=utf8mb4 COMMENT='特定版本的 API 集市资源绑定的实例表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `dice_api_asset_version_specs`
--

DROP TABLE IF EXISTS `dice_api_asset_version_specs`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `dice_api_asset_version_specs` (
  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT COMMENT 'primary key',
  `org_id` bigint(20) DEFAULT NULL COMMENT 'organization id',
  `asset_id` varchar(191) DEFAULT NULL COMMENT 'API 集市资源 id',
  `version_id` bigint(20) DEFAULT NULL COMMENT 'dice_api_asset_versions primary key',
  `spec_protocol` varchar(32) DEFAULT NULL COMMENT 'swagger protocol',
  `spec` longtext COMMENT 'swagger text',
  `creator_id` varchar(191) DEFAULT NULL COMMENT 'creator user id',
  `updater_id` varchar(191) DEFAULT NULL COMMENT 'updater user id',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  `asset_name` varchar(191) DEFAULT NULL COMMENT 'asset name',
  PRIMARY KEY (`id`),
  FULLTEXT KEY `ft_specs` (`spec`)
) ENGINE=InnoDB AUTO_INCREMENT=374 DEFAULT CHARSET=utf8mb4 COMMENT='特定版本的 API 集市资源的 swagger specification 内容';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `dice_api_asset_versions`
--

DROP TABLE IF EXISTS `dice_api_asset_versions`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `dice_api_asset_versions` (
  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT COMMENT 'primary key',
  `org_id` bigint(20) DEFAULT NULL COMMENT 'organization id',
  `asset_id` varchar(191) DEFAULT NULL COMMENT 'API 集市资源 id',
  `major` int(11) DEFAULT NULL COMMENT 'version major number',
  `minor` int(11) DEFAULT NULL COMMENT 'version minor number',
  `patch` int(11) DEFAULT NULL COMMENT 'version patch number',
  `desc` varchar(1024) DEFAULT NULL COMMENT 'description',
  `spec_protocol` varchar(32) DEFAULT NULL COMMENT 'swagger protocol',
  `creator_id` varchar(191) DEFAULT NULL COMMENT 'creator user id',
  `updater_id` varchar(191) DEFAULT NULL COMMENT 'updater user id',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  `swagger_version` varchar(16) DEFAULT NULL COMMENT '用户自定义的版本号, 相当于一个 tag',
  `asset_name` varchar(191) DEFAULT NULL COMMENT 'asset name',
  `deprecated` tinyint(1) DEFAULT '0' COMMENT 'is the asset version deprecated',
  `source` varchar(16) NOT NULL COMMENT '该版本文档来源',
  `app_id` bigint(20) NOT NULL COMMENT '应用 id',
  `branch` varchar(191) NOT NULL COMMENT '分支名',
  `service_name` varchar(191) NOT NULL COMMENT '服务名',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=347 DEFAULT CHARSET=utf8mb4 COMMENT='API 集市资源的版本列表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `dice_api_assets`
--

DROP TABLE IF EXISTS `dice_api_assets`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `dice_api_assets` (
  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT COMMENT 'primary key',
  `asset_id` varchar(191) DEFAULT NULL COMMENT 'API 集市资源 id',
  `asset_name` varchar(191) DEFAULT NULL COMMENT '集市名称',
  `desc` varchar(1024) DEFAULT NULL COMMENT '描述信息',
  `logo` varchar(1024) DEFAULT NULL COMMENT 'logo 地址',
  `org_id` bigint(20) DEFAULT NULL COMMENT 'organization id',
  `project_id` bigint(20) DEFAULT NULL COMMENT '项目 id',
  `app_id` bigint(20) DEFAULT NULL COMMENT '应用 id',
  `creator_id` varchar(191) DEFAULT NULL COMMENT 'creator user id',
  `updater_id` varchar(191) DEFAULT NULL COMMENT 'updater user id',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  `public` tinyint(1) DEFAULT '0' COMMENT 'public',
  `cur_version_id` bigint(20) DEFAULT NULL COMMENT 'latest version id',
  `cur_major` int(11) DEFAULT NULL COMMENT 'latest version major',
  `cur_minor` int(11) DEFAULT NULL COMMENT 'latest version minor',
  `cur_patch` int(11) DEFAULT NULL COMMENT 'latest version patch',
  `project_name` varchar(191) DEFAULT NULL COMMENT 'project name',
  `app_name` varchar(191) DEFAULT NULL COMMENT 'app name',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=49 DEFAULT CHARSET=utf8mb4 COMMENT='API 集市资源表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `dice_api_clients`
--

DROP TABLE IF EXISTS `dice_api_clients`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `dice_api_clients` (
  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT COMMENT 'primary key',
  `org_id` bigint(20) DEFAULT NULL COMMENT 'organization id',
  `name` varchar(64) DEFAULT NULL COMMENT 'client name',
  `desc` varchar(1024) DEFAULT NULL COMMENT 'describe',
  `client_id` varchar(32) DEFAULT NULL COMMENT 'client id',
  `creator_id` varchar(191) DEFAULT NULL COMMENT 'creator user id',
  `updater_id` varchar(191) DEFAULT NULL COMMENT 'updater user id',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  `alias_name` varchar(64) DEFAULT NULL COMMENT 'alias name',
  `display_name` varchar(191) DEFAULT NULL COMMENT 'client display name',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=22 DEFAULT CHARSET=utf8mb4 COMMENT='API 集市资源访问管理表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `dice_api_contract_records`
--

DROP TABLE IF EXISTS `dice_api_contract_records`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `dice_api_contract_records` (
  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT COMMENT 'primary key',
  `org_id` bigint(20) DEFAULT NULL COMMENT 'organization id',
  `contract_id` bigint(20) DEFAULT NULL COMMENT 'dice_api_contracts primary key',
  `action` varchar(64) DEFAULT NULL COMMENT 'operation describe',
  `creator_id` varchar(191) DEFAULT NULL COMMENT 'operation user id',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=55 DEFAULT CHARSET=utf8mb4 COMMENT='API 集市资源访问管理合约操作记录表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `dice_api_contracts`
--

DROP TABLE IF EXISTS `dice_api_contracts`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `dice_api_contracts` (
  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT COMMENT 'primary key',
  `asset_id` varchar(191) DEFAULT NULL COMMENT 'asset id',
  `asset_name` varchar(191) DEFAULT NULL COMMENT 'asset name',
  `org_id` bigint(20) DEFAULT NULL COMMENT 'organization id',
  `swagger_version` varchar(16) DEFAULT NULL COMMENT 'swagger version',
  `client_id` bigint(20) DEFAULT NULL COMMENT 'primary key of table dice_api_client',
  `status` varchar(16) DEFAULT NULL COMMENT 'proved:已授权, proving:待审批, disproved:已撤销',
  `creator_id` varchar(191) DEFAULT NULL COMMENT 'creator user id',
  `updater_id` varchar(191) DEFAULT NULL COMMENT 'updater user id',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  `cur_sla_id` bigint(20) DEFAULT NULL COMMENT 'contract current SLA id',
  `request_sla_id` bigint(20) DEFAULT NULL COMMENT 'contract request SLA',
  `sla_committed_at` datetime DEFAULT NULL COMMENT 'current SLA committed time',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=51 DEFAULT CHARSET=utf8mb4 COMMENT='API 集市资源访问管理合约表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `dice_api_doc_lock`
--

DROP TABLE IF EXISTS `dice_api_doc_lock`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `dice_api_doc_lock` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT COMMENT 'primary key',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  `session_id` char(36) NOT NULL COMMENT '会话标识',
  `is_locked` tinyint(1) NOT NULL DEFAULT '0' COMMENT '会话所有者是否持有文档锁',
  `expired_at` datetime NOT NULL COMMENT '会话过期时间',
  `application_id` bigint(20) NOT NULL COMMENT '应用 id',
  `branch_name` varchar(191) NOT NULL COMMENT '分支名',
  `doc_name` varchar(191) NOT NULL COMMENT '文档名, 也即服务名',
  `creator_id` varchar(191) NOT NULL COMMENT '创建者 id',
  `updater_id` varchar(191) NOT NULL COMMENT '更新者 id',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_doc` (`application_id`,`branch_name`,`doc_name`)
) ENGINE=InnoDB AUTO_INCREMENT=25 DEFAULT CHARSET=utf8mb4 COMMENT='API 设计中心文档锁表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `dice_api_doc_tmp_content`
--

DROP TABLE IF EXISTS `dice_api_doc_tmp_content`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `dice_api_doc_tmp_content` (
  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT COMMENT 'primary key',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT 'created time',
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT 'updated time',
  `application_id` bigint(20) NOT NULL COMMENT '应用 id',
  `branch_name` varchar(191) NOT NULL COMMENT '分支名',
  `doc_name` varchar(64) NOT NULL COMMENT '文档名',
  `content` longtext NOT NULL COMMENT 'API doc text',
  `creator_id` varchar(191) NOT NULL COMMENT 'creator id',
  `updater_id` varchar(191) NOT NULL COMMENT 'updater id',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_inode` (`application_id`,`branch_name`,`doc_name`)
) ENGINE=InnoDB AUTO_INCREMENT=10 DEFAULT CHARSET=utf8mb4 COMMENT='API 设计中心文档临时存储表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `dice_api_oas3_fragment`
--

DROP TABLE IF EXISTS `dice_api_oas3_fragment`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `dice_api_oas3_fragment` (
  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT COMMENT 'primary key',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT 'created time',
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT 'updated time',
  `index_id` bigint(20) NOT NULL COMMENT 'dice_api_oas3_index primary key',
  `version_id` bigint(20) NOT NULL COMMENT 'asset version primary key',
  `operation` text NOT NULL COMMENT '.paths.{path}.{method}.parameters, 序列化了的 parameters JSON 片段',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=9991 DEFAULT CHARSET=utf8mb4 COMMENT='API 集市 oas3 片段表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `dice_api_oas3_index`
--

DROP TABLE IF EXISTS `dice_api_oas3_index`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `dice_api_oas3_index` (
  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT COMMENT 'primary key',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT 'created time',
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT 'updated time',
  `asset_id` varchar(191) NOT NULL COMMENT 'asset id',
  `asset_name` varchar(191) NOT NULL COMMENT 'asset name',
  `info_version` varchar(191) NOT NULL COMMENT '.info.version value, 也即 swaggerVersion',
  `version_id` bigint(20) NOT NULL COMMENT 'asset version primary key',
  `path` varchar(191) NOT NULL COMMENT '.paths.{path}',
  `method` varchar(16) NOT NULL COMMENT '.paths.{path}.{method}',
  `operation_id` varchar(191) NOT NULL COMMENT '.paths.{path}.{method}.operationId',
  `description` text NOT NULL COMMENT '.path.{path}.{method}.description',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_path_method` (`version_id`,`path`,`method`) COMMENT '同一文档下, path + method 确定一个接口'
) ENGINE=InnoDB AUTO_INCREMENT=9991 DEFAULT CHARSET=utf8mb4 COMMENT='API 集市 operation 搜索索引表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `dice_api_sla_limits`
--

DROP TABLE IF EXISTS `dice_api_sla_limits`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `dice_api_sla_limits` (
  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT COMMENT 'primary key',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  `creator_id` varchar(191) DEFAULT NULL COMMENT 'creator id',
  `updater_id` varchar(191) DEFAULT NULL COMMENT 'creator id',
  `sla_id` bigint(20) DEFAULT NULL COMMENT 'SLA model id',
  `limit` bigint(20) DEFAULT NULL COMMENT 'request limit',
  `unit` varchar(16) DEFAULT NULL COMMENT 's: second, m: minute, h: hour, d: day',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='API 集市访问管理 SLA 限制条件表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `dice_api_slas`
--

DROP TABLE IF EXISTS `dice_api_slas`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `dice_api_slas` (
  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT COMMENT 'primary key',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  `creator_id` varchar(191) DEFAULT NULL COMMENT 'creator id',
  `updater_id` varchar(191) DEFAULT NULL COMMENT 'creator id',
  `name` varchar(191) DEFAULT NULL COMMENT 'SLA name',
  `desc` varchar(1024) DEFAULT NULL COMMENT 'description',
  `approval` varchar(16) DEFAULT NULL COMMENT 'auto, manual',
  `access_id` bigint(20) DEFAULT NULL COMMENT 'access id',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='API 集市访问管理 Service Level Agreements 表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `dice_api_test`
--

DROP TABLE IF EXISTS `dice_api_test`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `dice_api_test` (
  `id` int(11) unsigned NOT NULL AUTO_INCREMENT COMMENT '主键',
  `created_at` datetime DEFAULT NULL COMMENT '创建时间',
  `updated_at` datetime DEFAULT NULL COMMENT '更新时间',
  `usecase_id` int(11) DEFAULT NULL COMMENT '所属用例 ID',
  `usecase_order` int(11) DEFAULT NULL COMMENT '接口顺序',
  `status` varchar(16) DEFAULT NULL COMMENT '接口执行状态',
  `api_info` text COMMENT 'API 信息',
  `api_request` longtext COMMENT 'API 请求体',
  `api_response` longtext COMMENT 'API 响应',
  `assert_result` text COMMENT '断言接口',
  `project_id` int(11) DEFAULT NULL COMMENT '项目 ID',
  `pipeline_id` int(11) DEFAULT NULL COMMENT '关联的流水线 ID',
  PRIMARY KEY (`id`),
  KEY `idx_projectid_usercase` (`project_id`,`usecase_id`,`usecase_order`)
) ENGINE=InnoDB AUTO_INCREMENT=7093 DEFAULT CHARSET=utf8mb4 BLOCK_FORMAT=ENCRYPTED COMMENT='手动测试-接口信息表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `dice_api_test_env`
--

DROP TABLE IF EXISTS `dice_api_test_env`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `dice_api_test_env` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT COMMENT '主键',
  `created_at` timestamp NULL DEFAULT NULL COMMENT '创建时间',
  `updated_at` timestamp NULL DEFAULT NULL COMMENT '更新时间',
  `env_id` int(11) NOT NULL COMMENT '环境 ID',
  `env_type` varchar(64) DEFAULT NULL COMMENT '环境类型，分为项目级和用例级',
  `name` varchar(255) DEFAULT NULL COMMENT '配置名',
  `domain` varchar(255) DEFAULT NULL COMMENT '域名',
  `header` text COMMENT '公共请求头',
  `global` text COMMENT '全局变量配置',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=36 DEFAULT CHARSET=utf8mb4 BLOCK_FORMAT=ENCRYPTED COMMENT='手动测试-接口测试-环境配置表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `dice_app`
--

DROP TABLE IF EXISTS `dice_app`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `dice_app` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT,
  `created_at` timestamp NULL DEFAULT NULL,
  `updated_at` timestamp NULL DEFAULT NULL,
  `name` varchar(255) DEFAULT NULL,
  `desc` varchar(255) DEFAULT NULL,
  `config` varchar(255) DEFAULT NULL,
  `project_id` bigint(20) DEFAULT NULL,
  `project_name` varchar(255) DEFAULT NULL,
  `org_id` bigint(20) DEFAULT NULL,
  `mode` varchar(255) DEFAULT NULL,
  `git_repo` varchar(255) DEFAULT NULL,
  `git_repo_abbrev` varchar(255) DEFAULT NULL,
  `logo` varchar(255) DEFAULT NULL,
  `creator` varchar(255) DEFAULT NULL,
  `extra` varchar(255) DEFAULT NULL,
  `is_external_repo` tinyint(1) DEFAULT '0',
  `repo_config` text,
  `display_name` varchar(64) DEFAULT NULL COMMENT '应用展示名称',
  `unblock_start` timestamp NULL DEFAULT NULL COMMENT '解封开始时间',
  `unblock_end` timestamp NULL DEFAULT NULL COMMENT '解封结束时间',
  `is_public` tinyint(1) NOT NULL DEFAULT '0' COMMENT '是否公开',
  PRIMARY KEY (`id`),
  KEY `idx_project_id` (`project_id`)
) ENGINE=InnoDB AUTO_INCREMENT=117 DEFAULT CHARSET=utf8mb4 BLOCK_FORMAT=ENCRYPTED;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `dice_app_certificates`
--

DROP TABLE IF EXISTS `dice_app_certificates`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `dice_app_certificates` (
  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT COMMENT '主键',
  `app_id` bigint(20) NOT NULL COMMENT '所属应用',
  `certificate_id` bigint(20) NOT NULL COMMENT '证书',
  `status` varchar(45) NOT NULL DEFAULT '' COMMENT '证书审批状态',
  `operator` varchar(255) DEFAULT NULL COMMENT '操作者',
  `push_config` text COMMENT '证书推送信息',
  `created_at` datetime NOT NULL COMMENT '创建时间',
  `updated_at` datetime DEFAULT NULL COMMENT '更新时间',
  `approval_id` bigint(20) NOT NULL COMMENT '审批ID',
  PRIMARY KEY (`id`),
  KEY `certificate_id` (`certificate_id`),
  KEY `app_id` (`app_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='应用证书信息';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `dice_app_publish_item_relation`
--

DROP TABLE IF EXISTS `dice_app_publish_item_relation`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `dice_app_publish_item_relation` (
  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT COMMENT '主键',
  `app_id` bigint(20) NOT NULL COMMENT '应用ID',
  `publish_item_id` bigint(20) NOT NULL COMMENT '发布内容ID',
  `env` varchar(100) NOT NULL DEFAULT '' COMMENT '环境',
  `creator` varchar(255) NOT NULL DEFAULT '' COMMENT '创建用户',
  `created_at` datetime NOT NULL COMMENT '创建时间',
  `updated_at` datetime DEFAULT NULL COMMENT '更新时间',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='应用发布关联表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `dice_approves`
--

DROP TABLE IF EXISTS `dice_approves`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `dice_approves` (
  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT COMMENT '主键',
  `org_id` bigint(20) NOT NULL COMMENT '所属企业',
  `title` varchar(255) NOT NULL DEFAULT '' COMMENT '审批标题',
  `target_id` bigint(20) NOT NULL COMMENT '审批对象',
  `entity_id` bigint(20) NOT NULL COMMENT '审批实体',
  `target_name` varchar(255) NOT NULL DEFAULT '' COMMENT '审批对象名字',
  `extra` text COMMENT '其它字段',
  `status` varchar(45) NOT NULL DEFAULT '' COMMENT '审批状态',
  `priority` varchar(45) NOT NULL DEFAULT '' COMMENT '审批优先级',
  `type` varchar(64) NOT NULL DEFAULT '' COMMENT '审批类型',
  `desc` varchar(2048) DEFAULT NULL COMMENT '审批描述',
  `approval_time` datetime NOT NULL COMMENT '审批时间',
  `approver` varchar(64) DEFAULT NULL COMMENT '审批人',
  `submitter` varchar(64) DEFAULT NULL COMMENT '提交人',
  `created_at` datetime NOT NULL COMMENT '创建时间',
  `updated_at` datetime DEFAULT NULL COMMENT '更新时间',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=160 DEFAULT CHARSET=utf8mb4 COMMENT='审批信息';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `dice_audit`
--

DROP TABLE IF EXISTS `dice_audit`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `dice_audit` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT,
  `created_at` datetime DEFAULT NULL COMMENT '表记录创建时间',
  `updated_at` datetime DEFAULT NULL COMMENT '表记录更新时间',
  `start_time` datetime NOT NULL COMMENT '事件发生的时间',
  `end_time` datetime NOT NULL COMMENT '事件结束的时间',
  `user_id` varchar(40) NOT NULL COMMENT '事件的操作人',
  `scope_type` varchar(40) NOT NULL COMMENT '事件发生的scope类型',
  `scope_id` varchar(40) NOT NULL COMMENT '事件发生的scope类型',
  `app_id` bigint(20) DEFAULT NULL COMMENT '事件发生的上下文信息，appId，用于前端渲染',
  `project_id` bigint(20) DEFAULT NULL COMMENT '事件发生的上下文信息，projectId，用于前端渲染',
  `org_id` bigint(20) DEFAULT NULL COMMENT '事件发生的上下文信息，orgId',
  `context` text COMMENT '事件发生的自定义上下文信息，用于前端渲染',
  `template_name` varchar(40) NOT NULL COMMENT '前端渲染事件模版的key',
  `audit_level` varchar(40) DEFAULT NULL COMMENT '事件的等级',
  `result` varchar(40) DEFAULT NULL COMMENT '事件的结果',
  `error_msg` text COMMENT '事件的结果为失败时的错误信息',
  `client_ip` varchar(40) DEFAULT NULL COMMENT '事件的客户端地址',
  `user_agent` text COMMENT '事件的客户端类型',
  `deleted` varchar(40) DEFAULT '0' COMMENT '事件进入归档表前的软删除标记',
  `fdp_project_id` varchar(128) DEFAULT NULL COMMENT 'fdp项目id',
  PRIMARY KEY (`id`),
  KEY `start_time` (`start_time`),
  KEY `end_time` (`end_time`),
  KEY `org_id` (`org_id`),
  KEY `user_id` (`user_id`)
) ENGINE=InnoDB AUTO_INCREMENT=161283 DEFAULT CHARSET=utf8mb4 COMMENT='审计事件';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `dice_audit_history`
--

DROP TABLE IF EXISTS `dice_audit_history`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `dice_audit_history` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT,
  `created_at` datetime DEFAULT NULL COMMENT '表记录创建时间',
  `updated_at` datetime DEFAULT NULL COMMENT '表记录更新时间',
  `start_time` datetime NOT NULL COMMENT '事件发生的时间',
  `end_time` datetime NOT NULL COMMENT '事件结束的时间',
  `user_id` varchar(40) NOT NULL COMMENT '事件的操作人',
  `scope_type` varchar(40) NOT NULL COMMENT '事件发生的scope类型',
  `scope_id` varchar(40) NOT NULL COMMENT '事件发生的scope类型',
  `app_id` bigint(20) DEFAULT NULL COMMENT '事件发生的上下文信息，appId，用于前端渲染',
  `project_id` bigint(20) DEFAULT NULL COMMENT '事件发生的上下文信息，projectId，用于前端渲染',
  `org_id` bigint(20) DEFAULT NULL COMMENT '事件发生的上下文信息，orgId',
  `context` text COMMENT '事件发生的自定义上下文信息，用于前端渲染',
  `template_name` varchar(40) NOT NULL COMMENT '前端渲染事件模版的key',
  `audit_level` varchar(40) DEFAULT NULL COMMENT '事件的等级',
  `result` varchar(40) DEFAULT NULL COMMENT '事件的结果',
  `error_msg` text COMMENT '事件的结果为失败时的错误信息',
  `client_ip` varchar(40) DEFAULT NULL COMMENT '事件的客户端地址',
  `user_agent` tinytext COMMENT '事件的客户端类型',
  `deleted` varchar(40) DEFAULT '0' COMMENT '事件进入归档表前的软删除标记',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=95799 DEFAULT CHARSET=utf8mb4 COMMENT='审计事件历史';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `dice_autotest_filetree_nodes`
--

DROP TABLE IF EXISTS `dice_autotest_filetree_nodes`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `dice_autotest_filetree_nodes` (
  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT COMMENT '主键',
  `type` varchar(1) NOT NULL COMMENT '节点类型, f: 文件, d: 目录',
  `scope` varchar(191) NOT NULL COMMENT 'scope，例如 project-autotest, project-autotest-testplan',
  `scope_id` varchar(191) NOT NULL COMMENT 'scope 的具体 ID，例如 项目 ID，测试计划 ID',
  `pinode` bigint(20) NOT NULL COMMENT '父节点 inode',
  `inode` bigint(20) NOT NULL COMMENT 'inode',
  `name` varchar(191) NOT NULL COMMENT '节点名',
  `desc` varchar(512) DEFAULT NULL COMMENT '描述',
  `creator_id` varchar(191) DEFAULT NULL COMMENT '创建人',
  `updater_id` varchar(191) DEFAULT NULL COMMENT '更新人',
  `created_at` datetime NOT NULL COMMENT '创建时间',
  `updated_at` datetime NOT NULL COMMENT '更新时间',
  PRIMARY KEY (`id`),
  KEY `idx_inode` (`inode`),
  KEY `idx_pinode` (`pinode`),
  KEY `idx_type_scope_pinode_inode` (`type`,`scope`,`scope_id`,`pinode`,`inode`),
  KEY `idx_scope_pinode_inode` (`scope`,`scope_id`,`pinode`,`inode`),
  KEY `idx_pinode_inode` (`pinode`,`inode`),
  KEY `idx_pinode_name` (`pinode`,`name`)
) ENGINE=InnoDB AUTO_INCREMENT=134 DEFAULT CHARSET=utf8mb4 COMMENT='自动化测试目录树节点表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `dice_autotest_filetree_nodes_histories`
--

DROP TABLE IF EXISTS `dice_autotest_filetree_nodes_histories`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `dice_autotest_filetree_nodes_histories` (
  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT,
  `inode` bigint(20) NOT NULL COMMENT '节点的node id',
  `pinode` bigint(20) NOT NULL COMMENT '父节点的 node id',
  `pipeline_yml` mediumtext NOT NULL COMMENT '节点的yml标识',
  `snippet_action` mediumtext NOT NULL COMMENT 'snippet config 配置',
  `name` varchar(191) NOT NULL COMMENT '名称',
  `desc` varchar(512) NOT NULL COMMENT '描述',
  `creator_id` varchar(191) NOT NULL COMMENT '创建人',
  `updater_id` varchar(191) NOT NULL COMMENT '更新人',
  `extra` mediumtext NOT NULL COMMENT '其他信息',
  `created_at` datetime NOT NULL COMMENT '创建时间',
  `updated_at` datetime NOT NULL COMMENT '更新时间',
  PRIMARY KEY (`id`),
  KEY `idx_inode` (`inode`)
) ENGINE=InnoDB AUTO_INCREMENT=630 DEFAULT CHARSET=utf8mb4 COMMENT='自动化测试目录树节点历史表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `dice_autotest_filetree_nodes_meta`
--

DROP TABLE IF EXISTS `dice_autotest_filetree_nodes_meta`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `dice_autotest_filetree_nodes_meta` (
  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT,
  `inode` bigint(20) NOT NULL,
  `pipeline_yml` mediumtext,
  `snippet_action` mediumtext,
  `extra` mediumtext,
  `created_at` datetime NOT NULL,
  `updated_at` datetime NOT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_inode` (`inode`)
) ENGINE=InnoDB AUTO_INCREMENT=107 DEFAULT CHARSET=utf8mb4 COMMENT='自动化测试目录树节点元信息表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `dice_autotest_plan`
--

DROP TABLE IF EXISTS `dice_autotest_plan`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `dice_autotest_plan` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT COMMENT '主键',
  `created_at` timestamp NULL DEFAULT NULL COMMENT '创建时间',
  `updated_at` timestamp NULL DEFAULT NULL COMMENT '更新时间',
  `name` varchar(191) DEFAULT NULL COMMENT '测试计划名称',
  `desc` varchar(512) DEFAULT NULL COMMENT '测试计划描述',
  `creator_id` varchar(191) DEFAULT NULL COMMENT '创建人',
  `updater_id` varchar(191) DEFAULT NULL COMMENT '更新人',
  `space_id` bigint(20) NOT NULL COMMENT '测试空间id',
  `project_id` bigint(20) DEFAULT NULL COMMENT '项目id',
  PRIMARY KEY (`id`),
  KEY `idx_name` (`name`),
  KEY `idx_project` (`project_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='测试计划表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `dice_autotest_plan_members`
--

DROP TABLE IF EXISTS `dice_autotest_plan_members`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `dice_autotest_plan_members` (
  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT COMMENT '主键',
  `test_plan_id` bigint(20) DEFAULT NULL COMMENT '测试计划id',
  `role` varchar(32) DEFAULT NULL COMMENT '角色',
  `user_id` bigint(20) DEFAULT NULL COMMENT '用户id',
  `created_at` datetime DEFAULT NULL COMMENT '创建时间',
  `updated_at` datetime DEFAULT NULL COMMENT '更新时间',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='自动测试-测试计划成员表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `dice_autotest_plan_step`
--

DROP TABLE IF EXISTS `dice_autotest_plan_step`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `dice_autotest_plan_step` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT COMMENT '主键',
  `created_at` timestamp NULL DEFAULT NULL COMMENT '创建时间',
  `updated_at` timestamp NULL DEFAULT NULL COMMENT '更新时间',
  `plan_id` bigint(20) NOT NULL COMMENT '测试计划id',
  `scene_set_id` bigint(20) NOT NULL COMMENT '场景集id',
  `pre_id` bigint(20) NOT NULL COMMENT '前节点',
  PRIMARY KEY (`id`),
  KEY `idx_plan_id` (`plan_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='测试计划步骤表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `dice_autotest_scene`
--

DROP TABLE IF EXISTS `dice_autotest_scene`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `dice_autotest_scene` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT COMMENT '主键',
  `created_at` datetime NOT NULL COMMENT '创建时间',
  `updated_at` datetime NOT NULL COMMENT '更新时间',
  `name` varchar(191) NOT NULL COMMENT '名称',
  `description` text NOT NULL COMMENT '描述',
  `space_id` bigint(20) NOT NULL COMMENT '测试空间id',
  `set_id` bigint(20) NOT NULL COMMENT '场景集id',
  `pre_id` bigint(20) NOT NULL COMMENT '前节点',
  `creator_id` varchar(255) NOT NULL COMMENT '创建人',
  `updater_id` varchar(255) NOT NULL COMMENT '更新人',
  `status` varchar(255) DEFAULT NULL COMMENT '执行状态',
  `ref_set_id` bigint(20) NOT NULL DEFAULT '0' COMMENT '引用场景集的id',
  PRIMARY KEY (`id`),
  KEY `idx_set_id` (`set_id`),
  KEY `idx_name` (`name`)
) ENGINE=InnoDB AUTO_INCREMENT=74 DEFAULT CHARSET=utf8mb4 COMMENT='自动化测试场景表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `dice_autotest_scene_input`
--

DROP TABLE IF EXISTS `dice_autotest_scene_input`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `dice_autotest_scene_input` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT COMMENT '主键',
  `created_at` datetime NOT NULL COMMENT '创建时间',
  `updated_at` datetime NOT NULL COMMENT '更新时间',
  `name` varchar(255) NOT NULL COMMENT '名称',
  `value` text NOT NULL COMMENT '默认值',
  `temp` text NOT NULL COMMENT '当前值',
  `description` text NOT NULL COMMENT '描述',
  `scene_id` bigint(20) NOT NULL COMMENT '场景id',
  `space_id` bigint(20) NOT NULL COMMENT '空间id',
  `creator_id` varchar(255) NOT NULL COMMENT '创建人',
  `updater_id` varchar(255) NOT NULL COMMENT '更新人',
  PRIMARY KEY (`id`),
  KEY `idx_scene_id` (`scene_id`)
) ENGINE=InnoDB AUTO_INCREMENT=99 DEFAULT CHARSET=utf8mb4 COMMENT='自动化测试场景入参表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `dice_autotest_scene_output`
--

DROP TABLE IF EXISTS `dice_autotest_scene_output`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `dice_autotest_scene_output` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT COMMENT '主键',
  `created_at` datetime NOT NULL COMMENT '创建时间',
  `updated_at` datetime NOT NULL COMMENT '更新时间',
  `name` varchar(255) NOT NULL COMMENT '名称',
  `value` text NOT NULL COMMENT '值表达式',
  `description` text NOT NULL COMMENT '描述',
  `scene_id` bigint(20) NOT NULL COMMENT '场景id',
  `space_id` bigint(20) NOT NULL COMMENT '空间id',
  `creator_id` varchar(255) NOT NULL COMMENT '创建人',
  `updater_id` varchar(255) NOT NULL COMMENT '更新人',
  PRIMARY KEY (`id`),
  KEY `idx_scene_id` (`scene_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='自动化测试场景出参表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `dice_autotest_scene_set`
--

DROP TABLE IF EXISTS `dice_autotest_scene_set`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `dice_autotest_scene_set` (
  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT COMMENT '主键',
  `name` varchar(191) DEFAULT NULL,
  `description` varchar(255) DEFAULT NULL COMMENT '场景集描述',
  `space_id` bigint(20) unsigned NOT NULL COMMENT '测试空间id',
  `pre_id` bigint(20) unsigned DEFAULT NULL COMMENT '上一个节点id',
  `creator_id` varchar(255) NOT NULL COMMENT '创建人',
  `updater_id` varchar(255) NOT NULL COMMENT '更新人',
  `created_at` datetime DEFAULT NULL COMMENT '创建时间',
  `updated_at` datetime DEFAULT NULL COMMENT '更新时间',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=12 DEFAULT CHARSET=utf8mb4 COMMENT='场景集表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `dice_autotest_scene_step`
--

DROP TABLE IF EXISTS `dice_autotest_scene_step`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `dice_autotest_scene_step` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT COMMENT '主键',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '更新时间',
  `type` varchar(255) NOT NULL COMMENT '类型',
  `value` mediumtext,
  `name` varchar(255) NOT NULL COMMENT '名称',
  `pre_id` bigint(20) NOT NULL COMMENT '前节点',
  `scene_id` bigint(20) NOT NULL COMMENT '场景id',
  `space_id` bigint(20) NOT NULL COMMENT '空间id',
  `creator_id` varchar(255) NOT NULL COMMENT '创建人',
  `updater_id` varchar(255) NOT NULL COMMENT '更新人',
  `pre_type` varchar(255) NOT NULL COMMENT '排序类型',
  `api_spec_id` varchar(50) DEFAULT NULL COMMENT 'api集市id',
  PRIMARY KEY (`id`),
  KEY `idx_scene_id` (`scene_id`)
) ENGINE=InnoDB AUTO_INCREMENT=135 DEFAULT CHARSET=utf8mb4 COMMENT='自动化测试场景步骤表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `dice_autotest_space`
--

DROP TABLE IF EXISTS `dice_autotest_space`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `dice_autotest_space` (
  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT COMMENT '主键',
  `name` varchar(255) NOT NULL COMMENT '测试空间名称',
  `project_id` bigint(20) NOT NULL COMMENT '项目id',
  `description` varchar(1024) DEFAULT NULL,
  `creator_id` varchar(255) NOT NULL COMMENT '创建人',
  `created_at` datetime NOT NULL COMMENT '创建时间',
  `updated_at` datetime NOT NULL COMMENT '更新时间',
  `deleted_at` datetime DEFAULT NULL COMMENT '删除时间',
  `source_space_id` bigint(20) DEFAULT NULL COMMENT '被复制的源测试空间',
  `status` varchar(255) NOT NULL COMMENT '测试空间状态',
  `updater_id` varchar(255) NOT NULL COMMENT '更新人',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=3 DEFAULT CHARSET=utf8mb4 COMMENT='测试空间表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `dice_branch_rules`
--

DROP TABLE IF EXISTS `dice_branch_rules`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `dice_branch_rules` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT,
  `created_at` timestamp NULL DEFAULT NULL,
  `updated_at` timestamp NULL DEFAULT NULL,
  `rule` varchar(150) DEFAULT NULL,
  `desc` varchar(150) DEFAULT NULL,
  `workspace` varchar(150) DEFAULT NULL,
  `artifact_workspace` varchar(150) DEFAULT NULL,
  `is_protect` tinyint(1) DEFAULT NULL,
  `is_trigger_pipeline` tinyint(1) DEFAULT NULL,
  `need_approval` tinyint(1) DEFAULT NULL,
  `scope_type` varchar(50) DEFAULT NULL,
  `scope_id` bigint(20) DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=103 DEFAULT CHARSET=utf8mb4 COMMENT='分支规则';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `dice_certificates`
--

DROP TABLE IF EXISTS `dice_certificates`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `dice_certificates` (
  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT COMMENT '主键',
  `org_id` bigint(20) NOT NULL COMMENT '所属企业',
  `name` varchar(255) NOT NULL DEFAULT '' COMMENT '证书自定义名称',
  `message` text,
  `ios` text,
  `android` text,
  `status` varchar(45) NOT NULL DEFAULT '' COMMENT '企业审批状态',
  `type` varchar(64) NOT NULL DEFAULT '' COMMENT '企业类型',
  `desc` varchar(2048) DEFAULT NULL COMMENT 'publisher描述',
  `creator` varchar(255) NOT NULL DEFAULT '' COMMENT '创建用户',
  `operator` varchar(255) DEFAULT NULL COMMENT '操作者',
  `created_at` datetime NOT NULL COMMENT '创建时间',
  `updated_at` datetime DEFAULT NULL COMMENT '更新时间',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='证书信息';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `dice_cloud_accounts`
--

DROP TABLE IF EXISTS `dice_cloud_accounts`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `dice_cloud_accounts` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT,
  `created_at` timestamp NULL DEFAULT NULL,
  `updated_at` timestamp NULL DEFAULT NULL,
  `cloud_provider` varchar(255) DEFAULT NULL,
  `name` varchar(255) DEFAULT NULL,
  `access_key_id` varchar(255) DEFAULT NULL,
  `access_key_secret` varchar(255) DEFAULT NULL,
  `org_id` bigint(20) DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 BLOCK_FORMAT=ENCRYPTED COMMENT='云账号信息，老表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `dice_config_item`
--

DROP TABLE IF EXISTS `dice_config_item`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `dice_config_item` (
  `id` int(10) unsigned NOT NULL AUTO_INCREMENT COMMENT '配置项ID',
  `namespace_id` int(10) unsigned NOT NULL DEFAULT '0' COMMENT '配置命名空间ID',
  `item_key` varchar(128) NOT NULL DEFAULT 'default' COMMENT '配置项Key',
  `item_value` longtext NOT NULL COMMENT '配置项值',
  `item_comment` varchar(1024) DEFAULT '' COMMENT '注释',
  `is_sync` tinyint(1) DEFAULT '0' COMMENT '是否同步到配置中心',
  `delete_remote` tinyint(1) DEFAULT '0' COMMENT '是否删除远程配置',
  `create_time` datetime NOT NULL DEFAULT '1970-01-01 00:00:00' COMMENT '创建时间',
  `update_time` datetime NOT NULL DEFAULT '1970-01-01 00:00:00' COMMENT '修改时间',
  `is_deleted` varchar(1) NOT NULL DEFAULT 'N' COMMENT '逻辑删除',
  `status` varchar(32) DEFAULT 'PUBLISHED' COMMENT '配置状态',
  `source` varchar(32) DEFAULT 'DEPLOY_WEB' COMMENT '配置来源',
  `dynamic` tinyint(1) DEFAULT '1' COMMENT '是否为动态配置',
  `encrypt` tinyint(1) DEFAULT '0' COMMENT '配置项是否加密',
  `item_type` varchar(32) DEFAULT 'ENV' COMMENT '配置类型',
  PRIMARY KEY (`id`),
  KEY `idx_namespaceid` (`namespace_id`)
) ENGINE=InnoDB AUTO_INCREMENT=3549 DEFAULT CHARSET=utf8 BLOCK_FORMAT=ENCRYPTED COMMENT='配置项';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `dice_config_namespace`
--

DROP TABLE IF EXISTS `dice_config_namespace`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `dice_config_namespace` (
  `id` int(10) unsigned NOT NULL AUTO_INCREMENT COMMENT '自增主键',
  `name` varchar(255) NOT NULL COMMENT '配置命名空间名称',
  `dynamic` tinyint(1) DEFAULT '1' COMMENT '存储配置是否为动态配置',
  `is_default` tinyint(1) DEFAULT '0' COMMENT '是否为默认命名空间',
  `project_id` varchar(45) NOT NULL COMMENT '项目ID',
  `env` varchar(45) DEFAULT NULL COMMENT '所属部署环境',
  `application_id` varchar(45) DEFAULT NULL COMMENT '应用ID',
  `runtime_id` varchar(45) DEFAULT NULL COMMENT 'runtime ID',
  `create_time` datetime NOT NULL DEFAULT '1970-01-01 00:00:00' COMMENT '创建时间',
  `update_time` datetime NOT NULL DEFAULT '1970-01-01 00:00:00' COMMENT '修改时间',
  `is_deleted` varchar(1) NOT NULL DEFAULT 'N' COMMENT '逻辑删除',
  PRIMARY KEY (`id`),
  KEY `idx_name` (`name`)
) ENGINE=InnoDB AUTO_INCREMENT=1076 DEFAULT CHARSET=utf8 BLOCK_FORMAT=ENCRYPTED COMMENT='配置项namespace';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `dice_config_namespace_relation`
--

DROP TABLE IF EXISTS `dice_config_namespace_relation`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `dice_config_namespace_relation` (
  `id` int(10) unsigned NOT NULL AUTO_INCREMENT COMMENT '自增主键',
  `namespace` varchar(255) NOT NULL COMMENT '配置命名空间名称',
  `default_namespace` varchar(255) NOT NULL COMMENT '默认配置命名空间名称',
  `create_time` datetime NOT NULL DEFAULT '1970-01-01 00:00:00' COMMENT '创建时间',
  `update_time` datetime NOT NULL DEFAULT '1970-01-01 00:00:00' COMMENT '修改时间',
  `is_deleted` varchar(1) NOT NULL DEFAULT 'N' COMMENT '逻辑删除',
  PRIMARY KEY (`id`),
  UNIQUE KEY `namespace` (`namespace`),
  KEY `idx_default_namespace` (`default_namespace`)
) ENGINE=InnoDB AUTO_INCREMENT=465 DEFAULT CHARSET=utf8 BLOCK_FORMAT=ENCRYPTED COMMENT='配置项namespace关联表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `dice_db_migration_log`
--

DROP TABLE IF EXISTS `dice_db_migration_log`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `dice_db_migration_log` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT COMMENT '自增id',
  `project_id` bigint(20) NOT NULL COMMENT '项目id',
  `application_id` bigint(20) NOT NULL COMMENT '应用id',
  `runtime_id` bigint(20) NOT NULL COMMENT 'runtime id',
  `deployment_id` bigint(20) NOT NULL COMMENT 'deployment id',
  `release_id` varchar(64) NOT NULL DEFAULT '' COMMENT 'release id',
  `operator_id` bigint(20) NOT NULL COMMENT '执行人',
  `status` varchar(128) NOT NULL COMMENT '执行结果状态',
  `addon_instance_id` varchar(64) NOT NULL COMMENT '所要执行migration的addon实例Id',
  `addon_instance_config` varchar(4096) DEFAULT NULL COMMENT '需要使用的config',
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '修改时间',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8 COMMENT='migration执行日志记录表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `dice_error_box`
--

DROP TABLE IF EXISTS `dice_error_box`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `dice_error_box` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT,
  `created_at` datetime DEFAULT NULL COMMENT '表记录创建时间',
  `updated_at` datetime DEFAULT NULL COMMENT '表记录更新时间',
  `resource_type` varchar(40) NOT NULL COMMENT '资源类型',
  `resource_id` varchar(40) NOT NULL COMMENT '资源id',
  `occurrence_time` datetime NOT NULL COMMENT '日志发生时间',
  `human_log` text COMMENT '处理过的日志和提示',
  `primeval_log` text NOT NULL COMMENT '原生错误',
  `dedup_id` varchar(190) NOT NULL COMMENT '去重id',
  `level` varchar(50) DEFAULT 'error' COMMENT '日志级别',
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_rtype_rid_did` (`resource_type`,`resource_id`,`dedup_id`)
) ENGINE=InnoDB AUTO_INCREMENT=101736 DEFAULT CHARSET=utf8mb4 COMMENT='错误信息透出记录';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `dice_extension`
--

DROP TABLE IF EXISTS `dice_extension`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `dice_extension` (
  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT,
  `created_at` timestamp NULL DEFAULT NULL,
  `updated_at` timestamp NULL DEFAULT NULL,
  `type` varchar(128) DEFAULT NULL,
  `name` varchar(255) DEFAULT NULL,
  `category` varchar(255) DEFAULT NULL,
  `display_name` varchar(255) DEFAULT NULL,
  `logo_url` varchar(255) DEFAULT NULL,
  `desc` varchar(255) DEFAULT NULL,
  `public` tinyint(1) DEFAULT NULL,
  `labels` varchar(200) DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=185 DEFAULT CHARSET=utf8 BLOCK_FORMAT=ENCRYPTED COMMENT='action,addon扩展信息';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `dice_extension_version`
--

DROP TABLE IF EXISTS `dice_extension_version`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `dice_extension_version` (
  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT,
  `created_at` timestamp NULL DEFAULT NULL,
  `updated_at` timestamp NULL DEFAULT NULL,
  `extension_id` bigint(20) unsigned DEFAULT NULL,
  `name` varchar(128) DEFAULT NULL,
  `version` varchar(128) DEFAULT NULL,
  `spec` text,
  `dice` text,
  `readme` longtext,
  `public` tinyint(1) DEFAULT NULL,
  `is_default` tinyint(1) DEFAULT NULL,
  `swagger` longtext,
  PRIMARY KEY (`id`),
  KEY `idx_name` (`name`),
  KEY `idx_version` (`version`)
) ENGINE=InnoDB AUTO_INCREMENT=426 DEFAULT CHARSET=utf8 BLOCK_FORMAT=ENCRYPTED COMMENT='action,addon扩展版本信息';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `dice_files`
--

DROP TABLE IF EXISTS `dice_files`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `dice_files` (
  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT,
  `uuid` varchar(32) NOT NULL DEFAULT '',
  `display_name` varchar(1024) NOT NULL DEFAULT '',
  `ext` varchar(32) DEFAULT '',
  `byte_size` bigint(20) NOT NULL,
  `storage_type` varchar(32) NOT NULL DEFAULT '',
  `full_relative_path` varchar(2048) NOT NULL DEFAULT '',
  `from` varchar(64) DEFAULT NULL,
  `creator` varchar(255) DEFAULT NULL,
  `extra` varchar(2048) DEFAULT NULL,
  `created_at` datetime DEFAULT NULL,
  `updated_at` datetime DEFAULT NULL,
  `expired_at` datetime DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_uuid` (`uuid`),
  KEY `idx_storageType` (`storage_type`)
) ENGINE=InnoDB AUTO_INCREMENT=185 DEFAULT CHARSET=utf8mb4 COMMENT='Dice 文件表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `dice_init_sql_version`
--

DROP TABLE IF EXISTS `dice_init_sql_version`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `dice_init_sql_version` (
  `id` int(11) unsigned NOT NULL AUTO_INCREMENT COMMENT 'primary key',
  `version` varchar(64) DEFAULT NULL COMMENT 'init sql初始化版本',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='DICE init sql 初始化版本';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `dice_issue_app_relations`
--

DROP TABLE IF EXISTS `dice_issue_app_relations`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `dice_issue_app_relations` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT COMMENT '主键',
  `created_at` datetime DEFAULT NULL COMMENT '表记录创建时间',
  `updated_at` datetime DEFAULT NULL COMMENT '表记录更新时间',
  `issue_id` bigint(20) NOT NULL COMMENT '关联关系源id eg:issue_id',
  `comment_id` bigint(20) NOT NULL COMMENT 'MR评论 id',
  `app_id` bigint(20) NOT NULL COMMENT '应用 id',
  `mr_id` bigint(20) NOT NULL COMMENT '关联关系目标id eg:mr_id',
  PRIMARY KEY (`id`),
  KEY `idx_app` (`app_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='事件应用关联表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `dice_issue_panel`
--

DROP TABLE IF EXISTS `dice_issue_panel`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `dice_issue_panel` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT COMMENT '主键',
  `created_at` datetime DEFAULT NULL COMMENT '表记录创建时间',
  `updated_at` datetime DEFAULT NULL COMMENT '表记录更新时间',
  `project_id` bigint(20) DEFAULT NULL,
  `panel_name` varchar(255) DEFAULT NULL,
  `issue_id` bigint(20) DEFAULT NULL,
  `relation` bigint(20) DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='事件看板表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `dice_issue_property`
--

DROP TABLE IF EXISTS `dice_issue_property`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `dice_issue_property` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT COMMENT '主键',
  `created_at` datetime DEFAULT NULL COMMENT '表记录创建时间',
  `updated_at` datetime DEFAULT NULL COMMENT '表记录更新时间',
  `scope_type` varchar(255) DEFAULT NULL,
  `scope_id` bigint(20) DEFAULT NULL,
  `org_id` bigint(20) DEFAULT NULL,
  `required` tinyint(1) DEFAULT NULL,
  `property_type` varchar(255) DEFAULT NULL,
  `property_name` varchar(255) DEFAULT NULL,
  `display_name` varchar(255) DEFAULT NULL,
  `property_issue_type` varchar(255) DEFAULT NULL,
  `relation` bigint(20) DEFAULT NULL,
  `index` bigint(20) DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=127 DEFAULT CHARSET=utf8mb4 COMMENT='事件属性表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `dice_issue_property_relation`
--

DROP TABLE IF EXISTS `dice_issue_property_relation`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `dice_issue_property_relation` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT COMMENT '主键',
  `created_at` datetime DEFAULT NULL COMMENT '表记录创建时间',
  `updated_at` datetime DEFAULT NULL COMMENT '表记录更新时间',
  `org_id` bigint(20) DEFAULT NULL,
  `project_id` bigint(20) DEFAULT NULL,
  `issue_id` bigint(20) DEFAULT NULL,
  `property_id` bigint(20) DEFAULT NULL,
  `property_value_id` bigint(20) DEFAULT NULL,
  `arbitrary_value` varchar(255) DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=155 DEFAULT CHARSET=utf8mb4 COMMENT='事件属性关联表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `dice_issue_property_value`
--

DROP TABLE IF EXISTS `dice_issue_property_value`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `dice_issue_property_value` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT COMMENT '主键',
  `created_at` datetime DEFAULT NULL COMMENT '表记录创建时间',
  `updated_at` datetime DEFAULT NULL COMMENT '表记录更新时间',
  `property_id` bigint(20) DEFAULT NULL,
  `value` varchar(255) DEFAULT NULL,
  `name` varchar(255) DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=100 DEFAULT CHARSET=utf8mb4 COMMENT='事件属性值表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `dice_issue_relation`
--

DROP TABLE IF EXISTS `dice_issue_relation`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `dice_issue_relation` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT,
  `created_at` datetime DEFAULT NULL COMMENT '表记录创建时间',
  `updated_at` datetime DEFAULT NULL COMMENT '表记录更新时间',
  `issue_id` bigint(20) NOT NULL COMMENT '事件id',
  `related_issue` bigint(20) NOT NULL COMMENT '关联事件id',
  `comment` text COMMENT '关联描述',
  PRIMARY KEY (`id`),
  UNIQUE KEY `issue_related` (`issue_id`,`related_issue`),
  KEY `idx_issue_id` (`issue_id`),
  KEY `idx_related_issue` (`related_issue`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='事件关联表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `dice_issue_stage`
--

DROP TABLE IF EXISTS `dice_issue_stage`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `dice_issue_stage` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT,
  `created_at` datetime DEFAULT NULL COMMENT '表记录创建时间',
  `updated_at` datetime DEFAULT NULL COMMENT '表记录更新时间',
  `org_id` bigint(20) DEFAULT NULL,
  `name` varchar(255) DEFAULT NULL,
  `value` varchar(255) DEFAULT NULL,
  `issue_type` varchar(255) DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=82 DEFAULT CHARSET=utf8mb4 COMMENT='事件任务阶段+引入源';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `dice_issue_state`
--

DROP TABLE IF EXISTS `dice_issue_state`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `dice_issue_state` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT COMMENT '主键',
  `created_at` datetime DEFAULT NULL COMMENT '表记录创建时间',
  `updated_at` datetime DEFAULT NULL COMMENT '表记录更新时间',
  `project_id` bigint(20) DEFAULT NULL,
  `issue_type` varchar(255) DEFAULT NULL,
  `name` varchar(255) DEFAULT NULL,
  `belong` varchar(255) DEFAULT NULL,
  `index` bigint(20) DEFAULT NULL,
  `role` varchar(255) DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=413 DEFAULT CHARSET=utf8mb4 COMMENT='事件状态表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `dice_issue_state_relations`
--

DROP TABLE IF EXISTS `dice_issue_state_relations`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `dice_issue_state_relations` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT COMMENT '主键',
  `created_at` datetime DEFAULT NULL COMMENT '表记录创建时间',
  `updated_at` datetime DEFAULT NULL COMMENT '表记录更新时间',
  `start_state_id` bigint(20) DEFAULT NULL,
  `end_state_id` bigint(20) DEFAULT NULL,
  `project_id` bigint(20) DEFAULT NULL,
  `issue_type` varchar(255) DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=874 DEFAULT CHARSET=utf8mb4 COMMENT='事件状态关联表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `dice_issue_streams`
--

DROP TABLE IF EXISTS `dice_issue_streams`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `dice_issue_streams` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT COMMENT '主键',
  `created_at` datetime DEFAULT NULL COMMENT '表记录创建时间',
  `updated_at` datetime DEFAULT NULL COMMENT '表记录更新时间',
  `issue_id` bigint(20) NOT NULL COMMENT '所属 issue ID',
  `operator` varchar(255) DEFAULT NULL COMMENT '操作者',
  `stream_type` varchar(255) DEFAULT NULL COMMENT '操作记录类型',
  `stream_params` text COMMENT '操作记录参数',
  PRIMARY KEY (`id`),
  KEY `issue_id_index` (`issue_id`)
) ENGINE=InnoDB AUTO_INCREMENT=7 DEFAULT CHARSET=utf8mb4 COMMENT='事件活动记录表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `dice_issue_testcase_relations`
--

DROP TABLE IF EXISTS `dice_issue_testcase_relations`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `dice_issue_testcase_relations` (
  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT,
  `issue_id` bigint(20) DEFAULT NULL,
  `test_plan_id` bigint(20) DEFAULT NULL,
  `test_plan_case_rel_id` bigint(20) DEFAULT NULL,
  `test_case_id` bigint(20) DEFAULT NULL,
  `creator_id` varchar(191) DEFAULT NULL,
  `created_at` datetime DEFAULT NULL,
  `updated_at` datetime DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='事件测试用例关联表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `dice_issues`
--

DROP TABLE IF EXISTS `dice_issues`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `dice_issues` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT COMMENT '主键',
  `created_at` datetime DEFAULT NULL COMMENT '表记录创建时间',
  `updated_at` datetime DEFAULT NULL COMMENT '表记录更新时间',
  `plan_started_at` datetime DEFAULT NULL COMMENT '计划开始时间',
  `plan_finished_at` datetime DEFAULT NULL COMMENT '计划结束时间',
  `project_id` bigint(20) NOT NULL COMMENT '所属项目 ID',
  `iteration_id` bigint(20) NOT NULL COMMENT '所属迭代 ID',
  `app_id` bigint(20) DEFAULT NULL COMMENT '所属应用 ID',
  `requirement_id` bigint(20) DEFAULT NULL COMMENT '所属需求 ID',
  `type` varchar(32) DEFAULT NULL COMMENT 'issue 类型',
  `title` varchar(255) DEFAULT NULL COMMENT '标题',
  `content` text COMMENT '内容',
  `state` varchar(32) NOT NULL DEFAULT '' COMMENT '状态',
  `priority` varchar(32) DEFAULT NULL COMMENT '优先级',
  `complexity` varchar(32) DEFAULT NULL COMMENT '复杂度',
  `bug_type` varchar(32) DEFAULT NULL COMMENT '缺陷类型',
  `creator` varchar(255) DEFAULT NULL COMMENT '创建人',
  `assignee` varchar(255) NOT NULL DEFAULT '' COMMENT '处理人',
  `deleted` tinyint(4) DEFAULT '0',
  `man_hour` text COMMENT '事件工时信息',
  `source` varchar(32) DEFAULT 'user' COMMENT '事件来源',
  `severity` varchar(32) DEFAULT NULL COMMENT '事件严重程度',
  `external` tinyint(1) DEFAULT '1' COMMENT '是否是外部创建的issue',
  `stage` varchar(80) DEFAULT NULL COMMENT 'bug所属阶段和任务类型',
  `owner` varchar(255) DEFAULT NULL COMMENT 'bug责任人',
  `finish_time` datetime DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=6 DEFAULT CHARSET=utf8mb4 COMMENT='事件表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `dice_iterations`
--

DROP TABLE IF EXISTS `dice_iterations`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `dice_iterations` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT COMMENT '主键',
  `created_at` datetime DEFAULT NULL COMMENT '表记录创建时间',
  `updated_at` datetime DEFAULT NULL COMMENT '表记录更新时间',
  `started_at` datetime DEFAULT NULL COMMENT '迭代开始时间',
  `finished_at` datetime DEFAULT NULL COMMENT '迭代结束时间',
  `project_id` bigint(20) DEFAULT NULL COMMENT '所属项目 ID',
  `title` varchar(255) DEFAULT NULL COMMENT '标题',
  `content` text COMMENT '内容',
  `creator` varchar(255) DEFAULT NULL COMMENT '创建人',
  `state` varchar(255) NOT NULL DEFAULT 'UNFILED',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=2 DEFAULT CHARSET=utf8mb4 COMMENT='迭代表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `dice_label_relations`
--

DROP TABLE IF EXISTS `dice_label_relations`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `dice_label_relations` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT,
  `created_at` datetime DEFAULT NULL COMMENT '表记录创建时间',
  `updated_at` datetime DEFAULT NULL COMMENT '表记录更新时间',
  `label_id` bigint(20) NOT NULL,
  `ref_type` varchar(40) NOT NULL COMMENT '标签作用类型, 与 dice_labels type 相同, eg: issue',
  `ref_id` bigint(20) NOT NULL COMMENT '标签关联目标 id, eg: issue_id',
  PRIMARY KEY (`id`),
  KEY `idx_label_id` (`label_id`),
  KEY `idx_ref_id` (`ref_type`,`ref_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='标签关联表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `dice_labels`
--

DROP TABLE IF EXISTS `dice_labels`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `dice_labels` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT,
  `created_at` datetime DEFAULT NULL COMMENT '表记录创建时间',
  `updated_at` datetime DEFAULT NULL COMMENT '表记录更新时间',
  `name` varchar(50) NOT NULL COMMENT '标签名称',
  `type` varchar(40) NOT NULL COMMENT '标签作用类型, eg: issue',
  `color` varchar(40) NOT NULL COMMENT '标签颜色',
  `project_id` bigint(20) NOT NULL COMMENT '标签所属项目',
  `creator` varchar(255) DEFAULT NULL COMMENT '创建人',
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_project_name` (`project_id`,`name`)
) ENGINE=InnoDB AUTO_INCREMENT=2 DEFAULT CHARSET=utf8mb4 COMMENT='标签表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `dice_library_references`
--

DROP TABLE IF EXISTS `dice_library_references`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `dice_library_references` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT COMMENT '主键',
  `app_id` bigint(20) NOT NULL COMMENT '应用 id',
  `lib_id` bigint(20) NOT NULL COMMENT '库 id',
  `lib_name` varchar(255) NOT NULL COMMENT '库名称',
  `lib_desc` text COMMENT '库描述',
  `approval_id` bigint(20) NOT NULL COMMENT '审批流 id',
  `approval_status` varchar(100) NOT NULL COMMENT '状态: 待审核/已通过/已拒绝',
  `creator` varchar(255) NOT NULL COMMENT '创建者',
  `created_at` datetime NOT NULL COMMENT '创建时间',
  `updated_at` datetime DEFAULT NULL COMMENT '更新时间',
  PRIMARY KEY (`id`),
  KEY `idx_app_id` (`app_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8 COMMENT='库引用信息';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `dice_manual_review`
--

DROP TABLE IF EXISTS `dice_manual_review`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `dice_manual_review` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT COMMENT '审核Id',
  `build_id` bigint(20) NOT NULL COMMENT '流水线Id',
  `project_id` bigint(20) NOT NULL COMMENT '项目Id',
  `application_id` bigint(20) NOT NULL COMMENT '应用Id',
  `sponsor_id` bigint(20) NOT NULL COMMENT '发起人Id',
  `commit_id` varchar(50) NOT NULL COMMENT '提交Id',
  `org_id` bigint(20) NOT NULL COMMENT '企业Id',
  `task_id` bigint(20) NOT NULL COMMENT 'taskId 为action的唯一标示',
  `project_name` varchar(50) NOT NULL COMMENT '项目名字',
  `application_name` varchar(50) NOT NULL COMMENT '应用名字',
  `branch_name` varchar(50) NOT NULL COMMENT '代码分支',
  `approval_status` varchar(50) NOT NULL COMMENT '审查是否通过 初值为null,no是失败,yes是成功',
  `commit_message` varchar(50) DEFAULT NULL COMMENT '评论',
  `created_at` datetime NOT NULL COMMENT '创建时间',
  `updated_at` datetime DEFAULT NULL COMMENT '更新时间',
  `approval_reason` varchar(250) DEFAULT NULL COMMENT '拒绝原因',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='审批列表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `dice_manual_review_user`
--

DROP TABLE IF EXISTS `dice_manual_review_user`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `dice_manual_review_user` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT COMMENT '自增Id',
  `org_id` bigint(20) NOT NULL COMMENT '企业Id',
  `operator` bigint(20) NOT NULL COMMENT '用户id',
  `task_id` bigint(20) NOT NULL COMMENT 'taskId 为action的唯一标示',
  `created_at` datetime NOT NULL COMMENT '创建时间',
  `updated_at` datetime DEFAULT NULL COMMENT '更新时间',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='审批用户列表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `dice_mboxs`
--

DROP TABLE IF EXISTS `dice_mboxs`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `dice_mboxs` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT,
  `user_id` varchar(100) DEFAULT NULL COMMENT '用户id',
  `title` text COMMENT '站内信标题',
  `content` text COMMENT '站内信内容',
  `label` varchar(200) DEFAULT NULL COMMENT '站内信所属模块',
  `status` varchar(50) DEFAULT NULL COMMENT '状态 readed:已读 unread:未读',
  `org_id` bigint(20) DEFAULT NULL,
  `read_at` datetime DEFAULT NULL,
  `created_at` timestamp NULL DEFAULT NULL,
  `updated_at` timestamp NULL DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_user_id` (`user_id`),
  KEY `idx_org_id` (`org_id`),
  KEY `idx_status` (`status`),
  KEY `idx_created_at` (`created_at`)
) ENGINE=InnoDB AUTO_INCREMENT=2912 DEFAULT CHARSET=utf8mb4 COMMENT='站内信';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `dice_member`
--

DROP TABLE IF EXISTS `dice_member`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `dice_member` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT,
  `created_at` timestamp NULL DEFAULT NULL,
  `updated_at` timestamp NULL DEFAULT NULL,
  `scope_type` varchar(10) DEFAULT NULL,
  `scope_id` bigint(20) DEFAULT NULL,
  `scope_name` varchar(255) DEFAULT NULL,
  `parent_id` bigint(20) DEFAULT NULL,
  `role` varchar(20) NOT NULL DEFAULT '' COMMENT '角色: Manager/Developer/Tester/Guest',
  `user_id` varchar(128) DEFAULT NULL,
  `nick` varchar(255) DEFAULT NULL,
  `email` varchar(255) DEFAULT NULL,
  `mobile` varchar(40) DEFAULT NULL,
  `name` varchar(255) DEFAULT NULL COMMENT '用户名 (唯一)',
  `token` varchar(100) DEFAULT NULL,
  `avatar` varchar(255) DEFAULT NULL,
  `user_sync_at` timestamp NULL DEFAULT NULL,
  `org_id` bigint(20) DEFAULT NULL,
  `project_id` bigint(20) DEFAULT NULL,
  `project_name` varchar(255) DEFAULT NULL,
  `application_id` bigint(20) DEFAULT NULL,
  `application_name` varchar(255) DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_unique_scope_type_id_user_id` (`scope_type`,`scope_id`,`user_id`)
) ENGINE=InnoDB AUTO_INCREMENT=4679 DEFAULT CHARSET=utf8mb4 BLOCK_FORMAT=ENCRYPTED;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `dice_member_extra`
--

DROP TABLE IF EXISTS `dice_member_extra`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `dice_member_extra` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT,
  `created_at` datetime DEFAULT NULL COMMENT '表记录创建时间',
  `updated_at` datetime DEFAULT NULL COMMENT '表记录更新时间',
  `user_id` varchar(40) NOT NULL COMMENT '成员的用户id',
  `parent_id` varchar(40) DEFAULT '0' COMMENT '成员的父scope id',
  `scope_id` varchar(40) NOT NULL COMMENT '成员所属scope id',
  `scope_type` varchar(40) NOT NULL COMMENT '成员所属scope类型',
  `resource_key` varchar(40) NOT NULL COMMENT '成员关联资源的键',
  `resource_value` varchar(40) NOT NULL COMMENT '成员关联资源的值',
  PRIMARY KEY (`id`),
  KEY `idx_user_id_scope_id_scope_type` (`user_id`,`scope_id`,`scope_type`),
  KEY `idx_resource_key` (`resource_key`),
  KEY `idx_resource_value` (`resource_value`)
) ENGINE=InnoDB AUTO_INCREMENT=5342 DEFAULT CHARSET=utf8mb4 COMMENT='用户额外信息kv表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `dice_nexus_repositories`
--

DROP TABLE IF EXISTS `dice_nexus_repositories`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `dice_nexus_repositories` (
  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT,
  `org_id` bigint(20) DEFAULT NULL,
  `publisher_id` bigint(20) DEFAULT NULL,
  `cluster_name` varchar(128) DEFAULT NULL,
  `name` varchar(128) DEFAULT NULL,
  `format` varchar(32) DEFAULT NULL,
  `type` varchar(32) DEFAULT NULL,
  `config` text,
  `created_at` datetime NOT NULL,
  `updated_at` datetime NOT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=10 DEFAULT CHARSET=utf8mb4 COMMENT='Nexus 仓库表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `dice_nexus_users`
--

DROP TABLE IF EXISTS `dice_nexus_users`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `dice_nexus_users` (
  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT,
  `repo_id` bigint(20) DEFAULT NULL,
  `org_id` bigint(20) DEFAULT NULL,
  `publisher_id` bigint(20) DEFAULT NULL,
  `cluster_name` varchar(128) DEFAULT NULL,
  `name` varchar(128) NOT NULL DEFAULT '',
  `password` varchar(4096) NOT NULL DEFAULT '',
  `config` text,
  `created_at` datetime NOT NULL,
  `updated_at` datetime NOT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=13 DEFAULT CHARSET=utf8mb4 COMMENT='Nexus 仓库用户表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `dice_notices`
--

DROP TABLE IF EXISTS `dice_notices`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `dice_notices` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT COMMENT '主键',
  `org_id` bigint(20) NOT NULL COMMENT '企业 id',
  `content` text NOT NULL COMMENT '公告内容',
  `status` varchar(50) NOT NULL COMMENT '状态: 待发布/已发布/已停用',
  `creator` varchar(255) NOT NULL COMMENT '创建者',
  `created_at` datetime NOT NULL COMMENT '创建时间',
  `updated_at` datetime DEFAULT NULL COMMENT '更新时间',
  PRIMARY KEY (`id`),
  KEY `idx_org_id` (`org_id`),
  KEY `idx_status` (`status`)
) ENGINE=InnoDB AUTO_INCREMENT=25 DEFAULT CHARSET=utf8 COMMENT='平台公告';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `dice_notifies`
--

DROP TABLE IF EXISTS `dice_notifies`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `dice_notifies` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT,
  `created_at` timestamp NULL DEFAULT NULL,
  `updated_at` timestamp NULL DEFAULT NULL,
  `name` varchar(255) DEFAULT NULL,
  `scope_type` varchar(150) DEFAULT NULL,
  `scope_id` varchar(150) DEFAULT NULL,
  `label` varchar(150) DEFAULT NULL,
  `channels` text,
  `notify_group_id` bigint(20) DEFAULT NULL,
  `org_id` bigint(20) DEFAULT NULL,
  `creator` varchar(255) DEFAULT NULL,
  `enabled` tinyint(1) DEFAULT NULL,
  `data` text,
  `cluster_name` varchar(150) DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_scope_type` (`scope_type`),
  KEY `idx_scope_id` (`scope_id`),
  KEY `notify_group_id` (`notify_group_id`)
) ENGINE=InnoDB AUTO_INCREMENT=1023 DEFAULT CHARSET=utf8mb4 BLOCK_FORMAT=ENCRYPTED COMMENT='通知';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `dice_notify_groups`
--

DROP TABLE IF EXISTS `dice_notify_groups`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `dice_notify_groups` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT,
  `created_at` timestamp NULL DEFAULT NULL,
  `updated_at` timestamp NULL DEFAULT NULL,
  `name` varchar(255) DEFAULT NULL,
  `scope_type` varchar(150) DEFAULT NULL,
  `scope_id` varchar(150) DEFAULT NULL,
  `org_id` bigint(20) DEFAULT NULL,
  `workspace` varchar(150) DEFAULT NULL,
  `target_data` text,
  `label` varchar(200) DEFAULT NULL,
  `auto_create` tinyint(4) DEFAULT NULL,
  `creator` varchar(150) DEFAULT NULL,
  `cluster_name` varchar(150) DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_scope_type` (`scope_type`),
  KEY `idx_scope_id` (`scope_id`),
  KEY `idx_workspace` (`workspace`)
) ENGINE=InnoDB AUTO_INCREMENT=1043 DEFAULT CHARSET=utf8mb4 BLOCK_FORMAT=ENCRYPTED COMMENT='通知组';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `dice_notify_histories`
--

DROP TABLE IF EXISTS `dice_notify_histories`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `dice_notify_histories` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT,
  `created_at` timestamp NULL DEFAULT NULL,
  `updated_at` timestamp NULL DEFAULT NULL,
  `notify_name` varchar(255) DEFAULT NULL,
  `workspace` varchar(255) DEFAULT NULL,
  `notify_item_display_name` varchar(255) DEFAULT NULL,
  `channel` varchar(255) DEFAULT NULL,
  `target_data` text,
  `source_data` text,
  `content` text,
  `status` varchar(255) DEFAULT NULL,
  `error_msg` text,
  `org_id` bigint(20) DEFAULT NULL,
  `label` varchar(150) DEFAULT NULL,
  `source_type` varchar(150) DEFAULT NULL,
  `source_id` varchar(150) DEFAULT NULL,
  `cluster_name` varchar(150) DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=385347 DEFAULT CHARSET=utf8mb4 BLOCK_FORMAT=ENCRYPTED COMMENT='通知历史';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `dice_notify_item_relation`
--

DROP TABLE IF EXISTS `dice_notify_item_relation`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `dice_notify_item_relation` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT,
  `created_at` timestamp NULL DEFAULT NULL,
  `updated_at` timestamp NULL DEFAULT NULL,
  `notify_id` bigint(20) DEFAULT NULL,
  `notify_item_id` bigint(20) DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `notify_id` (`notify_id`),
  KEY `notify_item_id` (`notify_item_id`)
) ENGINE=InnoDB AUTO_INCREMENT=2410 DEFAULT CHARSET=utf8mb4 BLOCK_FORMAT=ENCRYPTED COMMENT='通知项通知关联关系';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `dice_notify_items`
--

DROP TABLE IF EXISTS `dice_notify_items`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `dice_notify_items` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT,
  `created_at` timestamp NULL DEFAULT NULL,
  `updated_at` timestamp NULL DEFAULT NULL,
  `name` varchar(150) DEFAULT NULL,
  `display_name` varchar(150) DEFAULT NULL,
  `category` varchar(150) DEFAULT NULL,
  `mobile_template` text,
  `mbox_template` text,
  `email_template` text,
  `dingding_template` text,
  `scope_type` varchar(150) DEFAULT NULL,
  `label` varchar(150) DEFAULT NULL,
  `params` text,
  `vms_template` text COMMENT '语音通知模版',
  `called_show_number` text COMMENT '语音通知被叫显号',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=450 DEFAULT CHARSET=utf8mb4 BLOCK_FORMAT=ENCRYPTED COMMENT='通知项';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `dice_notify_sources`
--

DROP TABLE IF EXISTS `dice_notify_sources`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `dice_notify_sources` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT,
  `created_at` timestamp NULL DEFAULT NULL,
  `updated_at` timestamp NULL DEFAULT NULL,
  `name` varchar(255) DEFAULT NULL,
  `notify_id` bigint(20) DEFAULT NULL,
  `source_type` varchar(255) DEFAULT NULL,
  `source_id` varchar(255) DEFAULT NULL,
  `org_id` bigint(20) DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `notify_id` (`notify_id`)
) ENGINE=InnoDB AUTO_INCREMENT=12189 DEFAULT CHARSET=utf8mb4 BLOCK_FORMAT=ENCRYPTED COMMENT='通知（dice_notifies）扩展表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `dice_org`
--

DROP TABLE IF EXISTS `dice_org`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `dice_org` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT,
  `created_at` timestamp NULL DEFAULT NULL,
  `updated_at` timestamp NULL DEFAULT NULL,
  `name` varchar(255) DEFAULT NULL,
  `desc` varchar(255) DEFAULT NULL,
  `logo` varchar(255) DEFAULT NULL,
  `config` text,
  `locale` varchar(50) DEFAULT 'zh_CN',
  `creator` varchar(255) DEFAULT NULL,
  `type` varchar(255) DEFAULT NULL,
  `status` varchar(255) DEFAULT NULL,
  `open_fdp` tinyint(1) DEFAULT '0' COMMENT '是否打开fdp服务，默认为false',
  `display_name` varchar(64) DEFAULT NULL COMMENT '企业展示名称',
  `blockout_config` text COMMENT '封网配置',
  `is_public` tinyint(1) NOT NULL DEFAULT '0' COMMENT '是否公开',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=3004 DEFAULT CHARSET=utf8mb4 BLOCK_FORMAT=ENCRYPTED;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `dice_org_cluster_relation`
--

DROP TABLE IF EXISTS `dice_org_cluster_relation`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `dice_org_cluster_relation` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT,
  `created_at` timestamp NULL DEFAULT NULL,
  `updated_at` timestamp NULL DEFAULT NULL,
  `org_id` bigint(20) unsigned DEFAULT NULL,
  `org_name` varchar(255) DEFAULT NULL,
  `cluster_id` bigint(20) unsigned DEFAULT NULL,
  `cluster_name` varchar(255) DEFAULT NULL,
  `creator` varchar(255) DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_org_cluster_id` (`org_id`,`cluster_id`)
) ENGINE=InnoDB AUTO_INCREMENT=9 DEFAULT CHARSET=utf8mb4 BLOCK_FORMAT=ENCRYPTED COMMENT='企业集群关联关系';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `dice_pipeline_cms_configs`
--

DROP TABLE IF EXISTS `dice_pipeline_cms_configs`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `dice_pipeline_cms_configs` (
  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT,
  `ns_id` bigint(20) NOT NULL,
  `key` varchar(191) NOT NULL DEFAULT '',
  `value` text,
  `encrypt` tinyint(1) NOT NULL,
  `type` varchar(32) DEFAULT NULL,
  `extra` text,
  `time_created` datetime NOT NULL,
  `time_updated` datetime NOT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_ns_key` (`ns_id`,`key`),
  KEY `idx_key` (`key`)
) ENGINE=InnoDB AUTO_INCREMENT=7613 DEFAULT CHARSET=utf8mb4 COMMENT='流水线配置项表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `dice_pipeline_cms_ns`
--

DROP TABLE IF EXISTS `dice_pipeline_cms_ns`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `dice_pipeline_cms_ns` (
  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT,
  `pipeline_source` varchar(191) NOT NULL DEFAULT '',
  `ns` varchar(191) NOT NULL DEFAULT '',
  `time_created` datetime NOT NULL,
  `time_updated` datetime NOT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_source_ns` (`pipeline_source`,`ns`),
  KEY `idx_source` (`pipeline_source`),
  KEY `idx_ns` (`ns`)
) ENGINE=InnoDB AUTO_INCREMENT=149 DEFAULT CHARSET=utf8mb4 COMMENT='流水线配置命名空间表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `dice_pipeline_lifecycle_hook_clients`
--

DROP TABLE IF EXISTS `dice_pipeline_lifecycle_hook_clients`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `dice_pipeline_lifecycle_hook_clients` (
  `id` bigint(20) NOT NULL COMMENT '主键',
  `host` varchar(255) NOT NULL COMMENT '域名',
  `name` varchar(255) NOT NULL COMMENT '来源名称',
  `prefix` varchar(255) NOT NULL COMMENT '访问前缀',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `dice_pipeline_reports`
--

DROP TABLE IF EXISTS `dice_pipeline_reports`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `dice_pipeline_reports` (
  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT COMMENT '主键',
  `pipeline_id` bigint(20) NOT NULL COMMENT '关联的流水线 ID',
  `type` varchar(32) NOT NULL DEFAULT '' COMMENT '报告类型',
  `meta` text NOT NULL COMMENT '报告元数据',
  `creator_id` varchar(191) DEFAULT '' COMMENT '创建人',
  `updater_id` varchar(191) DEFAULT NULL COMMENT '更新人',
  `created_at` datetime NOT NULL COMMENT '表记录创建时间',
  `updated_at` datetime NOT NULL COMMENT '表记录更新时间',
  PRIMARY KEY (`id`),
  KEY `idx_pipelineid_type` (`pipeline_id`,`type`)
) ENGINE=InnoDB AUTO_INCREMENT=373275 DEFAULT CHARSET=utf8mb4 COMMENT='流水线报告表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `dice_pipeline_snippet_clients`
--

DROP TABLE IF EXISTS `dice_pipeline_snippet_clients`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `dice_pipeline_snippet_clients` (
  `id` bigint(20) NOT NULL,
  `name` varchar(255) NOT NULL,
  `host` varchar(255) NOT NULL,
  `extra` text NOT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='流水线 snippet 客户端表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `dice_pipeline_template_versions`
--

DROP TABLE IF EXISTS `dice_pipeline_template_versions`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `dice_pipeline_template_versions` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT,
  `template_id` bigint(20) NOT NULL,
  `name` varchar(255) NOT NULL,
  `version` varchar(255) NOT NULL,
  `spec` text NOT NULL,
  `readme` text NOT NULL,
  `created_at` datetime NOT NULL,
  `updated_at` datetime NOT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=14 DEFAULT CHARSET=utf8mb4 COMMENT='流水线模板版本表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `dice_pipeline_templates`
--

DROP TABLE IF EXISTS `dice_pipeline_templates`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `dice_pipeline_templates` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT,
  `name` varchar(255) NOT NULL,
  `logo_url` varchar(255) NOT NULL,
  `desc` varchar(255) NOT NULL,
  `scope_type` varchar(10) NOT NULL,
  `scope_id` varchar(255) NOT NULL,
  `default_version` varchar(255) NOT NULL,
  `created_at` datetime NOT NULL,
  `updated_at` datetime NOT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=14 DEFAULT CHARSET=utf8mb4 COMMENT='流水线模板表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `dice_publish_item_h5_targets`
--

DROP TABLE IF EXISTS `dice_publish_item_h5_targets`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `dice_publish_item_h5_targets` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT COMMENT '主键',
  `created_at` datetime DEFAULT NULL COMMENT '表记录创建时间',
  `updated_at` datetime DEFAULT NULL COMMENT '表记录更新时间',
  `h5_version_id` bigint(20) NOT NULL COMMENT 'h5包版本的id',
  `target_version` varchar(40) DEFAULT NULL COMMENT 'h5的目标版本',
  `target_build_id` varchar(100) NOT NULL DEFAULT '' COMMENT 'h5目标版本的build id',
  `target_mobile_type` varchar(40) DEFAULT NULL COMMENT '目标app类型',
  PRIMARY KEY (`id`),
  KEY `idx_h5_version_id` (`h5_version_id`),
  KEY `idx_target` (`target_version`,`target_build_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='h5包适配的移动应用版本信息';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `dice_publish_item_versions`
--

DROP TABLE IF EXISTS `dice_publish_item_versions`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `dice_publish_item_versions` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT COMMENT '主键',
  `version` varchar(50) NOT NULL DEFAULT '' COMMENT '版本号',
  `meta` text COMMENT '元信息',
  `resources` text,
  `swagger` longtext,
  `spec` longtext,
  `readme` longtext,
  `logo` varchar(512) DEFAULT NULL COMMENT '版本logo',
  `desc` varchar(2048) DEFAULT NULL COMMENT '描述信息',
  `creator` varchar(255) DEFAULT NULL COMMENT '创建者',
  `org_id` bigint(20) DEFAULT NULL COMMENT '所属企业',
  `publish_item_id` bigint(20) NOT NULL COMMENT '所属发布仓库',
  `created_at` datetime DEFAULT NULL COMMENT '创建时间',
  `updated_at` datetime DEFAULT NULL COMMENT '更新时间',
  `public` tinyint(1) DEFAULT NULL,
  `is_default` tinyint(1) DEFAULT NULL,
  `version_states` varchar(20) DEFAULT NULL COMMENT '版本状态release or beta',
  `gray_level_percent` int(11) DEFAULT NULL COMMENT '灰度百分比',
  `mobile_type` varchar(40) DEFAULT NULL COMMENT '移动应用的类型',
  `build_id` varchar(255) DEFAULT '1' COMMENT '移动应用的构建id',
  `package_name` varchar(255) DEFAULT NULL COMMENT '包名',
  PRIMARY KEY (`id`),
  KEY `idx_org_id` (`org_id`),
  KEY `publish_item_id` (`publish_item_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8 COMMENT='发布版本';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `dice_publish_items`
--

DROP TABLE IF EXISTS `dice_publish_items`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `dice_publish_items` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT COMMENT '主键',
  `name` varchar(100) NOT NULL DEFAULT '' COMMENT '发布名',
  `display_name` varchar(100) DEFAULT NULL,
  `type` varchar(50) NOT NULL DEFAULT '' COMMENT '发布内容类型 ANDROID|IOS',
  `logo` varchar(512) DEFAULT NULL COMMENT 'logo',
  `desc` varchar(2048) DEFAULT NULL COMMENT '描述信息',
  `creator` varchar(255) DEFAULT NULL COMMENT '创建者',
  `org_id` bigint(20) NOT NULL COMMENT '所属企业',
  `publisher_id` bigint(20) NOT NULL COMMENT '所属发布仓库',
  `public` tinyint(1) NOT NULL DEFAULT '0',
  `ak` varchar(64) DEFAULT NULL COMMENT '离线包的监控AK',
  `created_at` datetime NOT NULL COMMENT '创建时间',
  `updated_at` datetime DEFAULT NULL COMMENT '更新时间',
  `no_jailbreak` tinyint(1) DEFAULT '0' COMMENT '是否禁止越狱配置',
  `geofence_lon` double DEFAULT NULL COMMENT '地理围栏，坐标经度',
  `geofence_lat` double DEFAULT NULL COMMENT '地理围栏，坐标纬度',
  `geofence_radius` int(20) DEFAULT NULL COMMENT '地理围栏，合理半径',
  `gray_level_percent` int(11) NOT NULL DEFAULT '0' COMMENT '灰度百分比，0-100',
  `is_migration` tinyint(4) DEFAULT '1' COMMENT '该item灰度逻辑是否已迁移',
  `preview_images` text COMMENT '预览图',
  `background_image` text COMMENT '背景图',
  `ai` varchar(50) DEFAULT NULL COMMENT '离线包的监控AI,一般是发布内容的名字',
  PRIMARY KEY (`id`),
  KEY `idx_org_id` (`org_id`),
  KEY `idx_publisher_id` (`publisher_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8 COMMENT='发布内容';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `dice_publish_items_blacklist`
--

DROP TABLE IF EXISTS `dice_publish_items_blacklist`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `dice_publish_items_blacklist` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT COMMENT '自增id',
  `user_id` varchar(256) NOT NULL COMMENT '用户id',
  `publish_item_id` bigint(20) NOT NULL COMMENT '发布内容id',
  `publish_item_key` varchar(64) DEFAULT NULL COMMENT '监控收集数据需要',
  `user_name` varchar(256) DEFAULT NULL COMMENT '用户名称',
  `device_no` varchar(512) NOT NULL DEFAULT '' COMMENT '设备号',
  `operator` varchar(255) NOT NULL COMMENT '操作人',
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '修改时间',
  PRIMARY KEY (`id`),
  KEY `idx_publish_item_id` (`publish_item_id`),
  KEY `idx_publish_item_key` (`publish_item_key`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8 COMMENT='发布内容黑名单';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `dice_publish_items_erase`
--

DROP TABLE IF EXISTS `dice_publish_items_erase`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `dice_publish_items_erase` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT COMMENT '自增id',
  `publish_item_id` bigint(20) NOT NULL COMMENT '发布内容id',
  `publish_item_key` varchar(64) DEFAULT NULL COMMENT '监控收集数据需要',
  `device_no` varchar(512) NOT NULL DEFAULT '' COMMENT '设备号',
  `erase_status` varchar(32) NOT NULL DEFAULT '' COMMENT '擦除状态',
  `operator` varchar(255) NOT NULL COMMENT '操作人',
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '修改时间',
  PRIMARY KEY (`id`),
  KEY `idx_publish_item_id` (`publish_item_id`),
  KEY `idx_publish_item_key` (`publish_item_key`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8 COMMENT='发布内容数据擦除列表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `dice_publishers`
--

DROP TABLE IF EXISTS `dice_publishers`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `dice_publishers` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT COMMENT '主键',
  `name` varchar(50) NOT NULL DEFAULT '' COMMENT 'publisher名称',
  `publisher_type` varchar(50) NOT NULL DEFAULT '' COMMENT 'publisher类型',
  `logo` varchar(512) DEFAULT NULL COMMENT 'publisher Logo',
  `desc` varchar(2048) DEFAULT NULL COMMENT 'publisher描述',
  `creator` varchar(255) NOT NULL COMMENT '创建者',
  `org_id` bigint(20) NOT NULL COMMENT '所属企业',
  `created_at` datetime NOT NULL COMMENT '创建时间',
  `updated_at` datetime DEFAULT NULL COMMENT '更新时间',
  `publisher_key` varchar(64) NOT NULL DEFAULT '' COMMENT 'publisher key',
  PRIMARY KEY (`id`),
  KEY `idx_org_id` (`org_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8 COMMENT='publisher信息';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `dice_release`
--

DROP TABLE IF EXISTS `dice_release`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `dice_release` (
  `release_id` varchar(64) NOT NULL DEFAULT '',
  `release_name` varchar(255) NOT NULL,
  `desc` text,
  `dice` text,
  `addon` text,
  `labels` varchar(1000) DEFAULT NULL,
  `version` varchar(100) DEFAULT NULL,
  `org_id` bigint(20) DEFAULT NULL,
  `project_id` bigint(20) DEFAULT NULL,
  `application_id` bigint(20) DEFAULT NULL,
  `project_name` varchar(80) DEFAULT NULL,
  `application_name` varchar(80) DEFAULT NULL,
  `user_id` varchar(50) DEFAULT NULL,
  `cluster_name` varchar(80) DEFAULT NULL,
  `cross_cluster` tinyint(4) NOT NULL DEFAULT '0',
  `resources` text,
  `reference` bigint(20) DEFAULT NULL,
  `created_at` timestamp NULL DEFAULT NULL,
  `updated_at` timestamp NULL DEFAULT NULL,
  PRIMARY KEY (`release_id`),
  KEY `idx_release_name` (`release_name`),
  KEY `idx_org_id` (`org_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8 BLOCK_FORMAT=ENCRYPTED COMMENT='Dice 版本表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `dice_repo_caches`
--

DROP TABLE IF EXISTS `dice_repo_caches`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `dice_repo_caches` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT,
  `type_name` varchar(150) DEFAULT NULL,
  `key_name` varchar(150) DEFAULT NULL,
  `value` text,
  `created_at` timestamp NULL DEFAULT NULL,
  `updated_at` timestamp NULL DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `type_name` (`type_name`),
  KEY `key_name` (`key_name`)
) ENGINE=InnoDB AUTO_INCREMENT=11048 DEFAULT CHARSET=utf8mb4 BLOCK_FORMAT=ENCRYPTED COMMENT='Gittar 仓库缓存表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `dice_repo_check_runs`
--

DROP TABLE IF EXISTS `dice_repo_check_runs`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `dice_repo_check_runs` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT,
  `repo_id` bigint(20) DEFAULT NULL,
  `name` varchar(100) COLLATE utf8mb4_bin DEFAULT NULL,
  `type` varchar(100) COLLATE utf8mb4_bin DEFAULT NULL,
  `external_id` varchar(100) COLLATE utf8mb4_bin DEFAULT NULL,
  `commit` varchar(100) COLLATE utf8mb4_bin DEFAULT NULL,
  `status` varchar(50) COLLATE utf8mb4_bin DEFAULT '',
  `output` text COLLATE utf8mb4_bin,
  `result` varchar(100) COLLATE utf8mb4_bin DEFAULT '',
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `completed_at` timestamp NULL DEFAULT NULL,
  `mr_id` int(11) NOT NULL DEFAULT '0',
  `pipeline_id` int(11) NOT NULL DEFAULT '0',
  PRIMARY KEY (`id`),
  KEY `idx_repo_id` (`repo_id`),
  KEY `idx_name` (`name`),
  KEY `idx_type` (`type`),
  KEY `idx_external_id` (`external_id`),
  KEY `idx_commit` (`commit`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin COMMENT='Gittar check-run 表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `dice_repo_files`
--

DROP TABLE IF EXISTS `dice_repo_files`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `dice_repo_files` (
  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT,
  `repo_id` bigint(20) DEFAULT NULL,
  `commit_id` varchar(64) DEFAULT NULL,
  `remark` text,
  `uuid` varchar(32) DEFAULT NULL,
  `deleted_at` datetime DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=64 DEFAULT CHARSET=utf8mb4 COMMENT='Gittar 文件表(目前只有备份)';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `dice_repo_merge_requests`
--

DROP TABLE IF EXISTS `dice_repo_merge_requests`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `dice_repo_merge_requests` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT,
  `repo_id` bigint(20) DEFAULT NULL,
  `title` varchar(255) DEFAULT NULL,
  `description` varchar(255) DEFAULT NULL,
  `state` varchar(150) DEFAULT NULL,
  `author_id` varchar(150) DEFAULT NULL,
  `assignee_id` varchar(150) DEFAULT NULL,
  `merge_user_id` varchar(255) DEFAULT NULL,
  `close_user_id` varchar(255) DEFAULT NULL,
  `merge_commit_sha` varchar(255) DEFAULT NULL,
  `repo_merge_id` int(11) DEFAULT NULL,
  `source_branch` varchar(255) DEFAULT NULL,
  `source_sha` varchar(255) DEFAULT NULL,
  `target_branch` varchar(255) DEFAULT NULL,
  `target_sha` varchar(255) DEFAULT NULL,
  `remove_source_branch` tinyint(1) DEFAULT NULL,
  `created_at` timestamp NULL DEFAULT NULL,
  `updated_at` timestamp NULL DEFAULT NULL,
  `merge_at` timestamp NULL DEFAULT NULL,
  `close_at` timestamp NULL DEFAULT NULL,
  `score` int(11) NOT NULL DEFAULT '0',
  `score_num` int(11) NOT NULL DEFAULT '0',
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_unique_merge_id` (`repo_id`,`repo_merge_id`),
  KEY `idx_repo_id` (`repo_id`),
  KEY `idx_state` (`state`),
  KEY `idx_author_id` (`author_id`),
  KEY `idx_assignee_id` (`assignee_id`)
) ENGINE=InnoDB AUTO_INCREMENT=2607 DEFAULT CHARSET=utf8mb4 BLOCK_FORMAT=ENCRYPTED COMMENT='Gittar 合并请求';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `dice_repo_notes`
--

DROP TABLE IF EXISTS `dice_repo_notes`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `dice_repo_notes` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT,
  `repo_id` bigint(20) DEFAULT NULL,
  `type` varchar(150) DEFAULT NULL,
  `discussion_id` varchar(255) DEFAULT NULL,
  `old_commit_id` varchar(255) DEFAULT NULL,
  `new_commit_id` varchar(255) DEFAULT NULL,
  `merge_id` bigint(20) DEFAULT NULL,
  `note` varchar(255) DEFAULT NULL,
  `data` text,
  `author_id` varchar(255) DEFAULT NULL,
  `created_at` timestamp NULL DEFAULT NULL,
  `updated_at` timestamp NULL DEFAULT NULL,
  `score` int(11) NOT NULL DEFAULT '0',
  PRIMARY KEY (`id`),
  KEY `idx_type` (`type`),
  KEY `idx_merge_id` (`merge_id`)
) ENGINE=InnoDB AUTO_INCREMENT=14 DEFAULT CHARSET=utf8mb4 BLOCK_FORMAT=ENCRYPTED COMMENT='Gittar 评论表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `dice_repo_web_hook_tasks`
--

DROP TABLE IF EXISTS `dice_repo_web_hook_tasks`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `dice_repo_web_hook_tasks` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT,
  `created_at` timestamp NULL DEFAULT NULL,
  `updated_at` timestamp NULL DEFAULT NULL,
  `deleted_at` timestamp NULL DEFAULT NULL,
  `hook_id` bigint(20) DEFAULT NULL,
  `url` varchar(255) DEFAULT NULL,
  `event` varchar(255) DEFAULT NULL,
  `is_delivered` tinyint(1) DEFAULT NULL,
  `is_succeed` tinyint(1) DEFAULT NULL,
  `request_content` text,
  `response_content` text,
  `response_status` varchar(255) DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_gittar_web_hook_tasks_deleted_at` (`deleted_at`),
  KEY `idx_hook_id` (`hook_id`)
) ENGINE=InnoDB AUTO_INCREMENT=45166 DEFAULT CHARSET=utf8mb4 BLOCK_FORMAT=ENCRYPTED COMMENT='Gittar webhook 任务表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `dice_repo_web_hooks`
--

DROP TABLE IF EXISTS `dice_repo_web_hooks`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `dice_repo_web_hooks` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT,
  `created_at` timestamp NULL DEFAULT NULL,
  `updated_at` timestamp NULL DEFAULT NULL,
  `deleted_at` timestamp NULL DEFAULT NULL,
  `hook_type` varchar(150) DEFAULT NULL,
  `name` varchar(150) DEFAULT NULL,
  `repo_id` bigint(20) DEFAULT NULL,
  `token` varchar(255) DEFAULT NULL,
  `url` varchar(255) DEFAULT NULL,
  `is_active` tinyint(1) DEFAULT NULL,
  `push_events` tinyint(1) DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_gittar_web_hooks_deleted_at` (`deleted_at`),
  KEY `idx_hook_type` (`hook_type`),
  KEY `idx_hook_name` (`name`),
  KEY `idx_repo_id` (`repo_id`)
) ENGINE=InnoDB AUTO_INCREMENT=5 DEFAULT CHARSET=utf8mb4 BLOCK_FORMAT=ENCRYPTED COMMENT='Gittar webhoob 表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `dice_repos`
--

DROP TABLE IF EXISTS `dice_repos`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `dice_repos` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT,
  `org_id` bigint(20) DEFAULT NULL,
  `project_id` bigint(20) DEFAULT NULL,
  `app_id` bigint(20) DEFAULT NULL,
  `org_name` varchar(150) DEFAULT NULL,
  `project_name` varchar(150) DEFAULT NULL,
  `app_name` varchar(150) DEFAULT NULL,
  `path` varchar(150) DEFAULT NULL,
  `size` bigint(20) DEFAULT NULL,
  `config` text,
  `is_external` tinyint(1) DEFAULT '0',
  `is_locked` tinyint(1) NOT NULL DEFAULT '0',
  PRIMARY KEY (`id`),
  KEY `idx_org_name` (`org_name`),
  KEY `idx_project_name` (`project_name`),
  KEY `idx_app_name` (`app_name`),
  KEY `idx_path` (`path`)
) ENGINE=InnoDB AUTO_INCREMENT=117 DEFAULT CHARSET=utf8mb4 BLOCK_FORMAT=ENCRYPTED COMMENT='Gittar 仓库表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `dice_role_permission`
--

DROP TABLE IF EXISTS `dice_role_permission`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `dice_role_permission` (
  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT,
  `created_at` timestamp NULL DEFAULT NULL,
  `updated_at` timestamp NULL DEFAULT NULL,
  `role` varchar(30) DEFAULT NULL,
  `resource` varchar(40) DEFAULT NULL,
  `action` varchar(30) DEFAULT NULL,
  `creator` varchar(255) DEFAULT NULL,
  `resource_role` varchar(30) DEFAULT NULL COMMENT '角色: Creator/Assignee',
  `scope` varchar(30) DEFAULT NULL COMMENT '角色所属的scope',
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_resource_action` (`role`,`resource`,`action`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 BLOCK_FORMAT=ENCRYPTED COMMENT='角色权限表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `dice_runner_tasks`
--

DROP TABLE IF EXISTS `dice_runner_tasks`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `dice_runner_tasks` (
  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT,
  `job_id` varchar(150) DEFAULT NULL,
  `status` varchar(150) DEFAULT NULL,
  `open_api_token` text,
  `context_data_url` varchar(255) DEFAULT NULL,
  `result_data_url` varchar(255) DEFAULT NULL,
  `commands` text,
  `targets` text,
  `work_dir` varchar(255) DEFAULT NULL,
  `created_at` datetime DEFAULT NULL,
  `updated_at` datetime DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_status` (`status`),
  KEY `idx_job_id` (`job_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='action runner 任务执行信息';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `dice_test_cases`
--

DROP TABLE IF EXISTS `dice_test_cases`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `dice_test_cases` (
  `id` int(10) unsigned NOT NULL AUTO_INCREMENT,
  `name` varchar(191) DEFAULT NULL,
  `project_id` bigint(20) DEFAULT NULL,
  `test_set_id` bigint(20) DEFAULT NULL,
  `priority` varchar(191) DEFAULT NULL,
  `pre_condition` text,
  `step_and_results` text,
  `desc` varchar(1024) DEFAULT NULL,
  `recycled` tinyint(1) DEFAULT NULL,
  `from` varchar(191) DEFAULT NULL,
  `creator_id` varchar(191) DEFAULT NULL,
  `updater_id` varchar(191) DEFAULT NULL,
  `created_at` datetime DEFAULT NULL,
  `updated_at` datetime DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=7504 DEFAULT CHARSET=utf8mb4 COMMENT='手动测试-测试用例表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `dice_test_plan_case_relations`
--

DROP TABLE IF EXISTS `dice_test_plan_case_relations`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `dice_test_plan_case_relations` (
  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT,
  `test_plan_id` bigint(20) DEFAULT NULL,
  `test_set_id` bigint(20) DEFAULT NULL,
  `test_case_id` bigint(20) DEFAULT NULL,
  `exec_status` varchar(191) DEFAULT NULL,
  `creator_id` varchar(191) DEFAULT NULL,
  `updater_id` varchar(191) DEFAULT NULL,
  `executor_id` varchar(191) DEFAULT NULL,
  `created_at` datetime DEFAULT NULL,
  `updated_at` datetime DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=2544 DEFAULT CHARSET=utf8mb4 COMMENT='手动测试-测试计划用例关联表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `dice_test_plan_members`
--

DROP TABLE IF EXISTS `dice_test_plan_members`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `dice_test_plan_members` (
  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT,
  `test_plan_id` bigint(20) DEFAULT NULL,
  `role` varchar(32) DEFAULT NULL,
  `user_id` bigint(20) DEFAULT NULL,
  `created_at` datetime DEFAULT NULL,
  `updated_at` datetime DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=246 DEFAULT CHARSET=utf8mb4 COMMENT='手动测试-测试计划成员表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `dice_test_plans`
--

DROP TABLE IF EXISTS `dice_test_plans`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `dice_test_plans` (
  `id` int(11) unsigned NOT NULL AUTO_INCREMENT,
  `name` varchar(191) DEFAULT NULL,
  `status` varchar(191) DEFAULT NULL,
  `project_id` bigint(20) DEFAULT NULL,
  `summary` text,
  `creator_id` varchar(191) DEFAULT NULL,
  `updater_id` varchar(191) DEFAULT NULL,
  `started_at` datetime DEFAULT NULL,
  `ended_at` datetime DEFAULT NULL,
  `type` varchar(1) DEFAULT 'm',
  `inode` varchar(20) DEFAULT NULL,
  `created_at` datetime DEFAULT NULL,
  `updated_at` datetime DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=66 DEFAULT CHARSET=utf8mb4 COMMENT='手动测试-测试计划表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `dice_test_sets`
--

DROP TABLE IF EXISTS `dice_test_sets`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `dice_test_sets` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT COMMENT '自增ID',
  `name` varchar(256) NOT NULL COMMENT '测试集的中文名,可重名',
  `parent_id` bigint(20) NOT NULL DEFAULT '0' COMMENT '上一级的所属测试集id,顶级时为0',
  `recycled` tinyint(1) NOT NULL DEFAULT '0' COMMENT '是否已进入回收站，默认0为否，1为是。在回收站的顶层显示',
  `directory` text NOT NULL COMMENT '当前节点+所有父级节点的name集合（参考值：新建测试集1/新建测试集2/测试集名称3），这里冗余是为了方便界面展示。',
  `project_id` bigint(20) NOT NULL DEFAULT '0' COMMENT '项目id，当前测试集所属的真正项目id',
  `order_num` int(4) NOT NULL DEFAULT '0' COMMENT '用例集展示的顺序',
  `creator_id` varchar(191) NOT NULL DEFAULT '' COMMENT '创建人',
  `updater_id` varchar(191) DEFAULT NULL COMMENT '修改人',
  `created_at` datetime DEFAULT NULL COMMENT '创建时间',
  `updated_at` datetime DEFAULT NULL COMMENT '创建时间',
  PRIMARY KEY (`id`),
  KEY `idx_test_project_id` (`project_id`,`parent_id`)
) ENGINE=InnoDB AUTO_INCREMENT=1259 DEFAULT CHARSET=utf8mb4 COMMENT='手动测试-测试集表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `dice_ucevent_sync_record`
--

DROP TABLE IF EXISTS `dice_ucevent_sync_record`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `dice_ucevent_sync_record` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT,
  `created_at` datetime DEFAULT NULL COMMENT '表记录创建时间',
  `updated_at` datetime DEFAULT NULL COMMENT '表记录更新时间',
  `uc_id` bigint(20) NOT NULL COMMENT 'uc事件id',
  `uc_eventtime` datetime NOT NULL COMMENT 'uc事件时间',
  `un_receiver` varchar(40) DEFAULT NULL COMMENT 'uc事件同步失败的接收者',
  PRIMARY KEY (`id`),
  KEY `idx_uc_id` (`uc_id`),
  KEY `idx_uc_eventtime` (`uc_eventtime`),
  KEY `idx_un_receiver` (`un_receiver`)
) ENGINE=InnoDB AUTO_INCREMENT=126915 DEFAULT CHARSET=utf8mb4 COMMENT='dice拉取uc事件的记录';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `edge_apps`
--

DROP TABLE IF EXISTS `edge_apps`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `edge_apps` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT COMMENT '自增Id',
  `org_id` bigint(20) NOT NULL COMMENT '企业Id',
  `cluster_id` bigint(20) NOT NULL COMMENT '关联集群ID',
  `name` varchar(50) NOT NULL COMMENT '应用名',
  `type` varchar(50) NOT NULL COMMENT '发布类型',
  `image` varchar(512) NOT NULL COMMENT '镜像',
  `registry_addr` varchar(512) NOT NULL COMMENT '镜像仓库地址',
  `registry_user` varchar(100) NOT NULL COMMENT '镜像仓库用户名',
  `registry_password` varchar(512) NOT NULL COMMENT '镜像仓库密码',
  `product_id` bigint(20) NOT NULL COMMENT '制品ID',
  `addon_name` varchar(50) NOT NULL COMMENT '中间件',
  `addon_version` varchar(50) NOT NULL COMMENT '中间件版本',
  `config_set_name` varchar(50) NOT NULL COMMENT '配置集',
  `replicas` bigint(20) NOT NULL COMMENT '副本',
  `health_check_type` varchar(50) NOT NULL COMMENT '健康检查类型',
  `health_check_http_port` varchar(50) NOT NULL COMMENT '健康检查http端口',
  `health_check_http_path` varchar(50) NOT NULL COMMENT '健康检查http路径',
  `health_check_exec` varchar(50) NOT NULL COMMENT '健康检查command',
  `edge_sites` varchar(2048) DEFAULT NULL COMMENT '发布站点',
  `depend_app` varchar(2048) DEFAULT NULL COMMENT '依赖应用',
  `port_maps` varchar(2048) DEFAULT NULL COMMENT '依赖应用',
  `extra_data` varchar(2048) DEFAULT NULL COMMENT '依赖应用',
  `limit_cpu` float DEFAULT NULL COMMENT 'CPU LIMIT',
  `request_cpu` float DEFAULT NULL COMMENT 'CPU REQUEST',
  `limit_mem` float DEFAULT NULL COMMENT 'MEMORY LIMIT',
  `request_mem` float DEFAULT NULL COMMENT 'MEMORY REQUEST',
  `description` varchar(100) DEFAULT NULL COMMENT '应用描述',
  `created_at` datetime NOT NULL COMMENT '创建时间',
  `updated_at` datetime DEFAULT NULL COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_org_edgeapp_name` (`org_id`,`name`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='边缘应用';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `edge_configsets`
--

DROP TABLE IF EXISTS `edge_configsets`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `edge_configsets` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT COMMENT '配置集ID',
  `org_id` bigint(20) NOT NULL COMMENT '企业ID',
  `cluster_id` bigint(20) NOT NULL COMMENT '关联集群ID',
  `name` varchar(50) NOT NULL COMMENT '配置集名称',
  `display_name` varchar(50) NOT NULL COMMENT '配置集显示名称',
  `description` varchar(2048) DEFAULT NULL COMMENT '配置集描述',
  `created_at` datetime NOT NULL COMMENT '创建时间',
  `updated_at` datetime DEFAULT NULL COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `edge_configsets_un` (`cluster_id`,`name`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='边缘配置集';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `edge_configsets_item`
--

DROP TABLE IF EXISTS `edge_configsets_item`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `edge_configsets_item` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT COMMENT '配置项ID',
  `configset_id` bigint(20) NOT NULL COMMENT '配置集ID',
  `scope` varchar(10) NOT NULL COMMENT '配置项范围',
  `site_id` bigint(20) DEFAULT NULL COMMENT '关联站点ID',
  `item_key` varchar(100) NOT NULL COMMENT '配置项Key',
  `item_value` varchar(2048) NOT NULL COMMENT '配置项Value',
  `created_at` datetime NOT NULL COMMENT '创建时间',
  `updated_at` datetime DEFAULT NULL COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `edge_configsets_item_un` (`configset_id`,`scope`,`site_id`,`item_key`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='边缘配置项';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `edge_sites`
--

DROP TABLE IF EXISTS `edge_sites`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `edge_sites` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT COMMENT '站点ID',
  `org_id` bigint(20) NOT NULL COMMENT '企业ID',
  `cluster_id` bigint(20) NOT NULL COMMENT '关联集群ID',
  `name` varchar(50) NOT NULL COMMENT '站点名称',
  `display_name` varchar(50) NOT NULL COMMENT '站点显示名称',
  `description` varchar(2048) DEFAULT NULL COMMENT '站点描述',
  `status` varchar(50) NOT NULL COMMENT '站点状态',
  `logo` varchar(500) DEFAULT NULL COMMENT '站点Logo',
  `created_at` datetime NOT NULL COMMENT '创建时间',
  `updated_at` datetime DEFAULT NULL COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `edge_sites_un` (`name`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='边缘站点';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `favorited_resources`
--

DROP TABLE IF EXISTS `favorited_resources`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `favorited_resources` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT,
  `created_at` timestamp NULL DEFAULT NULL,
  `updated_at` timestamp NULL DEFAULT NULL,
  `target` varchar(255) DEFAULT NULL,
  `target_id` bigint(20) unsigned DEFAULT NULL,
  `user_id` varchar(255) DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=10 DEFAULT CHARSET=utf8mb4 BLOCK_FORMAT=ENCRYPTED COMMENT='最喜欢的资源表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `fdp_agent_ad_hoc_cache_data`
--

DROP TABLE IF EXISTS `fdp_agent_ad_hoc_cache_data`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `fdp_agent_ad_hoc_cache_data` (
  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT COMMENT '主键id',
  `result` longtext COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '即席查询结果',
  `creator_id` varchar(256) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT '创建人 ID',
  `updater_id` varchar(256) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT '修改人 ID',
  `delete_yn` tinyint(1) NOT NULL DEFAULT '0' COMMENT '逻辑删除标记',
  `created_at` datetime NOT NULL COMMENT '表记录创建时间',
  `updated_at` datetime NOT NULL COMMENT '表记录更新时间',
  `cluster_id` bigint(20) DEFAULT NULL COMMENT '集群id',
  `project_id` bigint(20) DEFAULT '1' COMMENT '项目id',
  `org_id` bigint(20) DEFAULT NULL COMMENT '集群id',
  `application_id` bigint(20) DEFAULT NULL COMMENT '应用id',
  `cluster_name` varchar(256) COLLATE utf8mb4_unicode_ci DEFAULT '' COMMENT '集群名称',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=7 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='即席查询缓存记录表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `fdp_agent_algorithm_model_manage`
--

DROP TABLE IF EXISTS `fdp_agent_algorithm_model_manage`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `fdp_agent_algorithm_model_manage` (
  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT COMMENT '主键id',
  `model_name` varchar(128) NOT NULL COMMENT '模型名称',
  `algorithm_name` varchar(256) DEFAULT NULL COMMENT '算法名称',
  `feature_list` varchar(1024) DEFAULT NULL COMMENT '特征列',
  `target_list` varchar(1024) DEFAULT NULL COMMENT '目标列',
  `params` varchar(1024) DEFAULT NULL COMMENT '参数',
  `model_file_name` varchar(256) DEFAULT NULL COMMENT '模型文件名称',
  `model_file_url` varchar(1024) DEFAULT NULL COMMENT '模型文件地址',
  `model_version` varchar(45) NOT NULL DEFAULT '0' COMMENT '模型版本',
  `model_status` varchar(32) DEFAULT NULL COMMENT '发布状态',
  `api_url` varchar(1024) DEFAULT NULL COMMENT '调用api',
  `package_id` varchar(128) DEFAULT NULL COMMENT '流量入口ID',
  `consumer_id` varchar(128) DEFAULT NULL COMMENT '调用方ID',
  `creator_id` varchar(128) DEFAULT NULL COMMENT '创建人 ID',
  `updater_id` varchar(128) DEFAULT NULL COMMENT '更新者ID',
  `delete_yn` tinyint(1) NOT NULL DEFAULT '0' COMMENT '逻辑删除标记',
  `created_at` datetime NOT NULL COMMENT '表记录创建时间',
  `updated_at` datetime NOT NULL COMMENT '表记录更新时间',
  `cluster_id` bigint(20) DEFAULT NULL COMMENT '集群id',
  `project_id` bigint(20) DEFAULT '1' COMMENT '项目id',
  `org_id` bigint(20) DEFAULT NULL COMMENT '集群id',
  `application_id` bigint(20) DEFAULT NULL COMMENT '应用id',
  `cluster_name` varchar(256) DEFAULT '' COMMENT '集群名称',
  PRIMARY KEY (`id`),
  KEY `key_model_name` (`model_name`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8 COMMENT='模型管理';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `fdp_agent_data_audit`
--

DROP TABLE IF EXISTS `fdp_agent_data_audit`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `fdp_agent_data_audit` (
  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT COMMENT '主键id',
  `data_model_id` bigint(20) NOT NULL COMMENT '数据模型ID',
  `apply_name` varchar(128) DEFAULT '' COMMENT '申请名称',
  `apply_reason` varchar(256) DEFAULT '' COMMENT '申请原因',
  `apply_type` varchar(256) DEFAULT '' COMMENT '申请类型',
  `apply_result` varchar(256) DEFAULT '' COMMENT '申请结果',
  `flow_id` varchar(256) DEFAULT NULL COMMENT '流程ID',
  `audit_user_name` varchar(256) DEFAULT NULL COMMENT '审批人',
  `creator_id` varchar(255) NOT NULL DEFAULT '' COMMENT '创建人 ID',
  `delete_yn` tinyint(1) NOT NULL DEFAULT '0' COMMENT '逻辑删除标记',
  `created_at` datetime NOT NULL COMMENT '表记录创建时间',
  `updated_at` datetime NOT NULL COMMENT '表记录更新时间',
  `cluster_id` bigint(20) DEFAULT NULL COMMENT '集群id',
  `project_id` bigint(20) DEFAULT '1' COMMENT '项目id',
  `org_id` bigint(20) DEFAULT NULL COMMENT '集群id',
  `application_id` bigint(20) DEFAULT NULL COMMENT '应用id',
  `cluster_name` varchar(256) DEFAULT '' COMMENT '集群名称',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8 COMMENT='创建审批记录表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `fdp_agent_data_auth`
--

DROP TABLE IF EXISTS `fdp_agent_data_auth`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `fdp_agent_data_auth` (
  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT COMMENT '主键id',
  `data_model_id` bigint(20) NOT NULL COMMENT '数据模型ID',
  `apply_name` varchar(128) DEFAULT '' COMMENT '申请名称',
  `apply_user_id` varchar(128) DEFAULT '' COMMENT '申请人ID',
  `apply_user_name` varchar(256) DEFAULT '' COMMENT '申请人姓名',
  `apply_type` varchar(64) DEFAULT '' COMMENT '申请类型',
  `apply_time` datetime NOT NULL COMMENT '申请时间',
  `audit_status` varchar(128) NOT NULL DEFAULT '' COMMENT '审批状态',
  `last_audit_time` datetime DEFAULT NULL COMMENT '最后审批时间',
  `remark` text COMMENT '拒绝理由',
  `creator_id` varchar(255) NOT NULL DEFAULT '' COMMENT '创建人 ID',
  `delete_yn` tinyint(1) NOT NULL DEFAULT '0' COMMENT '逻辑删除标记',
  `created_at` datetime NOT NULL COMMENT '表记录创建时间',
  `updated_at` datetime NOT NULL COMMENT '表记录更新时间',
  `cluster_id` bigint(20) DEFAULT NULL COMMENT '集群id',
  `project_id` bigint(20) DEFAULT '1' COMMENT '项目id',
  `org_id` bigint(20) DEFAULT NULL COMMENT '集群id',
  `application_id` bigint(20) DEFAULT NULL COMMENT '应用id',
  `cluster_name` varchar(256) DEFAULT '' COMMENT '集群名称',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8 COMMENT='审批授权表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `fdp_agent_data_catalog`
--

DROP TABLE IF EXISTS `fdp_agent_data_catalog`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `fdp_agent_data_catalog` (
  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT COMMENT '主键id',
  `name` varchar(128) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT '目录名',
  `name_pinyin` varchar(128) CHARACTER SET latin1 NOT NULL DEFAULT '' COMMENT '目录拼音',
  `parent_id` bigint(20) NOT NULL COMMENT '上级目录id',
  `creator_id` varchar(256) CHARACTER SET latin1 NOT NULL DEFAULT '' COMMENT '创建人',
  `updater_id` varchar(256) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT '修改人',
  `created_at` datetime NOT NULL COMMENT '表记录创建时间',
  `updated_at` datetime NOT NULL COMMENT '表记录更新时间',
  `cluster_id` bigint(20) DEFAULT NULL COMMENT '集群id',
  `project_id` bigint(20) DEFAULT '1' COMMENT '项目id',
  `org_id` bigint(20) DEFAULT NULL COMMENT '集群id',
  `application_id` bigint(20) DEFAULT NULL COMMENT '应用id',
  `cluster_name` varchar(256) COLLATE utf8mb4_unicode_ci DEFAULT '' COMMENT '集群名称',
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_parent_id_name` (`name`,`parent_id`,`project_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='数据目录表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `fdp_agent_data_catalog_dependence`
--

DROP TABLE IF EXISTS `fdp_agent_data_catalog_dependence`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `fdp_agent_data_catalog_dependence` (
  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT COMMENT '主键id',
  `source` varchar(64) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT '平台来源：DL/CDP',
  `type_in_cdp` varchar(64) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT '数据模型在 CDP 下的类型：ods, dim, dwd, dws, ads',
  `name` varchar(128) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT '数据模型名称，在一个 source 下 name 需要唯一',
  `description` varchar(1024) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT '数据模型描述',
  `data_source_id` bigint(20) NOT NULL COMMENT '数据源 ID',
  `data_source_category` varchar(1024) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT '数据源分类: INTERNAL, EXTERNAL',
  `model_type` varchar(64) COLLATE utf8mb4_unicode_ci DEFAULT '' COMMENT '模型类型：BATCH/STREAMING',
  `file_source_id` bigint(20) DEFAULT NULL COMMENT '文件名规则 ID',
  `creator_id` varchar(256) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT '创建人 ID',
  `features` text COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '额外信息',
  `delete_yn` tinyint(1) NOT NULL DEFAULT '0' COMMENT '逻辑删除标记',
  `created_at` datetime NOT NULL COMMENT '表记录创建时间',
  `updated_at` datetime NOT NULL COMMENT '表记录更新时间',
  `data_catalog_id` bigint(20) NOT NULL DEFAULT '0' COMMENT '数据目录ID',
  `label0` bigint(20) NOT NULL DEFAULT '0' COMMENT '标签 ID',
  `label1` bigint(20) NOT NULL DEFAULT '0' COMMENT '标签 ID',
  `label2` bigint(20) NOT NULL DEFAULT '0' COMMENT '标签 ID',
  `label3` bigint(20) NOT NULL DEFAULT '0' COMMENT '标签 ID',
  `label4` bigint(20) NOT NULL DEFAULT '0' COMMENT '标签 ID',
  `label5` bigint(20) NOT NULL DEFAULT '0' COMMENT '标签 ID',
  `label6` bigint(20) NOT NULL DEFAULT '0' COMMENT '标签 ID',
  `label7` bigint(20) NOT NULL DEFAULT '0' COMMENT '标签 ID',
  `label8` bigint(20) NOT NULL DEFAULT '0' COMMENT '标签 ID',
  `label9` bigint(20) NOT NULL DEFAULT '0' COMMENT '标签 ID',
  `cluster_id` bigint(20) DEFAULT NULL COMMENT '集群id',
  `project_id` bigint(20) DEFAULT '1' COMMENT '项目id',
  `org_id` bigint(20) DEFAULT NULL COMMENT '集群id',
  `application_id` bigint(20) DEFAULT NULL COMMENT '应用id',
  `cluster_name` varchar(256) COLLATE utf8mb4_unicode_ci DEFAULT '' COMMENT '集群名称',
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_model_name_under_source` (`data_source_id`,`name`,`project_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='数据目录依赖表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `fdp_agent_data_model_metadata`
--

DROP TABLE IF EXISTS `fdp_agent_data_model_metadata`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `fdp_agent_data_model_metadata` (
  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT COMMENT '主键id',
  `data_model_id` bigint(20) NOT NULL DEFAULT '0' COMMENT '关联的数据模型 ID',
  `seq_in_model` int(11) NOT NULL COMMENT '字段在模型中的顺序',
  `column_name` varchar(128) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT '字段名',
  `column_type` varchar(128) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT '字段类型',
  `column_length` int(11) NOT NULL DEFAULT '-1' COMMENT '字段长度',
  `column_comment` varchar(1024) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT '字段说明',
  `is_primary_key` tinyint(1) NOT NULL DEFAULT '0' COMMENT '字段是否是主键',
  `allow_nullable` tinyint(1) NOT NULL DEFAULT '0' COMMENT '字段是否允许为空',
  `is_partition_key` tinyint(1) NOT NULL DEFAULT '0' COMMENT '字段是否是分区键',
  `creator_id` varchar(128) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT '创建者 ID',
  `updater_id` varchar(128) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT '更新者 ID',
  `quality_rules_ids` varchar(256) COLLATE utf8mb4_unicode_ci DEFAULT '' COMMENT '数据质量规则ID',
  `created_at` datetime NOT NULL COMMENT '创建时间',
  `updated_at` datetime NOT NULL COMMENT '更新时间',
  `is_encrypted` tinyint(1) NOT NULL DEFAULT '0' COMMENT '是否加密 0否 1是',
  `sensitive_rule_id` varchar(20) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '0' COMMENT '脱敏规>则id',
  `is_pii` tinyint(1) NOT NULL DEFAULT '0' COMMENT '是否为pii数据 0否 1是',
  `cluster_id` bigint(20) DEFAULT NULL COMMENT '集群id',
  `project_id` bigint(20) DEFAULT '1' COMMENT '项目id',
  `org_id` bigint(20) DEFAULT NULL COMMENT '集群id',
  `application_id` bigint(20) DEFAULT NULL COMMENT '应用id',
  `cluster_name` varchar(256) COLLATE utf8mb4_unicode_ci DEFAULT '' COMMENT '集群名称',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_model_column_name` (`data_model_id`,`column_name`,`project_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci BLOCK_FORMAT=ENCRYPTED COMMENT='数据模型元数据表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `fdp_agent_data_model_metadata_rel_quality`
--

DROP TABLE IF EXISTS `fdp_agent_data_model_metadata_rel_quality`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `fdp_agent_data_model_metadata_rel_quality` (
  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT COMMENT '主键id',
  `metadata_id` bigint(20) NOT NULL COMMENT '字段ID',
  `model_id` bigint(20) NOT NULL COMMENT '数据模型 id',
  `quality_id` bigint(20) NOT NULL COMMENT '数据质量ID',
  `created_at` datetime NOT NULL COMMENT '表记录创建时间',
  `updated_at` datetime NOT NULL COMMENT '表记录更新时间',
  `cluster_id` bigint(20) DEFAULT NULL COMMENT '集群id',
  `project_id` bigint(20) DEFAULT '1' COMMENT '项目id',
  `org_id` bigint(20) DEFAULT NULL COMMENT '集群id',
  `application_id` bigint(20) DEFAULT NULL COMMENT '应用id',
  `cluster_name` varchar(256) COLLATE utf8mb4_unicode_ci DEFAULT '' COMMENT '集群名称',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci BLOCK_FORMAT=ENCRYPTED COMMENT='数据模型元数据质量关联表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `fdp_agent_data_models`
--

DROP TABLE IF EXISTS `fdp_agent_data_models`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `fdp_agent_data_models` (
  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT COMMENT '主键id',
  `source` varchar(64) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT '平台来源：DL/CDP',
  `type_in_cdp` varchar(64) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT '数据模型在 CDP 下的类型：ods, dim, dwd, dws, ads',
  `name` varchar(128) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT '数据模型名称，在一个 source 下 name 需要唯一',
  `description` varchar(1024) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT '数据模型描述',
  `data_source_id` bigint(20) NOT NULL COMMENT '数据源 ID',
  `data_source_category` varchar(1024) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT '数据源分类: INTERNAL, EXTERNAL',
  `model_type` varchar(64) COLLATE utf8mb4_unicode_ci DEFAULT '' COMMENT '模型类型：BATCH/STREAMING',
  `file_source_id` bigint(20) DEFAULT NULL COMMENT '文件名规则 ID',
  `creator_id` varchar(256) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT '创建人 ID',
  `features` text COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '额外信息',
  `delete_yn` tinyint(1) NOT NULL DEFAULT '0' COMMENT '逻辑删除标记',
  `created_at` datetime NOT NULL COMMENT '表记录创建时间',
  `updated_at` datetime NOT NULL COMMENT '表记录更新时间',
  `label0` bigint(20) NOT NULL DEFAULT '0' COMMENT '标签 ID',
  `label1` bigint(20) NOT NULL DEFAULT '0' COMMENT '标签 ID',
  `label2` bigint(20) NOT NULL DEFAULT '0' COMMENT '标签 ID',
  `label3` bigint(20) NOT NULL DEFAULT '0' COMMENT '标签 ID',
  `label4` bigint(20) NOT NULL DEFAULT '0' COMMENT '标签 ID',
  `label5` bigint(20) NOT NULL DEFAULT '0' COMMENT '标签 ID',
  `label6` bigint(20) NOT NULL DEFAULT '0' COMMENT '标签 ID',
  `label7` bigint(20) NOT NULL DEFAULT '0' COMMENT '标签 ID',
  `label8` bigint(20) NOT NULL DEFAULT '0' COMMENT '标签 ID',
  `label9` bigint(20) NOT NULL DEFAULT '0' COMMENT '标签 ID',
  `data_catalog_id` bigint(20) DEFAULT '0' COMMENT '数据目录id',
  `subject_id` bigint(20) DEFAULT NULL COMMENT '主题域id',
  `is_physical` tinyint(1) DEFAULT NULL COMMENT '物理化标识',
  `cluster_id` bigint(20) DEFAULT NULL COMMENT '集群id',
  `project_id` bigint(20) DEFAULT '1' COMMENT '项目id',
  `org_id` bigint(20) DEFAULT NULL COMMENT '集群id',
  `application_id` bigint(20) DEFAULT NULL COMMENT '应用id',
  `cluster_name` varchar(256) COLLATE utf8mb4_unicode_ci DEFAULT '' COMMENT '集群名称',
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_model_name_under_source` (`data_source_id`,`name`,`delete_yn`,`project_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci BLOCK_FORMAT=ENCRYPTED COMMENT='数据模型表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `fdp_agent_data_service_generate_api_config`
--

DROP TABLE IF EXISTS `fdp_agent_data_service_generate_api_config`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `fdp_agent_data_service_generate_api_config` (
  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT COMMENT '主键id',
  `org_id` bigint(20) DEFAULT NULL COMMENT '企业 id',
  `api_name` varchar(64) NOT NULL COMMENT 'api 名称',
  `api_path` varchar(128) NOT NULL COMMENT 'api 路径',
  `request_type` varchar(32) NOT NULL COMMENT '请求方式：GET/POST/PUT/DELETE',
  `response_type` varchar(32) NOT NULL COMMENT '返回类型: 目前只支持JSON',
  `description` varchar(1024) DEFAULT NULL COMMENT '描述',
  `data_source_type` varchar(32) NOT NULL COMMENT '数据源类型: MYSQL/CASSANDRA',
  `data_source_name` varchar(128) NOT NULL COMMENT '数据源名称',
  `data_model_name` varchar(128) DEFAULT NULL COMMENT '数据表名称',
  `user_define_sql` text COMMENT '用户自定义sql',
  `request_params` text COMMENT '生成API 请求参数',
  `response_params` text COMMENT '生成API 返回参数',
  `is_paging` tinyint(1) DEFAULT '0' COMMENT '是否分页',
  `compute_engine` varchar(32) DEFAULT NULL COMMENT '计算引擎',
  `publish_status` varchar(32) DEFAULT NULL COMMENT '发布状态',
  `package_id` varchar(256) DEFAULT '' COMMENT '流量包id',
  `bind_domain` varchar(256) DEFAULT '' COMMENT '绑定域名',
  `api_call_times` bigint(20) DEFAULT NULL COMMENT 'api调用次数',
  `delete_yn` tinyint(1) NOT NULL DEFAULT '0' COMMENT '逻辑删除标记',
  `creator_id` varchar(128) DEFAULT NULL COMMENT '创建人 ID',
  `updater_id` varchar(128) DEFAULT NULL COMMENT '更新者ID',
  `created_at` datetime NOT NULL COMMENT '创建时间',
  `updated_at` datetime NOT NULL COMMENT '更新时间',
  `cluster_id` bigint(20) DEFAULT NULL COMMENT '集群id',
  `project_id` bigint(20) DEFAULT '1' COMMENT '项目id',
  `application_id` bigint(20) DEFAULT NULL COMMENT '应用id',
  `cluster_name` varchar(256) DEFAULT '' COMMENT '集群名称',
  PRIMARY KEY (`id`),
  KEY `key_api_path` (`api_path`),
  KEY `key_api_name` (`api_name`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8 COMMENT='Api配置表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `fdp_agent_data_sources`
--

DROP TABLE IF EXISTS `fdp_agent_data_sources`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `fdp_agent_data_sources` (
  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT COMMENT '主键id',
  `org_id` bigint(20) DEFAULT NULL COMMENT '组织 ID',
  `name` varchar(128) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT '数据源名称',
  `description` varchar(1024) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT '数据源描述',
  `source` varchar(64) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT '数据源来源: DL, CDP',
  `category` varchar(1024) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT '数据源分类: INTERNAL, EXTERNAL',
  `type` varchar(64) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT '数据源类型: MYSQL, FILE, SFTP, API',
  `host` varchar(1024) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT '数据源连接地址',
  `port` varchar(64) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT '连接端口',
  `db` varchar(256) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT '数据库',
  `user` varchar(256) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT '用户名',
  `pass` varchar(256) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT '密码',
  `endpoint` varchar(1024) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT 'API 连接终端地址',
  `app_key` varchar(64) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT 'API key',
  `app_secret` varchar(256) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT 'API secret',
  `public_key` text COLLATE utf8mb4_unicode_ci COMMENT '公钥',
  `private_key` text COLLATE utf8mb4_unicode_ci COMMENT '私钥',
  `file_source_id` bigint(20) NOT NULL DEFAULT '0' COMMENT '文件来源 ID',
  `creator_id` varchar(256) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT '创建人 ID',
  `delete_yn` tinyint(1) NOT NULL DEFAULT '0' COMMENT '逻辑删除标记',
  `created_at` datetime NOT NULL COMMENT '表记录创建时间',
  `updated_at` datetime NOT NULL COMMENT '表记录更新时间',
  `label0` bigint(20) NOT NULL DEFAULT '0' COMMENT '标签 ID',
  `label1` bigint(20) NOT NULL DEFAULT '0' COMMENT '标签 ID',
  `label2` bigint(20) NOT NULL DEFAULT '0' COMMENT '标签 ID',
  `label3` bigint(20) NOT NULL DEFAULT '0' COMMENT '标签 ID',
  `label4` bigint(20) NOT NULL DEFAULT '0' COMMENT '标签 ID',
  `label5` bigint(20) NOT NULL DEFAULT '0' COMMENT '标签 ID',
  `label6` bigint(20) NOT NULL DEFAULT '0' COMMENT '标签 ID',
  `label7` bigint(20) NOT NULL DEFAULT '0' COMMENT '标签 ID',
  `label8` bigint(20) NOT NULL DEFAULT '0' COMMENT '标签 ID',
  `label9` bigint(20) NOT NULL DEFAULT '0' COMMENT '标签 ID',
  `cluster_id` bigint(20) DEFAULT NULL COMMENT '集群id',
  `project_id` bigint(20) DEFAULT '1' COMMENT '项目id',
  `application_id` bigint(20) DEFAULT NULL COMMENT '应用id',
  `cluster_name` varchar(256) COLLATE utf8mb4_unicode_ci DEFAULT '' COMMENT '集群名称',
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_name` (`name`)
) ENGINE=InnoDB AUTO_INCREMENT=4 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci BLOCK_FORMAT=ENCRYPTED COMMENT='数据源表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `fdp_agent_dl_data_application`
--

DROP TABLE IF EXISTS `fdp_agent_dl_data_application`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `fdp_agent_dl_data_application` (
  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT COMMENT '主键id',
  `source` varchar(64) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT '平台来源：DL/CDP',
  `apply_name` varchar(64) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT '申请名称',
  `apply_type` varchar(64) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT '申请类型:DELETE/UPDATE/EXPORT',
  `apply_sql` text COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '相关sql',
  `data_source_id` bigint(20) NOT NULL COMMENT '数据源 ID',
  `expect_affected_lines` bigint(20) NOT NULL COMMENT '预计影响行数',
  `apply_reason` text COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '申请原因',
  `apply_result` varchar(64) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT '申请结果: BE_AUDITED/PASSED/REJECTED',
  `execute_status` varchar(64) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT '执行状态 WAIT_START/RUNNING/SUCCESS/FAILED',
  `execute_result` text COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '执行结果',
  `flow_id` varchar(256) COLLATE utf8mb4_unicode_ci DEFAULT '' COMMENT '审批流程ID',
  `attachments` text COLLATE utf8mb4_unicode_ci COMMENT '附件信息',
  `creator_id` varchar(255) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT '创建人 ID',
  `created_at` datetime NOT NULL COMMENT '表记录创建时间',
  `updated_at` datetime NOT NULL COMMENT '表记录更新时间',
  `cluster_id` bigint(20) DEFAULT NULL COMMENT '集群id',
  `project_id` bigint(20) DEFAULT '1' COMMENT '项目id',
  `org_id` bigint(20) DEFAULT NULL COMMENT '集群id',
  `application_id` bigint(20) DEFAULT NULL COMMENT '应用id',
  `cluster_name` varchar(256) COLLATE utf8mb4_unicode_ci DEFAULT '' COMMENT '集群名称',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='数据应用申请表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `fdp_agent_dl_data_model_indexes`
--

DROP TABLE IF EXISTS `fdp_agent_dl_data_model_indexes`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `fdp_agent_dl_data_model_indexes` (
  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT COMMENT '主键id',
  `data_model_id` bigint(20) NOT NULL COMMENT '关联的数据模型 ID',
  `index_name` varchar(255) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT '索引名',
  `index_column` varchar(128) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT '索引字段名',
  `is_unique` tinyint(1) NOT NULL COMMENT '是否是唯一索引',
  `seq_in_index` int(11) NOT NULL COMMENT '字段在联合索引中的顺序',
  `creator_id` varchar(128) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT '创建者 ID',
  `updater_id` varchar(128) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT '更新者 ID',
  `created_at` datetime NOT NULL COMMENT '创建时间',
  `updated_at` datetime NOT NULL COMMENT '更新时间',
  `cluster_id` bigint(20) DEFAULT NULL COMMENT '集群id',
  `project_id` bigint(20) DEFAULT '1' COMMENT '项目id',
  `org_id` bigint(20) DEFAULT NULL COMMENT '集群id',
  `application_id` bigint(20) DEFAULT NULL COMMENT '应用id',
  `cluster_name` varchar(256) COLLATE utf8mb4_unicode_ci DEFAULT '' COMMENT '集群名称',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci BLOCK_FORMAT=ENCRYPTED COMMENT='数据模型索引表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `fdp_agent_dl_data_model_metadata`
--

DROP TABLE IF EXISTS `fdp_agent_dl_data_model_metadata`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `fdp_agent_dl_data_model_metadata` (
  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT COMMENT '主键id',
  `data_model_id` bigint(20) NOT NULL DEFAULT '0' COMMENT '关联的数据模型 ID',
  `seq_in_model` int(11) NOT NULL DEFAULT '0' COMMENT '字段在模型中的顺序',
  `column_name` varchar(128) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT '字段名',
  `column_type` varchar(128) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT '字段类型',
  `column_length` int(11) NOT NULL DEFAULT '-1' COMMENT '字段长度',
  `column_comment` varchar(1024) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT '字段说明',
  `is_primary_key` tinyint(1) NOT NULL DEFAULT '0' COMMENT '字段是否是主键',
  `allow_nullable` tinyint(1) NOT NULL DEFAULT '0' COMMENT '字段是否允许为空',
  `is_partition_key` tinyint(1) NOT NULL DEFAULT '0' COMMENT '字段是否是分区键',
  `column_default` varchar(1024) COLLATE utf8mb4_unicode_ci DEFAULT '' COMMENT '字段默认值',
  `column_extra` varchar(1024) COLLATE utf8mb4_unicode_ci DEFAULT '' COMMENT '字段额外信息',
  `creator_id` varchar(128) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT '创建者 ID',
  `updater_id` varchar(128) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT '更新者 ID',
  `created_at` datetime NOT NULL COMMENT '创建时间',
  `updated_at` datetime NOT NULL COMMENT '更新时间',
  `is_encrypted` tinyint(1) NOT NULL DEFAULT '0' COMMENT '是否加密 0否 1是',
  `sensitive_rule_id` varchar(20) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '0' COMMENT '脱敏规则id',
  `is_pii` tinyint(1) NOT NULL DEFAULT '0' COMMENT '是否为pii数据 0否 1是',
  `cluster_id` bigint(20) DEFAULT NULL COMMENT '集群id',
  `project_id` bigint(20) DEFAULT '1' COMMENT '项目id',
  `org_id` bigint(20) DEFAULT NULL COMMENT '集群id',
  `application_id` bigint(20) DEFAULT NULL COMMENT '应用id',
  `cluster_name` varchar(256) COLLATE utf8mb4_unicode_ci DEFAULT '' COMMENT '集群名称',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_model_column_name` (`data_model_id`,`column_name`,`project_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci BLOCK_FORMAT=ENCRYPTED COMMENT='数据模型元数据表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `fdp_agent_dl_data_model_sensitive_rule`
--

DROP TABLE IF EXISTS `fdp_agent_dl_data_model_sensitive_rule`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `fdp_agent_dl_data_model_sensitive_rule` (
  `id` bigint(20) NOT NULL COMMENT '主键id',
  `name` varchar(255) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '脱敏名称',
  `describe` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '描述',
  `seq` int(11) NOT NULL COMMENT '序号',
  `regex` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '正则匹配',
  `replacement` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '替换规则',
  `tenant_id` bigint(20) DEFAULT NULL COMMENT '商户id',
  `delete_yn` tinyint(1) NOT NULL DEFAULT '0' COMMENT '删除标记 0:未删除 1:已删除',
  `cluster_id` bigint(20) DEFAULT NULL COMMENT '集群id',
  `project_id` bigint(20) DEFAULT '1' COMMENT '项目id',
  `org_id` bigint(20) DEFAULT NULL COMMENT '集群id',
  `application_id` bigint(20) DEFAULT NULL COMMENT '应用id',
  `cluster_name` varchar(256) COLLATE utf8mb4_unicode_ci DEFAULT '' COMMENT '集群名称',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='脱敏规则表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `fdp_agent_dl_data_models`
--

DROP TABLE IF EXISTS `fdp_agent_dl_data_models`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `fdp_agent_dl_data_models` (
  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT COMMENT '主键id',
  `source` varchar(64) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT '平台来源：DL/CDP',
  `type_in_cdp` varchar(64) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT '数据模型在 CDP 下的类型：ods, dim, dwd, dws, ads',
  `name` varchar(128) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT '数据模型名称，在一个 source 下 name 需要唯一',
  `description` varchar(1024) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT '数据模型描述',
  `data_source_id` bigint(20) NOT NULL COMMENT '数据源 ID',
  `data_source_category` varchar(1024) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT '数据源分类: INTERNAL, EXTERNAL',
  `file_source_id` bigint(20) DEFAULT NULL COMMENT '文件名规则 ID',
  `creator_id` varchar(256) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT '创建人 ID',
  `features` text COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '额外信息',
  `delete_yn` tinyint(1) NOT NULL DEFAULT '0' COMMENT '逻辑删除标记',
  `delete_duration` bigint(20) DEFAULT NULL COMMENT '删除时长',
  `created_at` datetime NOT NULL COMMENT '表记录创建时间',
  `updated_at` datetime NOT NULL COMMENT '表记录更新时间',
  `label0` bigint(20) NOT NULL DEFAULT '0' COMMENT '标签 ID',
  `label1` bigint(20) NOT NULL DEFAULT '0' COMMENT '标签 ID',
  `label2` bigint(20) NOT NULL DEFAULT '0' COMMENT '标签 ID',
  `label3` bigint(20) NOT NULL DEFAULT '0' COMMENT '标签 ID',
  `label4` bigint(20) NOT NULL DEFAULT '0' COMMENT '标签 ID',
  `label5` bigint(20) NOT NULL DEFAULT '0' COMMENT '标签 ID',
  `label6` bigint(20) NOT NULL DEFAULT '0' COMMENT '标签 ID',
  `label7` bigint(20) NOT NULL DEFAULT '0' COMMENT '标签 ID',
  `label8` bigint(20) NOT NULL DEFAULT '0' COMMENT '标签 ID',
  `label9` bigint(20) NOT NULL DEFAULT '0' COMMENT '标签 ID',
  `cluster_id` bigint(20) DEFAULT NULL COMMENT '集群id',
  `project_id` bigint(20) DEFAULT '1' COMMENT '项目id',
  `org_id` bigint(20) DEFAULT NULL COMMENT '集群id',
  `application_id` bigint(20) DEFAULT NULL COMMENT '应用id',
  `cluster_name` varchar(256) COLLATE utf8mb4_unicode_ci DEFAULT '' COMMENT '集群名称',
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_model_name_under_source` (`data_source_id`,`name`,`delete_yn`,`project_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci BLOCK_FORMAT=ENCRYPTED COMMENT='数据模型表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `fdp_agent_dl_data_models_config`
--

DROP TABLE IF EXISTS `fdp_agent_dl_data_models_config`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `fdp_agent_dl_data_models_config` (
  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT COMMENT '主键id',
  `data_model_id` bigint(20) NOT NULL COMMENT '关联的数据模型 ID',
  `delete_yn` tinyint(1) NOT NULL DEFAULT '0' COMMENT '逻辑删除标记',
  `creator_id` varchar(255) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT '创建人 ID',
  `updater_id` varchar(255) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT '修改人 ID',
  `created_at` datetime NOT NULL COMMENT '表记录创建时间',
  `updated_at` datetime NOT NULL COMMENT '表记录更新时间',
  `cluster_id` bigint(20) DEFAULT NULL COMMENT '集群id',
  `project_id` bigint(20) DEFAULT '1' COMMENT '项目id',
  `org_id` bigint(20) DEFAULT NULL COMMENT '集群id',
  `application_id` bigint(20) DEFAULT NULL COMMENT '应用id',
  `cluster_name` varchar(256) COLLATE utf8mb4_unicode_ci DEFAULT '' COMMENT '集群名称',
  PRIMARY KEY (`id`),
  KEY `indexs` (`data_model_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='数据模型配置表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `fdp_agent_dl_data_models_statistics`
--

DROP TABLE IF EXISTS `fdp_agent_dl_data_models_statistics`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `fdp_agent_dl_data_models_statistics` (
  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT COMMENT '主键id',
  `data_model_id` bigint(20) NOT NULL COMMENT '数据模型 ID',
  `statistical_date` varchar(64) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT '统计日期',
  `data_model_volume` bigint(20) NOT NULL COMMENT '数据模型统计量',
  `delete_yn` tinyint(1) NOT NULL DEFAULT '0' COMMENT '逻辑删除标记',
  `created_at` datetime NOT NULL COMMENT '表记录创建时间',
  `updated_at` datetime NOT NULL COMMENT '表记录更新时间',
  `cluster_id` bigint(20) DEFAULT NULL COMMENT '集群id',
  `project_id` bigint(20) DEFAULT '1' COMMENT '项目id',
  `org_id` bigint(20) DEFAULT NULL COMMENT '集群id',
  `application_id` bigint(20) DEFAULT NULL COMMENT '应用id',
  `cluster_name` varchar(256) COLLATE utf8mb4_unicode_ci DEFAULT '' COMMENT '集群名称',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_model_id_date` (`data_model_id`,`statistical_date`,`project_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='数据模型统计表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `fdp_agent_dl_data_sources`
--

DROP TABLE IF EXISTS `fdp_agent_dl_data_sources`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `fdp_agent_dl_data_sources` (
  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT COMMENT '主键id',
  `name` varchar(128) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT '数据源名称',
  `description` varchar(1024) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT '数据源描述',
  `source` varchar(64) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT '数据源来源: DL, CDP',
  `category` varchar(1024) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT '数据源分类: INTERNAL, EXTERNAL',
  `type` varchar(64) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT '数据源类型: MYSQL, FILE, SFTP, API',
  `host` varchar(1024) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT '数据源连接地址',
  `port` varchar(64) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT '连接端口',
  `db` varchar(256) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT '数据库',
  `user` varchar(256) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT '用户名',
  `pass` varchar(256) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT '密码',
  `endpoint` varchar(1024) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT 'API 连接终端地址',
  `app_key` varchar(64) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT 'API key',
  `app_secret` varchar(256) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT 'API secret',
  `public_key` text COLLATE utf8mb4_unicode_ci COMMENT '公钥',
  `private_key` text COLLATE utf8mb4_unicode_ci COMMENT '私钥',
  `server_id` varchar(256) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '服务器ID',
  `server_name` varchar(256) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '服务器名称',
  `file_source_id` bigint(20) NOT NULL DEFAULT '0' COMMENT '文件来源 ID',
  `creator_id` varchar(256) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT '创建人 ID',
  `delete_yn` tinyint(1) NOT NULL DEFAULT '0' COMMENT '逻辑删除标记',
  `created_at` datetime NOT NULL COMMENT '表记录创建时间',
  `updated_at` datetime NOT NULL COMMENT '表记录更新时间',
  `label0` bigint(20) NOT NULL DEFAULT '0' COMMENT '标签 ID',
  `label1` bigint(20) NOT NULL DEFAULT '0' COMMENT '标签 ID',
  `label2` bigint(20) NOT NULL DEFAULT '0' COMMENT '标签 ID',
  `label3` bigint(20) NOT NULL DEFAULT '0' COMMENT '标签 ID',
  `label4` bigint(20) NOT NULL DEFAULT '0' COMMENT '标签 ID',
  `label5` bigint(20) NOT NULL DEFAULT '0' COMMENT '标签 ID',
  `label6` bigint(20) NOT NULL DEFAULT '0' COMMENT '标签 ID',
  `label7` bigint(20) NOT NULL DEFAULT '0' COMMENT '标签 ID',
  `label8` bigint(20) NOT NULL DEFAULT '0' COMMENT '标签 ID',
  `label9` bigint(20) NOT NULL DEFAULT '0' COMMENT '标签 ID',
  `custom` varchar(1024) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '额外配置',
  `cluster_id` bigint(20) DEFAULT NULL COMMENT '集群id',
  `project_id` bigint(20) DEFAULT '1' COMMENT '项目id',
  `org_id` bigint(20) DEFAULT NULL COMMENT '集群id',
  `application_id` bigint(20) DEFAULT NULL COMMENT '应用id',
  `cluster_name` varchar(256) COLLATE utf8mb4_unicode_ci DEFAULT '' COMMENT '集群名称',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci BLOCK_FORMAT=ENCRYPTED COMMENT='数据源表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `fdp_agent_dl_disks`
--

DROP TABLE IF EXISTS `fdp_agent_dl_disks`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `fdp_agent_dl_disks` (
  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT COMMENT '主键id',
  `name` varchar(128) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT '磁盘名',
  `name_pinyin` varchar(128) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT '磁盘名拼音',
  `type` varchar(64) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT '磁盘类型: SYSTEM, NORMAL',
  `creator_id` varchar(256) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT '创建人 ID',
  `delete_yn` tinyint(1) NOT NULL DEFAULT '0' COMMENT '逻辑删除标记',
  `created_at` datetime NOT NULL COMMENT '表记录创建时间',
  `updated_at` datetime NOT NULL COMMENT '表记录更新时间',
  `cluster_id` bigint(20) DEFAULT NULL COMMENT '集群id',
  `project_id` bigint(20) DEFAULT '1' COMMENT '项目id',
  `org_id` bigint(20) DEFAULT NULL COMMENT '集群id',
  `application_id` bigint(20) DEFAULT NULL COMMENT '应用id',
  `cluster_name` varchar(256) COLLATE utf8mb4_unicode_ci DEFAULT '' COMMENT '集群名称',
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_name` (`name`)
) ENGINE=InnoDB AUTO_INCREMENT=3 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci BLOCK_FORMAT=ENCRYPTED COMMENT='磁盘表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `fdp_agent_dl_file_sources`
--

DROP TABLE IF EXISTS `fdp_agent_dl_file_sources`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `fdp_agent_dl_file_sources` (
  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT COMMENT '主键id',
  `name` varchar(128) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT '文件来源名',
  `creator_id` varchar(256) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT '创建人 ID',
  `delete_yn` tinyint(1) NOT NULL DEFAULT '0' COMMENT '逻辑删除标记',
  `created_at` datetime NOT NULL COMMENT '表记录创建时间',
  `updated_at` datetime NOT NULL COMMENT '表记录更新时间',
  `cluster_id` bigint(20) DEFAULT NULL COMMENT '集群id',
  `project_id` bigint(20) DEFAULT '1' COMMENT '项目id',
  `org_id` bigint(20) DEFAULT NULL COMMENT '集群id',
  `application_id` bigint(20) DEFAULT NULL COMMENT '应用id',
  `cluster_name` varchar(256) COLLATE utf8mb4_unicode_ci DEFAULT '' COMMENT '集群名称',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci BLOCK_FORMAT=ENCRYPTED COMMENT='文件来源表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `fdp_agent_dl_file_upload_rules`
--

DROP TABLE IF EXISTS `fdp_agent_dl_file_upload_rules`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `fdp_agent_dl_file_upload_rules` (
  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT COMMENT '主键id',
  `type` varchar(64) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT '文件上传规则类型',
  `value` varchar(64) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT '规则内容',
  `creator_id` varchar(256) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT '创建人 ID',
  `delete_yn` tinyint(1) NOT NULL DEFAULT '0' COMMENT '逻辑删除标记',
  `created_at` datetime NOT NULL COMMENT '表记录创建时间',
  `updated_at` datetime NOT NULL COMMENT '表记录更新时间',
  `cluster_id` bigint(20) DEFAULT NULL COMMENT '集群id',
  `project_id` bigint(20) DEFAULT '1' COMMENT '项目id',
  `org_id` bigint(20) DEFAULT NULL COMMENT '集群id',
  `application_id` bigint(20) DEFAULT NULL COMMENT '应用id',
  `cluster_name` varchar(256) COLLATE utf8mb4_unicode_ci DEFAULT '' COMMENT '集群名称',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci BLOCK_FORMAT=ENCRYPTED COMMENT='文件上传规则表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `fdp_agent_dl_files`
--

DROP TABLE IF EXISTS `fdp_agent_dl_files`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `fdp_agent_dl_files` (
  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT COMMENT '主键id',
  `name` varchar(128) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT '文件名',
  `name_pinyin` varchar(128) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT '文件名拼音',
  `disk_id` bigint(20) NOT NULL COMMENT '磁盘 ID',
  `type` varchar(64) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT '磁盘类型: SYSTEM, NORMAL',
  `folder_id` bigint(20) NOT NULL COMMENT '目录 ID',
  `url` varchar(1024) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT '文件下载链接',
  `size` bigint(20) NOT NULL COMMENT '文件大小',
  `suffix` varchar(64) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT '文件后缀',
  `creator_id` varchar(256) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT '创建人 ID',
  `delete_yn` tinyint(1) NOT NULL DEFAULT '0' COMMENT '逻辑删除标记',
  `created_at` datetime NOT NULL COMMENT '表记录创建时间',
  `updated_at` datetime NOT NULL COMMENT '表记录更新时间',
  `cluster_id` bigint(20) DEFAULT NULL COMMENT '集群id',
  `project_id` bigint(20) DEFAULT '1' COMMENT '项目id',
  `org_id` bigint(20) DEFAULT NULL COMMENT '集群id',
  `application_id` bigint(20) DEFAULT NULL COMMENT '应用id',
  `cluster_name` varchar(256) COLLATE utf8mb4_unicode_ci DEFAULT '' COMMENT '集群名称',
  PRIMARY KEY (`id`),
  KEY `idx_disk_id_folder_id_name` (`disk_id`,`folder_id`,`name`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci BLOCK_FORMAT=ENCRYPTED COMMENT='文件表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `fdp_agent_dl_labels`
--

DROP TABLE IF EXISTS `fdp_agent_dl_labels`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `fdp_agent_dl_labels` (
  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT COMMENT '主键id',
  `name` varchar(64) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT '标签名',
  `color` varchar(64) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT '标签颜色',
  `module` varchar(64) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT '标签所属模块: DATA_SOURCE, DATA_MODEL',
  `creator_id` varchar(256) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT '创建人 ID',
  `delete_yn` tinyint(1) NOT NULL DEFAULT '0' COMMENT '逻辑删除标记',
  `created_at` datetime NOT NULL COMMENT '表记录创建时间',
  `updated_at` datetime NOT NULL COMMENT '表记录更新时间',
  `cluster_id` bigint(20) DEFAULT NULL COMMENT '集群id',
  `project_id` bigint(20) DEFAULT '1' COMMENT '项目id',
  `org_id` bigint(20) DEFAULT NULL COMMENT '集群id',
  `application_id` bigint(20) DEFAULT NULL COMMENT '应用id',
  `cluster_name` varchar(256) COLLATE utf8mb4_unicode_ci DEFAULT '' COMMENT '集群名称',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci BLOCK_FORMAT=ENCRYPTED COMMENT='标签表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `fdp_agent_dl_quality_rules`
--

DROP TABLE IF EXISTS `fdp_agent_dl_quality_rules`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `fdp_agent_dl_quality_rules` (
  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT COMMENT '主键id',
  `name` varchar(64) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT '规则名称',
  `rule_type` varchar(64) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT '规则类型',
  `rule_desc` varchar(64) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '规则描述',
  `rule_content` varchar(64) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT '规则内容',
  `creator_id` varchar(255) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT '创建人 ID',
  `delete_yn` tinyint(1) NOT NULL DEFAULT '0' COMMENT '逻辑删除标记',
  `created_at` datetime NOT NULL COMMENT '表记录创建时间',
  `updated_at` datetime NOT NULL COMMENT '表记录更新时间',
  `cluster_id` bigint(20) DEFAULT NULL COMMENT '集群id',
  `project_id` bigint(20) DEFAULT '1' COMMENT '项目id',
  `org_id` bigint(20) DEFAULT NULL COMMENT '集群id',
  `application_id` bigint(20) DEFAULT NULL COMMENT '应用id',
  `cluster_name` varchar(256) COLLATE utf8mb4_unicode_ci DEFAULT '' COMMENT '集群名称',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci BLOCK_FORMAT=ENCRYPTED COMMENT='质量规则表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `fdp_agent_dl_sql_histories`
--

DROP TABLE IF EXISTS `fdp_agent_dl_sql_histories`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `fdp_agent_dl_sql_histories` (
  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT COMMENT '主键id',
  `data_model_id` bigint(20) NOT NULL COMMENT '数据模型 ID',
  `name` varchar(256) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT '执行记录名称',
  `sql` text COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '执行的 SQL',
  `result` text COLLATE utf8mb4_unicode_ci NOT NULL COMMENT 'SQL 执行结果',
  `creator_id` varchar(256) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT '创建人 ID',
  `created_at` datetime NOT NULL COMMENT '表记录创建时间',
  `updated_at` datetime NOT NULL COMMENT '表记录更新时间',
  `cluster_id` bigint(20) DEFAULT NULL COMMENT '集群id',
  `project_id` bigint(20) DEFAULT '1' COMMENT '项目id',
  `org_id` bigint(20) DEFAULT NULL COMMENT '集群id',
  `application_id` bigint(20) DEFAULT NULL COMMENT '应用id',
  `cluster_name` varchar(256) COLLATE utf8mb4_unicode_ci DEFAULT '' COMMENT '集群名称',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci BLOCK_FORMAT=ENCRYPTED COMMENT='sql历史表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `fdp_agent_dl_task_categories`
--

DROP TABLE IF EXISTS `fdp_agent_dl_task_categories`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `fdp_agent_dl_task_categories` (
  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT COMMENT '主键id',
  `source` varchar(64) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT '平台来源：DL/CDP',
  `name` varchar(128) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT '目录名',
  `name_pinyin` varchar(128) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT '目录名拼音',
  `parent_id` bigint(20) NOT NULL COMMENT '上级目录 ID',
  `creator_id` varchar(256) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT '创建人 ID',
  `created_at` datetime NOT NULL COMMENT '表记录创建时间',
  `updated_at` datetime NOT NULL COMMENT '表记录更新时间',
  `cluster_id` bigint(20) DEFAULT NULL COMMENT '集群id',
  `project_id` bigint(20) DEFAULT '1' COMMENT '项目id',
  `org_id` bigint(20) DEFAULT NULL COMMENT '集群id',
  `application_id` bigint(20) DEFAULT NULL COMMENT '应用id',
  `cluster_name` varchar(256) COLLATE utf8mb4_unicode_ci DEFAULT '' COMMENT '集群名称',
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_parent_id_name` (`parent_id`,`source`,`name`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci BLOCK_FORMAT=ENCRYPTED COMMENT='任务目录表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `fdp_agent_dl_task_runs`
--

DROP TABLE IF EXISTS `fdp_agent_dl_task_runs`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `fdp_agent_dl_task_runs` (
  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT COMMENT '主键id',
  `task_def_id` bigint(20) NOT NULL COMMENT '任务定义 ID',
  `instance_id` bigint(20) NOT NULL COMMENT '任务实例 ID',
  `name` varchar(256) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT '任务名',
  `name_pinyin` varchar(256) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT '任务名拼音',
  `type` varchar(64) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT '类型: 数据集成 Ingestion',
  `sync_type` varchar(64) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT '同步类型: API 同步、文件同步',
  `run_type` varchar(64) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT '运行类型: ONCE、CRON',
  `cron` varchar(128) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT '任务周期',
  `from_data_model_id` bigint(20) NOT NULL COMMENT '模型同步来源：数据模型 ID',
  `to_data_model_id` bigint(20) NOT NULL COMMENT '模型同步目标：数据模型 ID',
  `status` varchar(64) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT '任务执行状态',
  `success_count` int(11) NOT NULL DEFAULT '0' COMMENT '成功数',
  `failure_count` int(11) NOT NULL DEFAULT '0' COMMENT '失败数',
  `total_count` int(11) NOT NULL DEFAULT '0' COMMENT '总数',
  `link` varchar(1024) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '错误内容文本(TXT/JSON)的地址',
  `creator_id` varchar(256) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT '创建者 ID',
  `features` text COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '额外信息',
  `started_at` datetime DEFAULT NULL COMMENT '任务实例开始时间',
  `ended_at` datetime DEFAULT NULL COMMENT '任务实例结束时间',
  `created_at` datetime NOT NULL COMMENT '表记录创建时间',
  `updated_at` datetime NOT NULL COMMENT '表记录更新时间',
  `cluster_id` bigint(20) DEFAULT NULL COMMENT '集群id',
  `project_id` bigint(20) DEFAULT '1' COMMENT '项目id',
  `org_id` bigint(20) DEFAULT NULL COMMENT '集群id',
  `application_id` bigint(20) DEFAULT NULL COMMENT '应用id',
  `cluster_name` varchar(256) COLLATE utf8mb4_unicode_ci DEFAULT '' COMMENT '集群名称',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_task_def_id_instance_id` (`task_def_id`,`instance_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci BLOCK_FORMAT=ENCRYPTED COMMENT='任务运行记录表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `fdp_agent_dl_tasks`
--

DROP TABLE IF EXISTS `fdp_agent_dl_tasks`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `fdp_agent_dl_tasks` (
  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT COMMENT '主键id',
  `name` varchar(128) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT '任务名',
  `name_pinyin` varchar(256) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT '任务名拼音',
  `description` varchar(256) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT '任务描述',
  `category_id` bigint(20) NOT NULL COMMENT '任务目录 ID',
  `type` varchar(64) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT '类型: 数据集成 Ingestion',
  `sync_type` varchar(64) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT '同步类型: API 同步、文件同步',
  `run_type` varchar(64) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT '运行类型: ONCE、CRON',
  `cron` varchar(128) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT '任务周期',
  `from_data_model_id` bigint(20) NOT NULL COMMENT '模型同步来源：数据模型 ID',
  `to_data_model_id` bigint(20) NOT NULL COMMENT '模型同步目标：数据模型 ID',
  `creator_id` varchar(256) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT '创建者 ID',
  `updater_id` varchar(256) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT '修改者 ID',
  `features` text COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '额外信息',
  `created_at` datetime NOT NULL COMMENT '表记录创建时间',
  `updated_at` datetime NOT NULL COMMENT '表记录更新时间',
  `update_field` varchar(64) CHARACTER SET utf8 COLLATE utf8_unicode_ci DEFAULT '' COMMENT '时间更新字段',
  `cluster_id` bigint(20) DEFAULT NULL COMMENT '集群id',
  `project_id` bigint(20) DEFAULT '1' COMMENT '项目id',
  `org_id` bigint(20) DEFAULT NULL COMMENT '集群id',
  `application_id` bigint(20) DEFAULT NULL COMMENT '应用id',
  `cluster_name` varchar(256) COLLATE utf8mb4_unicode_ci DEFAULT '' COMMENT '集群名称',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_category_id_name` (`category_id`,`name`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci BLOCK_FORMAT=ENCRYPTED COMMENT='任务表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `fdp_agent_indicators`
--

DROP TABLE IF EXISTS `fdp_agent_indicators`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `fdp_agent_indicators` (
  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT COMMENT '主键id',
  `workflow_id` bigint(20) NOT NULL COMMENT '工作流Id',
  `workflow_name` varchar(128) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '工作流名称',
  `name` varchar(128) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT '指标名称',
  `table_id` bigint(20) DEFAULT NULL COMMENT '模型id',
  `table_name` varchar(128) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '表名称',
  `field_name` varchar(128) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '表字段名称',
  `created_at` datetime NOT NULL COMMENT '表记录创建时间',
  `updated_at` datetime NOT NULL COMMENT '表记录更新时间',
  `cluster_id` bigint(20) DEFAULT NULL COMMENT '集群id',
  `project_id` bigint(20) DEFAULT '1' COMMENT '项目id',
  `org_id` bigint(20) DEFAULT NULL COMMENT '集群id',
  `application_id` bigint(20) DEFAULT NULL COMMENT '应用id',
  `cluster_name` varchar(256) COLLATE utf8mb4_unicode_ci DEFAULT '' COMMENT '集群名称',
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_name` (`name`,`project_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci BLOCK_FORMAT=ENCRYPTED COMMENT='指标表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `fdp_agent_label_work_relation`
--

DROP TABLE IF EXISTS `fdp_agent_label_work_relation`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `fdp_agent_label_work_relation` (
  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT COMMENT '主键id',
  `label_id` bigint(20) NOT NULL COMMENT '标签名',
  `workflow_id` bigint(20) NOT NULL COMMENT '标签颜色',
  `module` varchar(64) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT '标签所属模块: DATA_SOURCE, DATA_MODEL',
  `creator_id` varchar(256) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT '创建人 ID',
  `delete_yn` tinyint(1) NOT NULL DEFAULT '0' COMMENT '逻辑删除标记',
  `created_at` datetime NOT NULL COMMENT '表记录创建时间',
  `updated_at` datetime NOT NULL COMMENT '表记录更新时间',
  `cluster_id` bigint(20) DEFAULT NULL COMMENT '集群id',
  `project_id` bigint(20) DEFAULT '1' COMMENT '项目id',
  `org_id` bigint(20) DEFAULT NULL COMMENT '集群id',
  `application_id` bigint(20) DEFAULT NULL COMMENT '应用id',
  `cluster_name` varchar(256) COLLATE utf8mb4_unicode_ci DEFAULT '' COMMENT '集群名称',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=226 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='标签工作流关联表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `fdp_agent_labels`
--

DROP TABLE IF EXISTS `fdp_agent_labels`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `fdp_agent_labels` (
  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT COMMENT '主键id',
  `name` varchar(64) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT '标签名',
  `color` varchar(64) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT '标签颜色',
  `module` varchar(64) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT '标签所属模块: DATA_SOURCE, DATA_MODEL',
  `creator_id` varchar(256) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT '创建人 ID',
  `delete_yn` tinyint(1) NOT NULL DEFAULT '0' COMMENT '逻辑删除标记',
  `created_at` datetime NOT NULL COMMENT '表记录创建时间',
  `updated_at` datetime NOT NULL COMMENT '表记录更新时间',
  `cluster_id` bigint(20) DEFAULT NULL COMMENT '集群id',
  `project_id` bigint(20) DEFAULT '1' COMMENT '项目id',
  `org_id` bigint(20) DEFAULT NULL COMMENT '集群id',
  `application_id` bigint(20) DEFAULT NULL COMMENT '应用id',
  `cluster_name` varchar(256) COLLATE utf8mb4_unicode_ci DEFAULT '' COMMENT '集群名称',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci BLOCK_FORMAT=ENCRYPTED COMMENT='标签表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `fdp_agent_model_record_detail`
--

DROP TABLE IF EXISTS `fdp_agent_model_record_detail`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `fdp_agent_model_record_detail` (
  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT COMMENT '主键id',
  `model_id` bigint(20) NOT NULL COMMENT '模型id',
  `model_name` varchar(128) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT '模型名称',
  `record_type` varchar(128) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT '明细类型:表容量等',
  `record_value` double NOT NULL COMMENT '类型值明细值,容量kb',
  `day_value` bigint(20) NOT NULL COMMENT '日期',
  `day_type` varchar(128) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT '日期类型',
  `created_at` datetime NOT NULL COMMENT '表记录创建时间',
  `updated_at` datetime NOT NULL COMMENT '表记录更新时间',
  `cluster_id` bigint(20) DEFAULT NULL COMMENT '集群id',
  `project_id` bigint(20) DEFAULT '1' COMMENT '项目id',
  `org_id` bigint(20) DEFAULT NULL COMMENT '集群id',
  `application_id` bigint(20) DEFAULT NULL COMMENT '应用id',
  `cluster_name` varchar(256) COLLATE utf8mb4_unicode_ci DEFAULT '' COMMENT '集群名称',
  PRIMARY KEY (`id`),
  KEY `idx_model_record` (`model_name`,`day_value`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='模型记录明细表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `fdp_agent_model_workflows`
--

DROP TABLE IF EXISTS `fdp_agent_model_workflows`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `fdp_agent_model_workflows` (
  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT COMMENT '主键id',
  `org_id` bigint(20) DEFAULT NULL COMMENT '企业 id',
  `model_manage_id` bigint(20) NOT NULL COMMENT '模型管理ID',
  `process_type` varchar(128) NOT NULL COMMENT '处理类型',
  `name` varchar(128) NOT NULL DEFAULT '' COMMENT '工作流名',
  `name_pinyin` varchar(128) NOT NULL DEFAULT '' COMMENT '工作流拼音',
  `description` varchar(1024) NOT NULL DEFAULT '' COMMENT '工作流描述',
  `source` varchar(32) DEFAULT '' COMMENT 'workflow 来源: dl/cdp',
  `run_type` varchar(64) NOT NULL DEFAULT '' COMMENT '运行类型: ATONCE、SCHEDULE',
  `cron` varchar(128) DEFAULT '' COMMENT '任务周期',
  `cron_start_from` datetime DEFAULT NULL COMMENT '延时执行时间',
  `pipeline_name` varchar(128) DEFAULT '' COMMENT 'pipeline 名称',
  `pipeline` mediumtext NOT NULL COMMENT 'pipeline 内容',
  `pipeline_id` bigint(20) DEFAULT NULL COMMENT 'pipeline id',
  `node_params` text COMMENT '工作流节点参数',
  `locations` varchar(1024) DEFAULT NULL,
  `creator_id` varchar(128) DEFAULT NULL COMMENT '创建者 ID',
  `updater_id` varchar(128) DEFAULT NULL COMMENT '更新者ID',
  `extra` mediumtext COMMENT '扩展信息',
  `delete_yn` tinyint(1) NOT NULL DEFAULT '0' COMMENT '逻辑删除标记',
  `created_at` datetime DEFAULT NULL COMMENT '创建时间',
  `updated_at` datetime DEFAULT NULL COMMENT '更新时间',
  `cluster_name` varchar(256) DEFAULT NULL COMMENT '集群名',
  `cluster_id` bigint(20) DEFAULT NULL COMMENT '集群id',
  `project_id` bigint(20) DEFAULT NULL COMMENT '项目id',
  `application_id` bigint(20) DEFAULT NULL COMMENT '应用id',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8 COMMENT='模型管理工作流表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `fdp_agent_onedata_atomic_metrics`
--

DROP TABLE IF EXISTS `fdp_agent_onedata_atomic_metrics`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `fdp_agent_onedata_atomic_metrics` (
  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT COMMENT '主键id',
  `metrics_type` varchar(64) NOT NULL DEFAULT '' COMMENT '指标类型',
  `metrics_name` varchar(64) NOT NULL DEFAULT '' COMMENT '指标名称',
  `metrics_label` varchar(64) NOT NULL DEFAULT '' COMMENT '指标命名',
  `data_type` varchar(64) NOT NULL DEFAULT '' COMMENT '数据类型',
  `updated_by` bigint(20) NOT NULL COMMENT '最近修改人',
  `updated_at` datetime NOT NULL COMMENT '最近修改时间',
  `created_by` bigint(20) NOT NULL COMMENT '创建人',
  `created_at` datetime NOT NULL COMMENT '创建时间',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='onedata原子指标表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `fdp_agent_onedata_derived_metrics`
--

DROP TABLE IF EXISTS `fdp_agent_onedata_derived_metrics`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `fdp_agent_onedata_derived_metrics` (
  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT COMMENT '主键id',
  `metrics_name` varchar(64) NOT NULL DEFAULT '' COMMENT '指标名称',
  `metrics_label` varchar(64) NOT NULL DEFAULT '' COMMENT '指标命名',
  `metrics_desc` varchar(64) NOT NULL DEFAULT '' COMMENT '指标描述',
  `data_type` varchar(64) NOT NULL DEFAULT '' COMMENT '数据类型',
  `updated_by` bigint(20) NOT NULL COMMENT '最近修改人',
  `updated_at` datetime NOT NULL COMMENT '最近修改时间',
  `created_by` bigint(20) NOT NULL COMMENT '创建人',
  `created_at` datetime NOT NULL COMMENT '创建时间',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='onedata派生指标表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `fdp_agent_onedata_frequency`
--

DROP TABLE IF EXISTS `fdp_agent_onedata_frequency`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `fdp_agent_onedata_frequency` (
  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT COMMENT '主键id',
  `frequency_name` varchar(64) NOT NULL DEFAULT '' COMMENT '频率名称',
  `frequency_label` varchar(64) NOT NULL DEFAULT '' COMMENT '频率标识',
  `frequency_desc` varchar(64) NOT NULL DEFAULT '' COMMENT '频率描述',
  `updated_by` bigint(20) NOT NULL COMMENT '最近修改人',
  `updated_at` datetime NOT NULL COMMENT '最近修改时间',
  `created_by` bigint(20) NOT NULL COMMENT '创建人',
  `created_at` datetime NOT NULL COMMENT '创建时间',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='onedata频率表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `fdp_agent_onedata_increment`
--

DROP TABLE IF EXISTS `fdp_agent_onedata_increment`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `fdp_agent_onedata_increment` (
  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT COMMENT '主键id',
  `increment_name` varchar(64) NOT NULL DEFAULT '' COMMENT '增量名称',
  `increment_label` varchar(64) NOT NULL DEFAULT '' COMMENT '增量标识',
  `increment_desc` varchar(64) NOT NULL DEFAULT '' COMMENT '增量描述',
  `updated_by` bigint(20) NOT NULL COMMENT '最近修改人',
  `updated_at` datetime NOT NULL COMMENT '最近修改时间',
  `created_by` bigint(20) NOT NULL COMMENT '创建人',
  `created_at` datetime NOT NULL COMMENT '创建时间',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='onedata增量表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `fdp_agent_onedata_metrics`
--

DROP TABLE IF EXISTS `fdp_agent_onedata_metrics`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `fdp_agent_onedata_metrics` (
  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT,
  `category_id` bigint(20) unsigned NOT NULL COMMENT '文件目录ID',
  `workflow_category_id` bigint(20) DEFAULT NULL COMMENT '派生指标创建工作流目录ID',
  `workflow_id` bigint(20) DEFAULT NULL COMMENT '工作流ID',
  `workflow_location` varchar(255) DEFAULT NULL COMMENT '工作流节点坐标',
  `workflow_node_id` bigint(20) DEFAULT NULL COMMENT '工作流节点ID',
  `metrics_name` varchar(256) DEFAULT '' COMMENT '指标名称',
  `metrics_type` varchar(128) DEFAULT '' COMMENT '指标类型, 原子指标/派生指标',
  `metrics_name_en` varchar(256) DEFAULT '' COMMENT '指标英文名称',
  `metrics_def` varchar(256) DEFAULT '' COMMENT '指标定义',
  `metrics_caliber` varchar(128) DEFAULT '' COMMENT '指标口径',
  `selected_method` varchar(256) DEFAULT '' COMMENT '选择方法',
  `metrics_measure` varchar(256) DEFAULT '' COMMENT '计量单位',
  `model_name` varchar(256) DEFAULT '' COMMENT '模型名称',
  `column_name` varchar(256) DEFAULT '' COMMENT '字段名称',
  `process_logic` varchar(256) DEFAULT '' COMMENT '处理逻辑',
  `metrics_sample` varchar(256) DEFAULT '' COMMENT '示例数据',
  `metrics_period` varchar(256) DEFAULT '' COMMENT '统计周期',
  `metrics_restrict` varchar(256) DEFAULT '' COMMENT '业务限定',
  `metrics_granularity` varchar(1024) DEFAULT NULL COMMENT '统计粒度',
  `rel_metrics_id` bigint(20) DEFAULT NULL COMMENT '派生指标关联的原子指标ID',
  `is_referred` tinyint(1) DEFAULT NULL COMMENT '是否被引用',
  `creator_id` varchar(255) DEFAULT NULL COMMENT '创建人 ID',
  `updater_id` varchar(128) DEFAULT NULL COMMENT '更新者 ID',
  `delete_yn` tinyint(1) DEFAULT NULL COMMENT '逻辑删除标记',
  `created_at` datetime NOT NULL COMMENT '表记录创建时间',
  `updated_at` datetime NOT NULL COMMENT '表记录更新时间',
  `org_id` bigint(20) DEFAULT NULL COMMENT '企业 id',
  `cluster_id` bigint(20) DEFAULT NULL COMMENT '集群id',
  `project_id` bigint(20) DEFAULT NULL COMMENT '项目id',
  `application_id` bigint(20) DEFAULT NULL COMMENT '应用id',
  `cluster_name` varchar(256) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '集群名',
  PRIMARY KEY (`id`),
  KEY `FK_category_id` (`category_id`),
  KEY `idx_rel_metrics_id` (`rel_metrics_id`),
  CONSTRAINT `FK_category_id` FOREIGN KEY (`category_id`) REFERENCES `fdp_agent_onedata_metrics_category` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8 COMMENT='指标定义表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `fdp_agent_onedata_metrics_category`
--

DROP TABLE IF EXISTS `fdp_agent_onedata_metrics_category`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `fdp_agent_onedata_metrics_category` (
  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT,
  `subject_id` bigint(20) unsigned NOT NULL COMMENT '主题域 ID',
  `name` varchar(128) NOT NULL DEFAULT '' COMMENT '目录名',
  `name_pinyin` varchar(128) NOT NULL DEFAULT '' COMMENT '目录名拼音',
  `parent_id` bigint(20) NOT NULL COMMENT '上级目录 ID',
  `creator_id` varchar(255) DEFAULT '' COMMENT '创建人 ID',
  `updater_id` varchar(128) DEFAULT '' COMMENT '更新者 ID',
  `delete_yn` tinyint(1) DEFAULT '0' COMMENT '逻辑删除标记',
  `created_at` datetime NOT NULL COMMENT '表记录创建时间',
  `updated_at` datetime NOT NULL COMMENT '表记录更新时间',
  `org_id` bigint(20) DEFAULT NULL COMMENT '企业 id',
  `cluster_id` bigint(20) DEFAULT NULL COMMENT '集群id',
  `project_id` bigint(20) DEFAULT NULL COMMENT '项目id',
  `application_id` bigint(20) DEFAULT NULL COMMENT '应用id',
  `cluster_name` varchar(256) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '集群名',
  PRIMARY KEY (`id`),
  KEY `FK_subject_id` (`subject_id`),
  CONSTRAINT `FK_subject_id` FOREIGN KEY (`subject_id`) REFERENCES `fdp_agent_onedata_subject` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8 COMMENT='指标定义目录表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `fdp_agent_onedata_model_level`
--

DROP TABLE IF EXISTS `fdp_agent_onedata_model_level`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `fdp_agent_onedata_model_level` (
  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT COMMENT '主键id',
  `level_code` bigint(20) NOT NULL DEFAULT '0' COMMENT '层级编号',
  `level_name` varchar(64) NOT NULL DEFAULT '' COMMENT '层级名称',
  `level_desc` varchar(64) NOT NULL DEFAULT '' COMMENT '层级说明',
  `level_prefix` varchar(64) NOT NULL DEFAULT '' COMMENT '层级前缀',
  `level_lifycycle` bigint(20) DEFAULT '365' COMMENT '层级生命周期',
  `is_marked_depend` tinyint(4) DEFAULT '0' COMMENT '是否记入层级依赖',
  `updated_by` bigint(20) NOT NULL COMMENT '最近修改人',
  `updated_at` datetime NOT NULL COMMENT '最近修改时间',
  `created_by` bigint(20) NOT NULL COMMENT '创建人',
  `created_at` datetime NOT NULL COMMENT '创建时间',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='onedata层级表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `fdp_agent_onedata_model_relation`
--

DROP TABLE IF EXISTS `fdp_agent_onedata_model_relation`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `fdp_agent_onedata_model_relation` (
  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT,
  `org_id` bigint(20) DEFAULT NULL COMMENT '企业 id',
  `name` varchar(256) NOT NULL COMMENT '关系网名称',
  `main_model_id` bigint(20) DEFAULT NULL COMMENT '主模型ID',
  `creator_id` varchar(255) DEFAULT '' COMMENT '创建人 ID',
  `updater_id` varchar(128) DEFAULT '' COMMENT '更新者 ID',
  `delete_yn` tinyint(1) NOT NULL DEFAULT '0' COMMENT '逻辑删除标记',
  `created_at` datetime NOT NULL COMMENT '表记录创建时间',
  `updated_at` datetime NOT NULL COMMENT '表记录更新时间',
  `application_id` bigint(20) DEFAULT NULL COMMENT '应用id',
  `platform_id` bigint(20) DEFAULT '0' COMMENT '平台 id',
  `cluster_id` bigint(20) DEFAULT NULL COMMENT '集群id',
  `project_id` bigint(20) DEFAULT NULL COMMENT '项目id',
  `cluster_name` varchar(256) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '集群名',
  PRIMARY KEY (`id`),
  KEY `idx_main_model_id` (`main_model_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `fdp_agent_onedata_model_relation_rel_model_field`
--

DROP TABLE IF EXISTS `fdp_agent_onedata_model_relation_rel_model_field`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `fdp_agent_onedata_model_relation_rel_model_field` (
  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT,
  `org_id` bigint(20) DEFAULT NULL COMMENT '企业 id',
  `model_relation_id` bigint(20) unsigned NOT NULL COMMENT '模型关系网ID',
  `main_model_field` varchar(256) DEFAULT NULL COMMENT '主模型字段',
  `rel_model_id` bigint(20) DEFAULT NULL COMMENT '关联数据模型ID',
  `rel_model_field` varchar(256) DEFAULT NULL COMMENT '关联数据模型字段名称',
  `rel_type` varchar(255) DEFAULT NULL COMMENT '关联类型',
  `process_logic` varchar(256) DEFAULT NULL COMMENT '处理逻辑',
  `creator_id` varchar(256) DEFAULT '' COMMENT '创建人 ID',
  `updater_id` varchar(128) DEFAULT '' COMMENT '更新者 ID',
  `delete_yn` tinyint(1) NOT NULL DEFAULT '0' COMMENT '逻辑删除标记',
  `created_at` datetime NOT NULL COMMENT '表记录创建时间',
  `updated_at` datetime NOT NULL COMMENT '表记录更新时间',
  `project_id` bigint(20) DEFAULT NULL COMMENT '项目id',
  `application_id` bigint(20) DEFAULT NULL COMMENT '应用id',
  `platform_id` bigint(20) DEFAULT '0' COMMENT '平台 id',
  `cluster_id` bigint(20) DEFAULT NULL COMMENT '集群id',
  `cluster_name` varchar(256) DEFAULT '' COMMENT '集群名称',
  PRIMARY KEY (`id`),
  KEY `FK_model_relation_id` (`model_relation_id`),
  KEY `idx_rel_model_field` (`main_model_field`(255),`rel_model_id`,`rel_model_field`(255)),
  KEY `idx_rel_model_id` (`rel_model_id`),
  CONSTRAINT `FK_model_relation_id` FOREIGN KEY (`model_relation_id`) REFERENCES `fdp_agent_onedata_model_relation` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `fdp_agent_onedata_period`
--

DROP TABLE IF EXISTS `fdp_agent_onedata_period`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `fdp_agent_onedata_period` (
  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT,
  `name` varchar(128) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT '限定名称',
  `english` varchar(128) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT '英文名',
  `start_time` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '开始时间',
  `end_time` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '结束时间',
  `method` varchar(128) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT '计算逻辑',
  `creator_id` varchar(128) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '创建者id',
  `updated_id` varchar(128) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '修改者id',
  `created_at` datetime NOT NULL COMMENT '表记录创建时间',
  `updated_at` datetime NOT NULL COMMENT '表记录更新时间',
  `org_id` bigint(20) DEFAULT NULL COMMENT '企业 id',
  `cluster_id` bigint(20) DEFAULT NULL COMMENT '集群id',
  `project_id` bigint(20) DEFAULT NULL COMMENT '项目id',
  `application_id` bigint(20) DEFAULT NULL COMMENT '应用id',
  `cluster_name` varchar(256) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '集群名',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='统计周期表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `fdp_agent_onedata_restrict`
--

DROP TABLE IF EXISTS `fdp_agent_onedata_restrict`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `fdp_agent_onedata_restrict` (
  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT,
  `name` varchar(128) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT '限定名称',
  `english` varchar(128) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT '英文名',
  `method` varchar(128) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT '计算逻辑',
  `real_model` varchar(128) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '引用模型',
  `real_field` varchar(128) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '引用字段',
  `creator_id` varchar(128) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '创建者id',
  `data_model_id` bigint(20) DEFAULT NULL,
  `updated_id` varchar(128) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '修改者id',
  `created_at` datetime NOT NULL COMMENT '表记录创建时间',
  `updated_at` datetime NOT NULL COMMENT '表记录更新时间',
  `data_field_id` bigint(20) DEFAULT NULL,
  `org_id` bigint(20) DEFAULT NULL COMMENT '企业 id',
  `cluster_id` bigint(20) DEFAULT NULL COMMENT '集群id',
  `project_id` bigint(20) DEFAULT NULL COMMENT '项目id',
  `application_id` bigint(20) DEFAULT NULL COMMENT '应用id',
  `cluster_name` varchar(256) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '集群名',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='业务限定表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `fdp_agent_onedata_subject`
--

DROP TABLE IF EXISTS `fdp_agent_onedata_subject`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `fdp_agent_onedata_subject` (
  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT COMMENT '主键id',
  `subject_name` varchar(64) NOT NULL DEFAULT '' COMMENT '主题域名称',
  `subject_desc` varchar(255) NOT NULL DEFAULT '' COMMENT '主题域说明',
  `subject_prefix` varchar(64) NOT NULL DEFAULT '' COMMENT '主题域前缀',
  `updater_id` varchar(128) NOT NULL COMMENT '更新者 ID',
  `updated_at` datetime NOT NULL COMMENT '最近修改时间',
  `delete_yn` tinyint(1) NOT NULL DEFAULT '0' COMMENT '逻辑删除标记',
  `creator_id` varchar(128) NOT NULL COMMENT '创建人 ID',
  `created_at` datetime NOT NULL COMMENT '创建时间',
  `org_id` bigint(20) DEFAULT NULL COMMENT '企业id',
  `cluster_id` bigint(20) DEFAULT NULL COMMENT '集群id',
  `project_id` bigint(20) DEFAULT NULL COMMENT '项目id',
  `application_id` bigint(20) DEFAULT NULL COMMENT '应用id',
  `cluster_name` varchar(256) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '集群名',
  `platform_id` bigint(20) DEFAULT NULL COMMENT '平台id',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='onedata主题域表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `fdp_agent_quality_alarm_summary`
--

DROP TABLE IF EXISTS `fdp_agent_quality_alarm_summary`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `fdp_agent_quality_alarm_summary` (
  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT COMMENT '主键id',
  `today_cnt` bigint(20) NOT NULL DEFAULT '0' COMMENT '今日告警数',
  `last_sevenday_cnt` bigint(20) NOT NULL DEFAULT '0' COMMENT '最近7天',
  `last_thirtyday_cnt` bigint(20) NOT NULL DEFAULT '0' COMMENT '最近30天',
  `updater_id` bigint(20) NOT NULL COMMENT '最近修改人',
  `updated_at` datetime NOT NULL COMMENT '最近修改时间',
  `creator_id` bigint(20) NOT NULL COMMENT '创建人',
  `created_at` datetime NOT NULL COMMENT '创建时间',
  `cluster_id` bigint(20) DEFAULT NULL COMMENT '集群id',
  `project_id` bigint(20) DEFAULT '1' COMMENT '项目id',
  `org_id` bigint(20) DEFAULT NULL COMMENT '集群id',
  `application_id` bigint(20) DEFAULT NULL COMMENT '应用id',
  `cluster_name` varchar(256) DEFAULT '' COMMENT '集群名称',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='质量告警表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `fdp_agent_quality_score_detail`
--

DROP TABLE IF EXISTS `fdp_agent_quality_score_detail`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `fdp_agent_quality_score_detail` (
  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT COMMENT '主键id',
  `model_id` bigint(20) NOT NULL DEFAULT '0' COMMENT '模型id',
  `model_name` varchar(256) NOT NULL DEFAULT '0' COMMENT '模型名称',
  `model_type` varchar(256) NOT NULL DEFAULT '0' COMMENT '模型类型',
  `score` double(10,2) NOT NULL DEFAULT '0.00' COMMENT '全局得分',
  `updater_id` bigint(20) NOT NULL COMMENT '最近修改人',
  `updated_at` datetime NOT NULL COMMENT '最近修改时间',
  `creator_id` bigint(20) NOT NULL COMMENT '创建人',
  `created_at` datetime NOT NULL COMMENT '创建时间',
  `total_model_score` bigint(20) DEFAULT '0' COMMENT '总得分',
  `score_column_cnt` bigint(20) DEFAULT '0' COMMENT '总字段数',
  `score_column_list` varchar(256) DEFAULT '' COMMENT '得分总字段名',
  `cluster_id` bigint(20) DEFAULT NULL COMMENT '集群id',
  `project_id` bigint(20) DEFAULT '1' COMMENT '项目id',
  `org_id` bigint(20) DEFAULT NULL COMMENT '集群id',
  `application_id` bigint(20) DEFAULT NULL COMMENT '应用id',
  `cluster_name` varchar(256) DEFAULT '' COMMENT '集群名称',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='质量得分详细表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `fdp_agent_quality_score_trend`
--

DROP TABLE IF EXISTS `fdp_agent_quality_score_trend`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `fdp_agent_quality_score_trend` (
  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT COMMENT '主键id',
  `score` double(10,2) NOT NULL DEFAULT '0.00' COMMENT '全局得分',
  `updater_id` bigint(20) NOT NULL COMMENT '最近修改人',
  `updated_at` datetime NOT NULL COMMENT '最近修改时间',
  `creator_id` bigint(20) NOT NULL COMMENT '创建人',
  `created_at` datetime NOT NULL COMMENT '创建时间',
  `cluster_id` bigint(20) DEFAULT NULL COMMENT '集群id',
  `project_id` bigint(20) DEFAULT '1' COMMENT '项目id',
  `org_id` bigint(20) DEFAULT NULL COMMENT '集群id',
  `application_id` bigint(20) DEFAULT NULL COMMENT '应用id',
  `cluster_name` varchar(256) DEFAULT '' COMMENT '集群名称',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='质量得分趋势表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `fdp_agent_quality_utility`
--

DROP TABLE IF EXISTS `fdp_agent_quality_utility`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `fdp_agent_quality_utility` (
  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT COMMENT '主键id',
  `model_cnt` bigint(20) NOT NULL DEFAULT '0' COMMENT '模型数',
  `rule_cnt` bigint(20) NOT NULL DEFAULT '0' COMMENT '规则数',
  `new_model_cnt` bigint(20) NOT NULL DEFAULT '0' COMMENT '新增模型数',
  `new_rule_cnt` bigint(20) NOT NULL DEFAULT '0' COMMENT '规则数',
  `updater_id` bigint(20) NOT NULL COMMENT '最近修改人',
  `updated_at` datetime NOT NULL COMMENT '最近修改时间',
  `creator_id` bigint(20) NOT NULL COMMENT '创建人',
  `created_at` datetime NOT NULL COMMENT '创建时间',
  `cluster_id` bigint(20) DEFAULT NULL COMMENT '集群id',
  `project_id` bigint(20) DEFAULT '1' COMMENT '项目id',
  `org_id` bigint(20) DEFAULT NULL COMMENT '集群id',
  `application_id` bigint(20) DEFAULT NULL COMMENT '应用id',
  `cluster_name` varchar(256) DEFAULT '' COMMENT '集群名称',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='质量得分效用表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `fdp_agent_reco_scenario_workflows`
--

DROP TABLE IF EXISTS `fdp_agent_reco_scenario_workflows`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `fdp_agent_reco_scenario_workflows` (
  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT COMMENT '主键id',
  `org_id` bigint(20) DEFAULT NULL COMMENT '企业 id',
  `scenario_name` varchar(128) NOT NULL DEFAULT '',
  `scenario_code` varchar(128) NOT NULL DEFAULT '',
  `process_type` varchar(128) NOT NULL COMMENT '处理类型',
  `name` varchar(128) NOT NULL DEFAULT '' COMMENT '工作流名',
  `name_pinyin` varchar(128) NOT NULL DEFAULT '' COMMENT '工作流拼音',
  `description` varchar(1024) NOT NULL DEFAULT '' COMMENT '工作流描述',
  `source` varchar(32) DEFAULT '' COMMENT 'workflow 来源: dl/cdp',
  `run_type` varchar(64) NOT NULL DEFAULT '' COMMENT '运行类型: ATONCE、SCHEDULE',
  `cron` varchar(128) DEFAULT '' COMMENT '任务周期',
  `cron_start_from` datetime DEFAULT NULL COMMENT '延时执行时间',
  `category_id` bigint(20) NOT NULL COMMENT '工作流目录 ID',
  `pipeline_name` varchar(128) DEFAULT '' COMMENT 'pipeline 名称',
  `pipeline` mediumtext NOT NULL COMMENT 'pipeline 内容',
  `pipeline_id` bigint(20) DEFAULT NULL COMMENT 'pipeline id',
  `node_params` text COMMENT '工作流节点参数',
  `locations` varchar(1024) DEFAULT NULL,
  `creator_id` varchar(128) CHARACTER SET utf8mb4 DEFAULT '' COMMENT '创建者 ID',
  `updater_id` varchar(128) DEFAULT NULL COMMENT '更新者ID',
  `extra` mediumtext COMMENT '扩展信息',
  `delete_yn` tinyint(1) NOT NULL DEFAULT '0' COMMENT '逻辑删除标记',
  `created_at` datetime DEFAULT NULL COMMENT '创建时间',
  `updated_at` datetime DEFAULT NULL COMMENT '更新时间',
  PRIMARY KEY (`id`),
  KEY `idx_category_id_name` (`scenario_code`,`process_type`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8 COMMENT='场景工作流表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `fdp_agent_service_consumers`
--

DROP TABLE IF EXISTS `fdp_agent_service_consumers`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `fdp_agent_service_consumers` (
  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT COMMENT '主键id',
  `org_id` bigint(20) DEFAULT NULL COMMENT '企业 id',
  `name` varchar(128) NOT NULL COMMENT '调用方名称',
  `description` varchar(1024) DEFAULT NULL COMMENT '描述',
  `dice_consumer_id` varchar(128) NOT NULL COMMENT 'dice API网关返回调用方id',
  `delete_yn` tinyint(1) NOT NULL DEFAULT '0' COMMENT '逻辑删除标记',
  `creator_id` varchar(128) DEFAULT NULL COMMENT '创建人 ID',
  `updater_id` varchar(128) DEFAULT NULL COMMENT '更新者ID',
  `created_at` datetime NOT NULL COMMENT '表记录创建时间',
  `updated_at` datetime NOT NULL COMMENT '表记录更新时间',
  `cluster_id` bigint(20) DEFAULT NULL COMMENT '集群id',
  `project_id` bigint(20) DEFAULT '1' COMMENT '项目id',
  `application_id` bigint(20) DEFAULT NULL COMMENT '应用id',
  `cluster_name` varchar(256) DEFAULT '' COMMENT '集群名称',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8 COMMENT='调用方列表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `fdp_agent_service_consumers_rel_ak`
--

DROP TABLE IF EXISTS `fdp_agent_service_consumers_rel_ak`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `fdp_agent_service_consumers_rel_ak` (
  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT COMMENT '主键id',
  `org_id` bigint(20) DEFAULT NULL COMMENT '企业 id',
  `consumer_id` bigint(20) NOT NULL COMMENT '调用方ID',
  `access_key` varchar(256) NOT NULL COMMENT 'access key',
  `created_at` datetime NOT NULL COMMENT '表记录创建时间',
  `updated_at` datetime NOT NULL COMMENT '表记录更新时间',
  `cluster_id` bigint(20) DEFAULT NULL COMMENT '集群id',
  `project_id` bigint(20) DEFAULT '1' COMMENT '项目id',
  `application_id` bigint(20) DEFAULT NULL COMMENT '应用id',
  `cluster_name` varchar(256) DEFAULT '' COMMENT '集群名称',
  PRIMARY KEY (`id`),
  KEY `idx_access_key` (`access_key`(255))
) ENGINE=InnoDB DEFAULT CHARSET=utf8 COMMENT='调用方与ak关联表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `fdp_agent_service_consumers_rel_data_api`
--

DROP TABLE IF EXISTS `fdp_agent_service_consumers_rel_data_api`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `fdp_agent_service_consumers_rel_data_api` (
  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT COMMENT '主键id',
  `org_id` bigint(20) DEFAULT NULL COMMENT '企业 id',
  `consumer_id` bigint(20) NOT NULL COMMENT '调用方ID',
  `data_api_id` bigint(20) NOT NULL COMMENT 'api id',
  `grant_auth_status` tinyint(1) NOT NULL DEFAULT '0' COMMENT '是否被授权',
  `creator_id` varchar(128) DEFAULT NULL COMMENT '创建人 ID',
  `updater_id` varchar(128) DEFAULT NULL COMMENT '更新者ID',
  `created_at` datetime NOT NULL COMMENT '表记录创建时间',
  `updated_at` datetime NOT NULL COMMENT '表记录更新时间',
  `cluster_id` bigint(20) DEFAULT NULL COMMENT '集群id',
  `project_id` bigint(20) DEFAULT '1' COMMENT '项目id',
  `application_id` bigint(20) DEFAULT NULL COMMENT '应用id',
  `cluster_name` varchar(256) DEFAULT '' COMMENT '集群名称',
  PRIMARY KEY (`id`),
  KEY `idx_consumer_scenario_id` (`consumer_id`,`data_api_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8 COMMENT='调用方与api表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `fdp_agent_sql_histories`
--

DROP TABLE IF EXISTS `fdp_agent_sql_histories`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `fdp_agent_sql_histories` (
  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT COMMENT '主键id',
  `data_model_id` bigint(20) NOT NULL COMMENT '数据模型 ID',
  `name` varchar(256) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT '执行记录名称',
  `sql` text COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '执行的 SQL',
  `result` text COLLATE utf8mb4_unicode_ci NOT NULL COMMENT 'SQL 执行结果',
  `creator_id` varchar(256) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT '创建人 ID',
  `created_at` datetime NOT NULL COMMENT '表记录创建时间',
  `updated_at` datetime NOT NULL COMMENT '表记录更新时间',
  `cluster_id` bigint(20) DEFAULT NULL COMMENT '集群id',
  `project_id` bigint(20) DEFAULT '1' COMMENT '项目id',
  `org_id` bigint(20) DEFAULT NULL COMMENT '集群id',
  `application_id` bigint(20) DEFAULT NULL COMMENT '应用id',
  `cluster_name` varchar(256) COLLATE utf8mb4_unicode_ci DEFAULT '' COMMENT '集群名称',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci BLOCK_FORMAT=ENCRYPTED COMMENT='sql执行历史表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `fdp_agent_sql_history`
--

DROP TABLE IF EXISTS `fdp_agent_sql_history`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `fdp_agent_sql_history` (
  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT COMMENT '主键id',
  `cache_data_id` bigint(20) unsigned NOT NULL COMMENT '数据缓存id',
  `name` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT '执行记录名称',
  `query_status` varchar(128) NOT NULL COMMENT '查询响应状态',
  `query_time` bigint(20) NOT NULL COMMENT '查询耗时',
  `sql` text NOT NULL COMMENT '查询sql',
  `data_model_name` varchar(255) NOT NULL COMMENT '数据模型 名称',
  `data_source_id` bigint(20) NOT NULL COMMENT '数据源 ID',
  `delete_yn` tinyint(1) NOT NULL DEFAULT '0' COMMENT '逻辑删除标记',
  `creator_id` varchar(128) DEFAULT NULL COMMENT '创建人 ID',
  `updater_id` varchar(128) DEFAULT NULL COMMENT '更新者ID',
  `created_at` datetime NOT NULL COMMENT '创建时间',
  `updated_at` datetime NOT NULL COMMENT '更新时间',
  `code` varchar(255) DEFAULT NULL COMMENT '数据模型 名称',
  `error` text COMMENT '数据模型 名称',
  `sourceStack` text COMMENT '数据模型 名称',
  `columns` text COMMENT '数据模型 名称',
  `logs` text COMMENT '数据模型 名称',
  `cluster_id` bigint(20) DEFAULT NULL COMMENT '集群id',
  `project_id` bigint(20) DEFAULT '1' COMMENT '项目id',
  `org_id` bigint(20) DEFAULT NULL COMMENT '集群id',
  `application_id` bigint(20) DEFAULT NULL COMMENT '应用id',
  `cluster_name` varchar(256) DEFAULT '' COMMENT '集群名称',
  PRIMARY KEY (`id`),
  KEY `abc` (`cache_data_id`),
  CONSTRAINT `abc` FOREIGN KEY (`cache_data_id`) REFERENCES `fdp_agent_ad_hoc_cache_data` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8 COMMENT='sql执行历史表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `fdp_agent_tag_sync_logs`
--

DROP TABLE IF EXISTS `fdp_agent_tag_sync_logs`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `fdp_agent_tag_sync_logs` (
  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT COMMENT '主键id',
  `tag_id` bigint(20) NOT NULL COMMENT '标签 Id',
  `tag_name` varchar(128) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT '标签名称',
  `pipeline_ids` varchar(128) COLLATE utf8mb4_unicode_ci DEFAULT '' COMMENT 'pipeline ids',
  `status` varchar(32) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '标签同步状态',
  `created_at` datetime NOT NULL COMMENT '表记录创建时间',
  `updated_at` datetime NOT NULL COMMENT '表记录更新时间',
  `cluster_id` bigint(20) DEFAULT NULL COMMENT '集群id',
  `project_id` bigint(20) DEFAULT '1' COMMENT '项目id',
  `org_id` bigint(20) DEFAULT NULL COMMENT '集群id',
  `application_id` bigint(20) DEFAULT NULL COMMENT '应用id',
  `cluster_name` varchar(256) COLLATE utf8mb4_unicode_ci DEFAULT '' COMMENT '集群名称',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci BLOCK_FORMAT=ENCRYPTED COMMENT='标签同步日志表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `fdp_agent_tag_values`
--

DROP TABLE IF EXISTS `fdp_agent_tag_values`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `fdp_agent_tag_values` (
  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT COMMENT '主键id',
  `tag_id` bigint(20) NOT NULL COMMENT '标签 Id',
  `name` varchar(128) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT '标签值名称',
  `description` text COLLATE utf8mb4_unicode_ci COMMENT '描述',
  `logic_type` varchar(32) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '标签值关联关系',
  `created_at` datetime NOT NULL COMMENT '表记录创建时间',
  `updated_at` datetime NOT NULL COMMENT '表记录更新时间',
  `cluster_id` bigint(20) DEFAULT NULL COMMENT '集群id',
  `project_id` bigint(20) DEFAULT '1' COMMENT '项目id',
  `org_id` bigint(20) DEFAULT NULL COMMENT '集群id',
  `application_id` bigint(20) DEFAULT NULL COMMENT '应用id',
  `cluster_name` varchar(256) COLLATE utf8mb4_unicode_ci DEFAULT '' COMMENT '集群名称',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci BLOCK_FORMAT=ENCRYPTED COMMENT='标签值表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `fdp_agent_tags`
--

DROP TABLE IF EXISTS `fdp_agent_tags`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `fdp_agent_tags` (
  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT COMMENT '主键id',
  `name` varchar(128) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT '标签名称',
  `description` text COLLATE utf8mb4_unicode_ci COMMENT '描述',
  `code` varchar(32) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '标签码',
  `type` varchar(32) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '标签类型',
  `value_type` varchar(40) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '标签值类型',
  `status` varchar(32) COLLATE utf8mb4_unicode_ci DEFAULT '' COMMENT '标签同步状态',
  `disabled` tinyint(1) DEFAULT NULL COMMENT '停用',
  `created_at` datetime NOT NULL COMMENT '表记录创建时间',
  `updated_at` datetime NOT NULL COMMENT '表记录更新时间',
  `cluster_id` bigint(20) DEFAULT NULL COMMENT '集群id',
  `project_id` bigint(20) DEFAULT '1' COMMENT '项目id',
  `org_id` bigint(20) DEFAULT NULL COMMENT '集群id',
  `application_id` bigint(20) DEFAULT NULL COMMENT '应用id',
  `cluster_name` varchar(256) COLLATE utf8mb4_unicode_ci DEFAULT '' COMMENT '集群名称',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci BLOCK_FORMAT=ENCRYPTED COMMENT='标签表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `fdp_agent_task_queue`
--

DROP TABLE IF EXISTS `fdp_agent_task_queue`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `fdp_agent_task_queue` (
  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT COMMENT '主键id',
  `name` varchar(128) DEFAULT NULL COMMENT '名字',
  `cpu` double(20,2) DEFAULT NULL COMMENT 'cpu',
  `memory` double(20,2) DEFAULT NULL COMMENT 'memory',
  `strategy` varchar(128) DEFAULT 'FIFO' COMMENT '调度策略',
  `level` bigint(20) DEFAULT NULL COMMENT '队列权重',
  `pipeline_create_id` bigint(20) DEFAULT NULL COMMENT 'pipeline组件创建队列的id',
  `creator_id` varchar(128) DEFAULT NULL COMMENT '创建人 ID',
  `updater_id` varchar(128) DEFAULT NULL COMMENT '更新者ID',
  `created_at` datetime NOT NULL COMMENT '创建时间',
  `updated_at` datetime NOT NULL COMMENT '更新时间',
  `org_id` bigint(20) DEFAULT NULL COMMENT '企业 id',
  `project_id` bigint(20) DEFAULT NULL COMMENT '项目 id',
  `application_id` bigint(20) DEFAULT NULL COMMENT '应用 id',
  `cluster_id` bigint(20) DEFAULT NULL COMMENT '集群 id',
  `cluster_name` varchar(256) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '集群名',
  `status` tinyint(4) DEFAULT NULL COMMENT 'pipeline队列创建情况',
  `concurrency` bigint(20) DEFAULT NULL COMMENT '并行度',
  `mode` varchar(256) DEFAULT '' COMMENT '模式',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=2 DEFAULT CHARSET=utf8 COMMENT='任务队列表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `fdp_agent_user_action_log_detail`
--

DROP TABLE IF EXISTS `fdp_agent_user_action_log_detail`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `fdp_agent_user_action_log_detail` (
  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT COMMENT '主键id',
  `user_id` varchar(64) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT '用户id',
  `action_type` varchar(64) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT '动作类型：搜索/查看/引用',
  `action_content` varchar(64) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT '动作内容:搜索条件/查看条件/引用条件',
  `result_model_name` varchar(128) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT '涉及结果模型表名称',
  `result_model_id` bigint(20) NOT NULL COMMENT '涉及结果模型表id',
  `action_result` varchar(64) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT '动作结果：搜索结果/查看结果/引用结果',
  `created_at` datetime NOT NULL COMMENT '表记录创建时间',
  `updated_at` datetime NOT NULL COMMENT '表记录更新时间',
  `cluster_id` bigint(20) DEFAULT NULL COMMENT '集群id',
  `project_id` bigint(20) DEFAULT '1' COMMENT '项目id',
  `org_id` bigint(20) DEFAULT NULL COMMENT '集群id',
  `application_id` bigint(20) DEFAULT NULL COMMENT '应用id',
  `cluster_name` varchar(256) COLLATE utf8mb4_unicode_ci DEFAULT '' COMMENT '集群名称',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户动作日志表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `fdp_agent_value_indicators`
--

DROP TABLE IF EXISTS `fdp_agent_value_indicators`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `fdp_agent_value_indicators` (
  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT COMMENT '主键id',
  `tag_id` bigint(20) NOT NULL COMMENT '标签 Id',
  `value_id` bigint(20) NOT NULL COMMENT '标签值 Id',
  `indicator_id` bigint(20) NOT NULL COMMENT '指标 Id',
  `tag_compute_type` varchar(64) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '标签计算类型',
  `min_range` varchar(32) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '最小值',
  `max_range` varchar(32) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '最大值',
  `created_at` datetime NOT NULL COMMENT '表记录创建时间',
  `updated_at` datetime NOT NULL COMMENT '表记录更新时间',
  `cluster_id` bigint(20) DEFAULT NULL COMMENT '集群id',
  `project_id` bigint(20) DEFAULT '1' COMMENT '项目id',
  `org_id` bigint(20) DEFAULT NULL COMMENT '集群id',
  `application_id` bigint(20) DEFAULT NULL COMMENT '应用id',
  `cluster_name` varchar(256) COLLATE utf8mb4_unicode_ci DEFAULT '' COMMENT '集群名称',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci BLOCK_FORMAT=ENCRYPTED COMMENT='指标值表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `fdp_agent_work_project`
--

DROP TABLE IF EXISTS `fdp_agent_work_project`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `fdp_agent_work_project` (
  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT COMMENT '主键id',
  `name` varchar(128) DEFAULT NULL COMMENT 'gittar project id',
  `alias` varchar(128) DEFAULT NULL COMMENT 'gittar project 名字',
  `delete_yn` tinyint(1) NOT NULL DEFAULT '0' COMMENT '逻辑删除标记',
  `creator_id` varchar(128) DEFAULT NULL COMMENT '创建人 ID',
  `updater_id` varchar(128) DEFAULT NULL COMMENT '更新者ID',
  `created_at` datetime NOT NULL COMMENT '创建时间',
  `updated_at` datetime NOT NULL COMMENT '更新时间',
  `org_id` bigint(20) DEFAULT NULL COMMENT '企业 id',
  `project_id` bigint(20) DEFAULT NULL COMMENT '项目 id',
  `application_id` bigint(20) DEFAULT NULL COMMENT '应用 id',
  `cluster_id` bigint(20) DEFAULT NULL COMMENT '集群 id',
  `cluster_name` varchar(256) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '集群名',
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_name_project` (`name`,`delete_yn`)
) ENGINE=InnoDB AUTO_INCREMENT=2 DEFAULT CHARSET=utf8 COMMENT='工作空间表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `fdp_agent_work_template_yaml`
--

DROP TABLE IF EXISTS `fdp_agent_work_template_yaml`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `fdp_agent_work_template_yaml` (
  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT COMMENT '主键id',
  `name` varchar(128) DEFAULT NULL COMMENT '第三方服务名称',
  `template` longtext,
  `org_id` bigint(20) DEFAULT NULL COMMENT '企业 id',
  `project_id` bigint(20) DEFAULT NULL COMMENT '项目 id',
  `application_id` bigint(20) DEFAULT NULL COMMENT '应用 id',
  `cluster_id` bigint(20) DEFAULT NULL COMMENT '集群 id',
  `cluster_name` varchar(256) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '集群名',
  `creator_id` varchar(128) DEFAULT NULL COMMENT '创建者id',
  `created_at` datetime NOT NULL COMMENT '创建时间',
  `updated_at` datetime NOT NULL COMMENT '更新时间',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8 COMMENT='提交spark，flinkyaml模版';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `fdp_agent_workflow_categories`
--

DROP TABLE IF EXISTS `fdp_agent_workflow_categories`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `fdp_agent_workflow_categories` (
  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT COMMENT '主键id',
  `org_id` bigint(20) DEFAULT NULL COMMENT '企业 id',
  `source` varchar(64) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT '平台来源：DL/CDP',
  `name` varchar(128) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT '目录名',
  `name_pinyin` varchar(128) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT '目录名拼音',
  `parent_id` bigint(20) NOT NULL COMMENT '上级目录 ID',
  `creator_id` varchar(256) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT '创建人 ID',
  `created_at` datetime NOT NULL COMMENT '表记录创建时间',
  `updated_at` datetime NOT NULL COMMENT '表记录更新时间',
  `cluster_id` bigint(20) DEFAULT NULL COMMENT '集群id',
  `project_id` bigint(20) DEFAULT NULL COMMENT '项目id',
  `application_id` bigint(20) DEFAULT NULL COMMENT '应用id',
  `cluster_name` varchar(128) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `workflow_category_unique_index1` (`parent_id`,`source`,`name`,`cluster_name`,`project_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='工作流目录表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `fdp_agent_workflow_copy_detail`
--

DROP TABLE IF EXISTS `fdp_agent_workflow_copy_detail`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `fdp_agent_workflow_copy_detail` (
  `id` int(11) NOT NULL AUTO_INCREMENT COMMENT '主键',
  `copy_log_id` int(11) NOT NULL COMMENT '关联字段',
  `workflow_name` varchar(256) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '工作流名称',
  `operation_type` varchar(128) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '操作类型',
  `copy_success` tinyint(1) DEFAULT '0' COMMENT '复制是否成功',
  `error_reason` text COLLATE utf8mb4_unicode_ci COMMENT '失败原因',
  `begin_date` datetime DEFAULT NULL COMMENT '复制开始时间',
  `end_date` datetime DEFAULT NULL COMMENT '复制结束时间',
  PRIMARY KEY (`id`),
  KEY `fdp_agent_workflow_copy_fk1` (`copy_log_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `fdp_agent_workflow_copy_log`
--

DROP TABLE IF EXISTS `fdp_agent_workflow_copy_log`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `fdp_agent_workflow_copy_log` (
  `id` int(11) NOT NULL AUTO_INCREMENT COMMENT '主键',
  `name` varchar(256) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '名称',
  `created_date` datetime NOT NULL COMMENT '创建时间',
  `file_success` tinyint(1) DEFAULT '0' COMMENT '文件生成是否成功',
  `error_reason` text COLLATE utf8mb4_unicode_ci COMMENT '失败原因',
  `file_path` varchar(256) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '文件路径',
  `file_date` datetime DEFAULT NULL COMMENT '文件生成成功时间',
  `complete` tinyint(1) DEFAULT '0' COMMENT '复制是否完成',
  `complete_date` datetime DEFAULT NULL COMMENT '复制完成时间',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `fdp_agent_workflow_dependencies`
--

DROP TABLE IF EXISTS `fdp_agent_workflow_dependencies`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `fdp_agent_workflow_dependencies` (
  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT COMMENT '主键id',
  `workflow_id` bigint(20) NOT NULL COMMENT '工作流 ID',
  `depend_workflow_id` bigint(20) NOT NULL COMMENT '依赖的工作流 ID',
  `creator_id` varchar(128) NOT NULL DEFAULT '' COMMENT '创建者 ID',
  `updater_id` varchar(128) NOT NULL DEFAULT '' COMMENT '更新着 ID',
  `created_at` datetime NOT NULL COMMENT '表记录创建时间',
  `updated_at` datetime NOT NULL COMMENT '表记录更新时间',
  `cluster_id` bigint(20) DEFAULT NULL COMMENT '集群id',
  `project_id` bigint(20) DEFAULT NULL COMMENT '项目id',
  `application_id` bigint(20) DEFAULT NULL COMMENT '应用id',
  `cluster_name` varchar(256) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '集群名',
  `org_id` bigint(20) DEFAULT NULL COMMENT '企业 id',
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_dependWorkflowId_under_workflowId` (`workflow_id`,`depend_workflow_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='工作流依赖表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `fdp_agent_workflow_indicators`
--

DROP TABLE IF EXISTS `fdp_agent_workflow_indicators`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `fdp_agent_workflow_indicators` (
  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT COMMENT '主键id',
  `execute_time_top_daily_result` text COMMENT '执行时长排行top10',
  `failed_time_top_30_days_result` text COMMENT '最近30天执行top10',
  `statistical_date` varchar(128) DEFAULT NULL COMMENT '统计日期',
  `created_at` datetime NOT NULL COMMENT '表记录创建时间',
  `updated_at` datetime NOT NULL COMMENT '表记录更新时间',
  `cluster_id` bigint(20) DEFAULT NULL COMMENT '集群id',
  `project_id` bigint(20) DEFAULT '1' COMMENT '项目id',
  `org_id` bigint(20) DEFAULT NULL COMMENT '集群id',
  `application_id` bigint(20) DEFAULT NULL COMMENT '应用id',
  `cluster_name` varchar(256) DEFAULT '' COMMENT '集群名称',
  PRIMARY KEY (`id`),
  KEY `idx_statistical_date` (`statistical_date`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8 COMMENT='工作流指标值表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `fdp_agent_workflow_node`
--

DROP TABLE IF EXISTS `fdp_agent_workflow_node`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `fdp_agent_workflow_node` (
  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT COMMENT '主键id',
  `workflow_id` bigint(20) NOT NULL COMMENT 'workflow id',
  `name` varchar(128) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '节点名称',
  `created_at` datetime NOT NULL COMMENT '表记录创建时间',
  `x` varchar(128) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT 'x轴坐标',
  `y` varchar(128) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT 'y轴坐标',
  `node_type` varchar(20) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '节点类型',
  `node_action` text COLLATE utf8mb4_unicode_ci,
  `org_id` bigint(20) DEFAULT NULL COMMENT '企业id',
  `cluster_id` bigint(20) DEFAULT NULL COMMENT '集群id',
  `project_id` bigint(20) DEFAULT NULL COMMENT '项目id',
  `application_id` bigint(20) DEFAULT NULL COMMENT '应用id',
  `cluster_name` varchar(256) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '集群名',
  `persist_type` varchar(20) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '存储方式',
  `updated_at` datetime NOT NULL COMMENT '表记录创建时间',
  `creator_id` varchar(256) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '创建人id',
  PRIMARY KEY (`id`),
  KEY `workflow_id` (`workflow_id`),
  KEY `workflow_x_y` (`workflow_id`,`x`,`y`)
) ENGINE=InnoDB AUTO_INCREMENT=401 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='节点详情';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `fdp_agent_workflow_node_cleaning`
--

DROP TABLE IF EXISTS `fdp_agent_workflow_node_cleaning`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `fdp_agent_workflow_node_cleaning` (
  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT COMMENT '主键id',
  `workflow_id` bigint(20) NOT NULL COMMENT ' workflow id',
  `name` varchar(128) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '节点名称',
  `mode` varchar(20) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '节点添加模式: ui/sql',
  `description` text COLLATE utf8mb4_unicode_ci COMMENT '节点描述',
  `clean_type` varchar(20) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '清洗类型: 维度表(dimension)/事实表(fact)',
  `tb_en_name` varchar(100) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '表英文名称',
  `tb_cn_name` varchar(100) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '表中文名称',
  `tb_desc` text COLLATE utf8mb4_unicode_ci COMMENT '表描述',
  `load_strategy` varchar(20) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '加载策略: increment/total',
  `custom_content` text COLLATE utf8mb4_unicode_ci COMMENT '自定义编辑内容',
  `main_model_id` bigint(20) DEFAULT NULL COMMENT '主模型 id',
  `main_model_alias` varchar(100) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '主模型别名',
  `main_model_filter_cond` text COLLATE utf8mb4_unicode_ci COMMENT '主模型过滤条件',
  `target_model_id` bigint(20) DEFAULT NULL COMMENT '目标模型 id',
  `target_model_alias` varchar(100) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '目标模型别名',
  `target_model_filter_cond` text COLLATE utf8mb4_unicode_ci COMMENT '目标模型过滤条件',
  `created_at` datetime NOT NULL COMMENT '表记录创建时间',
  `updated_at` datetime NOT NULL COMMENT '表记录更新时间',
  `org_id` bigint(20) DEFAULT NULL COMMENT '企业id',
  `cluster_id` bigint(20) DEFAULT NULL COMMENT '集群id',
  `project_id` bigint(20) DEFAULT NULL COMMENT '项目id',
  `application_id` bigint(20) DEFAULT NULL COMMENT '应用id',
  `cluster_name` varchar(256) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '集群名',
  `persist_type` varchar(20) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '存储方式',
  `cpu` double DEFAULT '1' COMMENT '节点cpu',
  `memory` double DEFAULT '1024' COMMENT '节点memory',
  `create_type` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT 'CREATE' COMMENT '清洗节点创建方式',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='工作流清洗节点表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `fdp_agent_workflow_node_cleaning_rt`
--

DROP TABLE IF EXISTS `fdp_agent_workflow_node_cleaning_rt`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `fdp_agent_workflow_node_cleaning_rt` (
  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT COMMENT '主键id',
  `workflow_id` bigint(20) NOT NULL COMMENT ' workflow id',
  `name` varchar(128) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '节点名称',
  `mode` varchar(20) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '节点添加模式: ui/sql',
  `description` text COLLATE utf8mb4_unicode_ci COMMENT '节点描述',
  `clean_type` varchar(20) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '清洗类型: 维度表(dimension)/事实表(fact)',
  `tb_en_name` varchar(100) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '表英文名称',
  `tb_cn_name` varchar(100) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '表中文名称',
  `tb_desc` text COLLATE utf8mb4_unicode_ci COMMENT '表描述',
  `load_strategy` varchar(20) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '加载策略: increment/total',
  `custom_content` text COLLATE utf8mb4_unicode_ci COMMENT '自定义编辑内容',
  `main_model_id` bigint(20) DEFAULT NULL COMMENT '主模型 id',
  `main_model_alias` varchar(100) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '主模型别名',
  `main_model_filter_cond` text COLLATE utf8mb4_unicode_ci COMMENT '主模型过滤条件',
  `target_model_id` bigint(20) DEFAULT NULL COMMENT '目标模型 id',
  `target_model_alias` varchar(100) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '目标模型别名',
  `target_model_filter_cond` text COLLATE utf8mb4_unicode_ci COMMENT '目标模型过滤条件',
  `created_at` datetime NOT NULL COMMENT '表记录创建时间',
  `updated_at` datetime NOT NULL COMMENT '表记录更新时间',
  `org_id` bigint(20) DEFAULT NULL COMMENT '企业id',
  `cluster_id` bigint(20) DEFAULT NULL COMMENT '集群id',
  `project_id` bigint(20) DEFAULT NULL COMMENT '项目id',
  `application_id` bigint(20) DEFAULT NULL COMMENT '应用id',
  `cluster_name` varchar(256) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '集群名',
  `persist_type` varchar(20) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '存储方式',
  `cpu` double DEFAULT '1' COMMENT '节点cpu',
  `memory` double DEFAULT '1024' COMMENT '节点memory',
  `create_type` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT 'CREATE' COMMENT '清洗节点创建方式:SELECT,CREATE',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='工作流实时清洗节点表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `fdp_agent_workflow_node_data_cancel`
--

DROP TABLE IF EXISTS `fdp_agent_workflow_node_data_cancel`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `fdp_agent_workflow_node_data_cancel` (
  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT COMMENT '主键id',
  `workflow_id` bigint(20) NOT NULL COMMENT 'workflow id',
  `created_at` datetime NOT NULL COMMENT '表记录创建时间',
  `updated_at` datetime NOT NULL COMMENT '表记录更新时间',
  `org_id` bigint(20) DEFAULT NULL COMMENT '企业id',
  `cluster_id` bigint(20) DEFAULT NULL COMMENT '集群id',
  `project_id` bigint(20) DEFAULT NULL COMMENT '项目id',
  `application_id` bigint(20) DEFAULT NULL COMMENT '应用id',
  `cluster_name` varchar(256) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '集群名',
  `persist_type` varchar(20) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '存储方式',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='工作流取消节点表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `fdp_agent_workflow_node_detail`
--

DROP TABLE IF EXISTS `fdp_agent_workflow_node_detail`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `fdp_agent_workflow_node_detail` (
  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT COMMENT '主键id',
  `workflow_id` bigint(20) NOT NULL COMMENT 'workflow id',
  `name` varchar(128) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '节点名称',
  `created_at` datetime NOT NULL COMMENT '表记录创建时间',
  `x` varchar(128) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT 'x轴坐标',
  `y` varchar(128) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT 'y轴坐标',
  `description` text COLLATE utf8mb4_unicode_ci COMMENT '节点描述',
  `node_type` varchar(20) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '节点类型',
  `tb_en_name` varchar(100) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '表英文名称',
  `tb_cn_name` varchar(100) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '表中文名称',
  `tb_desc` text COLLATE utf8mb4_unicode_ci COMMENT '表描述',
  `load_strategy` varchar(20) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '加载策略: increment/total',
  `jar_path` varchar(256) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `node_id` bigint(20) DEFAULT NULL,
  `node_script` text COLLATE utf8mb4_unicode_ci COMMENT '自定义编辑内容',
  `main_model_id` bigint(20) DEFAULT NULL COMMENT '主模型 id',
  `main_model_alias` varchar(100) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '主模型别名',
  `main_model_filter_cond` text COLLATE utf8mb4_unicode_ci COMMENT '主模型过滤条件',
  `target_model_id` bigint(20) DEFAULT NULL COMMENT '目标模型 id',
  `target_model_alias` varchar(100) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '目标模型别名',
  `target_model_filter_cond` text COLLATE utf8mb4_unicode_ci COMMENT '目标模型过滤条件',
  `org_id` bigint(20) DEFAULT NULL COMMENT '企业id',
  `cluster_id` bigint(20) DEFAULT NULL COMMENT '集群id',
  `project_id` bigint(20) DEFAULT NULL COMMENT '项目id',
  `application_id` bigint(20) DEFAULT NULL COMMENT '应用id',
  `cluster_name` varchar(256) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '集群名',
  `creator_id` varchar(20) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '创建者id',
  `updated_at` datetime NOT NULL COMMENT '表记录创建时间',
  `node_extended` text COLLATE utf8mb4_unicode_ci COMMENT '其余信息',
  `script_path` varchar(256) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `workflow_id` (`workflow_id`),
  KEY `workflow_x_y` (`workflow_id`,`x`,`y`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='节点详情';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `fdp_agent_workflow_node_export`
--

DROP TABLE IF EXISTS `fdp_agent_workflow_node_export`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `fdp_agent_workflow_node_export` (
  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT COMMENT '主键id',
  `workflow_id` bigint(20) NOT NULL COMMENT ' workflow id',
  `name` varchar(128) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '节点名称',
  `mode` varchar(20) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '节点添加模式: ui/sql',
  `description` text COLLATE utf8mb4_unicode_ci COMMENT '节点描述',
  `custom_content` text COLLATE utf8mb4_unicode_ci COMMENT '自定义编辑内容',
  `src_ds_type` varchar(20) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '源数据源类型: DL',
  `src_ds_id` bigint(20) DEFAULT NULL COMMENT '源数据源 id',
  `src_model_id` bigint(20) DEFAULT NULL COMMENT '源数据模型 id',
  `load_strategy` varchar(20) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '加载策略: increment/total',
  `load_interval` bigint(20) DEFAULT NULL COMMENT '增量加载时间间隔',
  `load_interval_unit` varchar(20) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '增量加载时间间隔单位: MINUTE/HOUR/DAY',
  `target_ds_type` varchar(20) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '目标数据源类型: CDP',
  `target_ds_id` bigint(20) DEFAULT NULL COMMENT '目标数据源 id',
  `target_model_id` bigint(20) DEFAULT NULL COMMENT '目标数据模型 id',
  `preconditions` text COLLATE utf8mb4_unicode_ci COMMENT '前置条件',
  `filter_cond` text COLLATE utf8mb4_unicode_ci COMMENT '过滤条件',
  `write_mode` varchar(32) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '写入模式',
  `created_at` datetime NOT NULL COMMENT '表记录创建时间',
  `updated_at` datetime NOT NULL COMMENT '表记录更新时间',
  `org_id` bigint(20) DEFAULT NULL COMMENT '企业id',
  `cluster_id` bigint(20) DEFAULT NULL COMMENT '集群id',
  `project_id` bigint(20) DEFAULT NULL COMMENT '项目id',
  `application_id` bigint(20) DEFAULT NULL COMMENT '应用id',
  `cluster_name` varchar(256) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '集群名',
  `persist_type` varchar(20) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '存储方式',
  `cpu` double DEFAULT '1' COMMENT '节点cpu',
  `memory` double DEFAULT '1024' COMMENT '节点memory',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='工作流导出节点表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `fdp_agent_workflow_node_export_rt`
--

DROP TABLE IF EXISTS `fdp_agent_workflow_node_export_rt`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `fdp_agent_workflow_node_export_rt` (
  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT COMMENT '主键id',
  `workflow_id` bigint(20) NOT NULL COMMENT ' workflow id',
  `name` varchar(128) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '节点名称',
  `mode` varchar(20) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '节点添加模式: ui/sql',
  `description` text COLLATE utf8mb4_unicode_ci COMMENT '节点描述',
  `custom_content` text COLLATE utf8mb4_unicode_ci COMMENT '自定义编辑内容',
  `src_ds_type` varchar(20) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '源数据源类型: DL',
  `src_ds_id` bigint(20) DEFAULT NULL COMMENT '源数据源 id',
  `src_model_id` bigint(20) DEFAULT NULL COMMENT '源数据模型 id',
  `target_ds_type` varchar(20) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '目标数据源类型: CDP',
  `target_ds_id` bigint(20) DEFAULT NULL COMMENT '目标数据源 id',
  `target_model_id` bigint(20) DEFAULT NULL COMMENT '目标数据模型 id',
  `preconditions` text COLLATE utf8mb4_unicode_ci COMMENT '前置条件',
  `filter_cond` text COLLATE utf8mb4_unicode_ci COMMENT '过滤条件',
  `write_mode` varchar(32) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '写入模式',
  `created_at` datetime NOT NULL COMMENT '表记录创建时间',
  `updated_at` datetime NOT NULL COMMENT '表记录更新时间',
  `org_id` bigint(20) DEFAULT NULL COMMENT '企业id',
  `cluster_id` bigint(20) DEFAULT NULL COMMENT '集群id',
  `project_id` bigint(20) DEFAULT NULL COMMENT '项目id',
  `application_id` bigint(20) DEFAULT NULL COMMENT '应用id',
  `cluster_name` varchar(256) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '集群名',
  `persist_type` varchar(20) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '存储方式',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='工作流实时导出节点表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `fdp_agent_workflow_node_extracting`
--

DROP TABLE IF EXISTS `fdp_agent_workflow_node_extracting`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `fdp_agent_workflow_node_extracting` (
  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT COMMENT '主键id',
  `workflow_id` bigint(20) NOT NULL COMMENT ' workflow id',
  `name` varchar(128) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '节点名称',
  `mode` varchar(20) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '节点添加模式: ui/sql',
  `description` text COLLATE utf8mb4_unicode_ci COMMENT '节点描述',
  `extract_type` varchar(20) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '萃取类型: 汇总表(summary)/应用数据表(application)',
  `tb_en_name` varchar(100) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '表英文名称',
  `tb_cn_name` varchar(100) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '表中文名称',
  `tb_desc` text COLLATE utf8mb4_unicode_ci COMMENT '表描述',
  `load_strategy` varchar(20) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '加载策略: increment/total',
  `exported` tinyint(1) DEFAULT NULL COMMENT '是否导出',
  `custom_content` text COLLATE utf8mb4_unicode_ci COMMENT '自定义编辑内容',
  `main_model_id` bigint(20) DEFAULT NULL COMMENT '主模型 id',
  `main_model_alias` varchar(100) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '主模型别名',
  `main_model_filter_cond` text COLLATE utf8mb4_unicode_ci COMMENT '主模型过滤条件',
  `target_model_id` bigint(20) DEFAULT NULL COMMENT '目标模型 id',
  `target_model_alias` varchar(100) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '目标模型别名',
  `target_model_filter_cond` text COLLATE utf8mb4_unicode_ci COMMENT '目标模型过滤条件',
  `created_at` datetime NOT NULL COMMENT '表记录创建时间',
  `updated_at` datetime NOT NULL COMMENT '表记录更新时间',
  `org_id` bigint(20) DEFAULT NULL COMMENT '企业id',
  `cluster_id` bigint(20) DEFAULT NULL COMMENT '集群id',
  `project_id` bigint(20) DEFAULT NULL COMMENT '项目id',
  `application_id` bigint(20) DEFAULT NULL COMMENT '应用id',
  `cluster_name` varchar(256) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '集群名',
  `persist_type` varchar(20) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '存储方式',
  `cpu` double DEFAULT '1' COMMENT '节点cpu',
  `memory` double DEFAULT '1024' COMMENT '节点memory',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='工作流萃取节点表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `fdp_agent_workflow_node_extracting_rt`
--

DROP TABLE IF EXISTS `fdp_agent_workflow_node_extracting_rt`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `fdp_agent_workflow_node_extracting_rt` (
  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT COMMENT '主键id',
  `workflow_id` bigint(20) NOT NULL COMMENT ' workflow id',
  `name` varchar(128) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '节点名称',
  `mode` varchar(20) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '节点添加模式: ui/sql',
  `description` text COLLATE utf8mb4_unicode_ci COMMENT '节点描述',
  `extract_type` varchar(20) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '萃取类型: 汇总表(summary)/应用数据表(application)',
  `tb_en_name` varchar(100) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '表英文名称',
  `tb_cn_name` varchar(100) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '表中文名称',
  `tb_desc` text COLLATE utf8mb4_unicode_ci COMMENT '表描述',
  `load_strategy` varchar(20) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '加载策略: increment/total',
  `exported` tinyint(1) DEFAULT NULL COMMENT '是否导出',
  `custom_content` text COLLATE utf8mb4_unicode_ci COMMENT '自定义编辑内容',
  `main_model_id` bigint(20) DEFAULT NULL COMMENT '主模型 id',
  `main_model_alias` varchar(100) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '主模型别名',
  `main_model_filter_cond` text COLLATE utf8mb4_unicode_ci COMMENT '主模型过滤条件',
  `target_model_id` bigint(20) DEFAULT NULL COMMENT '目标模型 id',
  `target_model_alias` varchar(100) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '目标模型别名',
  `target_model_filter_cond` text COLLATE utf8mb4_unicode_ci COMMENT '目标模型过滤条件',
  `created_at` datetime NOT NULL COMMENT '表记录创建时间',
  `updated_at` datetime NOT NULL COMMENT '表记录更新时间',
  `org_id` bigint(20) DEFAULT NULL COMMENT '企业id',
  `cluster_id` bigint(20) DEFAULT NULL COMMENT '集群id',
  `project_id` bigint(20) DEFAULT NULL COMMENT '项目id',
  `application_id` bigint(20) DEFAULT NULL COMMENT '应用id',
  `cluster_name` varchar(256) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '集群名',
  `persist_type` varchar(20) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '存储方式',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='工作流实时萃取节点表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `fdp_agent_workflow_node_git_checkout`
--

DROP TABLE IF EXISTS `fdp_agent_workflow_node_git_checkout`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `fdp_agent_workflow_node_git_checkout` (
  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT COMMENT '主键id',
  `workflow_id` bigint(20) NOT NULL COMMENT 'workflow id',
  `name` varchar(128) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT 'git checkout',
  `git_app_id` varchar(128) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT ' git app id',
  `git_url` varchar(128) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT 'git地址',
  `created_at` datetime NOT NULL COMMENT '表记录创建时间',
  `updated_at` datetime NOT NULL COMMENT '表记录更新时间',
  `cluster_name` varchar(256) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '集群名',
  `cluster_id` bigint(20) DEFAULT NULL COMMENT '集群id',
  `project_id` bigint(20) DEFAULT NULL COMMENT '项目id',
  `application_id` bigint(20) DEFAULT NULL COMMENT '应用id',
  `org_id` bigint(20) DEFAULT NULL COMMENT '应用id',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='工作流git切换节点表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `fdp_agent_workflow_node_integration`
--

DROP TABLE IF EXISTS `fdp_agent_workflow_node_integration`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `fdp_agent_workflow_node_integration` (
  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT COMMENT '主键id',
  `workflow_id` bigint(20) NOT NULL COMMENT ' workflow id',
  `name` varchar(128) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '节点名称',
  `mode` varchar(20) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '节点添加模式: ui/sql',
  `description` text COLLATE utf8mb4_unicode_ci COMMENT '节点描述',
  `custom_content` text COLLATE utf8mb4_unicode_ci COMMENT '自定义编辑内容',
  `src_ds_type` varchar(20) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '源数据源类型: DL',
  `src_ds_id` bigint(20) DEFAULT NULL COMMENT '源数据源 id',
  `src_model_id` bigint(20) DEFAULT NULL COMMENT '源数据模型 id',
  `src_column_ids` text COLLATE utf8mb4_unicode_ci COMMENT '源数据列 ids',
  `load_strategy` varchar(20) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '加载策略: increment/total',
  `load_interval` bigint(20) DEFAULT NULL COMMENT '增量加载时间间隔',
  `load_interval_unit` varchar(20) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '增量加载时间间隔单位: MINUTE/HOUR/DAY',
  `filter_cond` text COLLATE utf8mb4_unicode_ci COMMENT '源数据模型过滤条件',
  `target_ds_type` varchar(20) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '目标数据源类型: CDP',
  `target_ds_id` bigint(20) DEFAULT NULL COMMENT '目标数据源 id',
  `target_model_id` bigint(20) DEFAULT NULL COMMENT '目标数据模型 id',
  `target_distinct` tinyint(4) NOT NULL DEFAULT '0' COMMENT '目标表去重',
  `column_default` text COLLATE utf8mb4_unicode_ci COMMENT '模型列默认值，map结构',
  `created_at` datetime NOT NULL COMMENT '表记录创建时间',
  `updated_at` datetime NOT NULL COMMENT '表记录更新时间',
  `org_id` bigint(20) DEFAULT NULL COMMENT '企业id',
  `cluster_id` bigint(20) DEFAULT NULL COMMENT '集群id',
  `project_id` bigint(20) DEFAULT NULL COMMENT '项目id',
  `application_id` bigint(20) DEFAULT NULL COMMENT '应用id',
  `cluster_name` varchar(256) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '集群名',
  `persist_type` varchar(20) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '存储方式',
  `cpu` double DEFAULT '1' COMMENT '节点cpu',
  `memory` double DEFAULT '1024' COMMENT '节点memory',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='工作流集成节点表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `fdp_agent_workflow_node_integration_rt`
--

DROP TABLE IF EXISTS `fdp_agent_workflow_node_integration_rt`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `fdp_agent_workflow_node_integration_rt` (
  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT COMMENT '主键id',
  `workflow_id` bigint(20) NOT NULL COMMENT ' workflow id',
  `name` varchar(128) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '节点名称',
  `mode` varchar(20) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '节点添加模式: ui/sql',
  `description` text COLLATE utf8mb4_unicode_ci COMMENT '节点描述',
  `custom_content` text COLLATE utf8mb4_unicode_ci COMMENT '自定义编辑内容',
  `src_ds_type` varchar(20) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '源数据源类型: DL',
  `src_ds_id` bigint(20) DEFAULT NULL COMMENT '源数据源 id',
  `src_model_id` bigint(20) DEFAULT NULL COMMENT '源数据模型 id',
  `src_column_ids` text COLLATE utf8mb4_unicode_ci COMMENT '源列ids',
  `file_path` varchar(128) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '文件路径',
  `file_type` varchar(20) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '文件类型：JSON/XML/CSV/REGULAR',
  `file_splitter` varchar(20) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '文件分隔符',
  `filter_cond` text COLLATE utf8mb4_unicode_ci COMMENT '源数据模型过滤条件',
  `target_ds_type` varchar(20) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '目标数据源类型: CDP',
  `target_ds_id` bigint(20) DEFAULT NULL COMMENT '目标数据源 id',
  `target_model_id` bigint(20) DEFAULT NULL COMMENT '目标数据模型 id',
  `target_distinct` tinyint(4) NOT NULL DEFAULT '0' COMMENT '目标表去重',
  `column_default` text COLLATE utf8mb4_unicode_ci COMMENT '模型列默认值，map结构',
  `created_at` datetime NOT NULL COMMENT '表记录创建时间',
  `updated_at` datetime NOT NULL COMMENT '表记录更新时间',
  `org_id` bigint(20) DEFAULT NULL COMMENT '企业id',
  `cluster_id` bigint(20) DEFAULT NULL COMMENT '集群id',
  `project_id` bigint(20) DEFAULT NULL COMMENT '项目id',
  `application_id` bigint(20) DEFAULT NULL COMMENT '应用id',
  `cluster_name` varchar(256) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '集群名',
  `persist_type` varchar(20) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '存储方式',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='工作流实时集成节点表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `fdp_agent_workflow_node_labelcomputing`
--

DROP TABLE IF EXISTS `fdp_agent_workflow_node_labelcomputing`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `fdp_agent_workflow_node_labelcomputing` (
  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT COMMENT '主键id',
  `workflow_id` bigint(20) NOT NULL COMMENT 'workflow id',
  `created_at` datetime NOT NULL COMMENT '表记录创建时间',
  `updated_at` datetime NOT NULL COMMENT '表记录更新时间',
  `org_id` bigint(20) DEFAULT NULL COMMENT '企业id',
  `cluster_id` bigint(20) DEFAULT NULL COMMENT '集群id',
  `project_id` bigint(20) DEFAULT NULL COMMENT '项目id',
  `application_id` bigint(20) DEFAULT NULL COMMENT '应用id',
  `cluster_name` varchar(256) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '集群名',
  `cpu` double DEFAULT '1' COMMENT '节点cpu',
  `memory` double DEFAULT '1024' COMMENT '节点memory',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='工作流标签计算节点表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `fdp_agent_workflow_node_labelcomputing_rt`
--

DROP TABLE IF EXISTS `fdp_agent_workflow_node_labelcomputing_rt`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `fdp_agent_workflow_node_labelcomputing_rt` (
  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT COMMENT '主键id',
  `workflow_id` bigint(20) NOT NULL COMMENT 'workflow id',
  `created_at` datetime NOT NULL COMMENT '表记录创建时间',
  `updated_at` datetime NOT NULL COMMENT '表记录更新时间',
  `org_id` bigint(20) DEFAULT NULL COMMENT '企业id',
  `cluster_id` bigint(20) DEFAULT NULL COMMENT '集群id',
  `project_id` bigint(20) DEFAULT NULL COMMENT '项目id',
  `application_id` bigint(20) DEFAULT NULL COMMENT '应用id',
  `cluster_name` varchar(256) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '集群名',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='工作流实时标签计算节点表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `fdp_agent_workflow_node_model_mapping`
--

DROP TABLE IF EXISTS `fdp_agent_workflow_node_model_mapping`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `fdp_agent_workflow_node_model_mapping` (
  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT COMMENT '主键id',
  `node_id` bigint(20) NOT NULL COMMENT '节点 id',
  `node_type` varchar(20) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '节点类型: cleaning/extraction',
  `mode` varchar(20) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '配置方式: ui/sql',
  `model_id` bigint(20) NOT NULL COMMENT '目标模型 id',
  `model_field` varchar(40) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '目标模型字段名称',
  `rel_model_id` bigint(20) DEFAULT NULL COMMENT '关联模型 id',
  `rel_model_field` varchar(40) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '关联模型字段名称',
  `custom_content` text COLLATE utf8mb4_unicode_ci COMMENT '自定义编辑内容',
  `created_at` datetime NOT NULL COMMENT '表记录创建时间',
  `updated_at` datetime NOT NULL COMMENT '表记录更新时间',
  `org_id` bigint(20) DEFAULT NULL COMMENT '企业id',
  `cluster_id` bigint(20) DEFAULT NULL COMMENT '集群id',
  `project_id` bigint(20) DEFAULT NULL COMMENT '项目id',
  `application_id` bigint(20) DEFAULT NULL COMMENT '应用id',
  `cluster_name` varchar(256) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '集群名',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='工作流模型节点映射表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `fdp_agent_workflow_node_one_id`
--

DROP TABLE IF EXISTS `fdp_agent_workflow_node_one_id`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `fdp_agent_workflow_node_one_id` (
  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT COMMENT '主键id',
  `workflow_id` bigint(20) NOT NULL COMMENT 'workflow id',
  `name` varchar(128) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT 'oneID 节点名称',
  `created_at` datetime NOT NULL COMMENT '表记录创建时间',
  `updated_at` datetime NOT NULL COMMENT '表记录更新时间',
  `org_id` bigint(20) DEFAULT NULL COMMENT '企业id',
  `cluster_id` bigint(20) DEFAULT NULL COMMENT '集群id',
  `project_id` bigint(20) DEFAULT NULL COMMENT '项目id',
  `application_id` bigint(20) DEFAULT NULL COMMENT '应用id',
  `cluster_name` varchar(256) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '集群名',
  `persist_type` varchar(20) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '存储方式',
  `cpu` double DEFAULT '1' COMMENT '节点cpu',
  `memory` double DEFAULT '1024' COMMENT '节点memory',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='工作流oneid节点表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `fdp_agent_workflow_node_one_id_rel_model_field`
--

DROP TABLE IF EXISTS `fdp_agent_workflow_node_one_id_rel_model_field`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `fdp_agent_workflow_node_one_id_rel_model_field` (
  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT COMMENT '主键id',
  `node_id` bigint(20) NOT NULL COMMENT '节点 id',
  `model_id` bigint(20) NOT NULL COMMENT '数据模型 id',
  `model_field` varchar(40) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '数据模型字段名称',
  `note` varchar(40) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '业务说明, 可选值: phone/appID/openID/memberID',
  `created_at` datetime NOT NULL COMMENT '表记录创建时间',
  `updated_at` datetime NOT NULL COMMENT '表记录更新时间',
  `org_id` bigint(20) DEFAULT NULL COMMENT '企业id',
  `cluster_id` bigint(20) DEFAULT NULL COMMENT '集群id',
  `project_id` bigint(20) DEFAULT NULL COMMENT '项目id',
  `application_id` bigint(20) DEFAULT NULL COMMENT '应用id',
  `cluster_name` varchar(256) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '集群名',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='工作流oneid节点模型关联表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `fdp_agent_workflow_node_python`
--

DROP TABLE IF EXISTS `fdp_agent_workflow_node_python`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `fdp_agent_workflow_node_python` (
  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT COMMENT '主键id',
  `workflow_id` bigint(20) NOT NULL COMMENT 'workflow id',
  `name` varchar(128) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT 'python任务名',
  `description` text COLLATE utf8mb4_unicode_ci COMMENT 'python描述',
  `version` varchar(128) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT 'python版本',
  `content_git_path` varchar(128) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT 'content存在路径',
  `package_git_path` varchar(128) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '安装包存放路径',
  `created_at` datetime NOT NULL COMMENT '表记录创建时间',
  `updated_at` datetime NOT NULL COMMENT '表记录更新时间',
  `cluster_name` varchar(256) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '集群名',
  `cluster_id` bigint(20) DEFAULT NULL COMMENT '集群id',
  `git_app_id` bigint(20) DEFAULT NULL COMMENT '存放的git应用id',
  `relation_git_node_id` bigint(20) DEFAULT NULL COMMENT '关联的gitnodeid',
  `project_id` bigint(20) DEFAULT NULL COMMENT '项目id',
  `application_id` bigint(20) DEFAULT NULL COMMENT '应用id',
  `org_id` bigint(20) DEFAULT NULL COMMENT '应用id',
  `persist_type` varchar(20) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '存储方式',
  `cpu` double DEFAULT '1' COMMENT '节点cpu',
  `memory` double DEFAULT '1024' COMMENT '节点memory',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='工作流python节点表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `fdp_agent_workflow_node_rel_model`
--

DROP TABLE IF EXISTS `fdp_agent_workflow_node_rel_model`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `fdp_agent_workflow_node_rel_model` (
  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT COMMENT '主键id',
  `workflow_id` bigint(20) NOT NULL COMMENT '工作流ID',
  `node_id` bigint(20) NOT NULL COMMENT '节点 id',
  `node_type` varchar(20) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '节点类型: cleaning/extraction',
  `model_id` bigint(20) NOT NULL COMMENT '关联模型 id',
  `model_alias` varchar(100) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '关联模型别名',
  `join_type` varchar(20) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT 'join类型, LEFT/RIGHT/INNER/FULL',
  `filter_cond` text COLLATE utf8mb4_unicode_ci COMMENT '关联模型过滤条件',
  `created_at` datetime NOT NULL COMMENT '表记录创建时间',
  `updated_at` datetime NOT NULL COMMENT '表记录更新时间',
  `org_id` bigint(20) DEFAULT NULL COMMENT '企业id',
  `cluster_id` bigint(20) DEFAULT NULL COMMENT '集群id',
  `project_id` bigint(20) DEFAULT NULL COMMENT '项目id',
  `application_id` bigint(20) DEFAULT NULL COMMENT '应用id',
  `cluster_name` varchar(256) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '集群名',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='工作流节点模型关联表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `fdp_agent_workflow_node_rel_model_field`
--

DROP TABLE IF EXISTS `fdp_agent_workflow_node_rel_model_field`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `fdp_agent_workflow_node_rel_model_field` (
  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT COMMENT '主键id',
  `node_id` bigint(20) NOT NULL COMMENT '节点 id',
  `node_type` varchar(20) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '节点类型: cleaning/extraction',
  `mode` varchar(20) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '节点添加模式: ui/sql',
  `model_id` bigint(20) NOT NULL COMMENT '关联模型 id',
  `model_field` varchar(40) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '关联模型字段名称',
  `rel_model_id` bigint(20) NOT NULL COMMENT '被关联模型 id',
  `rel_model_field` varchar(40) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '被关联模型字段名称',
  `custom_content` text COLLATE utf8mb4_unicode_ci COMMENT '自定义编辑内容',
  `created_at` datetime NOT NULL COMMENT '表记录创建时间',
  `updated_at` datetime NOT NULL COMMENT '表记录更新时间',
  `org_id` bigint(20) DEFAULT NULL COMMENT '企业id',
  `cluster_id` bigint(20) DEFAULT NULL COMMENT '集群id',
  `project_id` bigint(20) DEFAULT NULL COMMENT '项目id',
  `application_id` bigint(20) DEFAULT NULL COMMENT '应用id',
  `cluster_name` varchar(256) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '集群名',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='工作流节点模型字段关联表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `fdp_agent_workflow_period_statistics`
--

DROP TABLE IF EXISTS `fdp_agent_workflow_period_statistics`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `fdp_agent_workflow_period_statistics` (
  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT COMMENT '主键id',
  `statistics` varchar(1024) DEFAULT NULL COMMENT '每日完成数量',
  `statistical_date` varchar(128) DEFAULT NULL COMMENT '统计日期',
  `created_at` datetime NOT NULL COMMENT '表记录创建时间',
  `updated_at` datetime NOT NULL COMMENT '表记录更新时间',
  `cluster_id` bigint(20) DEFAULT NULL COMMENT '集群id',
  `project_id` bigint(20) DEFAULT '1' COMMENT '项目id',
  `org_id` bigint(20) DEFAULT NULL COMMENT '集群id',
  `application_id` bigint(20) DEFAULT NULL COMMENT '应用id',
  `cluster_name` varchar(256) DEFAULT '' COMMENT '集群名称',
  PRIMARY KEY (`id`),
  KEY `idx_statistical_date` (`statistical_date`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8 COMMENT='工作流日期统计表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `fdp_agent_workflows`
--

DROP TABLE IF EXISTS `fdp_agent_workflows`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `fdp_agent_workflows` (
  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT COMMENT '主键id',
  `org_id` bigint(20) DEFAULT NULL COMMENT '企业 id',
  `source` varchar(64) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT '平台来源：DL/CDP',
  `name` varchar(128) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT '工作流名',
  `name_pinyin` varchar(128) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT '工作流拼音',
  `description` varchar(1024) COLLATE utf8mb4_unicode_ci DEFAULT '' COMMENT '工作流描述',
  `run_type` varchar(64) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT '运行类型: ATONCE、SCHEDULE',
  `cron` varchar(128) COLLATE utf8mb4_unicode_ci DEFAULT '' COMMENT '任务周期',
  `cron_start_from` datetime DEFAULT NULL COMMENT '延时执行时间',
  `category_id` bigint(20) NOT NULL COMMENT '工作流目录 ID',
  `pipeline_name` varchar(128) COLLATE utf8mb4_unicode_ci DEFAULT '' COMMENT 'pipeline 名称',
  `pipeline` mediumtext COLLATE utf8mb4_unicode_ci COMMENT 'pipeline 内容, 最大16MB',
  `locations` text COLLATE utf8mb4_unicode_ci COMMENT '工作流节点坐标二维数组',
  `node_params` text COLLATE utf8mb4_unicode_ci COMMENT '工作流节点参数',
  `creator_id` varchar(128) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT '创建者 ID',
  `updater_id` varchar(128) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '创建者 ID',
  `extra` mediumtext COLLATE utf8mb4_unicode_ci COMMENT '扩展信息',
  `delete_yn` tinyint(1) NOT NULL DEFAULT '0' COMMENT '逻辑删除标记',
  `created_at` datetime NOT NULL COMMENT '创建时间',
  `updated_at` datetime NOT NULL COMMENT '更新时间',
  `cluster_id` bigint(20) DEFAULT NULL COMMENT '集群id',
  `project_id` bigint(20) DEFAULT NULL COMMENT '项目id',
  `application_id` bigint(20) DEFAULT NULL COMMENT '应用id',
  `cluster_name` varchar(256) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '集群名',
  `pipeline_status` varchar(128) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT 'pipeline最近一次执行状态',
  `pipeline_begin_time` datetime DEFAULT NULL COMMENT 'pipeline最近一次执行时间',
  `task_queue_id` bigint(20) DEFAULT '1' COMMENT '队列id',
  `queue_level` varchar(128) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '队列优先级',
  PRIMARY KEY (`id`),
  KEY `idx_category_id_name` (`category_id`,`name`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='工作流表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `fdp_master_data_sync`
--

DROP TABLE IF EXISTS `fdp_master_data_sync`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `fdp_master_data_sync` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT,
  `type` varchar(128) NOT NULL,
  `sync` tinyint(1) NOT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `fdp_master_event_box_log`
--

DROP TABLE IF EXISTS `fdp_master_event_box_log`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `fdp_master_event_box_log` (
  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT,
  `queue_id` varchar(128) DEFAULT NULL COMMENT '队列id',
  `project_id` varchar(128) DEFAULT NULL COMMENT '空间id',
  `status` varchar(255) DEFAULT NULL COMMENT '状态',
  `pipeline_yml_name` varchar(255) DEFAULT NULL COMMENT 'yaml',
  `cluster_name` varchar(128) DEFAULT NULL COMMENT '集群名',
  `h_key` varchar(128) DEFAULT NULL COMMENT 'hkey',
  `count` varchar(128) DEFAULT NULL COMMENT 'count值',
  `request_content` longtext COMMENT '回调原文',
  `pipeline_id` bigint(20) DEFAULT NULL COMMENT 'pipeline_id',
  `created_at` datetime NOT NULL COMMENT '表记录创建时间',
  PRIMARY KEY (`id`),
  KEY `hkey` (`h_key`,`project_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='回调表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `fdp_master_gittar_repo`
--

DROP TABLE IF EXISTS `fdp_master_gittar_repo`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `fdp_master_gittar_repo` (
  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT COMMENT '主键id',
  `gittar_project_id` bigint(20) DEFAULT NULL COMMENT 'gittar project id',
  `gittar_project_name` varchar(128) DEFAULT NULL COMMENT 'gittar project 名字',
  `gittar_app_id` bigint(20) DEFAULT NULL COMMENT 'gittar应用 id',
  `gittar_app_name` varchar(128) DEFAULT NULL COMMENT 'gittar应用名',
  `repo_url` varchar(256) DEFAULT NULL COMMENT '仓库地址',
  `repo_abbrev` varchar(128) DEFAULT NULL COMMENT '相对地址',
  `token` varchar(128) DEFAULT NULL COMMENT '仓库token',
  `org_name` varchar(128) DEFAULT NULL COMMENT '企业id名字',
  `delete_yn` tinyint(1) NOT NULL DEFAULT '0' COMMENT '逻辑删除标记',
  `creator_id` varchar(128) DEFAULT NULL COMMENT '创建人 ID',
  `updater_id` varchar(128) DEFAULT NULL COMMENT '更新者ID',
  `created_at` datetime NOT NULL COMMENT '创建时间',
  `updated_at` datetime NOT NULL COMMENT '更新时间',
  `org_id` bigint(20) DEFAULT NULL COMMENT '企业 id',
  `project_id` bigint(20) DEFAULT NULL COMMENT '项目 id',
  `application_id` bigint(20) DEFAULT NULL COMMENT '应用 id',
  `cluster_id` bigint(20) DEFAULT NULL COMMENT '集群 id',
  `cluster_name` varchar(256) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '集群名',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8 COMMENT='gittar仓库表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `fdp_master_integration_service_log`
--

DROP TABLE IF EXISTS `fdp_master_integration_service_log`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `fdp_master_integration_service_log` (
  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT COMMENT '主键id',
  `name` varchar(128) DEFAULT NULL COMMENT '第三方服务名称',
  `request_header` longtext COMMENT '请求透',
  `request_body` longtext COMMENT '请求体',
  `request_ok` tinyint(4) DEFAULT NULL COMMENT '是否请求成功',
  `request_url` longtext COMMENT '请求url',
  `request_params` longtext COMMENT 'query的参数',
  `response_data` longtext COMMENT '返回的data',
  `response_body` longtext COMMENT '返回body全文',
  `remark` varchar(256) DEFAULT NULL COMMENT '备注信息',
  `err` longtext COMMENT '错误信息',
  `creator_id` varchar(128) DEFAULT NULL COMMENT '创建者id',
  `created_at` datetime NOT NULL COMMENT '创建时间',
  `updated_at` datetime NOT NULL COMMENT '更新时间',
  PRIMARY KEY (`id`),
  KEY `request_url` (`request_url`(128)),
  KEY `request_ok` (`request_ok`)
) ENGINE=InnoDB AUTO_INCREMENT=40 DEFAULT CHARSET=utf8 COMMENT='集成的第三方服务日志';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `fdp_master_model_workflows`
--

DROP TABLE IF EXISTS `fdp_master_model_workflows`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `fdp_master_model_workflows` (
  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT COMMENT '主键id',
  `org_id` bigint(20) DEFAULT NULL COMMENT '企业 id',
  `model_manage_id` bigint(20) NOT NULL COMMENT '模型管理ID',
  `process_type` varchar(128) NOT NULL COMMENT '处理类型',
  `name` varchar(128) NOT NULL DEFAULT '' COMMENT '工作流名',
  `name_pinyin` varchar(128) NOT NULL DEFAULT '' COMMENT '工作流拼音',
  `description` varchar(1024) NOT NULL DEFAULT '' COMMENT '工作流描述',
  `source` varchar(32) DEFAULT '' COMMENT 'workflow 来源: dl/cdp',
  `run_type` varchar(64) NOT NULL DEFAULT '' COMMENT '运行类型: ATONCE、SCHEDULE',
  `cron` varchar(128) DEFAULT '' COMMENT '任务周期',
  `cron_start_from` datetime DEFAULT NULL COMMENT '延时执行时间',
  `pipeline_name` varchar(128) DEFAULT '' COMMENT 'pipeline 名称',
  `pipeline` mediumtext NOT NULL COMMENT 'pipeline 内容',
  `pipeline_id` bigint(20) DEFAULT NULL COMMENT 'pipeline id',
  `node_params` text COMMENT '工作流节点参数',
  `locations` varchar(1024) DEFAULT NULL,
  `creator_id` varchar(128) DEFAULT NULL COMMENT '创建者 ID',
  `updater_id` varchar(128) DEFAULT NULL COMMENT '更新者ID',
  `extra` mediumtext COMMENT '扩展信息',
  `delete_yn` tinyint(1) NOT NULL DEFAULT '0' COMMENT '逻辑删除标记',
  `created_at` datetime DEFAULT NULL COMMENT '创建时间',
  `updated_at` datetime DEFAULT NULL COMMENT '更新时间',
  `cluster_name` varchar(256) DEFAULT NULL COMMENT '集群名',
  `cluster_id` bigint(20) DEFAULT NULL COMMENT '集群id',
  `project_id` bigint(20) DEFAULT NULL COMMENT '项目id',
  `application_id` bigint(20) DEFAULT NULL COMMENT '应用id',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8 COMMENT='模型管理工作流表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `fdp_master_oneid_rules`
--

DROP TABLE IF EXISTS `fdp_master_oneid_rules`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `fdp_master_oneid_rules` (
  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT COMMENT '主键id',
  `rule_code` varchar(64) DEFAULT NULL,
  `rule_desc` varchar(64) NOT NULL COMMENT '规则描述',
  `is_unique` tinyint(1) NOT NULL DEFAULT '1' COMMENT '是否唯一',
  `weight` int(11) NOT NULL COMMENT '权重',
  `delete_yn` tinyint(1) NOT NULL DEFAULT '0' COMMENT '逻辑删除标记',
  `creator_id` varchar(128) DEFAULT NULL COMMENT '创建人 ID',
  `updater_id` varchar(128) DEFAULT NULL COMMENT '更新者ID',
  `created_at` datetime NOT NULL COMMENT '创建时间',
  `updated_at` datetime NOT NULL COMMENT '更新时间',
  `cluster_id` bigint(20) DEFAULT NULL COMMENT '集群id',
  `project_id` bigint(20) DEFAULT NULL COMMENT '项目id',
  `application_id` bigint(20) DEFAULT NULL COMMENT '应用id',
  `cluster_name` varchar(256) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '集群名',
  `org_id` bigint(20) DEFAULT NULL COMMENT '应用id',
  PRIMARY KEY (`id`),
  UNIQUE KEY `rule_code` (`rule_code`)
) ENGINE=InnoDB AUTO_INCREMENT=26 DEFAULT CHARSET=utf8 COMMENT='oneid规则表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `fdp_master_reco_scenario_workflows`
--

DROP TABLE IF EXISTS `fdp_master_reco_scenario_workflows`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `fdp_master_reco_scenario_workflows` (
  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT COMMENT '主键id',
  `org_id` bigint(20) DEFAULT NULL COMMENT '企业 id',
  `scenario_name` varchar(128) NOT NULL DEFAULT '',
  `scenario_code` varchar(128) NOT NULL DEFAULT '',
  `process_type` varchar(128) NOT NULL COMMENT '处理类型',
  `name` varchar(128) NOT NULL DEFAULT '' COMMENT '工作流名',
  `name_pinyin` varchar(128) NOT NULL DEFAULT '' COMMENT '工作流拼音',
  `description` varchar(1024) NOT NULL DEFAULT '' COMMENT '工作流描述',
  `source` varchar(32) DEFAULT '' COMMENT 'workflow 来源: dl/cdp',
  `run_type` varchar(64) NOT NULL DEFAULT '' COMMENT '运行类型: ATONCE、SCHEDULE',
  `cron` varchar(128) DEFAULT '' COMMENT '任务周期',
  `cron_start_from` datetime DEFAULT NULL COMMENT '延时执行时间',
  `category_id` bigint(20) NOT NULL COMMENT '工作流目录 ID',
  `pipeline_name` varchar(128) DEFAULT '' COMMENT 'pipeline 名称',
  `pipeline` mediumtext NOT NULL COMMENT 'pipeline 内容',
  `pipeline_id` bigint(20) DEFAULT NULL COMMENT 'pipeline id',
  `node_params` text COMMENT '工作流节点参数',
  `locations` varchar(1024) DEFAULT NULL,
  `creator_id` varchar(128) CHARACTER SET utf8mb4 DEFAULT '' COMMENT '创建者 ID',
  `updater_id` varchar(128) DEFAULT NULL COMMENT '更新者ID',
  `extra` mediumtext COMMENT '扩展信息',
  `delete_yn` tinyint(1) NOT NULL DEFAULT '0' COMMENT '逻辑删除标记',
  `created_at` datetime DEFAULT NULL COMMENT '创建时间',
  `updated_at` datetime DEFAULT NULL COMMENT '更新时间',
  PRIMARY KEY (`id`),
  KEY `idx_category_id_name` (`scenario_code`,`process_type`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8 COMMENT='场景工作流表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `fdp_master_sql_history`
--

DROP TABLE IF EXISTS `fdp_master_sql_history`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `fdp_master_sql_history` (
  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT COMMENT '主键id',
  `org_id` bigint(20) DEFAULT NULL COMMENT '企业 id',
  `project_id` bigint(20) DEFAULT NULL COMMENT '项目 id',
  `application_id` bigint(20) DEFAULT NULL COMMENT '应用 id',
  `cluster_id` bigint(20) DEFAULT NULL COMMENT '集群 id',
  `name` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT '执行记录名称',
  `query_status` varchar(128) NOT NULL COMMENT '查询响应状态',
  `query_time` bigint(20) NOT NULL COMMENT '查询耗时',
  `sql` text NOT NULL COMMENT '查询sql',
  `data_model_name` varchar(255) NOT NULL COMMENT '数据模型 名称',
  `data_source_id` bigint(20) NOT NULL COMMENT '数据源 ID',
  `delete_yn` tinyint(1) NOT NULL DEFAULT '0' COMMENT '逻辑删除标记',
  `creator_id` varchar(128) DEFAULT NULL COMMENT '创建人 ID',
  `updater_id` varchar(128) DEFAULT NULL COMMENT '更新者ID',
  `created_at` datetime NOT NULL COMMENT '创建时间',
  `updated_at` datetime NOT NULL COMMENT '更新时间',
  `cluster_name` varchar(256) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '集群名',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8 COMMENT='sql执行历史表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `fdp_master_workflow_categories`
--

DROP TABLE IF EXISTS `fdp_master_workflow_categories`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `fdp_master_workflow_categories` (
  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT COMMENT '主键id',
  `org_id` bigint(20) DEFAULT NULL COMMENT '企业 id',
  `source` varchar(64) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT '平台来源：DL/CDP',
  `name` varchar(128) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT '目录名',
  `name_pinyin` varchar(128) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT '目录名拼音',
  `parent_id` bigint(20) NOT NULL COMMENT '上级目录 ID',
  `creator_id` varchar(256) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT '创建人 ID',
  `created_at` datetime NOT NULL COMMENT '表记录创建时间',
  `updated_at` datetime NOT NULL COMMENT '表记录更新时间',
  `cluster_id` bigint(20) DEFAULT NULL COMMENT '集群id',
  `project_id` bigint(20) DEFAULT NULL COMMENT '项目id',
  `application_id` bigint(20) DEFAULT NULL COMMENT '应用id',
  `cluster_name` varchar(128) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `workflow_category_unique_index1` (`parent_id`,`source`,`name`,`cluster_name`,`project_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci BLOCK_FORMAT=ENCRYPTED COMMENT='工作流目录表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `fdp_master_workflow_copy_detail`
--

DROP TABLE IF EXISTS `fdp_master_workflow_copy_detail`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `fdp_master_workflow_copy_detail` (
  `id` int(11) NOT NULL AUTO_INCREMENT COMMENT '主键',
  `copy_log_id` int(11) NOT NULL COMMENT '关联字段',
  `workflow_name` varchar(256) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '工作流名称',
  `operation_type` varchar(128) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '操作类型',
  `copy_success` tinyint(1) DEFAULT '0' COMMENT '复制是否成功',
  `error_reason` text COLLATE utf8mb4_unicode_ci COMMENT '失败原因',
  `begin_date` datetime DEFAULT NULL COMMENT '复制开始时间',
  `end_date` datetime DEFAULT NULL COMMENT '复制结束时间',
  PRIMARY KEY (`id`),
  KEY `workflow_copy_fk1` (`copy_log_id`),
  CONSTRAINT `workflow_copy_fk1` FOREIGN KEY (`copy_log_id`) REFERENCES `fdp_master_workflow_copy_log` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `fdp_master_workflow_copy_log`
--

DROP TABLE IF EXISTS `fdp_master_workflow_copy_log`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `fdp_master_workflow_copy_log` (
  `id` int(11) NOT NULL AUTO_INCREMENT COMMENT '主键',
  `name` varchar(256) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '名称',
  `created_date` datetime NOT NULL COMMENT '创建时间',
  `file_success` tinyint(1) DEFAULT '0' COMMENT '文件生成是否成功',
  `error_reason` text COLLATE utf8mb4_unicode_ci COMMENT '失败原因',
  `file_path` varchar(256) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '文件路径',
  `file_date` datetime DEFAULT NULL COMMENT '文件生成成功时间',
  `complete` tinyint(1) DEFAULT '0' COMMENT '复制是否完成',
  `complete_date` datetime DEFAULT NULL COMMENT '复制完成时间',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `fdp_master_workflow_dependencies`
--

DROP TABLE IF EXISTS `fdp_master_workflow_dependencies`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `fdp_master_workflow_dependencies` (
  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT COMMENT '主键id',
  `workflow_id` bigint(20) NOT NULL COMMENT '工作流 ID',
  `depend_workflow_id` bigint(20) NOT NULL COMMENT '依赖的工作流 ID',
  `creator_id` varchar(128) NOT NULL DEFAULT '' COMMENT '创建者 ID',
  `updater_id` varchar(128) NOT NULL DEFAULT '' COMMENT '更新着 ID',
  `created_at` datetime NOT NULL COMMENT '表记录创建时间',
  `updated_at` datetime NOT NULL COMMENT '表记录更新时间',
  `cluster_id` bigint(20) DEFAULT NULL COMMENT '集群id',
  `project_id` bigint(20) DEFAULT NULL COMMENT '项目id',
  `application_id` bigint(20) DEFAULT NULL COMMENT '应用id',
  `cluster_name` varchar(256) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '集群名',
  `org_id` bigint(20) DEFAULT NULL COMMENT '企业 id',
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_dependWorkflowId_under_workflowId` (`workflow_id`,`depend_workflow_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 BLOCK_FORMAT=ENCRYPTED COMMENT='工作流依赖表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `fdp_master_workflow_node_cleaning`
--

DROP TABLE IF EXISTS `fdp_master_workflow_node_cleaning`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `fdp_master_workflow_node_cleaning` (
  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT COMMENT '主键id',
  `workflow_id` bigint(20) NOT NULL COMMENT ' workflow id',
  `name` varchar(128) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '节点名称',
  `mode` varchar(20) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '节点添加模式: ui/sql',
  `description` text COLLATE utf8mb4_unicode_ci COMMENT '节点描述',
  `clean_type` varchar(20) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '清洗类型: 维度表(dimension)/事实表(fact)',
  `tb_en_name` varchar(100) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '表英文名称',
  `tb_cn_name` varchar(100) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '表中文名称',
  `tb_desc` text COLLATE utf8mb4_unicode_ci COMMENT '表描述',
  `load_strategy` varchar(20) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '加载策略: increment/total',
  `custom_content` text COLLATE utf8mb4_unicode_ci COMMENT '自定义编辑内容',
  `main_model_id` bigint(20) DEFAULT NULL COMMENT '主模型 id',
  `main_model_alias` varchar(100) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '主模型别名',
  `main_model_filter_cond` text COLLATE utf8mb4_unicode_ci COMMENT '主模型过滤条件',
  `target_model_id` bigint(20) DEFAULT NULL COMMENT '目标模型 id',
  `target_model_alias` varchar(100) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '目标模型别名',
  `target_model_filter_cond` text COLLATE utf8mb4_unicode_ci COMMENT '目标模型过滤条件',
  `created_at` datetime NOT NULL COMMENT '表记录创建时间',
  `updated_at` datetime NOT NULL COMMENT '表记录更新时间',
  `org_id` bigint(20) DEFAULT NULL COMMENT '企业id',
  `cluster_id` bigint(20) DEFAULT NULL COMMENT '集群id',
  `project_id` bigint(20) DEFAULT NULL COMMENT '项目id',
  `application_id` bigint(20) DEFAULT NULL COMMENT '应用id',
  `cluster_name` varchar(256) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '集群名',
  `persist_type` varchar(20) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '存储方式',
  `cpu` double DEFAULT '1' COMMENT '节点cpu',
  `memory` double DEFAULT '1024' COMMENT '节点memory',
  `create_type` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT 'CREATE' COMMENT '清洗节点创建方式',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci BLOCK_FORMAT=ENCRYPTED COMMENT='工作流清洗节点表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `fdp_master_workflow_node_cleaning_rt`
--

DROP TABLE IF EXISTS `fdp_master_workflow_node_cleaning_rt`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `fdp_master_workflow_node_cleaning_rt` (
  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT COMMENT '主键id',
  `workflow_id` bigint(20) NOT NULL COMMENT ' workflow id',
  `name` varchar(128) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '节点名称',
  `mode` varchar(20) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '节点添加模式: ui/sql',
  `description` text COLLATE utf8mb4_unicode_ci COMMENT '节点描述',
  `clean_type` varchar(20) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '清洗类型: 维度表(dimension)/事实表(fact)',
  `tb_en_name` varchar(100) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '表英文名称',
  `tb_cn_name` varchar(100) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '表中文名称',
  `tb_desc` text COLLATE utf8mb4_unicode_ci COMMENT '表描述',
  `load_strategy` varchar(20) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '加载策略: increment/total',
  `custom_content` text COLLATE utf8mb4_unicode_ci COMMENT '自定义编辑内容',
  `main_model_id` bigint(20) DEFAULT NULL COMMENT '主模型 id',
  `main_model_alias` varchar(100) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '主模型别名',
  `main_model_filter_cond` text COLLATE utf8mb4_unicode_ci COMMENT '主模型过滤条件',
  `target_model_id` bigint(20) DEFAULT NULL COMMENT '目标模型 id',
  `target_model_alias` varchar(100) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '目标模型别名',
  `target_model_filter_cond` text COLLATE utf8mb4_unicode_ci COMMENT '目标模型过滤条件',
  `created_at` datetime NOT NULL COMMENT '表记录创建时间',
  `updated_at` datetime NOT NULL COMMENT '表记录更新时间',
  `org_id` bigint(20) DEFAULT NULL COMMENT '企业id',
  `cluster_id` bigint(20) DEFAULT NULL COMMENT '集群id',
  `project_id` bigint(20) DEFAULT NULL COMMENT '项目id',
  `application_id` bigint(20) DEFAULT NULL COMMENT '应用id',
  `cluster_name` varchar(256) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '集群名',
  `persist_type` varchar(20) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '存储方式',
  `cpu` double DEFAULT '1' COMMENT '节点cpu',
  `memory` double DEFAULT '1024' COMMENT '节点memory',
  `create_type` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT 'CREATE' COMMENT '清洗节点创建方式:SELECT,CREATE',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='工作流实时清洗节点表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `fdp_master_workflow_node_data_cancel`
--

DROP TABLE IF EXISTS `fdp_master_workflow_node_data_cancel`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `fdp_master_workflow_node_data_cancel` (
  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT COMMENT '主键id',
  `workflow_id` bigint(20) NOT NULL COMMENT 'workflow id',
  `created_at` datetime NOT NULL COMMENT '表记录创建时间',
  `updated_at` datetime NOT NULL COMMENT '表记录更新时间',
  `org_id` bigint(20) DEFAULT NULL COMMENT '企业id',
  `cluster_id` bigint(20) DEFAULT NULL COMMENT '集群id',
  `project_id` bigint(20) DEFAULT NULL COMMENT '项目id',
  `application_id` bigint(20) DEFAULT NULL COMMENT '应用id',
  `cluster_name` varchar(256) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '集群名',
  `persist_type` varchar(20) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '存储方式',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=12 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='工作流取消节点表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `fdp_master_workflow_node_export`
--

DROP TABLE IF EXISTS `fdp_master_workflow_node_export`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `fdp_master_workflow_node_export` (
  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT COMMENT '主键id',
  `workflow_id` bigint(20) NOT NULL COMMENT ' workflow id',
  `name` varchar(128) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '节点名称',
  `mode` varchar(20) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '节点添加模式: ui/sql',
  `description` text COLLATE utf8mb4_unicode_ci COMMENT '节点描述',
  `custom_content` text COLLATE utf8mb4_unicode_ci COMMENT '自定义编辑内容',
  `src_ds_type` varchar(20) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '源数据源类型: DL',
  `src_ds_id` bigint(20) DEFAULT NULL COMMENT '源数据源 id',
  `src_model_id` bigint(20) DEFAULT NULL COMMENT '源数据模型 id',
  `load_strategy` varchar(20) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '加载策略: increment/total',
  `load_interval` bigint(20) DEFAULT NULL COMMENT '增量加载时间间隔',
  `load_interval_unit` varchar(20) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '增量加载时间间隔单位: MINUTE/HOUR/DAY',
  `target_ds_type` varchar(20) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '目标数据源类型: CDP',
  `target_ds_id` bigint(20) DEFAULT NULL COMMENT '目标数据源 id',
  `target_model_id` bigint(20) DEFAULT NULL COMMENT '目标数据模型 id',
  `preconditions` text COLLATE utf8mb4_unicode_ci COMMENT '前置条件',
  `filter_cond` text COLLATE utf8mb4_unicode_ci COMMENT '过滤条件',
  `write_mode` varchar(32) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '写入模式',
  `created_at` datetime NOT NULL COMMENT '表记录创建时间',
  `updated_at` datetime NOT NULL COMMENT '表记录更新时间',
  `org_id` bigint(20) DEFAULT NULL COMMENT '企业id',
  `cluster_id` bigint(20) DEFAULT NULL COMMENT '集群id',
  `project_id` bigint(20) DEFAULT NULL COMMENT '项目id',
  `application_id` bigint(20) DEFAULT NULL COMMENT '应用id',
  `cluster_name` varchar(256) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '集群名',
  `persist_type` varchar(20) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '存储方式',
  `cpu` double DEFAULT '1' COMMENT '节点cpu',
  `memory` double DEFAULT '1024' COMMENT '节点memory',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci BLOCK_FORMAT=ENCRYPTED COMMENT='工作流导出节点表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `fdp_master_workflow_node_export_rt`
--

DROP TABLE IF EXISTS `fdp_master_workflow_node_export_rt`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `fdp_master_workflow_node_export_rt` (
  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT COMMENT '主键id',
  `workflow_id` bigint(20) NOT NULL COMMENT ' workflow id',
  `name` varchar(128) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '节点名称',
  `mode` varchar(20) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '节点添加模式: ui/sql',
  `description` text COLLATE utf8mb4_unicode_ci COMMENT '节点描述',
  `custom_content` text COLLATE utf8mb4_unicode_ci COMMENT '自定义编辑内容',
  `src_ds_type` varchar(20) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '源数据源类型: DL',
  `src_ds_id` bigint(20) DEFAULT NULL COMMENT '源数据源 id',
  `src_model_id` bigint(20) DEFAULT NULL COMMENT '源数据模型 id',
  `target_ds_type` varchar(20) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '目标数据源类型: CDP',
  `target_ds_id` bigint(20) DEFAULT NULL COMMENT '目标数据源 id',
  `target_model_id` bigint(20) DEFAULT NULL COMMENT '目标数据模型 id',
  `preconditions` text COLLATE utf8mb4_unicode_ci COMMENT '前置条件',
  `filter_cond` text COLLATE utf8mb4_unicode_ci COMMENT '过滤条件',
  `write_mode` varchar(32) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '写入模式',
  `created_at` datetime NOT NULL COMMENT '表记录创建时间',
  `updated_at` datetime NOT NULL COMMENT '表记录更新时间',
  `org_id` bigint(20) DEFAULT NULL COMMENT '企业id',
  `cluster_id` bigint(20) DEFAULT NULL COMMENT '集群id',
  `project_id` bigint(20) DEFAULT NULL COMMENT '项目id',
  `application_id` bigint(20) DEFAULT NULL COMMENT '应用id',
  `cluster_name` varchar(256) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '集群名',
  `persist_type` varchar(20) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '存储方式',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='工作流实时导出节点表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `fdp_master_workflow_node_extracting`
--

DROP TABLE IF EXISTS `fdp_master_workflow_node_extracting`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `fdp_master_workflow_node_extracting` (
  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT COMMENT '主键id',
  `workflow_id` bigint(20) NOT NULL COMMENT ' workflow id',
  `name` varchar(128) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '节点名称',
  `mode` varchar(20) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '节点添加模式: ui/sql',
  `description` text COLLATE utf8mb4_unicode_ci COMMENT '节点描述',
  `extract_type` varchar(20) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '萃取类型: 汇总表(summary)/应用数据表(application)',
  `tb_en_name` varchar(100) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '表英文名称',
  `tb_cn_name` varchar(100) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '表中文名称',
  `tb_desc` text COLLATE utf8mb4_unicode_ci COMMENT '表描述',
  `load_strategy` varchar(20) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '加载策略: increment/total',
  `exported` tinyint(1) DEFAULT NULL COMMENT '是否导出',
  `custom_content` text COLLATE utf8mb4_unicode_ci COMMENT '自定义编辑内容',
  `main_model_id` bigint(20) DEFAULT NULL COMMENT '主模型 id',
  `main_model_alias` varchar(100) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '主模型别名',
  `main_model_filter_cond` text COLLATE utf8mb4_unicode_ci COMMENT '主模型过滤条件',
  `target_model_id` bigint(20) DEFAULT NULL COMMENT '目标模型 id',
  `target_model_alias` varchar(100) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '目标模型别名',
  `target_model_filter_cond` text COLLATE utf8mb4_unicode_ci COMMENT '目标模型过滤条件',
  `created_at` datetime NOT NULL COMMENT '表记录创建时间',
  `updated_at` datetime NOT NULL COMMENT '表记录更新时间',
  `org_id` bigint(20) DEFAULT NULL COMMENT '企业id',
  `cluster_id` bigint(20) DEFAULT NULL COMMENT '集群id',
  `project_id` bigint(20) DEFAULT NULL COMMENT '项目id',
  `application_id` bigint(20) DEFAULT NULL COMMENT '应用id',
  `cluster_name` varchar(256) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '集群名',
  `persist_type` varchar(20) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '存储方式',
  `cpu` double DEFAULT '1' COMMENT '节点cpu',
  `memory` double DEFAULT '1024' COMMENT '节点memory',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci BLOCK_FORMAT=ENCRYPTED COMMENT='工作流萃取节点表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `fdp_master_workflow_node_extracting_rt`
--

DROP TABLE IF EXISTS `fdp_master_workflow_node_extracting_rt`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `fdp_master_workflow_node_extracting_rt` (
  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT COMMENT '主键id',
  `workflow_id` bigint(20) NOT NULL COMMENT ' workflow id',
  `name` varchar(128) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '节点名称',
  `mode` varchar(20) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '节点添加模式: ui/sql',
  `description` text COLLATE utf8mb4_unicode_ci COMMENT '节点描述',
  `extract_type` varchar(20) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '萃取类型: 汇总表(summary)/应用数据表(application)',
  `tb_en_name` varchar(100) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '表英文名称',
  `tb_cn_name` varchar(100) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '表中文名称',
  `tb_desc` text COLLATE utf8mb4_unicode_ci COMMENT '表描述',
  `load_strategy` varchar(20) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '加载策略: increment/total',
  `exported` tinyint(1) DEFAULT NULL COMMENT '是否导出',
  `custom_content` text COLLATE utf8mb4_unicode_ci COMMENT '自定义编辑内容',
  `main_model_id` bigint(20) DEFAULT NULL COMMENT '主模型 id',
  `main_model_alias` varchar(100) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '主模型别名',
  `main_model_filter_cond` text COLLATE utf8mb4_unicode_ci COMMENT '主模型过滤条件',
  `target_model_id` bigint(20) DEFAULT NULL COMMENT '目标模型 id',
  `target_model_alias` varchar(100) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '目标模型别名',
  `target_model_filter_cond` text COLLATE utf8mb4_unicode_ci COMMENT '目标模型过滤条件',
  `created_at` datetime NOT NULL COMMENT '表记录创建时间',
  `updated_at` datetime NOT NULL COMMENT '表记录更新时间',
  `org_id` bigint(20) DEFAULT NULL COMMENT '企业id',
  `cluster_id` bigint(20) DEFAULT NULL COMMENT '集群id',
  `project_id` bigint(20) DEFAULT NULL COMMENT '项目id',
  `application_id` bigint(20) DEFAULT NULL COMMENT '应用id',
  `cluster_name` varchar(256) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '集群名',
  `persist_type` varchar(20) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '存储方式',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='工作流实时萃取节点表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `fdp_master_workflow_node_git_checkout`
--

DROP TABLE IF EXISTS `fdp_master_workflow_node_git_checkout`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `fdp_master_workflow_node_git_checkout` (
  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT COMMENT '主键id',
  `workflow_id` bigint(20) NOT NULL COMMENT 'workflow id',
  `name` varchar(128) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT 'git checkout',
  `git_app_id` varchar(128) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT ' git app id',
  `git_url` varchar(128) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT 'git地址',
  `created_at` datetime NOT NULL COMMENT '表记录创建时间',
  `updated_at` datetime NOT NULL COMMENT '表记录更新时间',
  `cluster_name` varchar(256) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '集群名',
  `cluster_id` bigint(20) DEFAULT NULL COMMENT '集群id',
  `project_id` bigint(20) DEFAULT NULL COMMENT '项目id',
  `application_id` bigint(20) DEFAULT NULL COMMENT '应用id',
  `org_id` bigint(20) DEFAULT NULL COMMENT '应用id',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='工作流git切换节点表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `fdp_master_workflow_node_groupcomputing`
--

DROP TABLE IF EXISTS `fdp_master_workflow_node_groupcomputing`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `fdp_master_workflow_node_groupcomputing` (
  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT COMMENT '主键id',
  `workflow_id` bigint(20) NOT NULL COMMENT 'workflow id',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '表记录创建时间',
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '表记录更新时间',
  `org_id` bigint(20) NOT NULL DEFAULT '0' COMMENT '企业id',
  `cluster_id` bigint(20) NOT NULL DEFAULT '0' COMMENT '集群id',
  `project_id` bigint(20) NOT NULL DEFAULT '0' COMMENT '项目id',
  `application_id` bigint(20) DEFAULT '0' COMMENT '应用id',
  `cluster_name` varchar(256) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT '集群名',
  `cpu` decimal(10,0) NOT NULL DEFAULT '1' COMMENT '节点cpu',
  `memory` decimal(10,0) NOT NULL DEFAULT '1024' COMMENT '节点memory',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='工作流标签计算节点表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `fdp_master_workflow_node_integration`
--

DROP TABLE IF EXISTS `fdp_master_workflow_node_integration`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `fdp_master_workflow_node_integration` (
  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT COMMENT '主键id',
  `workflow_id` bigint(20) NOT NULL COMMENT ' workflow id',
  `name` varchar(128) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '节点名称',
  `mode` varchar(20) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '节点添加模式: ui/sql',
  `description` text COLLATE utf8mb4_unicode_ci COMMENT '节点描述',
  `custom_content` text COLLATE utf8mb4_unicode_ci COMMENT '自定义编辑内容',
  `src_ds_type` varchar(20) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '源数据源类型: DL',
  `src_ds_id` bigint(20) DEFAULT NULL COMMENT '源数据源 id',
  `src_model_id` bigint(20) DEFAULT NULL COMMENT '源数据模型 id',
  `src_column_ids` text COLLATE utf8mb4_unicode_ci COMMENT '源数据列 ids',
  `load_strategy` varchar(20) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '加载策略: increment/total',
  `load_interval` bigint(20) DEFAULT NULL COMMENT '增量加载时间间隔',
  `load_interval_unit` varchar(20) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '增量加载时间间隔单位: MINUTE/HOUR/DAY',
  `filter_cond` text COLLATE utf8mb4_unicode_ci COMMENT '源数据模型过滤条件',
  `target_ds_type` varchar(20) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '目标数据源类型: CDP',
  `target_ds_id` bigint(20) DEFAULT NULL COMMENT '目标数据源 id',
  `target_model_id` bigint(20) DEFAULT NULL COMMENT '目标数据模型 id',
  `target_distinct` tinyint(4) NOT NULL DEFAULT '0' COMMENT '目标表去重',
  `column_default` text COLLATE utf8mb4_unicode_ci COMMENT '模型列默认值，map结构',
  `created_at` datetime NOT NULL COMMENT '表记录创建时间',
  `updated_at` datetime NOT NULL COMMENT '表记录更新时间',
  `org_id` bigint(20) DEFAULT NULL COMMENT '企业id',
  `cluster_id` bigint(20) DEFAULT NULL COMMENT '集群id',
  `project_id` bigint(20) DEFAULT NULL COMMENT '项目id',
  `application_id` bigint(20) DEFAULT NULL COMMENT '应用id',
  `cluster_name` varchar(256) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '集群名',
  `persist_type` varchar(20) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '存储方式',
  `cpu` double DEFAULT '1' COMMENT '节点cpu',
  `memory` double DEFAULT '1024' COMMENT '节点memory',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci BLOCK_FORMAT=ENCRYPTED COMMENT='工作流集成节点表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `fdp_master_workflow_node_integration_rt`
--

DROP TABLE IF EXISTS `fdp_master_workflow_node_integration_rt`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `fdp_master_workflow_node_integration_rt` (
  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT COMMENT '主键id',
  `workflow_id` bigint(20) NOT NULL COMMENT ' workflow id',
  `name` varchar(128) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '节点名称',
  `mode` varchar(20) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '节点添加模式: ui/sql',
  `description` text COLLATE utf8mb4_unicode_ci COMMENT '节点描述',
  `custom_content` text COLLATE utf8mb4_unicode_ci COMMENT '自定义编辑内容',
  `src_ds_type` varchar(20) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '源数据源类型: DL',
  `src_ds_id` bigint(20) DEFAULT NULL COMMENT '源数据源 id',
  `src_model_id` bigint(20) DEFAULT NULL COMMENT '源数据模型 id',
  `src_column_ids` text COLLATE utf8mb4_unicode_ci COMMENT '源列ids',
  `file_path` varchar(128) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '文件路径',
  `file_type` varchar(20) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '文件类型：JSON/XML/CSV/REGULAR',
  `file_splitter` varchar(20) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '文件分隔符',
  `filter_cond` text COLLATE utf8mb4_unicode_ci COMMENT '源数据模型过滤条件',
  `target_ds_type` varchar(20) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '目标数据源类型: CDP',
  `target_ds_id` bigint(20) DEFAULT NULL COMMENT '目标数据源 id',
  `target_model_id` bigint(20) DEFAULT NULL COMMENT '目标数据模型 id',
  `target_distinct` tinyint(4) NOT NULL DEFAULT '0' COMMENT '目标表去重',
  `column_default` text COLLATE utf8mb4_unicode_ci COMMENT '模型列默认值，map结构',
  `created_at` datetime NOT NULL COMMENT '表记录创建时间',
  `updated_at` datetime NOT NULL COMMENT '表记录更新时间',
  `org_id` bigint(20) DEFAULT NULL COMMENT '企业id',
  `cluster_id` bigint(20) DEFAULT NULL COMMENT '集群id',
  `project_id` bigint(20) DEFAULT NULL COMMENT '项目id',
  `application_id` bigint(20) DEFAULT NULL COMMENT '应用id',
  `cluster_name` varchar(256) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '集群名',
  `persist_type` varchar(20) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '存储方式',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='工作流实时集成节点表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `fdp_master_workflow_node_labelcomputing`
--

DROP TABLE IF EXISTS `fdp_master_workflow_node_labelcomputing`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `fdp_master_workflow_node_labelcomputing` (
  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT COMMENT '主键id',
  `workflow_id` bigint(20) NOT NULL COMMENT 'workflow id',
  `created_at` datetime NOT NULL COMMENT '表记录创建时间',
  `updated_at` datetime NOT NULL COMMENT '表记录更新时间',
  `org_id` bigint(20) DEFAULT NULL COMMENT '企业id',
  `cluster_id` bigint(20) DEFAULT NULL COMMENT '集群id',
  `project_id` bigint(20) DEFAULT NULL COMMENT '项目id',
  `application_id` bigint(20) DEFAULT NULL COMMENT '应用id',
  `cluster_name` varchar(256) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '集群名',
  `cpu` double DEFAULT '1' COMMENT '节点cpu',
  `memory` double DEFAULT '1024' COMMENT '节点memory',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci BLOCK_FORMAT=ENCRYPTED COMMENT='工作流标签计算节点表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `fdp_master_workflow_node_labelcomputing_rt`
--

DROP TABLE IF EXISTS `fdp_master_workflow_node_labelcomputing_rt`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `fdp_master_workflow_node_labelcomputing_rt` (
  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT COMMENT '主键id',
  `workflow_id` bigint(20) NOT NULL COMMENT 'workflow id',
  `created_at` datetime NOT NULL COMMENT '表记录创建时间',
  `updated_at` datetime NOT NULL COMMENT '表记录更新时间',
  `org_id` bigint(20) DEFAULT NULL COMMENT '企业id',
  `cluster_id` bigint(20) DEFAULT NULL COMMENT '集群id',
  `project_id` bigint(20) DEFAULT NULL COMMENT '项目id',
  `application_id` bigint(20) DEFAULT NULL COMMENT '应用id',
  `cluster_name` varchar(256) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '集群名',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='工作流实时标签计算节点表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `fdp_master_workflow_node_model_mapping`
--

DROP TABLE IF EXISTS `fdp_master_workflow_node_model_mapping`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `fdp_master_workflow_node_model_mapping` (
  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT COMMENT '主键id',
  `node_id` bigint(20) NOT NULL COMMENT '节点 id',
  `node_type` varchar(20) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '节点类型: cleaning/extraction',
  `mode` varchar(20) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '配置方式: ui/sql',
  `model_id` bigint(20) NOT NULL COMMENT '目标模型 id',
  `model_field` varchar(40) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '目标模型字段名称',
  `rel_model_id` bigint(20) DEFAULT NULL COMMENT '关联模型 id',
  `rel_model_field` varchar(40) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '关联模型字段名称',
  `custom_content` text COLLATE utf8mb4_unicode_ci COMMENT '自定义编辑内容',
  `created_at` datetime NOT NULL COMMENT '表记录创建时间',
  `updated_at` datetime NOT NULL COMMENT '表记录更新时间',
  `org_id` bigint(20) DEFAULT NULL COMMENT '企业id',
  `cluster_id` bigint(20) DEFAULT NULL COMMENT '集群id',
  `project_id` bigint(20) DEFAULT NULL COMMENT '项目id',
  `application_id` bigint(20) DEFAULT NULL COMMENT '应用id',
  `cluster_name` varchar(256) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '集群名',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci BLOCK_FORMAT=ENCRYPTED COMMENT='工作流模型节点映射表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `fdp_master_workflow_node_one_id`
--

DROP TABLE IF EXISTS `fdp_master_workflow_node_one_id`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `fdp_master_workflow_node_one_id` (
  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT COMMENT '主键id',
  `workflow_id` bigint(20) NOT NULL COMMENT 'workflow id',
  `name` varchar(128) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT 'oneID 节点名称',
  `created_at` datetime NOT NULL COMMENT '表记录创建时间',
  `updated_at` datetime NOT NULL COMMENT '表记录更新时间',
  `org_id` bigint(20) DEFAULT NULL COMMENT '企业id',
  `cluster_id` bigint(20) DEFAULT NULL COMMENT '集群id',
  `project_id` bigint(20) DEFAULT NULL COMMENT '项目id',
  `application_id` bigint(20) DEFAULT NULL COMMENT '应用id',
  `cluster_name` varchar(256) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '集群名',
  `persist_type` varchar(20) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '存储方式',
  `cpu` double DEFAULT '1' COMMENT '节点cpu',
  `memory` double DEFAULT '1024' COMMENT '节点memory',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci BLOCK_FORMAT=ENCRYPTED COMMENT='工作流oneid节点表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `fdp_master_workflow_node_one_id_rel_model_field`
--

DROP TABLE IF EXISTS `fdp_master_workflow_node_one_id_rel_model_field`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `fdp_master_workflow_node_one_id_rel_model_field` (
  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT COMMENT '主键id',
  `node_id` bigint(20) NOT NULL COMMENT '节点 id',
  `model_id` bigint(20) NOT NULL COMMENT '数据模型 id',
  `model_field` varchar(40) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '数据模型字段名称',
  `note` varchar(40) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '业务说明, 可选值: phone/appID/openID/memberID',
  `created_at` datetime NOT NULL COMMENT '表记录创建时间',
  `updated_at` datetime NOT NULL COMMENT '表记录更新时间',
  `org_id` bigint(20) DEFAULT NULL COMMENT '企业id',
  `cluster_id` bigint(20) DEFAULT NULL COMMENT '集群id',
  `project_id` bigint(20) DEFAULT NULL COMMENT '项目id',
  `application_id` bigint(20) DEFAULT NULL COMMENT '应用id',
  `cluster_name` varchar(256) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '集群名',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci BLOCK_FORMAT=ENCRYPTED COMMENT='工作流oneid节点模型关联表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `fdp_master_workflow_node_python`
--

DROP TABLE IF EXISTS `fdp_master_workflow_node_python`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `fdp_master_workflow_node_python` (
  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT COMMENT '主键id',
  `workflow_id` bigint(20) NOT NULL COMMENT 'workflow id',
  `name` varchar(128) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT 'python任务名',
  `description` text COLLATE utf8mb4_unicode_ci COMMENT 'python描述',
  `version` varchar(128) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT 'python版本',
  `content_git_path` varchar(128) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT 'content存在路径',
  `package_git_path` varchar(128) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '安装包存放路径',
  `created_at` datetime NOT NULL COMMENT '表记录创建时间',
  `updated_at` datetime NOT NULL COMMENT '表记录更新时间',
  `cluster_name` varchar(256) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '集群名',
  `cluster_id` bigint(20) DEFAULT NULL COMMENT '集群id',
  `git_app_id` bigint(20) DEFAULT NULL COMMENT '存放的git应用id',
  `relation_git_node_id` bigint(20) DEFAULT NULL COMMENT '关联的gitnodeid',
  `project_id` bigint(20) DEFAULT NULL COMMENT '项目id',
  `application_id` bigint(20) DEFAULT NULL COMMENT '应用id',
  `org_id` bigint(20) DEFAULT NULL COMMENT '应用id',
  `persist_type` varchar(20) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '存储方式',
  `cpu` double DEFAULT '1' COMMENT '节点cpu',
  `memory` double DEFAULT '1024' COMMENT '节点memory',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='工作流python节点表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `fdp_master_workflow_node_rel_model`
--

DROP TABLE IF EXISTS `fdp_master_workflow_node_rel_model`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `fdp_master_workflow_node_rel_model` (
  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT COMMENT '主键id',
  `workflow_id` bigint(20) NOT NULL COMMENT '工作流ID',
  `node_id` bigint(20) NOT NULL COMMENT '节点 id',
  `node_type` varchar(20) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '节点类型: cleaning/extraction',
  `model_id` bigint(20) NOT NULL COMMENT '关联模型 id',
  `model_alias` varchar(100) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '关联模型别名',
  `join_type` varchar(20) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT 'join类型, LEFT/RIGHT/INNER/FULL',
  `filter_cond` text COLLATE utf8mb4_unicode_ci COMMENT '关联模型过滤条件',
  `created_at` datetime NOT NULL COMMENT '表记录创建时间',
  `updated_at` datetime NOT NULL COMMENT '表记录更新时间',
  `org_id` bigint(20) DEFAULT NULL COMMENT '企业id',
  `cluster_id` bigint(20) DEFAULT NULL COMMENT '集群id',
  `project_id` bigint(20) DEFAULT NULL COMMENT '项目id',
  `application_id` bigint(20) DEFAULT NULL COMMENT '应用id',
  `cluster_name` varchar(256) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '集群名',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci BLOCK_FORMAT=ENCRYPTED COMMENT='工作流节点模型关联表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `fdp_master_workflow_node_rel_model_field`
--

DROP TABLE IF EXISTS `fdp_master_workflow_node_rel_model_field`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `fdp_master_workflow_node_rel_model_field` (
  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT COMMENT '主键id',
  `node_id` bigint(20) NOT NULL COMMENT '节点 id',
  `node_type` varchar(20) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '节点类型: cleaning/extraction',
  `mode` varchar(20) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '节点添加模式: ui/sql',
  `model_id` bigint(20) NOT NULL COMMENT '关联模型 id',
  `model_field` varchar(40) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '关联模型字段名称',
  `rel_model_id` bigint(20) NOT NULL COMMENT '被关联模型 id',
  `rel_model_field` varchar(40) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '被关联模型字段名称',
  `custom_content` text COLLATE utf8mb4_unicode_ci COMMENT '自定义编辑内容',
  `created_at` datetime NOT NULL COMMENT '表记录创建时间',
  `updated_at` datetime NOT NULL COMMENT '表记录更新时间',
  `org_id` bigint(20) DEFAULT NULL COMMENT '企业id',
  `cluster_id` bigint(20) DEFAULT NULL COMMENT '集群id',
  `project_id` bigint(20) DEFAULT NULL COMMENT '项目id',
  `application_id` bigint(20) DEFAULT NULL COMMENT '应用id',
  `cluster_name` varchar(256) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '集群名',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci BLOCK_FORMAT=ENCRYPTED COMMENT='工作流节点模型字段关联表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `fdp_master_workflows`
--

DROP TABLE IF EXISTS `fdp_master_workflows`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `fdp_master_workflows` (
  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT COMMENT '主键id',
  `org_id` bigint(20) DEFAULT NULL COMMENT '企业 id',
  `source` varchar(64) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT '平台来源：DL/CDP',
  `name` varchar(128) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT '工作流名',
  `name_pinyin` varchar(128) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT '工作流拼音',
  `description` varchar(1024) COLLATE utf8mb4_unicode_ci DEFAULT '' COMMENT '工作流描述',
  `run_type` varchar(64) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT '运行类型: ATONCE、SCHEDULE',
  `cron` varchar(128) COLLATE utf8mb4_unicode_ci DEFAULT '' COMMENT '任务周期',
  `cron_start_from` datetime DEFAULT NULL COMMENT '延时执行时间',
  `category_id` bigint(20) NOT NULL COMMENT '工作流目录 ID',
  `pipeline_name` varchar(128) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `pipeline` mediumtext COLLATE utf8mb4_unicode_ci COMMENT 'pipeline 内容, 最大16MB',
  `locations` text COLLATE utf8mb4_unicode_ci,
  `node_params` text COLLATE utf8mb4_unicode_ci COMMENT '工作流节点参数',
  `creator_id` varchar(128) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT '创建者 ID',
  `updater_id` varchar(128) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '创建者 ID',
  `extra` mediumtext COLLATE utf8mb4_unicode_ci COMMENT '扩展信息',
  `delete_yn` tinyint(1) NOT NULL DEFAULT '0' COMMENT '逻辑删除标记',
  `created_at` datetime NOT NULL COMMENT '创建时间',
  `updated_at` datetime NOT NULL COMMENT '更新时间',
  `cluster_id` bigint(20) DEFAULT NULL COMMENT '集群id',
  `project_id` bigint(20) DEFAULT NULL COMMENT '项目id',
  `application_id` bigint(20) DEFAULT NULL COMMENT '应用id',
  `cluster_name` varchar(256) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '集群名',
  `pipeline_status` varchar(128) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT 'pipeline最近一次执行状态',
  `pipeline_begin_time` datetime DEFAULT NULL COMMENT 'pipeline最近一次执行时间',
  `task_queue_id` bigint(20) DEFAULT '1' COMMENT '队列id',
  `queue_level` varchar(128) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '队列优先级',
  PRIMARY KEY (`id`),
  KEY `idx_category_id_name` (`category_id`,`name`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci BLOCK_FORMAT=ENCRYPTED COMMENT='工作流表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `migration_records`
--

DROP TABLE IF EXISTS `migration_records`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `migration_records` (
  `version` varchar(10) NOT NULL COMMENT '服务版本号',
  `module` varchar(50) NOT NULL COMMENT '服务名',
  `version_b` varchar(10) NOT NULL COMMENT '服务 B 版本号',
  `done` varchar(1) NOT NULL DEFAULT '0',
  `create_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  PRIMARY KEY (`version`,`module`,`version_b`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 BLOCK_FORMAT=ENCRYPTED COMMENT='Dice各模块 db migration 执行记录表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `openapi_oauth2_token_clients`
--

DROP TABLE IF EXISTS `openapi_oauth2_token_clients`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `openapi_oauth2_token_clients` (
  `id` varchar(191) NOT NULL,
  `secret` varchar(191) NOT NULL,
  `domain` varchar(4096) NOT NULL DEFAULT '',
  `created_at` datetime NOT NULL,
  `updated_at` datetime NOT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='openapi oauth2 客户端表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `openapi_oauth2_tokens`
--

DROP TABLE IF EXISTS `openapi_oauth2_tokens`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `openapi_oauth2_tokens` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT,
  `code` varchar(191) DEFAULT NULL,
  `access` varchar(4096) NOT NULL DEFAULT '',
  `refresh` varchar(4096) NOT NULL DEFAULT '',
  `data` text NOT NULL,
  `created_at` datetime NOT NULL,
  `expired_at` datetime DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_expired_at` (`expired_at`),
  KEY `idx_created_at` (`created_at`),
  KEY `idx_code` (`code`),
  FULLTEXT KEY `idx_access` (`access`),
  FULLTEXT KEY `idx_refresh` (`refresh`)
) ENGINE=InnoDB AUTO_INCREMENT=5270009 DEFAULT CHARSET=utf8mb4 COMMENT='openapi oauth2 token 表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `ops_orgak`
--

DROP TABLE IF EXISTS `ops_orgak`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `ops_orgak` (
  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT COMMENT 'primary key',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  `org_id` varchar(64) DEFAULT NULL COMMENT '企业ID',
  `vendor` varchar(64) DEFAULT NULL COMMENT '云供应商',
  `access_key` mediumtext COMMENT '云供应商access_key',
  `secret_key` mediumtext COMMENT '云供应商secret_key',
  `description` mediumtext COMMENT '云供应商ak,sk描述',
  PRIMARY KEY (`id`),
  KEY `idx_ops_orgak_org_id` (`org_id`),
  KEY `idx_ops_orgak_vendor` (`vendor`)
) ENGINE=InnoDB AUTO_INCREMENT=5 DEFAULT CHARSET=utf8mb4 COMMENT='云账号记录表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `ops_record`
--

DROP TABLE IF EXISTS `ops_record`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `ops_record` (
  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT COMMENT 'primary key',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  `record_type` varchar(64) DEFAULT NULL COMMENT '操作记录类型',
  `user_id` varchar(64) DEFAULT NULL COMMENT '操作用户ID',
  `org_id` varchar(64) DEFAULT NULL COMMENT '操作用户所属企业ID',
  `cluster_name` varchar(64) DEFAULT NULL COMMENT '操作相关集群名',
  `status` varchar(64) DEFAULT NULL COMMENT '操作结果状态，是否成功',
  `detail` text COMMENT '操作详情',
  `pipeline_id` bigint(20) unsigned DEFAULT NULL COMMENT '操作相关流水线ID',
  PRIMARY KEY (`id`),
  KEY `idx_ops_record_org_id` (`org_id`),
  KEY `idx_ops_record_cluster_name` (`cluster_name`)
) ENGINE=InnoDB AUTO_INCREMENT=192 DEFAULT CHARSET=utf8mb4 COMMENT='云管操作日志记录表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `pipeline_archives`
--

DROP TABLE IF EXISTS `pipeline_archives`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `pipeline_archives` (
  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT,
  `time_created` datetime NOT NULL,
  `time_updated` datetime NOT NULL,
  `pipeline_id` bigint(20) NOT NULL,
  `pipeline_source` varchar(32) NOT NULL DEFAULT '',
  `pipeline_yml_name` varchar(191) NOT NULL DEFAULT '',
  `status` varchar(191) NOT NULL DEFAULT '',
  `dice_version` varchar(32) NOT NULL DEFAULT '',
  `content` longtext NOT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_pipelineID` (`pipeline_id`),
  KEY `idx_source_ymlName` (`pipeline_source`,`pipeline_yml_name`),
  KEY `idx_source_ymlName_status` (`pipeline_source`,`pipeline_yml_name`,`status`),
  KEY `idx_source_status` (`pipeline_source`,`status`),
  KEY `idx_source_pipelineID` (`pipeline_source`,`pipeline_id`),
  KEY `idx_source_ymlName_pipelineID` (`pipeline_source`,`pipeline_yml_name`,`pipeline_id`),
  KEY `idx_source_status_pipelineID` (`pipeline_source`,`status`,`pipeline_id`)
) ENGINE=InnoDB AUTO_INCREMENT=893930 DEFAULT CHARSET=utf8mb4 BLOCK_FORMAT=ENCRYPTED COMMENT='流水线归档表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `pipeline_bases`
--

DROP TABLE IF EXISTS `pipeline_bases`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `pipeline_bases` (
  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT,
  `pipeline_source` varchar(191) NOT NULL DEFAULT '',
  `pipeline_yml_name` varchar(191) NOT NULL DEFAULT '',
  `cluster_name` varchar(191) NOT NULL DEFAULT '',
  `status` varchar(32) NOT NULL DEFAULT '',
  `type` varchar(32) NOT NULL DEFAULT '',
  `trigger_mode` varchar(32) NOT NULL DEFAULT '',
  `cron_id` bigint(20) DEFAULT NULL,
  `is_snippet` tinyint(1) NOT NULL DEFAULT '0' COMMENT '是否是嵌套流水线',
  `parent_pipeline_id` bigint(20) DEFAULT NULL COMMENT '当前嵌套流水线对应的父流水线 ID',
  `parent_task_id` bigint(20) DEFAULT NULL COMMENT '当前嵌套流水线对应的父流水线任务 ID',
  `cost_time_sec` bigint(20) NOT NULL,
  `time_begin` datetime DEFAULT NULL,
  `time_end` datetime DEFAULT NULL,
  `time_created` datetime NOT NULL,
  `time_updated` datetime NOT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_source_ymlName` (`pipeline_source`,`pipeline_yml_name`),
  KEY `idx_status` (`status`),
  KEY `idx_source_status` (`pipeline_source`,`status`),
  KEY `idx_source_ymlName_status` (`pipeline_source`,`pipeline_yml_name`,`status`),
  KEY `idx_id_source_cluster_status` (`id`,`pipeline_source`,`cluster_name`,`status`),
  KEY `idx_source_status_cluster_timebegin_timeend_id` (`pipeline_source`,`status`,`cluster_name`,`time_begin`,`time_end`,`id`)
) ENGINE=InnoDB AUTO_INCREMENT=11657177 DEFAULT CHARSET=utf8mb4 BLOCK_FORMAT=ENCRYPTED COMMENT='流水线基础信息表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `pipeline_configs`
--

DROP TABLE IF EXISTS `pipeline_configs`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `pipeline_configs` (
  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT,
  `type` varchar(255) NOT NULL DEFAULT '',
  `value` text NOT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=2 DEFAULT CHARSET=utf8mb4 BLOCK_FORMAT=ENCRYPTED COMMENT='(内部)流水线内部配置表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `pipeline_crons`
--

DROP TABLE IF EXISTS `pipeline_crons`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `pipeline_crons` (
  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT,
  `pipeline_source` varchar(32) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '',
  `application_id` bigint(20) NOT NULL,
  `branch` varchar(191) COLLATE utf8mb4_unicode_ci NOT NULL,
  `cron_expr` varchar(255) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '',
  `enable` tinyint(1) NOT NULL,
  `pipeline_yml_name` varchar(191) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '',
  `base_pipeline_id` bigint(20) NOT NULL,
  `time_created` datetime NOT NULL,
  `time_updated` datetime NOT NULL,
  `extra` mediumtext COLLATE utf8mb4_unicode_ci,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=2684 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci BLOCK_FORMAT=ENCRYPTED COMMENT='定时流水线';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `pipeline_extras`
--

DROP TABLE IF EXISTS `pipeline_extras`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `pipeline_extras` (
  `pipeline_id` bigint(20) unsigned NOT NULL,
  `pipeline_yml` mediumtext NOT NULL,
  `extra` mediumtext NOT NULL,
  `normal_labels` mediumtext NOT NULL COMMENT '这里存储的 label 仅用于展示，不做筛选。用于筛选的 label 存储在 pipeline_labels 表中',
  `snapshot` mediumtext NOT NULL,
  `commit_detail` text NOT NULL,
  `progress` int(3) NOT NULL DEFAULT '-1' COMMENT '0-100，-1 表示未设置',
  `time_created` datetime NOT NULL,
  `time_updated` datetime NOT NULL,
  `commit` varchar(64) DEFAULT NULL,
  `org_name` varchar(191) DEFAULT NULL,
  `snippets` text COMMENT 'snippet 历史',
  PRIMARY KEY (`pipeline_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 BLOCK_FORMAT=ENCRYPTED COMMENT='流水线额外信息表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `pipeline_labels`
--

DROP TABLE IF EXISTS `pipeline_labels`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `pipeline_labels` (
  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT,
  `type` varchar(16) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `pipeline_source` varchar(32) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '',
  `pipeline_yml_name` varchar(191) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '',
  `target_id` bigint(20) DEFAULT NULL,
  `key` varchar(191) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '',
  `value` varchar(191) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT '标签值',
  `time_created` datetime NOT NULL,
  `time_updated` datetime NOT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_source` (`pipeline_source`),
  KEY `idx_pipeline_yml_name` (`pipeline_yml_name`),
  KEY `idx_namespace` (`pipeline_source`,`pipeline_yml_name`),
  KEY `idx_key` (`key`),
  KEY `idx_target_id` (`target_id`),
  KEY `idx_type_source_key_value_targetid` (`type`,`pipeline_source`,`key`,`value`,`target_id`),
  KEY `idx_type_source_ymlname_key_value_targetid` (`type`,`pipeline_source`,`pipeline_yml_name`,`key`,`value`,`target_id`)
) ENGINE=InnoDB AUTO_INCREMENT=18416494 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci BLOCK_FORMAT=ENCRYPTED COMMENT='流水线标签表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `pipeline_stages`
--

DROP TABLE IF EXISTS `pipeline_stages`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `pipeline_stages` (
  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT,
  `pipeline_id` bigint(20) NOT NULL,
  `name` varchar(255) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '',
  `extra` text COLLATE utf8mb4_unicode_ci NOT NULL,
  `status` varchar(128) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '',
  `cost_time_sec` bigint(20) NOT NULL,
  `time_begin` datetime DEFAULT NULL,
  `time_end` datetime DEFAULT NULL,
  `time_created` datetime NOT NULL,
  `time_updated` datetime NOT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_pipeline_id` (`pipeline_id`)
) ENGINE=InnoDB AUTO_INCREMENT=4706134 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci BLOCK_FORMAT=ENCRYPTED COMMENT='流水线阶段(stage)表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `pipeline_tasks`
--

DROP TABLE IF EXISTS `pipeline_tasks`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `pipeline_tasks` (
  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT,
  `pipeline_id` bigint(20) NOT NULL,
  `stage_id` bigint(20) NOT NULL,
  `name` varchar(255) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '',
  `op_type` varchar(255) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '',
  `type` varchar(255) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '',
  `executor_kind` varchar(255) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '',
  `status` varchar(128) COLLATE utf8mb4_unicode_ci NOT NULL,
  `extra` mediumtext COLLATE utf8mb4_unicode_ci NOT NULL,
  `context` text COLLATE utf8mb4_unicode_ci NOT NULL,
  `result` text COLLATE utf8mb4_unicode_ci NOT NULL,
  `is_snippet` tinyint(1) NOT NULL DEFAULT '0' COMMENT '是否是嵌套流水线任务',
  `snippet_pipeline_id` bigint(20) DEFAULT NULL COMMENT '当前任务对应的嵌套流水线 ID',
  `snippet_pipeline_detail` mediumtext COLLATE utf8mb4_unicode_ci COMMENT '当前任务对应的嵌套流水线汇总后的详情',
  `cost_time_sec` bigint(20) NOT NULL,
  `queue_time_sec` bigint(20) NOT NULL,
  `time_begin` datetime DEFAULT NULL,
  `time_end` datetime DEFAULT NULL,
  `time_created` datetime NOT NULL,
  `time_updated` datetime NOT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_pipeline_id` (`pipeline_id`),
  KEY `idx_stage_id` (`stage_id`),
  KEY `idx_status` (`status`)
) ENGINE=InnoDB AUTO_INCREMENT=11983604 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci BLOCK_FORMAT=ENCRYPTED COMMENT='流水线任务(task)表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `ps_activities`
--

DROP TABLE IF EXISTS `ps_activities`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `ps_activities` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT,
  `created_at` timestamp NULL DEFAULT NULL,
  `updated_at` timestamp NULL DEFAULT NULL,
  `org_id` bigint(20) DEFAULT NULL,
  `project_id` bigint(20) DEFAULT NULL,
  `application_id` bigint(20) DEFAULT NULL,
  `build_id` bigint(20) DEFAULT NULL,
  `runtime_id` bigint(20) DEFAULT NULL,
  `operator` varchar(255) DEFAULT NULL,
  `type` varchar(255) DEFAULT NULL,
  `action` varchar(255) DEFAULT NULL,
  `desc` varchar(255) DEFAULT NULL,
  `context` text,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=48869 DEFAULT CHARSET=utf8mb4 BLOCK_FORMAT=ENCRYPTED COMMENT='Dice 活动表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `ps_comments`
--

DROP TABLE IF EXISTS `ps_comments`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `ps_comments` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT,
  `created_at` timestamp NULL DEFAULT NULL,
  `updated_at` timestamp NULL DEFAULT NULL,
  `ticket_id` bigint(20) DEFAULT NULL,
  `comment_type` varchar(20) DEFAULT 'normal' COMMENT '评论类型: normal/issueRelation',
  `content` text,
  `ir_comment` text COMMENT '关联事件评论内容',
  `user_id` varchar(255) DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 BLOCK_FORMAT=ENCRYPTED;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `ps_group_projects`
--

DROP TABLE IF EXISTS `ps_group_projects`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `ps_group_projects` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT,
  `created_at` timestamp NULL DEFAULT NULL,
  `updated_at` timestamp NULL DEFAULT NULL,
  `name` varchar(255) DEFAULT NULL,
  `desc` varchar(255) DEFAULT NULL,
  `logo` varchar(255) DEFAULT NULL,
  `org_id` bigint(20) DEFAULT NULL,
  `creator` varchar(255) DEFAULT NULL,
  `dd_hook` varchar(255) DEFAULT NULL,
  `cluster_config` varchar(255) DEFAULT NULL,
  `cpu_quota` decimal(65,2) DEFAULT NULL COMMENT '项目 cpu 配额',
  `mem_quota` decimal(65,2) DEFAULT NULL COMMENT '项目 mem 配额',
  `functions` varchar(255) DEFAULT '{"projectCooperative": true, "testManagement": true, "codeQuality": true, "codeBase": true,"branchRule": true, "cicd": true, "productLibManagement": true, "notify": true}',
  `active_time` datetime DEFAULT NULL,
  `rollback_config` varchar(1000) DEFAULT NULL COMMENT '回滚点配置',
  `display_name` varchar(50) DEFAULT NULL COMMENT '项目显示名称',
  `enable_ns` tinyint(1) DEFAULT '0' COMMENT '是否开启项目级命名空间',
  `is_public` tinyint(1) NOT NULL DEFAULT '0' COMMENT '是否公开',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=12 DEFAULT CHARSET=utf8mb4 BLOCK_FORMAT=ENCRYPTED;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `ps_images`
--

DROP TABLE IF EXISTS `ps_images`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `ps_images` (
  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT,
  `created_at` timestamp NULL DEFAULT NULL,
  `updated_at` timestamp NULL DEFAULT NULL,
  `release_id` varchar(255) DEFAULT NULL,
  `image_name` varchar(128) NOT NULL,
  `image_tag` varchar(64) DEFAULT NULL,
  `image` varchar(255) NOT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_release_id` (`release_id`)
) ENGINE=InnoDB AUTO_INCREMENT=73067 DEFAULT CHARSET=utf8 BLOCK_FORMAT=ENCRYPTED COMMENT='Dice 镜像表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `ps_runtime_instances`
--

DROP TABLE IF EXISTS `ps_runtime_instances`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `ps_runtime_instances` (
  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT,
  `created_at` timestamp NULL DEFAULT NULL,
  `updated_at` timestamp NULL DEFAULT NULL,
  `instance_id` varchar(255) NOT NULL,
  `runtime_id` bigint(20) unsigned NOT NULL,
  `service_id` bigint(20) unsigned NOT NULL,
  `ip` varchar(255) DEFAULT NULL,
  `status` varchar(255) DEFAULT NULL,
  `stage` varchar(255) DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_instance_id` (`instance_id`),
  KEY `idx_service_id` (`service_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8 BLOCK_FORMAT=ENCRYPTED COMMENT='runtime对应instance信息';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `ps_runtime_services`
--

DROP TABLE IF EXISTS `ps_runtime_services`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `ps_runtime_services` (
  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT,
  `created_at` timestamp NULL DEFAULT NULL,
  `updated_at` timestamp NULL DEFAULT NULL,
  `runtime_id` bigint(20) unsigned NOT NULL,
  `service_name` varchar(255) NOT NULL,
  `cpu` varchar(255) DEFAULT NULL,
  `mem` int(11) DEFAULT NULL,
  `environment` text,
  `ports` text COMMENT '端口',
  `replica` int(11) DEFAULT NULL,
  `status` varchar(255) DEFAULT NULL,
  `errors` text,
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_runtime_id_service_name` (`runtime_id`,`service_name`)
) ENGINE=InnoDB AUTO_INCREMENT=1383 DEFAULT CHARSET=utf8 BLOCK_FORMAT=ENCRYPTED COMMENT='runtime对应service信息';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `ps_tickets`
--

DROP TABLE IF EXISTS `ps_tickets`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `ps_tickets` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT,
  `created_at` timestamp NULL DEFAULT NULL,
  `updated_at` timestamp NULL DEFAULT NULL,
  `title` varchar(255) DEFAULT NULL,
  `content` text,
  `type` varchar(20) DEFAULT NULL,
  `priority` varchar(255) DEFAULT NULL,
  `status` varchar(255) DEFAULT NULL,
  `request_id` varchar(60) DEFAULT NULL,
  `key` varchar(64) DEFAULT NULL,
  `org_id` varchar(255) DEFAULT NULL,
  `metric` varchar(255) DEFAULT NULL,
  `metric_id` varchar(255) DEFAULT NULL,
  `count` bigint(20) DEFAULT NULL,
  `creator` varchar(255) DEFAULT NULL,
  `last_operator` varchar(255) DEFAULT NULL,
  `label` text,
  `target_type` varchar(40) DEFAULT NULL,
  `target_id` varchar(255) DEFAULT NULL,
  `triggered_at` timestamp NULL DEFAULT NULL,
  `closed_at` timestamp NULL DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_type` (`type`),
  KEY `idx_request_id` (`request_id`),
  KEY `idx_key` (`key`),
  KEY `idx_target_type` (`target_type`)
) ENGINE=InnoDB AUTO_INCREMENT=14006 DEFAULT CHARSET=utf8mb4 BLOCK_FORMAT=ENCRYPTED;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `ps_user_current_org`
--

DROP TABLE IF EXISTS `ps_user_current_org`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `ps_user_current_org` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT,
  `created_at` timestamp NULL DEFAULT NULL,
  `updated_at` timestamp NULL DEFAULT NULL,
  `user_id` varchar(255) DEFAULT NULL,
  `org_id` bigint(20) DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=19 DEFAULT CHARSET=utf8mb4 BLOCK_FORMAT=ENCRYPTED COMMENT='用户当前所属企业';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `ps_v2_deployments`
--

DROP TABLE IF EXISTS `ps_v2_deployments`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `ps_v2_deployments` (
  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT,
  `created_at` timestamp NULL DEFAULT NULL,
  `updated_at` timestamp NULL DEFAULT NULL,
  `runtime_id` bigint(20) unsigned NOT NULL,
  `release_id` varchar(255) DEFAULT NULL,
  `outdated` tinyint(1) DEFAULT NULL,
  `dice` text,
  `built_docker_images` text,
  `operator` varchar(255) NOT NULL,
  `status` varchar(255) NOT NULL,
  `step` varchar(255) DEFAULT NULL,
  `fail_cause` text,
  `extra` text,
  `finished_at` timestamp NULL DEFAULT NULL,
  `build_id` bigint(20) unsigned DEFAULT NULL,
  `dice_type` int(1) DEFAULT '0' COMMENT 'dice字段类型，0: legace, 1: diceyml',
  `type` varchar(32) DEFAULT '' COMMENT 'build类型，REDEPLOY、RELEASE、BUILD',
  `need_approval` tinyint(1) DEFAULT NULL,
  `approved_by_user` varchar(255) DEFAULT NULL,
  `approved_at` timestamp NULL DEFAULT NULL,
  `approval_status` varchar(255) DEFAULT NULL,
  `approval_reason` varchar(255) DEFAULT NULL,
  `skip_push_by_orch` tinyint(1) DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_status` (`status`),
  KEY `idx_runtime_id` (`runtime_id`),
  KEY `idx_operator` (`operator`)
) ENGINE=InnoDB AUTO_INCREMENT=32293 DEFAULT CHARSET=utf8 BLOCK_FORMAT=ENCRYPTED COMMENT='部署单';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `ps_v2_domains`
--

DROP TABLE IF EXISTS `ps_v2_domains`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `ps_v2_domains` (
  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT,
  `created_at` timestamp NULL DEFAULT NULL,
  `updated_at` timestamp NULL DEFAULT NULL,
  `runtime_id` bigint(20) unsigned NOT NULL,
  `domain` varchar(255) DEFAULT NULL,
  `domain_type` varchar(255) DEFAULT NULL,
  `endpoint_name` varchar(255) DEFAULT NULL,
  `use_https` tinyint(1) DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `unique_domain_key` (`domain`)
) ENGINE=InnoDB AUTO_INCREMENT=934 DEFAULT CHARSET=utf8 BLOCK_FORMAT=ENCRYPTED COMMENT='Dice 域名表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `ps_v2_pre_builds`
--

DROP TABLE IF EXISTS `ps_v2_pre_builds`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `ps_v2_pre_builds` (
  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT,
  `created_at` timestamp NULL DEFAULT NULL,
  `updated_at` timestamp NULL DEFAULT NULL,
  `project_id` bigint(20) unsigned DEFAULT NULL,
  `env` varchar(255) DEFAULT NULL,
  `git_branch` varchar(255) DEFAULT NULL,
  `dice` text,
  `dice_overlay` text,
  `dice_type` int(1) DEFAULT '0' COMMENT 'dice字段类型，0: legace, 1: diceyml',
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_unique_project_env_branch` (`project_id`,`env`,`git_branch`)
) ENGINE=InnoDB AUTO_INCREMENT=372 DEFAULT CHARSET=utf8 BLOCK_FORMAT=ENCRYPTED COMMENT='ps_v2_pre_builds(废弃)';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `ps_v2_project_runtimes`
--

DROP TABLE IF EXISTS `ps_v2_project_runtimes`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `ps_v2_project_runtimes` (
  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT,
  `created_at` timestamp NULL DEFAULT NULL,
  `updated_at` timestamp NULL DEFAULT NULL,
  `name` varchar(255) NOT NULL,
  `application_id` bigint(20) unsigned NOT NULL,
  `workspace` varchar(255) NOT NULL,
  `git_branch` varchar(255) DEFAULT NULL,
  `project_id` bigint(20) unsigned NOT NULL,
  `env` varchar(255) DEFAULT NULL,
  `cluster_name` varchar(255) DEFAULT NULL,
  `cluster_id` bigint(20) unsigned DEFAULT NULL,
  `creator` varchar(255) NOT NULL,
  `schedule_name` varchar(255) DEFAULT NULL,
  `runtime_status` varchar(255) DEFAULT NULL,
  `status` varchar(255) DEFAULT NULL,
  `deployed` tinyint(1) DEFAULT NULL,
  `version` varchar(255) DEFAULT NULL,
  `source` varchar(255) DEFAULT NULL,
  `dice_version` varchar(255) DEFAULT NULL,
  `config_updated_date` timestamp NULL DEFAULT NULL,
  `readable_unique_id` varchar(255) DEFAULT NULL,
  `git_repo_abbrev` varchar(255) DEFAULT NULL,
  `cpu` double(8,2) NOT NULL COMMENT 'cpu核数',
  `mem` double(8,2) NOT NULL COMMENT '内存大小（M）',
  `org_id` bigint(20) NOT NULL COMMENT '企业ID',
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_unique_app_id_name` (`name`,`application_id`,`workspace`)
) ENGINE=InnoDB AUTO_INCREMENT=686 DEFAULT CHARSET=utf8 BLOCK_FORMAT=ENCRYPTED COMMENT='runtime信息';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `qa_sonar`
--

DROP TABLE IF EXISTS `qa_sonar`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `qa_sonar` (
  `id` int(11) unsigned NOT NULL AUTO_INCREMENT COMMENT '自增主键',
  `key` varchar(191) NOT NULL DEFAULT '' COMMENT '分析代码的key',
  `bugs` longtext COMMENT '代码bug数量',
  `code_smells` longtext COMMENT '代码异味数量',
  `vulnerabilities` longtext COMMENT '代码漏洞数量',
  `coverage` longtext COMMENT '代码覆盖率',
  `duplications` longtext COMMENT '代码重复率',
  `issues_statistics` text COMMENT '代码质量统计',
  `created_at` datetime DEFAULT NULL COMMENT '创建时间',
  `updated_at` datetime DEFAULT NULL COMMENT '更新时间',
  `app_id` bigint(20) DEFAULT NULL COMMENT '应用id',
  `operator_id` varchar(255) DEFAULT NULL COMMENT '用户id',
  `project_id` bigint(20) DEFAULT NULL COMMENT '项目id',
  `commit_id` varchar(50) DEFAULT NULL COMMENT '提交id',
  `branch` varchar(255) DEFAULT NULL COMMENT '代码分支',
  `git_repo` varchar(255) DEFAULT NULL COMMENT 'git仓库地址',
  `build_id` bigint(20) DEFAULT NULL COMMENT '创建id',
  `log_id` varchar(40) DEFAULT NULL COMMENT '日志id',
  `app_name` varchar(255) DEFAULT NULL COMMENT '应用名称',
  PRIMARY KEY (`id`),
  KEY `index_name` (`commit_id`),
  KEY `app_id` (`app_id`),
  KEY `idx_key` (`key`)
) ENGINE=InnoDB AUTO_INCREMENT=8139 DEFAULT CHARSET=utf8mb4 BLOCK_FORMAT=ENCRYPTED COMMENT='sonar 代码质量扫描结果表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `qa_sonar_metric_keys`
--

DROP TABLE IF EXISTS `qa_sonar_metric_keys`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `qa_sonar_metric_keys` (
  `id` int(20) NOT NULL AUTO_INCREMENT,
  `metric_key` varchar(50) NOT NULL COMMENT 'key',
  `value_type` varchar(50) NOT NULL COMMENT '值的类型',
  `name` varchar(50) NOT NULL COMMENT '名称',
  `metric_key_desc` varchar(255) NOT NULL COMMENT 'key的英文描述',
  `domain` varchar(50) NOT NULL COMMENT '所属类型',
  `operational` varchar(20) NOT NULL COMMENT '操作',
  `qualitative` tinyint(1) NOT NULL COMMENT '是否是增量',
  `hidden` tinyint(1) NOT NULL COMMENT '是否隐藏',
  `custom` tinyint(1) NOT NULL COMMENT '是否是本地',
  `decimal_scale` tinyint(5) NOT NULL COMMENT '小数点后几位',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=61 DEFAULT CHARSET=utf8mb4 COMMENT='代码质量扫描规则项';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `qa_sonar_metric_rules`
--

DROP TABLE IF EXISTS `qa_sonar_metric_rules`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `qa_sonar_metric_rules` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT COMMENT '主键id',
  `description` varchar(150) NOT NULL COMMENT '描述',
  `created_at` datetime NOT NULL COMMENT '创建时间',
  `updated_at` datetime NOT NULL COMMENT '更新时间',
  `scope_type` varchar(50) NOT NULL COMMENT '所属类型',
  `scope_id` varchar(50) NOT NULL COMMENT '所属类型的id',
  `metric_key_id` varchar(255) NOT NULL COMMENT '指标的id',
  `metric_value` varchar(255) NOT NULL COMMENT '指标的值',
  PRIMARY KEY (`id`),
  KEY `scope` (`scope_type`,`scope_id`) USING BTREE COMMENT '指标规则的所属类型和id'
) ENGINE=InnoDB AUTO_INCREMENT=53 DEFAULT CHARSET=utf8mb4 COMMENT='代码质量扫描规则配置表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `qa_test_records`
--

DROP TABLE IF EXISTS `qa_test_records`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `qa_test_records` (
  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT COMMENT '自增主键',
  `created_at` datetime DEFAULT NULL COMMENT '创建时间',
  `updated_at` datetime DEFAULT NULL COMMENT '更新时间',
  `name` varchar(255) DEFAULT NULL COMMENT '名称',
  `app_id` bigint(20) DEFAULT NULL COMMENT '应用id',
  `operator_id` varchar(255) DEFAULT NULL COMMENT '操作者id',
  `output` varchar(1024) DEFAULT NULL COMMENT '测试结果输出地址',
  `type` varchar(20) DEFAULT NULL COMMENT '测试类型',
  `parser_type` varchar(255) DEFAULT NULL COMMENT '测试 parser 类型， JUNIT/TESTNG',
  `totals` mediumtext COMMENT '测试用例执行结果及耗时分布',
  `desc` varchar(1024) DEFAULT NULL COMMENT '描述',
  `extra` varchar(1024) DEFAULT NULL COMMENT '附加信息',
  `project_id` bigint(20) DEFAULT NULL COMMENT '项目id',
  `commit_id` varchar(191) DEFAULT NULL COMMENT '测试代码 commit id',
  `branch` varchar(255) DEFAULT NULL COMMENT '代码分支',
  `git_repo` varchar(255) DEFAULT NULL COMMENT 'git仓库地址',
  `envs` varchar(1024) DEFAULT NULL COMMENT '环境变量',
  `case_dir` varchar(255) DEFAULT NULL COMMENT '执行目录',
  `workspace` varchar(255) DEFAULT NULL COMMENT '应用对应环境',
  `build_id` bigint(20) DEFAULT NULL COMMENT '构建id',
  `app_name` varchar(255) DEFAULT NULL COMMENT '应用名字',
  `uuid` varchar(40) DEFAULT NULL COMMENT '用户id',
  `suites` longtext COMMENT '测试结果数据',
  `operator_name` varchar(255) DEFAULT NULL COMMENT '操作者名字',
  PRIMARY KEY (`id`),
  KEY `app_id` (`app_id`),
  KEY `test_type` (`type`),
  KEY `commit_id` (`commit_id`)
) ENGINE=InnoDB AUTO_INCREMENT=9157 DEFAULT CHARSET=utf8mb4 BLOCK_FORMAT=ENCRYPTED COMMENT='单元测试执行记录表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `runtime_instances`
--

DROP TABLE IF EXISTS `runtime_instances`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `runtime_instances` (
  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT,
  `created_at` timestamp NULL DEFAULT NULL,
  `updated_at` timestamp NULL DEFAULT NULL,
  `instance_id` varchar(255) NOT NULL,
  `runtime_id` bigint(20) unsigned NOT NULL,
  `service_id` bigint(20) unsigned NOT NULL,
  `ip` varchar(255) DEFAULT NULL,
  `status` varchar(255) DEFAULT NULL,
  `stage` varchar(255) DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_instance_id` (`instance_id`),
  KEY `idx_service_id` (`service_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8 BLOCK_FORMAT=ENCRYPTED;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `runtime_services`
--

DROP TABLE IF EXISTS `runtime_services`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `runtime_services` (
  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT,
  `created_at` timestamp NULL DEFAULT NULL,
  `updated_at` timestamp NULL DEFAULT NULL,
  `runtime_id` bigint(20) unsigned NOT NULL,
  `service_name` varchar(255) NOT NULL,
  `cpu` varchar(255) DEFAULT NULL,
  `mem` int(11) DEFAULT NULL,
  `environment` text,
  `ports` varchar(255) DEFAULT NULL,
  `replica` int(11) DEFAULT NULL,
  `status` varchar(255) DEFAULT NULL,
  `errors` text,
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_runtime_id_service_name` (`runtime_id`,`service_name`)
) ENGINE=InnoDB AUTO_INCREMENT=425 DEFAULT CHARSET=utf8 BLOCK_FORMAT=ENCRYPTED;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `s_instance_info`
--

DROP TABLE IF EXISTS `s_instance_info`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `s_instance_info` (
  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT,
  `created_at` timestamp NULL DEFAULT NULL,
  `updated_at` timestamp NULL DEFAULT NULL,
  `cluster` varchar(64) DEFAULT NULL,
  `namespace` varchar(64) DEFAULT NULL,
  `name` varchar(64) DEFAULT NULL,
  `org_name` varchar(64) DEFAULT NULL,
  `org_id` varchar(64) DEFAULT NULL,
  `project_name` varchar(64) DEFAULT NULL,
  `project_id` varchar(64) DEFAULT NULL,
  `application_name` varchar(255) DEFAULT NULL,
  `application_id` varchar(255) DEFAULT NULL,
  `runtime_name` varchar(255) DEFAULT NULL,
  `runtime_id` varchar(255) DEFAULT NULL,
  `service_name` varchar(255) DEFAULT NULL,
  `workspace` varchar(10) DEFAULT NULL,
  `service_type` varchar(64) DEFAULT NULL,
  `addon_id` varchar(255) DEFAULT NULL,
  `meta` varchar(255) DEFAULT NULL,
  `task_id` varchar(150) DEFAULT NULL,
  `phase` varchar(255) DEFAULT NULL,
  `message` varchar(1024) DEFAULT NULL,
  `container_id` varchar(100) DEFAULT NULL,
  `container_ip` varchar(255) DEFAULT NULL,
  `host_ip` varchar(255) DEFAULT NULL,
  `started_at` timestamp NULL DEFAULT NULL,
  `finished_at` timestamp NULL DEFAULT NULL,
  `exit_code` int(11) DEFAULT NULL,
  `cpu_origin` double DEFAULT NULL,
  `mem_origin` int(11) DEFAULT NULL,
  `cpu_request` double DEFAULT NULL,
  `mem_request` int(11) DEFAULT NULL,
  `cpu_limit` double DEFAULT NULL,
  `mem_limit` int(11) DEFAULT NULL,
  `image` varchar(255) DEFAULT NULL,
  `edge_application_name` varchar(255) DEFAULT NULL COMMENT '边缘应用名',
  `edge_site` varchar(255) DEFAULT NULL COMMENT '边缘站点',
  PRIMARY KEY (`id`),
  KEY `idx_s_instance_info_namespace` (`namespace`),
  KEY `idx_s_instance_info_name` (`name`),
  KEY `idx_s_instance_info_org_name` (`org_name`),
  KEY `idx_s_instance_info_org_id` (`org_id`),
  KEY `idx_s_instance_info_project_name` (`project_name`),
  KEY `idx_s_instance_info_project_id` (`project_id`),
  KEY `index_container_id` (`container_id`),
  KEY `index_task_id` (`task_id`)
) ENGINE=InnoDB AUTO_INCREMENT=5750208 DEFAULT CHARSET=utf8mb4 BLOCK_FORMAT=ENCRYPTED COMMENT='实例信息';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `s_instance_info_bak`
--

DROP TABLE IF EXISTS `s_instance_info_bak`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `s_instance_info_bak` (
  `id` bigint(20) unsigned NOT NULL DEFAULT '0',
  `created_at` timestamp NULL DEFAULT NULL,
  `updated_at` timestamp NULL DEFAULT NULL,
  `cluster` varchar(255) DEFAULT NULL,
  `namespace` varchar(64) DEFAULT NULL,
  `name` varchar(64) DEFAULT NULL,
  `org_name` varchar(64) DEFAULT NULL,
  `org_id` varchar(64) DEFAULT NULL,
  `project_name` varchar(64) DEFAULT NULL,
  `project_id` varchar(64) DEFAULT NULL,
  `application_name` varchar(255) DEFAULT NULL,
  `application_id` varchar(255) DEFAULT NULL,
  `runtime_name` varchar(255) DEFAULT NULL,
  `runtime_id` varchar(255) DEFAULT NULL,
  `service_name` varchar(255) DEFAULT NULL,
  `workspace` varchar(10) DEFAULT NULL,
  `service_type` varchar(64) DEFAULT NULL,
  `addon_id` varchar(255) DEFAULT NULL,
  `meta` varchar(255) DEFAULT NULL,
  `task_id` varchar(255) DEFAULT NULL,
  `phase` varchar(255) DEFAULT NULL,
  `message` varchar(1024) DEFAULT NULL,
  `container_id` varchar(255) DEFAULT NULL,
  `container_ip` varchar(255) DEFAULT NULL,
  `host_ip` varchar(255) DEFAULT NULL,
  `started_at` timestamp NULL DEFAULT NULL,
  `finished_at` timestamp NULL DEFAULT NULL,
  `exit_code` int(11) DEFAULT NULL,
  `cpu_origin` double DEFAULT NULL,
  `mem_origin` int(11) DEFAULT NULL,
  `cpu_request` double DEFAULT NULL,
  `mem_request` int(11) DEFAULT NULL,
  `cpu_limit` double DEFAULT NULL,
  `mem_limit` int(11) DEFAULT NULL,
  `image` varchar(255) DEFAULT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `s_pod_info`
--

DROP TABLE IF EXISTS `s_pod_info`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `s_pod_info` (
  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT,
  `created_at` timestamp NULL DEFAULT NULL,
  `updated_at` timestamp NULL DEFAULT NULL,
  `cluster` varchar(64) DEFAULT NULL,
  `namespace` varchar(64) DEFAULT NULL,
  `name` varchar(64) DEFAULT NULL,
  `org_name` varchar(64) DEFAULT NULL,
  `org_id` varchar(64) DEFAULT NULL,
  `project_name` varchar(64) DEFAULT NULL,
  `project_id` varchar(64) DEFAULT NULL,
  `application_name` varchar(255) DEFAULT NULL,
  `application_id` varchar(255) DEFAULT NULL,
  `runtime_name` varchar(255) DEFAULT NULL,
  `runtime_id` varchar(255) DEFAULT NULL,
  `service_name` varchar(255) DEFAULT NULL,
  `workspace` varchar(10) DEFAULT NULL,
  `service_type` varchar(64) DEFAULT NULL,
  `addon_id` varchar(255) DEFAULT NULL,
  `uid` varchar(128) DEFAULT NULL,
  `k8s_namespace` varchar(128) DEFAULT NULL,
  `pod_name` varchar(128) DEFAULT NULL,
  `phase` varchar(255) DEFAULT NULL,
  `message` varchar(1024) DEFAULT NULL,
  `pod_ip` varchar(255) DEFAULT NULL,
  `host_ip` varchar(255) DEFAULT NULL,
  `started_at` timestamp NULL DEFAULT NULL,
  `cpu_request` double DEFAULT NULL,
  `mem_request` int(11) DEFAULT NULL,
  `cpu_limit` double DEFAULT NULL,
  `mem_limit` int(11) DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_s_pod_info_namespace` (`namespace`),
  KEY `idx_s_pod_info_name` (`name`),
  KEY `idx_s_pod_info_org_name` (`org_name`),
  KEY `idx_s_pod_info_org_id` (`org_id`),
  KEY `idx_s_pod_info_project_id` (`project_id`),
  KEY `idx_s_pod_info_cluster` (`cluster`),
  KEY `idx_s_pod_info_project_name` (`project_name`),
  KEY `idx_s_pod_info_k8s_namespace` (`k8s_namespace`),
  KEY `idx_s_pod_info_pod_name` (`pod_name`),
  KEY `idx_s_pod_info_uid` (`uid`)
) ENGINE=InnoDB AUTO_INCREMENT=1915237 DEFAULT CHARSET=utf8mb4 COMMENT='pod信息';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `s_service_info`
--

DROP TABLE IF EXISTS `s_service_info`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `s_service_info` (
  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT,
  `created_at` timestamp NULL DEFAULT NULL,
  `updated_at` timestamp NULL DEFAULT NULL,
  `cluster` varchar(255) DEFAULT NULL,
  `namespace` varchar(64) DEFAULT NULL,
  `name` varchar(64) DEFAULT NULL,
  `org_name` varchar(64) DEFAULT NULL,
  `org_id` varchar(64) DEFAULT NULL,
  `project_name` varchar(64) DEFAULT NULL,
  `project_id` varchar(64) DEFAULT NULL,
  `application_name` varchar(255) DEFAULT NULL,
  `application_id` varchar(255) DEFAULT NULL,
  `runtime_name` varchar(255) DEFAULT NULL,
  `runtime_id` varchar(255) DEFAULT NULL,
  `service_name` varchar(255) DEFAULT NULL,
  `workspace` varchar(10) DEFAULT NULL,
  `service_type` varchar(64) DEFAULT NULL,
  `meta` varchar(255) DEFAULT NULL,
  `phase` varchar(255) DEFAULT NULL,
  `message` varchar(255) DEFAULT NULL,
  `started_at` timestamp NULL DEFAULT NULL,
  `finished_at` timestamp NULL DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_s_service_info_org_name` (`org_name`),
  KEY `idx_s_service_info_org_id` (`org_id`),
  KEY `idx_s_service_info_project_name` (`project_name`),
  KEY `idx_s_service_info_project_id` (`project_id`),
  KEY `idx_s_service_info_namespace` (`namespace`),
  KEY `idx_s_service_info_name` (`name`)
) ENGINE=InnoDB AUTO_INCREMENT=427 DEFAULT CHARSET=utf8mb4 BLOCK_FORMAT=ENCRYPTED;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `s_service_info_bak`
--

DROP TABLE IF EXISTS `s_service_info_bak`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `s_service_info_bak` (
  `id` bigint(20) unsigned NOT NULL DEFAULT '0',
  `created_at` timestamp NULL DEFAULT NULL,
  `updated_at` timestamp NULL DEFAULT NULL,
  `cluster` varchar(255) DEFAULT NULL,
  `namespace` varchar(64) DEFAULT NULL,
  `name` varchar(64) DEFAULT NULL,
  `org_name` varchar(64) DEFAULT NULL,
  `org_id` varchar(64) DEFAULT NULL,
  `project_name` varchar(64) DEFAULT NULL,
  `project_id` varchar(64) DEFAULT NULL,
  `application_name` varchar(255) DEFAULT NULL,
  `application_id` varchar(255) DEFAULT NULL,
  `runtime_name` varchar(255) DEFAULT NULL,
  `runtime_id` varchar(255) DEFAULT NULL,
  `service_name` varchar(255) DEFAULT NULL,
  `workspace` varchar(10) DEFAULT NULL,
  `service_type` varchar(64) DEFAULT NULL,
  `meta` varchar(255) DEFAULT NULL,
  `phase` varchar(255) DEFAULT NULL,
  `message` varchar(255) DEFAULT NULL,
  `started_at` timestamp NULL DEFAULT NULL,
  `finished_at` timestamp NULL DEFAULT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `sp_account`
--

DROP TABLE IF EXISTS `sp_account`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `sp_account` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `auth_id` int(11) DEFAULT NULL,
  `username` varchar(255) DEFAULT NULL,
  `password` varchar(255) DEFAULT NULL,
  `create_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `update_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '修改时间',
  `is_deleted` varchar(1) NOT NULL DEFAULT 'N' COMMENT '逻辑删除',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8 BLOCK_FORMAT=ENCRYPTED COMMENT='主动监控存储用户账号表,已废弃';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `sp_alert`
--

DROP TABLE IF EXISTS `sp_alert`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `sp_alert` (
  `id` int(11) unsigned NOT NULL AUTO_INCREMENT COMMENT '主键',
  `name` varchar(255) NOT NULL DEFAULT '' COMMENT '告警名称',
  `alert_scope` varchar(255) NOT NULL DEFAULT 'micro_service' COMMENT '告警所在的组织域，包含org、project和app',
  `alert_scope_id` varchar(255) NOT NULL DEFAULT '' COMMENT '告警范围id',
  `alert_type` varchar(64) NOT NULL DEFAULT '' COMMENT '告警类型',
  `attributes` varchar(2048) NOT NULL DEFAULT '' COMMENT '告警扩展信息',
  `enable` int(1) NOT NULL DEFAULT '1' COMMENT '告警开关',
  `created` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '告警创建时间',
  `updated` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '告警更新时间',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=80 DEFAULT CHARSET=utf8mb4 BLOCK_FORMAT=ENCRYPTED COMMENT='监控告警策略记录表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `sp_alert_expression`
--

DROP TABLE IF EXISTS `sp_alert_expression`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `sp_alert_expression` (
  `id` int(11) unsigned NOT NULL AUTO_INCREMENT COMMENT '主键',
  `alert_id` int(11) NOT NULL COMMENT '告警外键ID',
  `attributes` varchar(2048) NOT NULL DEFAULT '' COMMENT '告警规则扩展信息',
  `expression` varchar(4096) NOT NULL DEFAULT '' COMMENT '告警规则表达式',
  `version` varchar(64) NOT NULL DEFAULT '1.0' COMMENT '告警规则版本',
  `enable` int(1) NOT NULL DEFAULT '1' COMMENT '告警规则开关',
  `created` datetime NOT NULL COMMENT '告警规则创建时间',
  `updated` datetime NOT NULL COMMENT '告警规则更新时间',
  PRIMARY KEY (`id`),
  KEY `INDEX_ALERT_ID` (`alert_id`)
) ENGINE=InnoDB AUTO_INCREMENT=308 DEFAULT CHARSET=utf8mb4 BLOCK_FORMAT=ENCRYPTED COMMENT='监控告警策略中记录表达式的表,monitor和analyzer都会使用';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `sp_alert_notify`
--

DROP TABLE IF EXISTS `sp_alert_notify`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `sp_alert_notify` (
  `id` int(11) unsigned NOT NULL AUTO_INCREMENT COMMENT '主键',
  `alert_id` int(11) NOT NULL COMMENT '告警外键ID',
  `notify_key` varchar(128) NOT NULL COMMENT '告警属性key',
  `notify_target` varchar(1024) NOT NULL COMMENT '通知信息JSON',
  `notify_target_id` varchar(1024) DEFAULT NULL COMMENT '通知地址 已废弃',
  `silence` int(11) NOT NULL DEFAULT '1800000' COMMENT '静默时间，单位ms',
  `silence_policy` varchar(128) NOT NULL DEFAULT 'fixed' COMMENT '静默时间策略',
  `enable` int(1) NOT NULL DEFAULT '1' COMMENT '通知开关',
  `created` datetime NOT NULL COMMENT '通知创建时间',
  `updated` datetime NOT NULL COMMENT '通知更新时间',
  PRIMARY KEY (`id`),
  KEY `INDEX_NOTIFY_KEY` (`notify_key`),
  KEY `INDEX_ALERT_ID` (`alert_id`)
) ENGINE=InnoDB AUTO_INCREMENT=169 DEFAULT CHARSET=utf8mb4 BLOCK_FORMAT=ENCRYPTED COMMENT='监控告警策略中记录通知对象的表,monitor和analyzer-alert都会使';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `sp_alert_notify_template`
--

DROP TABLE IF EXISTS `sp_alert_notify_template`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `sp_alert_notify_template` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `name` varchar(255) NOT NULL COMMENT '名称',
  `alert_type` varchar(32) NOT NULL COMMENT '类型',
  `alert_index` varchar(255) NOT NULL COMMENT '索引',
  `target` varchar(255) NOT NULL COMMENT '发送类型',
  `trigger` varchar(255) NOT NULL COMMENT '触发类似，alert| recovery',
  `title` varchar(4096) NOT NULL COMMENT '模板标题',
  `template` varchar(4096) NOT NULL COMMENT '模板JSON，包含cyle、agg、operator、value、unit等信息',
  `formats` varchar(4096) NOT NULL DEFAULT '{}' COMMENT '字段格式化定义',
  `version` varchar(255) NOT NULL DEFAULT '1.0' COMMENT '版本',
  `create_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `update_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '修改时间',
  `enable` tinyint(1) NOT NULL DEFAULT '1' COMMENT '逻辑删除',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=6790 DEFAULT CHARSET=utf8mb4 COMMENT='告警模板表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `sp_alert_record`
--

DROP TABLE IF EXISTS `sp_alert_record`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `sp_alert_record` (
  `group_id` varchar(191) NOT NULL COMMENT '分组ID，唯一键',
  `scope` varchar(16) NOT NULL COMMENT '范围',
  `scope_key` varchar(128) NOT NULL COMMENT '范围Key',
  `alert_group` varchar(128) NOT NULL COMMENT '目标Key',
  `title` varchar(4096) NOT NULL COMMENT '标题',
  `alert_state` varchar(16) NOT NULL COMMENT '告警状态',
  `alert_type` varchar(32) NOT NULL COMMENT '告警类型',
  `alert_index` varchar(128) NOT NULL COMMENT '告警索引',
  `expression_key` varchar(32) NOT NULL COMMENT '告警表达式Key',
  `alert_id` bigint(20) NOT NULL COMMENT '告警ID',
  `alert_name` varchar(32) NOT NULL COMMENT '告警名称',
  `rule_id` bigint(20) NOT NULL COMMENT '告警规则ID',
  `issue_id` bigint(20) DEFAULT NULL COMMENT 'issue ID',
  `handle_state` varchar(16) DEFAULT NULL COMMENT '处理状态',
  `handler_id` varchar(128) DEFAULT NULL COMMENT '处理人ID',
  `alert_time` datetime DEFAULT NULL COMMENT '告警时间',
  `handle_time` datetime DEFAULT NULL COMMENT '处理时间',
  `create_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `update_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '修改时间',
  PRIMARY KEY (`group_id`),
  KEY `idx_scope_key` (`scope_key`),
  KEY `idx_alert_type` (`alert_type`),
  KEY `idx_alert_group` (`alert_group`),
  KEY `idx_handler_id` (`handler_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='告警记录表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `sp_alert_rules`
--

DROP TABLE IF EXISTS `sp_alert_rules`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `sp_alert_rules` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `name` varchar(255) NOT NULL COMMENT '名称',
  `alert_scope` varchar(32) NOT NULL COMMENT '范围',
  `alert_type` varchar(32) NOT NULL COMMENT '类型',
  `alert_index` varchar(255) NOT NULL COMMENT '索引',
  `template` varchar(4096) NOT NULL COMMENT '模板JSON，包含cyle、agg、operator、value、unit等信息',
  `attributes` varchar(1024) NOT NULL COMMENT '扩展属性JSON，包含is_recovery是否恢复后告警',
  `enable` tinyint(1) NOT NULL DEFAULT '1' COMMENT '逻辑删除',
  `create_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `update_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '修改时间',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=1038 DEFAULT CHARSET=utf8mb4 COMMENT='告警规则表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `sp_authentication`
--

DROP TABLE IF EXISTS `sp_authentication`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `sp_authentication` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `project_id` int(11) unsigned zerofill DEFAULT NULL,
  `name` varchar(255) DEFAULT NULL,
  `extra` varchar(2048) NOT NULL DEFAULT '',
  `create_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `update_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '修改时间',
  `is_deleted` varchar(1) NOT NULL DEFAULT 'N' COMMENT '逻辑删除',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8 BLOCK_FORMAT=ENCRYPTED COMMENT='主动监控存储用户认证的表,已废弃';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `sp_chart_meta`
--

DROP TABLE IF EXISTS `sp_chart_meta`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `sp_chart_meta` (
  `id` int(10) NOT NULL AUTO_INCREMENT COMMENT '主键',
  `name` varchar(64) NOT NULL,
  `title` varchar(64) NOT NULL,
  `metricsName` varchar(127) NOT NULL,
  `fields` varchar(4096) NOT NULL,
  `parameters` varchar(4096) NOT NULL,
  `type` varchar(64) NOT NULL,
  `order` int(11) NOT NULL,
  `unit` varchar(16) NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `name_unique` (`name`)
) ENGINE=InnoDB AUTO_INCREMENT=116 DEFAULT CHARSET=utf8mb4 BLOCK_FORMAT=ENCRYPTED COMMENT='监控图表元数据配置';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `sp_chart_profile`
--

DROP TABLE IF EXISTS `sp_chart_profile`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `sp_chart_profile` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT,
  `created_at` datetime NOT NULL,
  `updated_at` datetime DEFAULT NULL,
  `unique_id` varchar(125) DEFAULT NULL,
  `layout` text,
  `drawer_info_map` text,
  `url_config` text,
  `name` varchar(125) NOT NULL,
  `category` varchar(125) NOT NULL,
  `cluster_name` varchar(125) DEFAULT '',
  `organization` varchar(125) DEFAULT '',
  `editable` tinyint(1) unsigned DEFAULT '0',
  `deletable` tinyint(1) DEFAULT '0',
  `cluster_level` tinyint(1) DEFAULT '0',
  `cluster_type` varchar(20) NOT NULL DEFAULT '',
  PRIMARY KEY (`id`),
  UNIQUE KEY `unique_id` (`unique_id`,`organization`,`cluster_name`)
) ENGINE=InnoDB AUTO_INCREMENT=14 DEFAULT CHARSET=utf8 BLOCK_FORMAT=ENCRYPTED COMMENT='告警消息跳转的页面配置';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `sp_customize_alert`
--

DROP TABLE IF EXISTS `sp_customize_alert`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `sp_customize_alert` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `name` varchar(64) NOT NULL COMMENT '告警名称',
  `alert_type` varchar(32) NOT NULL COMMENT '类型',
  `alert_scope` varchar(32) NOT NULL COMMENT '告警所在的组织域，包含org、project和app',
  `alert_scope_id` varchar(128) NOT NULL DEFAULT '' COMMENT '告警范围id',
  `attributes` varchar(1024) NOT NULL COMMENT '扩展属性JSON',
  `enable` int(1) NOT NULL DEFAULT '1' COMMENT '告警开关',
  `create_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '告警创建时间',
  `update_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '告警更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uniq_scope_X_name` (`alert_scope_id`,`name`,`alert_scope`)
) ENGINE=InnoDB AUTO_INCREMENT=33 DEFAULT CHARSET=utf8mb4 COMMENT='自定义告警';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `sp_customize_alert_notify_template`
--

DROP TABLE IF EXISTS `sp_customize_alert_notify_template`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `sp_customize_alert_notify_template` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `name` varchar(64) NOT NULL COMMENT '名称',
  `customize_alert_id` bigint(20) NOT NULL COMMENT '自定义告警ID',
  `alert_type` varchar(32) NOT NULL COMMENT '类型',
  `alert_index` varchar(128) NOT NULL COMMENT '索引',
  `target` varchar(255) NOT NULL COMMENT '发送类型',
  `trigger` varchar(255) NOT NULL COMMENT '触发类似，alert| recovery',
  `title` varchar(4096) NOT NULL COMMENT '模板标题',
  `template` varchar(4096) NOT NULL COMMENT '模板JSON，包含cyle、agg、operator、value、unit等信息',
  `formats` varchar(4096) NOT NULL DEFAULT '{}' COMMENT '字段格式化定义',
  `version` varchar(255) NOT NULL DEFAULT '1.0' COMMENT '版本',
  `create_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `update_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '修改时间',
  `enable` tinyint(1) NOT NULL DEFAULT '1' COMMENT '逻辑删除',
  PRIMARY KEY (`id`),
  KEY `idx_customize_alert_id` (`customize_alert_id`),
  KEY `idx_alert_index` (`alert_index`)
) ENGINE=InnoDB AUTO_INCREMENT=197 DEFAULT CHARSET=utf8mb4 COMMENT='自定义告警模板表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `sp_customize_alert_rule`
--

DROP TABLE IF EXISTS `sp_customize_alert_rule`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `sp_customize_alert_rule` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `name` varchar(64) NOT NULL COMMENT '名称',
  `customize_alert_id` bigint(20) NOT NULL COMMENT '自定义告警ID',
  `alert_scope` varchar(32) NOT NULL COMMENT '告警所在的组织域，包含org、project和app',
  `alert_scope_id` varchar(128) NOT NULL COMMENT '告警所在的组织ID',
  `alert_type` varchar(32) NOT NULL COMMENT '类型',
  `alert_index` varchar(128) NOT NULL COMMENT '索引',
  `template` varchar(4096) NOT NULL COMMENT '模板JSON，包含cyle、agg、operator、value、unit等信息',
  `attributes` varchar(1024) NOT NULL COMMENT '扩展属性JSON，包含is_recovery是否恢复后告警',
  `enable` tinyint(1) NOT NULL DEFAULT '1' COMMENT '逻辑删除',
  `create_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `update_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '修改时间',
  PRIMARY KEY (`id`),
  KEY `idx_customize_alert_id` (`customize_alert_id`),
  KEY `idx_scope_id` (`alert_scope_id`),
  KEY `idx_alert_index` (`alert_index`)
) ENGINE=InnoDB AUTO_INCREMENT=85 DEFAULT CHARSET=utf8mb4 COMMENT='自定义告警规则表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `sp_dashboard_block`
--

DROP TABLE IF EXISTS `sp_dashboard_block`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `sp_dashboard_block` (
  `id` varchar(64) NOT NULL COMMENT '主键ID',
  `name` varchar(32) NOT NULL COMMENT '名称',
  `desc` varchar(255) DEFAULT NULL COMMENT '描述',
  `domain` varchar(255) DEFAULT NULL,
  `scope` varchar(255) DEFAULT NULL,
  `scope_id` varchar(255) DEFAULT NULL,
  `view_config` text NOT NULL COMMENT '数据配置',
  `data_config` text NOT NULL COMMENT '数据配置',
  `created_at` datetime DEFAULT NULL COMMENT '创建时间',
  `updated_at` datetime DEFAULT NULL COMMENT '修改时间',
  `version` varchar(32) NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `id` (`name`,`scope`,`scope_id`,`id`),
  UNIQUE KEY `Scope` (`name`,`scope`,`scope_id`,`id`),
  UNIQUE KEY `ScopeID` (`name`,`scope`,`scope_id`,`id`),
  UNIQUE KEY `Name` (`name`,`scope`,`scope_id`,`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8 COMMENT='用户配置的自定义大盘表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `sp_dashboard_block_system`
--

DROP TABLE IF EXISTS `sp_dashboard_block_system`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `sp_dashboard_block_system` (
  `id` varchar(64) NOT NULL COMMENT '主键ID',
  `name` varchar(32) NOT NULL COMMENT '名称',
  `desc` varchar(255) DEFAULT NULL COMMENT '描述',
  `domain` varchar(255) DEFAULT NULL,
  `scope` varchar(255) DEFAULT NULL,
  `scope_id` varchar(255) DEFAULT NULL,
  `view_config` text NOT NULL COMMENT '图表配置',
  `data_config` text COMMENT '数据配置',
  `created_at` datetime DEFAULT NULL COMMENT '创建时间',
  `updated_at` datetime DEFAULT NULL COMMENT '修改时间',
  `version` varchar(32) NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `id` (`id`),
  UNIQUE KEY `Scope` (`name`,`scope`,`scope_id`),
  UNIQUE KEY `ScopeID` (`name`,`scope`,`scope_id`),
  UNIQUE KEY `Name` (`name`,`scope`,`scope_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8 COMMENT='系统内置的大盘配置表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `sp_diagnosis`
--

DROP TABLE IF EXISTS `sp_diagnosis`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `sp_diagnosis` (
  `id` varchar(191) NOT NULL COMMENT '主键ID',
  `scope` varchar(255) NOT NULL COMMENT 'diagnosis作用域',
  `scope_id` varchar(255) NOT NULL COMMENT 'diagnosis作用域Id',
  `diagnosis_object` varchar(191) NOT NULL COMMENT '被diagnosis的目标id',
  `request_token` varchar(766) NOT NULL COMMENT 'diagnosis请求token',
  `state` varchar(191) NOT NULL COMMENT '状态',
  `attributes` varchar(4096) NOT NULL COMMENT '附加属性，作为UI展示',
  `create_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `update_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '修改时间',
  PRIMARY KEY (`id`),
  KEY `DIAGNOSIS_OBJECT_INDEX` (`diagnosis_object`),
  KEY `STATE_INDEX` (`state`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='监控diagnosis记录表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `sp_diagnosis_actions`
--

DROP TABLE IF EXISTS `sp_diagnosis_actions`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `sp_diagnosis_actions` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `diagnosis_id` varchar(191) NOT NULL COMMENT 'diagnosis唯一标识',
  `action` varchar(255) NOT NULL COMMENT '操作',
  `state` varchar(255) NOT NULL COMMENT '状态',
  `config` varchar(4096) NOT NULL COMMENT 'config',
  `create_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `update_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '修改时间',
  PRIMARY KEY (`id`),
  KEY `DIAGNOSIS_ID_INDEX` (`diagnosis_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='监控diagnosis的子操作表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `sp_diagnosis_resource`
--

DROP TABLE IF EXISTS `sp_diagnosis_resource`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `sp_diagnosis_resource` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `resource_id` varchar(255) NOT NULL COMMENT '资源唯一标志',
  `diagnosis_id` varchar(255) NOT NULL COMMENT 'diagnosis唯一标识',
  `resource` varchar(255) NOT NULL COMMENT '资源',
  `state` varchar(255) NOT NULL COMMENT '状态',
  `config` varchar(4096) NOT NULL COMMENT '注入到资源的配置JSON',
  `condition` varchar(4096) NOT NULL COMMENT '记录资源最后一次状态的JSON详细数据',
  `create_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `update_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '修改时间',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='监控diagnosis资源管理表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `sp_diagnosis_state`
--

DROP TABLE IF EXISTS `sp_diagnosis_state`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `sp_diagnosis_state` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `diagnosis_id` varchar(255) NOT NULL COMMENT 'diagnosis唯一标识',
  `state` varchar(191) NOT NULL COMMENT '状态',
  `reason` varchar(255) NOT NULL COMMENT '当前状态正在执行的操作',
  `version` varchar(255) NOT NULL COMMENT '状态版本，用于记录锁',
  `create_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `update_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '修改时间',
  PRIMARY KEY (`id`),
  KEY `STATE_INDEX` (`state`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='监控diagnosis状态表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `sp_history`
--

DROP TABLE IF EXISTS `sp_history`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `sp_history` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `metric_id` int(11) DEFAULT NULL,
  `status_id` int(11) DEFAULT NULL,
  `latency` bigint(64) NOT NULL,
  `count` int(11) NOT NULL,
  `code` int(11) unsigned zerofill NOT NULL,
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `metricId` (`metric_id`),
  KEY `statusId` (`status_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8 BLOCK_FORMAT=ENCRYPTED COMMENT='主动监控存储历史记录项的表,已废弃';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `sp_log_deployment`
--

DROP TABLE IF EXISTS `sp_log_deployment`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `sp_log_deployment` (
  `id` int(10) NOT NULL AUTO_INCREMENT,
  `org_id` varchar(32) NOT NULL,
  `cluster_name` varchar(64) NOT NULL,
  `cluster_type` tinyint(4) NOT NULL,
  `es_url` varchar(1024) NOT NULL,
  `es_config` varchar(1024) NOT NULL,
  `kafka_servers` varchar(1024) NOT NULL,
  `kafka_config` varchar(1024) NOT NULL,
  `collector_url` varchar(255) NOT NULL DEFAULT '',
  `domain` varchar(255) NOT NULL,
  `created` datetime NOT NULL,
  `updated` datetime NOT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=4 DEFAULT CHARSET=utf8mb4 BLOCK_FORMAT=ENCRYPTED COMMENT='日志分析部署表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `sp_log_instance`
--

DROP TABLE IF EXISTS `sp_log_instance`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `sp_log_instance` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `log_key` varchar(64) NOT NULL,
  `org_id` varchar(32) NOT NULL,
  `org_name` varchar(64) NOT NULL,
  `cluster_name` varchar(64) NOT NULL,
  `project_id` varchar(32) NOT NULL,
  `project_name` varchar(255) NOT NULL,
  `workspace` varchar(64) NOT NULL,
  `application_id` varchar(32) NOT NULL,
  `application_name` varchar(255) NOT NULL,
  `runtime_id` varchar(32) NOT NULL,
  `runtime_name` varchar(255) NOT NULL,
  `config` varchar(1023) NOT NULL,
  `version` varchar(64) NOT NULL,
  `plan` varchar(255) NOT NULL,
  `is_delete` tinyint(1) NOT NULL,
  `created` datetime NOT NULL,
  `updated` datetime NOT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=6 DEFAULT CHARSET=utf8mb4 BLOCK_FORMAT=ENCRYPTED COMMENT='日志分析addon实例表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `sp_log_metric_config`
--

DROP TABLE IF EXISTS `sp_log_metric_config`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `sp_log_metric_config` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `org_id` int(11) NOT NULL,
  `org_name` varchar(128) NOT NULL DEFAULT '',
  `scope` varchar(32) NOT NULL,
  `scope_id` varchar(64) NOT NULL,
  `name` varchar(128) NOT NULL,
  `metric` varchar(128) NOT NULL,
  `filters` text NOT NULL,
  `processors` text NOT NULL,
  `enable` tinyint(1) NOT NULL,
  `create_time` datetime NOT NULL,
  `update_time` datetime NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `metric_unique` (`metric`) USING BTREE,
  UNIQUE KEY `scope_name_unique` (`scope`,`scope_id`,`name`) USING BTREE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='日志分析规则配置表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `sp_maintenance`
--

DROP TABLE IF EXISTS `sp_maintenance`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `sp_maintenance` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `project_id` int(11) DEFAULT NULL,
  `name` varchar(255) DEFAULT NULL,
  `duration` int(11) NOT NULL DEFAULT '0',
  `start_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `project_id` (`project_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8 BLOCK_FORMAT=ENCRYPTED COMMENT='主动监控存储配置项的表,已废弃';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `sp_metric`
--

DROP TABLE IF EXISTS `sp_metric`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `sp_metric` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `project_id` int(11) DEFAULT NULL,
  `service_id` int(11) DEFAULT NULL,
  `name` varchar(255) NOT NULL DEFAULT '',
  `url` varchar(255) DEFAULT NULL,
  `mode` varchar(255) NOT NULL DEFAULT '',
  `extra` varchar(1024) NOT NULL DEFAULT '',
  `account_id` int(11) NOT NULL,
  `status` int(11) DEFAULT '0',
  `env` varchar(36) NOT NULL DEFAULT 'PROD',
  `create_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `update_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '修改时间',
  `is_deleted` varchar(1) NOT NULL DEFAULT 'N' COMMENT '逻辑删除',
  PRIMARY KEY (`id`),
  KEY `projectId` (`project_id`),
  KEY `serviceId` (`service_id`)
) ENGINE=InnoDB AUTO_INCREMENT=6 DEFAULT CHARSET=utf8 BLOCK_FORMAT=ENCRYPTED COMMENT='主动监控存储配置项的表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `sp_metric_expression`
--

DROP TABLE IF EXISTS `sp_metric_expression`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `sp_metric_expression` (
  `id` int(11) unsigned NOT NULL AUTO_INCREMENT COMMENT '主键',
  `attributes` varchar(1024) NOT NULL DEFAULT '' COMMENT '扩展信息',
  `expression` varchar(4096) NOT NULL DEFAULT '' COMMENT '聚合表达式',
  `version` varchar(64) NOT NULL DEFAULT '2.0' COMMENT '版本',
  `enable` int(1) NOT NULL DEFAULT '1' COMMENT '开关',
  `created` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '修改时间',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=216 DEFAULT CHARSET=utf8mb4 BLOCK_FORMAT=ENCRYPTED COMMENT='系统内置metrics表达式表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `sp_metric_meta`
--

DROP TABLE IF EXISTS `sp_metric_meta`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `sp_metric_meta` (
  `id` int(11) NOT NULL AUTO_INCREMENT COMMENT '主键',
  `scope` varchar(64) NOT NULL,
  `scope_id` varchar(64) NOT NULL,
  `group` varchar(128) NOT NULL,
  `metric` varchar(128) NOT NULL,
  `name` varchar(128) NOT NULL,
  `tags` text NOT NULL,
  `fields` text NOT NULL,
  `create_time` datetime NOT NULL,
  `update_time` datetime NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `query_index` (`scope`,`scope_id`,`metric`) USING BTREE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='日志分析生成的metric元数据存储';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `sp_metric_metas`
--

DROP TABLE IF EXISTS `sp_metric_metas`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `sp_metric_metas` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `metric` varchar(64) NOT NULL,
  `meta_metric` varchar(64) NOT NULL,
  `name` varchar(64) NOT NULL,
  `type` char(16) NOT NULL,
  `unit` char(16) NOT NULL,
  `force_tags` text,
  `group_id` varchar(32) NOT NULL,
  `created_at` datetime DEFAULT NULL,
  `updated_at` datetime DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `metric_id` (`metric`,`meta_metric`,`group_id`),
  KEY `idx_group_id` (`group_id`)
) ENGINE=InnoDB AUTO_INCREMENT=273 DEFAULT CHARSET=utf8 COMMENT='监控图表元数据配置';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `sp_monitor`
--

DROP TABLE IF EXISTS `sp_monitor`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `sp_monitor` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `monitor_id` varchar(55) NOT NULL DEFAULT '',
  `terminus_key` varchar(55) NOT NULL DEFAULT '',
  `terminus_key_runtime` varchar(55) DEFAULT NULL,
  `workspace` varchar(55) DEFAULT NULL,
  `runtime_id` varchar(55) NOT NULL DEFAULT '',
  `runtime_name` varchar(255) DEFAULT NULL,
  `application_id` varchar(55) DEFAULT NULL,
  `application_name` varchar(255) DEFAULT NULL,
  `project_id` varchar(55) DEFAULT NULL,
  `project_name` varchar(255) DEFAULT NULL,
  `org_id` varchar(55) DEFAULT NULL,
  `org_name` varchar(255) DEFAULT NULL,
  `cluster_id` varchar(255) DEFAULT NULL,
  `cluster_name` varchar(125) DEFAULT NULL,
  `config` varchar(1023) NOT NULL DEFAULT '',
  `callback_url` varchar(255) DEFAULT NULL,
  `version` varchar(55) DEFAULT NULL,
  `plan` varchar(255) DEFAULT NULL,
  `is_delete` tinyint(1) NOT NULL DEFAULT '0',
  `created` datetime NOT NULL,
  `updated` datetime NOT NULL,
  PRIMARY KEY (`id`),
  KEY `sp_monitor_monitor_id` (`monitor_id`),
  KEY `sp_monitor_runtime_id` (`runtime_id`),
  KEY `sp_monitor_terminus_key` (`terminus_key`),
  KEY `sp_monitor_application_id` (`application_id`),
  KEY `sp_monitor_workspace` (`workspace`),
  KEY `sp_monitor_project_id` (`project_id`)
) ENGINE=InnoDB AUTO_INCREMENT=6578 DEFAULT CHARSET=utf8mb4 BLOCK_FORMAT=ENCRYPTED COMMENT='应用监控addon实例表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `sp_monitor_config`
--

DROP TABLE IF EXISTS `sp_monitor_config`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `sp_monitor_config` (
  `id` int(11) NOT NULL AUTO_INCREMENT COMMENT '主键',
  `org_id` int(11) NOT NULL COMMENT '企业ID',
  `org_name` varchar(64) NOT NULL DEFAULT '' COMMENT '企业名',
  `type` varchar(64) NOT NULL COMMENT 'log、metric',
  `names` varchar(255) NOT NULL COMMENT '名称,支持通配符*,支持逗号分隔',
  `filters` varchar(4096) NOT NULL DEFAULT '' COMMENT '数组,层级高的在前面,[{"key":"project_id","value":"1"}]',
  `config` varchar(1024) NOT NULL DEFAULT '' COMMENT '具体配置,{"ttl":"168h"}',
  `create_time` datetime NOT NULL COMMENT '创建时间',
  `update_time` datetime NOT NULL COMMENT '更新时间',
  `enable` tinyint(1) NOT NULL COMMENT '是否启用',
  `key` varchar(32) NOT NULL DEFAULT '' COMMENT '配置唯一标示',
  `hash` varchar(64) NOT NULL COMMENT 'hash(org_id,type,names+filters)',
  PRIMARY KEY (`id`),
  UNIQUE KEY `unique_keys` (`hash`) USING BTREE
) ENGINE=InnoDB AUTO_INCREMENT=9 DEFAULT CHARSET=utf8 COMMENT='实际监控配置表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `sp_monitor_config_register`
--

DROP TABLE IF EXISTS `sp_monitor_config_register`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `sp_monitor_config_register` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `scope` varchar(32) NOT NULL COMMENT 'org',
  `scope_id` varchar(64) NOT NULL DEFAULT '' COMMENT 'org_id',
  `namespace` varchar(255) NOT NULL COMMENT 'dev,test,staging,prod,other',
  `type` varchar(64) NOT NULL COMMENT 'metric、log',
  `names` varchar(255) NOT NULL,
  `filters` varchar(4096) NOT NULL DEFAULT '',
  `enable` tinyint(1) NOT NULL DEFAULT '1',
  `update_time` datetime NOT NULL,
  `desc` varchar(64) NOT NULL DEFAULT '',
  `hash` varchar(64) NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `hash_unique` (`hash`) USING BTREE
) ENGINE=InnoDB AUTO_INCREMENT=107 DEFAULT CHARSET=utf8 COMMENT='企业监控和日志过期时间配置表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `sp_monitor_global_settings`
--

DROP TABLE IF EXISTS `sp_monitor_global_settings`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `sp_monitor_global_settings` (
  `id` int(11) NOT NULL AUTO_INCREMENT COMMENT '主键',
  `org_id` int(11) NOT NULL COMMENT '企业id',
  `org_name` varchar(255) NOT NULL COMMENT '企业',
  `namespace` varchar(64) NOT NULL COMMENT '通用、开发、测试、预发、生产',
  `group` varchar(64) NOT NULL COMMENT '分组',
  `key` varchar(64) NOT NULL COMMENT '键',
  `type` varchar(32) NOT NULL COMMENT '值类型',
  `value` varchar(128) NOT NULL COMMENT '值',
  `unit` varchar(64) NOT NULL COMMENT '单位',
  PRIMARY KEY (`id`),
  UNIQUE KEY `unique_key` (`org_id`,`namespace`,`group`,`key`)
) ENGINE=InnoDB AUTO_INCREMENT=34 DEFAULT CHARSET=utf8 COMMENT='全局单一配置，面向用户的配置';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `sp_notify`
--

DROP TABLE IF EXISTS `sp_notify`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `sp_notify` (
  `id` int(11) unsigned NOT NULL AUTO_INCREMENT,
  `notify_id` varchar(64) NOT NULL COMMENT '模版ID唯一',
  `created_at` datetime DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime DEFAULT CURRENT_TIMESTAMP,
  `deleted_at` datetime DEFAULT NULL,
  `notify_name` varchar(128) DEFAULT NULL COMMENT '通知名称',
  `scope` varchar(128) DEFAULT NULL,
  `scope_id` int(11) DEFAULT NULL,
  `target` varchar(255) DEFAULT NULL COMMENT '存放群组相关的信息，json格式 {"group_id":1,"notify_style":"email,webhook"}',
  `attributes` varchar(1024) DEFAULT NULL,
  `enable` tinyint(1) DEFAULT '1' COMMENT '该告警通知模版是否启用',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=213 DEFAULT CHARSET=utf8mb4 COMMENT='用户告警通知';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `sp_notify_record`
--

DROP TABLE IF EXISTS `sp_notify_record`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `sp_notify_record` (
  `notify_id` varchar(64) NOT NULL COMMENT '模版ID唯一',
  `notify_name` varchar(512) NOT NULL COMMENT '通知名称',
  `scope_type` varchar(128) NOT NULL COMMENT '租户类型',
  `scope_id` int(11) NOT NULL COMMENT '租户id',
  `group_id` varchar(128) NOT NULL COMMENT '通知组id',
  `notify_group` varchar(256) DEFAULT NULL COMMENT '通知组名称',
  `title` varchar(256) DEFAULT NULL COMMENT '通知模版标题',
  `notify_time` datetime DEFAULT NULL COMMENT '通知处理时间',
  `create_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `update_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '修改时间',
  PRIMARY KEY (`group_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `sp_notify_user_define`
--

DROP TABLE IF EXISTS `sp_notify_user_define`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `sp_notify_user_define` (
  `id` int(11) unsigned NOT NULL AUTO_INCREMENT,
  `notify_id` varchar(64) NOT NULL COMMENT '模版id唯一',
  `metadata` varchar(2048) DEFAULT NULL COMMENT '模版字段',
  `behavior` varchar(2048) NOT NULL COMMENT '模版字段',
  `templates` varchar(4096) NOT NULL COMMENT '模版字段',
  `scope` varchar(128) NOT NULL DEFAULT 'org' COMMENT '租户类型',
  `scope_id` int(11) NOT NULL COMMENT '租户id',
  `updated_at` datetime DEFAULT CURRENT_TIMESTAMP,
  `created_at` datetime DEFAULT CURRENT_TIMESTAMP,
  `deleted_at` datetime DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `notify_id` (`notify_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='用户自定义告警通知模版';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `sp_project`
--

DROP TABLE IF EXISTS `sp_project`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `sp_project` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `identity` varchar(255) NOT NULL,
  `name` varchar(255) DEFAULT NULL,
  `description` varchar(255) DEFAULT NULL,
  `ats` varchar(255) NOT NULL,
  `callback` varchar(255) DEFAULT NULL,
  `project_id` int(11) NOT NULL DEFAULT '0',
  `create_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `update_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '修改时间',
  `is_deleted` varchar(1) NOT NULL DEFAULT 'N' COMMENT '逻辑删除',
  PRIMARY KEY (`id`),
  KEY `identity` (`identity`)
) ENGINE=InnoDB AUTO_INCREMENT=11 DEFAULT CHARSET=utf8 BLOCK_FORMAT=ENCRYPTED COMMENT='主动监控存储项目映射的表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `sp_report_history`
--

DROP TABLE IF EXISTS `sp_report_history`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `sp_report_history` (
  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `scope` varchar(255) DEFAULT NULL COMMENT 'scope',
  `scope_id` varchar(255) DEFAULT NULL COMMENT 'scope id',
  `task_id` bigint(20) unsigned NOT NULL COMMENT 'scope id',
  `dashboard_id` varchar(255) NOT NULL COMMENT 'dashboard id',
  `created_at` datetime DEFAULT NULL COMMENT '创建时间',
  `start` bigint(20) DEFAULT NULL COMMENT '报表开始统计时间',
  `end` bigint(20) NOT NULL COMMENT '报表截至统计时间',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=2113 DEFAULT CHARSET=utf8 COMMENT='报表记录表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `sp_report_settings`
--

DROP TABLE IF EXISTS `sp_report_settings`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `sp_report_settings` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `project_id` int(11) NOT NULL,
  `project_name` varchar(64) NOT NULL,
  `workspace` varchar(16) NOT NULL,
  `created` datetime NOT NULL,
  `weekly_report_enable` tinyint(1) NOT NULL,
  `daily_report_enable` tinyint(1) NOT NULL,
  `weekly_report_config` text NOT NULL,
  `daily_report_config` text NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `project_workspace` (`project_id`,`workspace`)
) ENGINE=InnoDB AUTO_INCREMENT=7 DEFAULT CHARSET=utf8mb4 BLOCK_FORMAT=ENCRYPTED COMMENT='应用监控项目报告配置表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `sp_report_task`
--

DROP TABLE IF EXISTS `sp_report_task`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `sp_report_task` (
  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `name` varchar(32) NOT NULL COMMENT '名称',
  `scope` varchar(255) DEFAULT NULL COMMENT 'scope',
  `scope_id` varchar(255) DEFAULT NULL COMMENT 'scope id',
  `type` varchar(255) NOT NULL COMMENT '报表类型',
  `dashboard_id` varchar(255) NOT NULL COMMENT 'dashboard id',
  `enable` tinyint(1) DEFAULT NULL COMMENT '通知开关',
  `notify_target` varchar(1024) NOT NULL COMMENT '通知组',
  `pipeline_cron_id` bigint(20) NOT NULL COMMENT 'pipelineCron id',
  `created_at` datetime DEFAULT NULL COMMENT '创建时间',
  `updated_at` datetime DEFAULT NULL COMMENT '修改时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `Name` (`name`,`scope`,`scope_id`),
  UNIQUE KEY `Scope` (`name`,`scope`,`scope_id`),
  UNIQUE KEY `ScopeID` (`name`,`scope`,`scope_id`)
) ENGINE=InnoDB AUTO_INCREMENT=17 DEFAULT CHARSET=utf8 COMMENT='报表任务表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `sp_reports`
--

DROP TABLE IF EXISTS `sp_reports`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `sp_reports` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `key` varchar(32) NOT NULL,
  `start` datetime NOT NULL,
  `end` datetime NOT NULL,
  `project_id` int(11) NOT NULL,
  `project_name` varchar(64) NOT NULL,
  `workspace` varchar(16) NOT NULL,
  `created` datetime NOT NULL,
  `version` varchar(16) NOT NULL,
  `data` text NOT NULL,
  `type` tinyint(4) NOT NULL,
  `terminus_key` varchar(64) NOT NULL DEFAULT '',
  PRIMARY KEY (`id`),
  UNIQUE KEY `key_unique` (`key`),
  KEY `project_id_workspace` (`project_id`,`workspace`,`type`)
) ENGINE=InnoDB AUTO_INCREMENT=434 DEFAULT CHARSET=utf8mb4 BLOCK_FORMAT=ENCRYPTED COMMENT='应用监控项目报告历史表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `sp_service`
--

DROP TABLE IF EXISTS `sp_service`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `sp_service` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `name` varchar(255) DEFAULT NULL,
  `project_id` int(11) unsigned zerofill NOT NULL,
  `create_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `update_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '修改时间',
  `is_deleted` varchar(1) NOT NULL DEFAULT 'N' COMMENT '逻辑删除',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=4 DEFAULT CHARSET=utf8 BLOCK_FORMAT=ENCRYPTED COMMENT='主动监控存储服务项的表,已废弃';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `sp_stage`
--

DROP TABLE IF EXISTS `sp_stage`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `sp_stage` (
  `id` int(11) NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `name` varchar(255) DEFAULT NULL COMMENT '名称',
  `color` varchar(255) DEFAULT NULL COMMENT '颜色',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uniqName` (`name`)
) ENGINE=InnoDB AUTO_INCREMENT=5 DEFAULT CHARSET=utf8 BLOCK_FORMAT=ENCRYPTED COMMENT='告警状态颜色枚举表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `sp_status`
--

DROP TABLE IF EXISTS `sp_status`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `sp_status` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `name` varchar(255) DEFAULT NULL,
  `color` varchar(255) DEFAULT NULL,
  `level` int(10) unsigned zerofill NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uniqName` (`name`)
) ENGINE=InnoDB AUTO_INCREMENT=5 DEFAULT CHARSET=utf8 BLOCK_FORMAT=ENCRYPTED COMMENT='主动监控状态枚举表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `sp_trace_request_history`
--

DROP TABLE IF EXISTS `sp_trace_request_history`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `sp_trace_request_history` (
  `request_id` varchar(128) NOT NULL,
  `terminus_key` varchar(55) NOT NULL DEFAULT '',
  `url` varchar(1024) NOT NULL,
  `query_string` text,
  `header` text,
  `body` text,
  `method` varchar(16) NOT NULL,
  `status` int(2) NOT NULL DEFAULT '0',
  `response_status` int(11) NOT NULL DEFAULT '200',
  `response_body` text,
  `create_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `update_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`request_id`),
  KEY `INDEX_TERMINUS_KEY` (`terminus_key`),
  KEY `INDEX_STATUS` (`status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 BLOCK_FORMAT=ENCRYPTED COMMENT='链路诊断历史表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `sp_user`
--

DROP TABLE IF EXISTS `sp_user`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `sp_user` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `username` varchar(255) DEFAULT NULL,
  `salt` varchar(255) DEFAULT NULL,
  `password` varchar(255) DEFAULT NULL,
  `created_at` timestamp NULL DEFAULT NULL ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uniqName` (`username`)
) ENGINE=InnoDB AUTO_INCREMENT=6 DEFAULT CHARSET=utf8 BLOCK_FORMAT=ENCRYPTED COMMENT='主动监控存储用户的表,已废弃';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `tb_addon_attachment`
--

DROP TABLE IF EXISTS `tb_addon_attachment`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `tb_addon_attachment` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT,
  `app_id` varchar(45) DEFAULT NULL COMMENT 'appID',
  `instance_id` varchar(64) DEFAULT NULL,
  `create_time` datetime NOT NULL DEFAULT '1970-01-01 00:00:00' COMMENT '创建时间',
  `update_time` datetime NOT NULL DEFAULT '1970-01-01 00:00:00' COMMENT '修改时间',
  `is_deleted` varchar(1) NOT NULL DEFAULT 'N' COMMENT '逻辑删除',
  `options` varchar(1024) DEFAULT NULL COMMENT '可选字段',
  `org_id` varchar(64) DEFAULT NULL COMMENT '组织id',
  `project_id` varchar(64) DEFAULT NULL COMMENT '项目id',
  `application_id` varchar(64) DEFAULT NULL COMMENT '应用id',
  `routing_instance_id` varchar(64) DEFAULT NULL COMMENT '路由表addon实例ID',
  `runtime_name` varchar(64) DEFAULT '' COMMENT 'runtime名称',
  `inside_addon` varchar(1) DEFAULT 'N' COMMENT '是否为内部依赖addon，N:否，Y:是',
  `tenant_instance_id` varchar(64) DEFAULT NULL COMMENT 'addon tenant ID',
  PRIMARY KEY (`id`),
  KEY `idx_app_id` (`app_id`,`is_deleted`),
  KEY `idx_application_id` (`application_id`,`is_deleted`),
  KEY `idx_instance_id` (`instance_id`,`is_deleted`)
) ENGINE=InnoDB AUTO_INCREMENT=4126 DEFAULT CHARSET=utf8 BLOCK_FORMAT=ENCRYPTED COMMENT='addon attach信息';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `tb_addon_audit`
--

DROP TABLE IF EXISTS `tb_addon_audit`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `tb_addon_audit` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT COMMENT '自增id',
  `org_id` varchar(16) NOT NULL COMMENT '企业ID',
  `project_id` varchar(16) DEFAULT NULL COMMENT '项目ID',
  `workspace` varchar(16) NOT NULL COMMENT '环境',
  `operator` varchar(255) NOT NULL COMMENT '操作人',
  `op_name` varchar(64) NOT NULL COMMENT '操作类型',
  `addon_name` varchar(128) NOT NULL COMMENT 'addon名称',
  `ins_id` varchar(64) NOT NULL COMMENT 'addon实例ID',
  `ins_name` varchar(128) NOT NULL COMMENT 'addon实例名称',
  `params` varchar(4096) DEFAULT NULL COMMENT '修改参数',
  `is_deleted` varchar(1) DEFAULT 'N' COMMENT '逻辑删除',
  `create_time` datetime DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `update_time` datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '删除时间',
  PRIMARY KEY (`id`),
  KEY `idx_projectid_insname` (`project_id`,`ins_name`)
) ENGINE=InnoDB AUTO_INCREMENT=185 DEFAULT CHARSET=utf8mb4 BLOCK_FORMAT=ENCRYPTED COMMENT='addon操作审计信息';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `tb_addon_extra`
--

DROP TABLE IF EXISTS `tb_addon_extra`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `tb_addon_extra` (
  `id` varchar(64) NOT NULL DEFAULT '' COMMENT '数据库唯一Id',
  `addon_id` varchar(64) NOT NULL DEFAULT '' COMMENT 'addon id',
  `field` varchar(64) NOT NULL DEFAULT '' COMMENT '字段名称',
  `value` text NOT NULL COMMENT '字段value',
  `create_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `update_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '最后更新时间',
  `is_deleted` varchar(1) NOT NULL DEFAULT 'N' COMMENT '逻辑删除',
  `addon_name` varchar(64) DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_addon_field` (`addon_id`,`field`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8 BLOCK_FORMAT=ENCRYPTED COMMENT='addon额外信息';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `tb_addon_instance`
--

DROP TABLE IF EXISTS `tb_addon_instance`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `tb_addon_instance` (
  `id` varchar(45) NOT NULL COMMENT 'addon实例ID',
  `name` varchar(128) NOT NULL COMMENT 'addon实例名称',
  `addon_id` varchar(64) DEFAULT '' COMMENT 'addon编号',
  `plan` varchar(128) DEFAULT NULL COMMENT 'addon规格',
  `version` varchar(128) DEFAULT NULL COMMENT 'addon版本',
  `app_id` varchar(45) DEFAULT NULL COMMENT 'appID',
  `project_id` varchar(45) DEFAULT NULL COMMENT '项目ID',
  `org_id` varchar(45) DEFAULT NULL COMMENT '组织ID',
  `share_scope` varchar(45) DEFAULT NULL COMMENT '共享范围',
  `env` varchar(45) NOT NULL COMMENT '所属部署环境',
  `options` varchar(1024) DEFAULT NULL COMMENT '请求参数中可选字段',
  `status` varchar(45) NOT NULL COMMENT 'instance状态',
  `config` varchar(4096) DEFAULT NULL COMMENT '需要使用的config',
  `az` varchar(45) DEFAULT NULL COMMENT '所属集群',
  `admin_create` tinyint(1) NOT NULL DEFAULT '0' COMMENT '是否后台创建',
  `category` varchar(32) NOT NULL COMMENT '分类名称',
  `create_time` datetime NOT NULL DEFAULT '1970-01-01 00:00:00' COMMENT '创建时间',
  `update_time` datetime NOT NULL DEFAULT '1970-01-01 00:00:00' COMMENT '修改时间',
  `is_deleted` varchar(1) NOT NULL DEFAULT 'N' COMMENT '逻辑删除',
  `is_migrate` varchar(1) DEFAULT 'N' COMMENT '是否迁移数据',
  `attach_count` int(11) DEFAULT '0' COMMENT '被引用数',
  `application_id` varchar(64) DEFAULT NULL COMMENT '应用id',
  `is_platform` tinyint(1) DEFAULT '0' COMMENT '是否为平台Addon实例',
  `is_default` int(1) DEFAULT '0' COMMENT '是否默认addon创建',
  `platform_service_type` int(1) NOT NULL DEFAULT '0' COMMENT '平台服务类型，0：非平台服务，1：微服务，2：平台组件',
  `addon_name` varchar(64) NOT NULL DEFAULT '' COMMENT 'addon名称，平台内唯一标识',
  `namespace` varchar(64) DEFAULT '' COMMENT 'scheduler namespace',
  `schedule_name` varchar(255) DEFAULT '' COMMENT 'scheduler name',
  `label` varchar(4096) DEFAULT NULL COMMENT '需要使用的label',
  `kms_key` varchar(64) DEFAULT NULL COMMENT 'kms key id',
  `cpu_request` double DEFAULT NULL,
  `cpu_limit` double DEFAULT NULL,
  `mem_request` int(11) DEFAULT NULL,
  `mem_limit` int(11) DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_appid_name` (`app_id`,`name`,`addon_id`,`az`),
  KEY `idx_org_status` (`org_id`,`status`,`share_scope`,`is_deleted`),
  KEY `idx_project_status` (`project_id`,`status`,`share_scope`,`is_deleted`),
  KEY `idx_project_addon` (`project_id`,`status`,`addon_id`,`is_deleted`),
  KEY `idx_appid` (`app_id`,`status`,`env`,`share_scope`,`is_deleted`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8 BLOCK_FORMAT=ENCRYPTED COMMENT='addon实例信息';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `tb_addon_instance_routing`
--

DROP TABLE IF EXISTS `tb_addon_instance_routing`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `tb_addon_instance_routing` (
  `id` varchar(64) NOT NULL COMMENT 'addon实例ID',
  `name` varchar(128) NOT NULL COMMENT 'addon实例名称',
  `addon_id` varchar(64) DEFAULT '' COMMENT 'addon编号',
  `plan` varchar(128) DEFAULT NULL COMMENT 'addon规格',
  `version` varchar(128) DEFAULT NULL COMMENT 'addon版本',
  `app_id` varchar(45) DEFAULT NULL COMMENT 'appID',
  `application_id` varchar(64) DEFAULT NULL COMMENT '应用id',
  `project_id` varchar(45) DEFAULT NULL COMMENT '项目ID',
  `org_id` varchar(45) DEFAULT NULL COMMENT '组织ID',
  `share_scope` varchar(45) DEFAULT NULL COMMENT '共享范围',
  `env` varchar(45) NOT NULL COMMENT '所属部署环境',
  `options` varchar(1024) DEFAULT NULL COMMENT '请求参数中可选字段',
  `status` varchar(45) NOT NULL COMMENT 'instance状态',
  `az` varchar(45) DEFAULT NULL COMMENT '所属集群',
  `category` varchar(32) NOT NULL COMMENT '分类名称',
  `is_migrate` varchar(1) DEFAULT 'N' COMMENT '是否迁移数据',
  `attach_count` int(11) DEFAULT '0' COMMENT '被引用数',
  `is_platform` tinyint(1) DEFAULT '0' COMMENT '是否为平台Addon实例',
  `real_instance` varchar(64) NOT NULL COMMENT '真实insId',
  `create_time` datetime NOT NULL DEFAULT '1970-01-01 00:00:00' COMMENT '创建时间',
  `update_time` datetime NOT NULL DEFAULT '1970-01-01 00:00:00' COMMENT '修改时间',
  `is_deleted` varchar(1) NOT NULL DEFAULT 'N' COMMENT '逻辑删除',
  `platform_service_type` int(1) NOT NULL DEFAULT '0' COMMENT '平台服务类型，0：非平台服务，1：微服务，2：平台组件',
  `addon_name` varchar(64) NOT NULL DEFAULT '' COMMENT 'addon名称，平台内唯一标识',
  `inside_addon` varchar(1) DEFAULT 'N' COMMENT '是否为内部依赖addon，N:否，Y:是',
  `tag` varchar(64) DEFAULT '' COMMENT '实例标签',
  PRIMARY KEY (`id`),
  KEY `idx_appid_name` (`app_id`,`name`,`addon_id`,`az`),
  KEY `idx_org_status` (`org_id`,`status`,`share_scope`,`is_deleted`),
  KEY `idx_addon_id_and_env` (`addon_id`,`env`,`project_id`,`status`),
  KEY `idx_project_status` (`project_id`,`status`,`share_scope`,`is_deleted`),
  KEY `idx_project_addon` (`project_id`,`status`,`addon_id`,`is_deleted`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8 BLOCK_FORMAT=ENCRYPTED COMMENT='addon实例信息';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `tb_addon_instance_tenant`
--

DROP TABLE IF EXISTS `tb_addon_instance_tenant`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `tb_addon_instance_tenant` (
  `id` varchar(45) NOT NULL,
  `name` varchar(128) NOT NULL COMMENT 'addon租户名称',
  `addon_instance_id` varchar(64) NOT NULL DEFAULT '' COMMENT '对应addon实例id',
  `addon_instance_routing_id` varchar(64) NOT NULL DEFAULT '' COMMENT '对应addonrouting id',
  `app_id` varchar(45) DEFAULT NULL COMMENT 'appID',
  `project_id` varchar(45) DEFAULT NULL COMMENT '项目ID',
  `org_id` varchar(45) DEFAULT NULL COMMENT 'orgID',
  `workspace` varchar(45) NOT NULL COMMENT '所属部署环境',
  `config` varchar(4096) DEFAULT NULL COMMENT '需要使用的config',
  `create_time` datetime NOT NULL DEFAULT '1970-01-01 00:00:00' COMMENT '创建时间',
  `update_time` datetime NOT NULL DEFAULT '1970-01-01 00:00:00' COMMENT '修改时间',
  `is_deleted` varchar(1) NOT NULL DEFAULT 'N' COMMENT '逻辑删除',
  `kms_key` varchar(64) DEFAULT NULL COMMENT 'kms key id',
  `reference` int(11) DEFAULT '0' COMMENT '被引用数',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8 COMMENT='addon租户';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `tb_addon_management`
--

DROP TABLE IF EXISTS `tb_addon_management`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `tb_addon_management` (
  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT,
  `addon_id` varchar(64) NOT NULL COMMENT 'addon实例ID',
  `name` varchar(128) NOT NULL COMMENT 'addon实例名称',
  `project_id` varchar(45) DEFAULT NULL COMMENT '项目ID',
  `org_id` varchar(45) DEFAULT NULL COMMENT '组织ID',
  `addon_config` text COMMENT 'addon参数配置',
  `cpu` double(8,2) NOT NULL COMMENT 'cpu核数',
  `mem` int(11) NOT NULL COMMENT '内存大小（M）',
  `nodes` int(4) NOT NULL COMMENT '节点数',
  `create_time` datetime NOT NULL DEFAULT '1970-01-01 00:00:00' COMMENT '创建时间',
  `update_time` datetime NOT NULL DEFAULT '1970-01-01 00:00:00' COMMENT '修改时间',
  PRIMARY KEY (`id`),
  KEY `idx_addon_id` (`addon_id`)
) ENGINE=InnoDB AUTO_INCREMENT=41 DEFAULT CHARSET=utf8 COMMENT='云addon信息(ops)';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `tb_addon_micro_attach`
--

DROP TABLE IF EXISTS `tb_addon_micro_attach`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `tb_addon_micro_attach` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT,
  `addon_name` varchar(64) NOT NULL DEFAULT '' COMMENT 'addon名称，平台内唯一标识',
  `routing_instance_id` varchar(64) DEFAULT NULL COMMENT '路由表addon实例ID',
  `instance_id` varchar(64) DEFAULT NULL,
  `project_id` varchar(64) DEFAULT NULL COMMENT '项目id',
  `env` varchar(16) DEFAULT NULL COMMENT '环境',
  `org_id` varchar(64) DEFAULT NULL COMMENT '组织id',
  `count` int(11) NOT NULL DEFAULT '1' COMMENT '引用数量',
  `create_time` datetime NOT NULL DEFAULT '1970-01-01 00:00:00' COMMENT '创建时间',
  `update_time` datetime NOT NULL DEFAULT '1970-01-01 00:00:00' COMMENT '修改时间',
  `is_deleted` varchar(1) NOT NULL DEFAULT 'N' COMMENT '逻辑删除',
  PRIMARY KEY (`id`),
  KEY `idx_addon_name` (`addon_name`,`is_deleted`),
  KEY `idx_routing_instance_id` (`routing_instance_id`,`is_deleted`),
  KEY `idx_project_id` (`project_id`,`is_deleted`)
) ENGINE=InnoDB AUTO_INCREMENT=29 DEFAULT CHARSET=utf8 BLOCK_FORMAT=ENCRYPTED COMMENT='microservice addon attach信息';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `tb_addon_prebuild`
--

DROP TABLE IF EXISTS `tb_addon_prebuild`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `tb_addon_prebuild` (
  `id` int(11) NOT NULL AUTO_INCREMENT COMMENT '数据库自增id',
  `application_id` varchar(32) NOT NULL COMMENT 'app id',
  `git_branch` varchar(128) NOT NULL COMMENT 'git分支',
  `env` varchar(10) NOT NULL COMMENT '环境',
  `runtime_id` varchar(32) DEFAULT NULL COMMENT 'runtimeId',
  `instance_id` varchar(64) DEFAULT NULL COMMENT 'addon实例id',
  `instance_name` varchar(128) DEFAULT NULL COMMENT 'addon实例名称',
  `addon_name` varchar(128) DEFAULT '' COMMENT 'addon名称',
  `addon_class` varchar(64) DEFAULT '' COMMENT '规格信息',
  `options` varchar(1024) DEFAULT '' COMMENT '额外信息',
  `config` varchar(1024) DEFAULT NULL COMMENT '环境变量信息',
  `build_from` int(1) NOT NULL DEFAULT '0' COMMENT '创建来源，0:dice.yml，1:重新分析',
  `delete_status` int(1) NOT NULL DEFAULT '0' COMMENT '删除状态，0:未删除，1:diceyml删除，2:重新分析删除',
  `create_time` datetime NOT NULL DEFAULT '1970-01-01 00:00:00' COMMENT '创建时间',
  `update_time` datetime NOT NULL DEFAULT '1970-01-01 00:00:00' COMMENT '修改时间',
  `is_deleted` varchar(1) NOT NULL DEFAULT 'N' COMMENT '逻辑删除',
  `routing_instance_id` varchar(64) DEFAULT NULL COMMENT '路由表addon实例ID',
  `use_type` varchar(32) DEFAULT 'NORMAL' COMMENT 'addon使用类型，NORMAL(添加使用), DEFAULT(默认使用)',
  PRIMARY KEY (`id`),
  KEY `idx_app_branch_env` (`application_id`,`git_branch`,`env`),
  KEY `idx_instance_id` (`instance_id`),
  KEY `idx_runtime_id` (`runtime_id`)
) ENGINE=InnoDB AUTO_INCREMENT=4433 DEFAULT CHARSET=utf8 BLOCK_FORMAT=ENCRYPTED COMMENT='addon创建流程记录信息';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `tb_gateway_api`
--

DROP TABLE IF EXISTS `tb_gateway_api`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `tb_gateway_api` (
  `id` varchar(32) NOT NULL DEFAULT '' COMMENT '唯一主键',
  `zone_id` varchar(32) DEFAULT '' COMMENT '所属的zone',
  `consumer_id` varchar(32) NOT NULL DEFAULT '' COMMENT '消费者id',
  `api_path` varchar(256) NOT NULL DEFAULT '' COMMENT 'api路径',
  `method` varchar(128) NOT NULL DEFAULT '' COMMENT '方法',
  `redirect_addr` varchar(256) NOT NULL DEFAULT '' COMMENT '转发地址',
  `description` varchar(256) DEFAULT NULL COMMENT '描述',
  `group_id` varchar(32) NOT NULL DEFAULT '' COMMENT '服务Id',
  `policies` varchar(1024) DEFAULT NULL COMMENT '策略配置',
  `upstream_api_id` varchar(32) NOT NULL DEFAULT '' COMMENT '对应的后端api',
  `dice_app` varchar(128) DEFAULT '' COMMENT 'dice应用名',
  `dice_service` varchar(128) DEFAULT '' COMMENT 'dice服务名',
  `register_type` varchar(16) NOT NULL DEFAULT 'auto' COMMENT '注册类型',
  `net_type` varchar(16) NOT NULL DEFAULT 'outer' COMMENT '网络类型',
  `need_auth` tinyint(1) NOT NULL DEFAULT '0' COMMENT '需要鉴权标识',
  `create_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `update_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  `is_deleted` varchar(1) NOT NULL DEFAULT 'N' COMMENT '逻辑删除',
  `redirect_type` varchar(32) NOT NULL DEFAULT 'url' COMMENT '转发类型',
  `runtime_service_id` varchar(32) NOT NULL DEFAULT '' COMMENT '关联的service的id',
  `swagger` blob COMMENT 'swagger文档',
  PRIMARY KEY (`id`),
  KEY `idx_service_id` (`runtime_service_id`,`is_deleted`),
  KEY `idx_consumer_id` (`consumer_id`,`is_deleted`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8 BLOCK_FORMAT=ENCRYPTED COMMENT='微服务 API';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `tb_gateway_api_in_package`
--

DROP TABLE IF EXISTS `tb_gateway_api_in_package`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `tb_gateway_api_in_package` (
  `id` varchar(32) NOT NULL DEFAULT '' COMMENT '唯一id',
  `dice_api_id` varchar(32) NOT NULL DEFAULT '' COMMENT 'dice服务api的id',
  `package_id` varchar(32) NOT NULL DEFAULT '' COMMENT '产品包id',
  `zone_id` varchar(32) DEFAULT NULL COMMENT '所属的zone',
  `create_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `update_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  `is_deleted` varchar(1) NOT NULL DEFAULT 'N' COMMENT '逻辑删除',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8 BLOCK_FORMAT=ENCRYPTED COMMENT='被流量入口引用的微服务 API';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `tb_gateway_az_info`
--

DROP TABLE IF EXISTS `tb_gateway_az_info`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `tb_gateway_az_info` (
  `id` varchar(32) NOT NULL DEFAULT '' COMMENT '唯一id',
  `create_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `update_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  `is_deleted` varchar(1) NOT NULL DEFAULT 'N' COMMENT '逻辑删除',
  `org_id` varchar(32) NOT NULL COMMENT '企业标识id',
  `project_id` varchar(32) NOT NULL COMMENT '项目标识id',
  `env` varchar(32) NOT NULL COMMENT '应用所属环境',
  `az` varchar(32) NOT NULL COMMENT '集群名',
  `type` varchar(16) NOT NULL DEFAULT '' COMMENT '集群类型',
  `wildcard_domain` varchar(1024) NOT NULL DEFAULT '' COMMENT '集群泛域名',
  `master_addr` varchar(1024) NOT NULL DEFAULT '' COMMENT '集群管控地址',
  `need_update` tinyint(1) DEFAULT '1' COMMENT '待更新标识',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8 BLOCK_FORMAT=ENCRYPTED COMMENT='API 网关集群信息';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `tb_gateway_consumer`
--

DROP TABLE IF EXISTS `tb_gateway_consumer`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `tb_gateway_consumer` (
  `id` varchar(32) NOT NULL DEFAULT '' COMMENT '唯一id',
  `consumer_id` varchar(128) NOT NULL DEFAULT '' COMMENT '消费者id',
  `consumer_name` varchar(128) NOT NULL DEFAULT '' COMMENT '消费者名称',
  `config` varchar(1024) DEFAULT NULL COMMENT '配置信息，存放key等',
  `endpoint` varchar(256) NOT NULL DEFAULT '' COMMENT '终端',
  `create_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `update_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  `is_deleted` varchar(1) NOT NULL DEFAULT 'N' COMMENT '逻辑删除',
  `org_id` varchar(32) NOT NULL COMMENT '企业id',
  `project_id` varchar(32) NOT NULL COMMENT '项目id',
  `env` varchar(32) NOT NULL COMMENT '环境',
  `az` varchar(32) NOT NULL COMMENT '集群名',
  `auth_config` blob COMMENT '鉴权配置',
  `description` varchar(256) DEFAULT NULL COMMENT '备注',
  `type` varchar(16) NOT NULL DEFAULT 'project' COMMENT '调用方类型',
  `cloudapi_app_id` varchar(128) NOT NULL DEFAULT '' COMMENT '阿里云APP id',
  `client_id` varchar(32) NOT NULL DEFAULT '' COMMENT '对应的客户端id',
  PRIMARY KEY (`id`),
  UNIQUE KEY `consumer_id` (`consumer_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8 BLOCK_FORMAT=ENCRYPTED COMMENT='API 网关调用方';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `tb_gateway_consumer_api`
--

DROP TABLE IF EXISTS `tb_gateway_consumer_api`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `tb_gateway_consumer_api` (
  `id` varchar(32) NOT NULL DEFAULT '' COMMENT '唯一id',
  `consumer_id` varchar(32) NOT NULL DEFAULT '' COMMENT '消费者id',
  `api_id` varchar(32) NOT NULL DEFAULT '' COMMENT 'apiId',
  `policies` varchar(512) DEFAULT NULL COMMENT '策略信息',
  `create_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `update_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  `is_deleted` varchar(1) NOT NULL DEFAULT 'N' COMMENT '逻辑删除',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8 BLOCK_FORMAT=ENCRYPTED COMMENT='API 的调用方授权信息(已废弃)';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `tb_gateway_default_policy`
--

DROP TABLE IF EXISTS `tb_gateway_default_policy`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `tb_gateway_default_policy` (
  `id` varchar(32) NOT NULL DEFAULT '' COMMENT '唯一id',
  `create_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `update_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  `is_deleted` varchar(1) NOT NULL DEFAULT 'N' COMMENT '逻辑删除',
  `name` varchar(32) NOT NULL DEFAULT '' COMMENT '名称',
  `level` varchar(32) NOT NULL COMMENT '策略级别',
  `tenant_id` varchar(128) NOT NULL DEFAULT '' COMMENT '租户id',
  `dice_app` varchar(128) DEFAULT '' COMMENT 'dice应用名',
  `config` blob COMMENT '具体配置',
  `package_id` varchar(32) NOT NULL DEFAULT '' COMMENT '流量入口id',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8 BLOCK_FORMAT=ENCRYPTED COMMENT='API 默认策略';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `tb_gateway_domain`
--

DROP TABLE IF EXISTS `tb_gateway_domain`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `tb_gateway_domain` (
  `id` varchar(32) NOT NULL DEFAULT '' COMMENT '唯一主键',
  `create_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `update_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  `is_deleted` varchar(1) NOT NULL DEFAULT 'N' COMMENT '逻辑删除',
  `domain` varchar(255) NOT NULL COMMENT '域名',
  `cluster_name` varchar(32) NOT NULL DEFAULT '' COMMENT '所属集群',
  `type` varchar(32) NOT NULL COMMENT '域名类型',
  `runtime_service_id` varchar(32) DEFAULT NULL COMMENT '所属服务id',
  `package_id` varchar(32) DEFAULT NULL COMMENT '所属流量入口id',
  `component_name` varchar(32) DEFAULT NULL COMMENT '所属平台组件的名称',
  `ingress_name` varchar(128) DEFAULT NULL COMMENT '所属平台组件的ingress的名称',
  `project_id` varchar(32) NOT NULL DEFAULT '' COMMENT '项目标识id',
  `project_name` varchar(50) NOT NULL DEFAULT '' COMMENT '项目名称',
  `workspace` varchar(32) NOT NULL DEFAULT '' COMMENT '所属环境',
  PRIMARY KEY (`id`),
  KEY `idx_runtime_service` (`runtime_service_id`,`is_deleted`),
  KEY `idx_package` (`package_id`,`is_deleted`),
  KEY `idx_cluster_domain` (`cluster_name`,`domain`,`is_deleted`),
  KEY `idx_cluster` (`is_deleted`,`cluster_name`,`domain`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8 COMMENT='域名管理';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `tb_gateway_ingress_policy`
--

DROP TABLE IF EXISTS `tb_gateway_ingress_policy`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `tb_gateway_ingress_policy` (
  `id` varchar(32) NOT NULL DEFAULT '' COMMENT '唯一id',
  `create_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `update_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  `is_deleted` varchar(1) NOT NULL DEFAULT 'N' COMMENT '逻辑删除',
  `name` varchar(32) NOT NULL DEFAULT '' COMMENT '名称',
  `regions` varchar(128) NOT NULL DEFAULT '' COMMENT '作用域',
  `az` varchar(32) NOT NULL COMMENT '集群名',
  `zone_id` varchar(32) DEFAULT NULL COMMENT '所属的zone',
  `config` blob COMMENT '具体配置',
  `configmap_option` blob COMMENT 'ingress configmap option',
  `main_snippet` blob COMMENT 'ingress configmap main 配置',
  `http_snippet` blob COMMENT 'ingress configmap http 配置',
  `server_snippet` blob COMMENT 'ingress configmap server 配置',
  `annotations` blob COMMENT '包含的annotations',
  `location_snippet` blob COMMENT 'nginx location 配置',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8 BLOCK_FORMAT=ENCRYPTED COMMENT='Ingress 策略管理';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `tb_gateway_kong_info`
--

DROP TABLE IF EXISTS `tb_gateway_kong_info`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `tb_gateway_kong_info` (
  `id` varchar(32) NOT NULL DEFAULT '' COMMENT '唯一id',
  `create_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `update_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  `is_deleted` varchar(1) NOT NULL DEFAULT 'N' COMMENT '逻辑删除',
  `az` varchar(32) NOT NULL COMMENT '集群名',
  `project_id` varchar(32) NOT NULL COMMENT '项目id',
  `project_name` varchar(256) NOT NULL COMMENT '项目名',
  `env` varchar(32) DEFAULT '' COMMENT '环境名',
  `kong_addr` varchar(256) NOT NULL COMMENT 'kong admin地址',
  `endpoint` varchar(256) NOT NULL COMMENT 'kong gateway地址',
  `inner_addr` varchar(1024) NOT NULL DEFAULT '' COMMENT 'kong内网地址',
  `service_name` varchar(32) NOT NULL DEFAULT '' COMMENT 'kong的服务名称',
  `addon_instance_id` varchar(64) NOT NULL DEFAULT '' COMMENT 'addon id',
  `need_update` tinyint(1) DEFAULT '1' COMMENT '待更新标识',
  `tenant_id` varchar(128) NOT NULL DEFAULT '' COMMENT '租户id',
  `tenant_group` varchar(128) NOT NULL DEFAULT '' COMMENT '租户分组',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8 BLOCK_FORMAT=ENCRYPTED COMMENT='Kong 实例信息';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `tb_gateway_org_client`
--

DROP TABLE IF EXISTS `tb_gateway_org_client`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `tb_gateway_org_client` (
  `id` varchar(32) NOT NULL DEFAULT '' COMMENT '唯一id',
  `org_id` varchar(32) NOT NULL COMMENT '企业id',
  `name` varchar(128) NOT NULL DEFAULT '' COMMENT '消费者名称',
  `client_secret` varchar(32) NOT NULL COMMENT '客户端凭证',
  `create_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `update_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  `is_deleted` varchar(1) NOT NULL DEFAULT 'N' COMMENT '逻辑删除',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8 COMMENT='企业级 API 网关客户端';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `tb_gateway_package`
--

DROP TABLE IF EXISTS `tb_gateway_package`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `tb_gateway_package` (
  `id` varchar(32) NOT NULL DEFAULT '' COMMENT '唯一主键',
  `dice_org_id` varchar(32) NOT NULL DEFAULT '' COMMENT 'dice企业标识id',
  `dice_project_id` varchar(32) DEFAULT '' COMMENT 'dice项目标识id',
  `dice_env` varchar(32) NOT NULL COMMENT 'dice环境',
  `dice_cluster_name` varchar(32) NOT NULL COMMENT 'dice集群名',
  `zone_id` varchar(32) DEFAULT NULL COMMENT '所属的zone',
  `package_name` varchar(1024) NOT NULL COMMENT '产品包名称',
  `bind_domain` varchar(1024) DEFAULT NULL COMMENT '绑定的域名',
  `description` varchar(256) DEFAULT NULL COMMENT '描述',
  `acl_type` varchar(16) NOT NULL DEFAULT 'off' COMMENT '授权方式',
  `auth_type` varchar(16) NOT NULL DEFAULT '' COMMENT '鉴权方式',
  `create_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `update_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  `is_deleted` varchar(1) NOT NULL DEFAULT 'N' COMMENT '逻辑删除',
  `scene` varchar(32) NOT NULL DEFAULT 'openapi' COMMENT '场景',
  `runtime_service_id` varchar(32) NOT NULL DEFAULT '' COMMENT '关联的service的id',
  `cloudapi_instance_id` varchar(128) NOT NULL DEFAULT '' COMMENT '阿里云API网关的实例id',
  `cloudapi_group_id` varchar(128) NOT NULL DEFAULT '' COMMENT '阿里云API网关的分组id',
  `cloudapi_domain` varchar(1024) NOT NULL DEFAULT '' COMMENT '阿里云API网关上的分组二级域名',
  `cloudapi_vpc_grant` varchar(128) NOT NULL DEFAULT '' COMMENT '阿里云API网关的VPC Grant',
  `cloudapi_need_bind` tinyint(1) NOT NULL DEFAULT '0' COMMENT '是否需要绑定阿里云API网关',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8 BLOCK_FORMAT=ENCRYPTED COMMENT='流量入口';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `tb_gateway_package_api`
--

DROP TABLE IF EXISTS `tb_gateway_package_api`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `tb_gateway_package_api` (
  `id` varchar(32) NOT NULL DEFAULT '' COMMENT '唯一主键',
  `package_id` varchar(32) DEFAULT '' COMMENT '所属的产品包id',
  `api_path` varchar(256) NOT NULL DEFAULT '' COMMENT 'api路径',
  `method` varchar(128) NOT NULL DEFAULT '' COMMENT '方法',
  `redirect_addr` varchar(256) NOT NULL DEFAULT '' COMMENT '转发地址',
  `description` varchar(256) DEFAULT NULL COMMENT '描述',
  `dice_app` varchar(128) DEFAULT '' COMMENT 'dice应用名',
  `dice_service` varchar(128) DEFAULT '' COMMENT 'dice服务名',
  `acl_type` varchar(16) DEFAULT NULL COMMENT '独立的授权类型',
  `origin` varchar(16) NOT NULL DEFAULT 'custom' COMMENT '来源',
  `dice_api_id` varchar(32) DEFAULT NULL COMMENT '对应dice服务api的id',
  `create_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `update_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  `is_deleted` varchar(1) NOT NULL DEFAULT 'N' COMMENT '逻辑删除',
  `zone_id` varchar(32) NOT NULL DEFAULT '' COMMENT '所属的zone',
  `redirect_type` varchar(32) NOT NULL DEFAULT 'url' COMMENT '转发类型',
  `runtime_service_id` varchar(32) NOT NULL DEFAULT '' COMMENT '关联的service的id',
  `redirect_path` varchar(256) NOT NULL DEFAULT '' COMMENT '转发路径',
  `cloudapi_api_id` varchar(128) NOT NULL DEFAULT '' COMMENT '阿里云API网关上的api id',
  PRIMARY KEY (`id`),
  KEY `idx_package_id` (`package_id`,`is_deleted`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8 BLOCK_FORMAT=ENCRYPTED COMMENT='流量入口下的路由规则';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `tb_gateway_package_api_in_consumer`
--

DROP TABLE IF EXISTS `tb_gateway_package_api_in_consumer`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `tb_gateway_package_api_in_consumer` (
  `id` varchar(32) NOT NULL DEFAULT '' COMMENT '唯一id',
  `consumer_id` varchar(32) NOT NULL DEFAULT '' COMMENT '消费者id',
  `package_id` varchar(32) NOT NULL DEFAULT '' COMMENT '产品包id',
  `package_api_id` varchar(32) NOT NULL DEFAULT '' COMMENT '产品包 api id',
  `create_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `update_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  `is_deleted` varchar(1) NOT NULL DEFAULT 'N' COMMENT '逻辑删除',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8 COMMENT='流量入口路由的授权信息';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `tb_gateway_package_in_consumer`
--

DROP TABLE IF EXISTS `tb_gateway_package_in_consumer`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `tb_gateway_package_in_consumer` (
  `id` varchar(32) NOT NULL DEFAULT '' COMMENT '唯一id',
  `consumer_id` varchar(32) NOT NULL DEFAULT '' COMMENT '消费者id',
  `package_id` varchar(32) NOT NULL DEFAULT '' COMMENT '产品包id',
  `create_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `update_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  `is_deleted` varchar(1) NOT NULL DEFAULT 'N' COMMENT '逻辑删除',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8 BLOCK_FORMAT=ENCRYPTED COMMENT='流量入口的授权信息';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `tb_gateway_package_rule`
--

DROP TABLE IF EXISTS `tb_gateway_package_rule`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `tb_gateway_package_rule` (
  `id` varchar(32) NOT NULL DEFAULT '' COMMENT '唯一id',
  `dice_org_id` varchar(32) NOT NULL DEFAULT '' COMMENT 'dice企业标识id',
  `dice_project_id` varchar(32) DEFAULT '' COMMENT 'dice项目标识id',
  `dice_env` varchar(32) NOT NULL COMMENT 'dice环境',
  `dice_cluster_name` varchar(32) NOT NULL COMMENT 'dice集群名',
  `category` varchar(32) NOT NULL DEFAULT '' COMMENT '插件类目',
  `enabled` tinyint(1) DEFAULT '1' COMMENT '插件开关',
  `plugin_id` varchar(128) NOT NULL DEFAULT '' COMMENT '插件id',
  `plugin_name` varchar(128) NOT NULL DEFAULT '' COMMENT '插件名称',
  `config` blob COMMENT '插件具体配置',
  `consumer_id` varchar(32) NOT NULL DEFAULT '' COMMENT '消费者id',
  `consumer_name` varchar(128) NOT NULL DEFAULT '' COMMENT '消费者名称',
  `package_id` varchar(32) NOT NULL DEFAULT '' COMMENT '产品包id',
  `package_name` varchar(1024) NOT NULL COMMENT '产品包名称',
  `api_id` varchar(32) DEFAULT NULL COMMENT '产品包api id',
  `create_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `update_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  `is_deleted` varchar(1) NOT NULL DEFAULT 'N' COMMENT '逻辑删除',
  `package_zone_need` tinyint(1) DEFAULT '1' COMMENT '是否在package的zone内生效',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8 BLOCK_FORMAT=ENCRYPTED COMMENT='流量入口的策略信息';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `tb_gateway_plugin_instance`
--

DROP TABLE IF EXISTS `tb_gateway_plugin_instance`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `tb_gateway_plugin_instance` (
  `id` varchar(32) NOT NULL DEFAULT '' COMMENT '唯一id',
  `plugin_id` varchar(128) NOT NULL DEFAULT '' COMMENT '插件id',
  `plugin_name` varchar(128) NOT NULL DEFAULT '' COMMENT '插件名称',
  `policy_id` varchar(32) NOT NULL DEFAULT '' COMMENT '策略id',
  `consumer_id` varchar(32) DEFAULT NULL COMMENT '消费者id',
  `group_id` varchar(32) DEFAULT NULL COMMENT '组id',
  `route_id` varchar(32) DEFAULT NULL COMMENT '路由id',
  `service_id` varchar(32) DEFAULT NULL COMMENT '服务id',
  `api_id` varchar(32) DEFAULT '' COMMENT 'apiID',
  `create_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `update_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  `is_deleted` varchar(1) NOT NULL DEFAULT 'N' COMMENT '逻辑删除',
  PRIMARY KEY (`id`),
  KEY `plugin_id` (`plugin_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8 BLOCK_FORMAT=ENCRYPTED COMMENT='Kong Plugin 实例信息';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `tb_gateway_policy`
--

DROP TABLE IF EXISTS `tb_gateway_policy`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `tb_gateway_policy` (
  `id` varchar(32) NOT NULL DEFAULT '' COMMENT '唯一id',
  `zone_id` varchar(32) NOT NULL DEFAULT '' COMMENT '所属的zone',
  `policy_name` varchar(128) DEFAULT '' COMMENT '策略名称',
  `display_name` varchar(128) NOT NULL DEFAULT '' COMMENT '策略展示名称',
  `category` varchar(128) NOT NULL DEFAULT '' COMMENT '策略类目',
  `description` varchar(128) NOT NULL DEFAULT '' COMMENT '描述类目',
  `plugin_id` varchar(128) DEFAULT '' COMMENT '插件id',
  `plugin_name` varchar(128) NOT NULL DEFAULT '' COMMENT '插件名称',
  `config` blob COMMENT 'plugin具体配置',
  `consumer_id` varchar(32) DEFAULT NULL COMMENT '消费者id',
  `create_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `update_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  `is_deleted` varchar(1) NOT NULL DEFAULT 'N' COMMENT '逻辑删除',
  `enabled` tinyint(1) DEFAULT '1' COMMENT '插件开关',
  `api_id` varchar(32) DEFAULT '' COMMENT 'api id',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8 BLOCK_FORMAT=ENCRYPTED COMMENT='Kong Plugin 元信息';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `tb_gateway_publish`
--

DROP TABLE IF EXISTS `tb_gateway_publish`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `tb_gateway_publish` (
  `id` varchar(32) NOT NULL DEFAULT '' COMMENT '唯一主键',
  `create_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `update_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  `is_deleted` varchar(1) NOT NULL DEFAULT 'N' COMMENT '逻辑删除',
  `dice_publish_id` varchar(32) NOT NULL DEFAULT '' COMMENT 'dice市场租户id',
  `dice_publish_item_id` varchar(32) NOT NULL DEFAULT '' COMMENT 'dice市场商品id',
  `dice_publish_item_name` varchar(128) NOT NULL DEFAULT '' COMMENT 'dice市场商品名称',
  `version` varchar(32) NOT NULL DEFAULT '' COMMENT '版本',
  `published` tinyint(1) DEFAULT '0' COMMENT '发布状态',
  `owner_email` varchar(1024) NOT NULL DEFAULT '' COMMENT '负责人邮箱地址',
  `api_register_id` varchar(32) NOT NULL DEFAULT '' COMMENT 'api register',
  `package_id` varchar(32) DEFAULT '' COMMENT '生成的流量入口id',
  PRIMARY KEY (`id`),
  KEY `idx_publish` (`dice_publish_id`,`dice_publish_item_id`,`is_deleted`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8 COMMENT='API 发布管理(已废弃)';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `tb_gateway_register`
--

DROP TABLE IF EXISTS `tb_gateway_register`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `tb_gateway_register` (
  `id` varchar(32) NOT NULL DEFAULT '' COMMENT '唯一主键',
  `create_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `update_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  `is_deleted` varchar(1) NOT NULL DEFAULT 'N' COMMENT '逻辑删除',
  `org_id` varchar(32) NOT NULL DEFAULT '' COMMENT '所属企业',
  `project_id` varchar(32) NOT NULL DEFAULT '' COMMENT '所属项目',
  `workspace` varchar(32) NOT NULL DEFAULT '' COMMENT '所属环境',
  `app_id` varchar(32) NOT NULL DEFAULT '' COMMENT '所属应用',
  `app_name` varchar(128) NOT NULL DEFAULT '' COMMENT '应用名称',
  `service_name` varchar(128) NOT NULL DEFAULT '' COMMENT '服务名称',
  `cluster_name` varchar(32) NOT NULL DEFAULT '' COMMENT '所属集群',
  `origin` varchar(16) NOT NULL DEFAULT 'action' COMMENT '注册来源',
  `runtime_name` varchar(128) NOT NULL DEFAULT '' COMMENT 'runtime名称/分支名称',
  `swagger` longblob COMMENT 'swagger文档',
  `md5_sum` varchar(128) NOT NULL DEFAULT '' COMMENT 'swagger摘要',
  `runtime_service_id` varchar(32) NOT NULL DEFAULT '' COMMENT 'runtime service',
  `registered` tinyint(1) DEFAULT '0' COMMENT '注册状态',
  `last_error` blob COMMENT '注册失败信息',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8 COMMENT='API 注册管理(已废弃)';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `tb_gateway_route`
--

DROP TABLE IF EXISTS `tb_gateway_route`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `tb_gateway_route` (
  `id` varchar(32) NOT NULL DEFAULT '' COMMENT '唯一id',
  `route_id` varchar(128) NOT NULL DEFAULT '' COMMENT '路由id',
  `protocols` varchar(128) DEFAULT NULL COMMENT '协议列表',
  `methods` varchar(128) DEFAULT NULL COMMENT '方法列表',
  `hosts` varchar(1024) DEFAULT NULL COMMENT '主机列表',
  `paths` varchar(1024) DEFAULT NULL COMMENT '路径列表',
  `service_id` varchar(128) NOT NULL DEFAULT '' COMMENT '绑定服务id',
  `config` varchar(1024) DEFAULT '' COMMENT '选填配置',
  `api_id` varchar(32) NOT NULL DEFAULT '' COMMENT 'apiid',
  `create_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `update_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  `is_deleted` varchar(1) NOT NULL DEFAULT 'N' COMMENT '逻辑删除',
  PRIMARY KEY (`id`),
  KEY `route_id` (`route_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8 BLOCK_FORMAT=ENCRYPTED COMMENT='Kong Route 配置信息';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `tb_gateway_runtime_service`
--

DROP TABLE IF EXISTS `tb_gateway_runtime_service`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `tb_gateway_runtime_service` (
  `id` varchar(32) NOT NULL DEFAULT '' COMMENT '唯一主键',
  `create_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `update_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  `is_deleted` varchar(1) NOT NULL DEFAULT 'N' COMMENT '逻辑删除',
  `project_id` varchar(32) NOT NULL DEFAULT '' COMMENT '所属项目',
  `workspace` varchar(32) NOT NULL DEFAULT '' COMMENT '所属环境',
  `cluster_name` varchar(32) NOT NULL DEFAULT '' COMMENT '所属集群',
  `runtime_id` varchar(32) NOT NULL DEFAULT '' COMMENT '所属runtime',
  `runtime_name` varchar(128) NOT NULL DEFAULT '' COMMENT 'runtime名称',
  `app_id` varchar(32) NOT NULL DEFAULT '' COMMENT '所属应用',
  `app_name` varchar(128) NOT NULL DEFAULT '' COMMENT '应用名称',
  `service_name` varchar(128) NOT NULL DEFAULT '' COMMENT '服务名称',
  `inner_address` varchar(1024) DEFAULT NULL COMMENT '服务内部地址',
  `use_apigw` tinyint(1) NOT NULL DEFAULT '0' COMMENT '是否使用api网关',
  `is_endpoint` tinyint(1) NOT NULL DEFAULT '0' COMMENT '是否是endpoint',
  `release_id` varchar(128) NOT NULL DEFAULT '' COMMENT '对应的releaseId',
  `group_namespace` varchar(128) NOT NULL DEFAULT '' COMMENT 'serviceGroup的namespace',
  `group_name` varchar(128) NOT NULL DEFAULT '' COMMENT 'serviceGroup的name',
  `service_port` int(11) NOT NULL DEFAULT '0' COMMENT '服务监听端口',
  `is_security` tinyint(1) NOT NULL DEFAULT '0' COMMENT '是否需要安全加密',
  `backend_protocol` varchar(16) NOT NULL DEFAULT '' COMMENT '后端协议',
  `project_namespace` varchar(128) NOT NULL DEFAULT '' COMMENT '项目级 namespace',
  PRIMARY KEY (`id`),
  KEY `idx_config_tenant` (`project_id`,`workspace`,`cluster_name`,`is_deleted`),
  KEY `idx_runtime_id` (`runtime_id`,`is_deleted`),
  KEY `idx_runtime_name` (`project_id`,`workspace`,`cluster_name`,`app_id`,`runtime_name`,`is_deleted`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8 COMMENT='Dice 部署服务实例信息';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `tb_gateway_service`
--

DROP TABLE IF EXISTS `tb_gateway_service`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `tb_gateway_service` (
  `id` varchar(32) NOT NULL DEFAULT '' COMMENT '唯一id',
  `service_id` varchar(128) NOT NULL DEFAULT '' COMMENT '服务id',
  `service_name` varchar(64) DEFAULT NULL COMMENT '服务名称',
  `url` varchar(1024) DEFAULT NULL COMMENT '具体路径',
  `protocol` varchar(32) DEFAULT NULL COMMENT '协议',
  `host` varchar(1024) DEFAULT NULL COMMENT '主机',
  `port` varchar(32) DEFAULT NULL COMMENT '端口',
  `path` varchar(1024) DEFAULT NULL COMMENT '路径',
  `config` varchar(1024) DEFAULT NULL COMMENT '选填配置',
  `api_id` varchar(32) NOT NULL DEFAULT '' COMMENT 'apiid',
  `create_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `update_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  `is_deleted` varchar(1) NOT NULL DEFAULT 'N' COMMENT '逻辑删除',
  PRIMARY KEY (`id`),
  KEY `service_id` (`service_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8 BLOCK_FORMAT=ENCRYPTED COMMENT='Kong Service 配置信息';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `tb_gateway_subscribe`
--

DROP TABLE IF EXISTS `tb_gateway_subscribe`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `tb_gateway_subscribe` (
  `id` varchar(32) NOT NULL DEFAULT '' COMMENT '唯一主键',
  `create_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `update_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  `is_deleted` varchar(1) NOT NULL DEFAULT 'N' COMMENT '逻辑删除',
  `confirmed` tinyint(1) DEFAULT '0' COMMENT '订阅状态',
  `subscriber_email` varchar(1024) NOT NULL DEFAULT '' COMMENT '订阅者邮箱地址',
  `dice_publish_id` varchar(32) NOT NULL DEFAULT '' COMMENT 'dice市场租户id',
  `dice_publish_item_id` varchar(32) NOT NULL DEFAULT '' COMMENT 'dice市场商品id',
  `publish_id` varchar(32) DEFAULT '' COMMENT '发布id',
  `consumer_id` varchar(32) NOT NULL DEFAULT '' COMMENT '消费者id',
  `description` varchar(256) DEFAULT NULL COMMENT '描述',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8 COMMENT='API 订阅管理(已废弃)';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `tb_gateway_upstream`
--

DROP TABLE IF EXISTS `tb_gateway_upstream`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `tb_gateway_upstream` (
  `id` varchar(32) NOT NULL DEFAULT '' COMMENT '唯一id',
  `create_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `update_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  `is_deleted` varchar(1) NOT NULL DEFAULT 'N' COMMENT '逻辑删除',
  `zone_id` varchar(32) NOT NULL DEFAULT '' COMMENT '所属的zone',
  `org_id` varchar(32) NOT NULL COMMENT '企业标识id',
  `project_id` varchar(32) NOT NULL COMMENT '项目标识id',
  `upstream_name` varchar(128) NOT NULL COMMENT '后端名称',
  `dice_app` varchar(128) DEFAULT '' COMMENT 'dice应用名',
  `dice_service` varchar(128) DEFAULT '' COMMENT 'dice服务名',
  `env` varchar(32) NOT NULL COMMENT '应用所属环境',
  `az` varchar(32) NOT NULL COMMENT '集群名',
  `last_register_id` varchar(64) NOT NULL COMMENT '应用最近一次注册id',
  `valid_register_id` varchar(64) NOT NULL COMMENT '应用当前生效的注册id',
  `auto_bind` tinyint(1) NOT NULL DEFAULT '1' COMMENT 'api是否自动绑定',
  `runtime_service_id` varchar(32) NOT NULL DEFAULT '' COMMENT '关联的service的id',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8 BLOCK_FORMAT=ENCRYPTED COMMENT='API 网关注册 SDK 的服务元信息';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `tb_gateway_upstream_api`
--

DROP TABLE IF EXISTS `tb_gateway_upstream_api`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `tb_gateway_upstream_api` (
  `id` varchar(32) NOT NULL DEFAULT '' COMMENT '唯一id',
  `create_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `update_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  `is_deleted` varchar(1) NOT NULL DEFAULT 'N' COMMENT '逻辑删除',
  `upstream_id` varchar(32) NOT NULL COMMENT '应用标识id',
  `register_id` varchar(64) NOT NULL COMMENT '应用注册id',
  `api_name` varchar(256) NOT NULL COMMENT '标识api的名称，应用下唯一',
  `path` varchar(256) NOT NULL COMMENT '注册的api路径',
  `method` varchar(256) NOT NULL COMMENT '注册的api方法',
  `address` varchar(256) NOT NULL COMMENT '注册的转发地址',
  `doc` blob COMMENT 'api描述',
  `api_id` varchar(32) DEFAULT '' COMMENT 'api标识id',
  `gateway_path` varchar(256) NOT NULL COMMENT 'gateway的api路径',
  `is_inner` tinyint(1) NOT NULL DEFAULT '0' COMMENT '是否是内部api',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8 BLOCK_FORMAT=ENCRYPTED COMMENT='API 网关注册 SDK 的服务 API 信息';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `tb_gateway_upstream_lb`
--

DROP TABLE IF EXISTS `tb_gateway_upstream_lb`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `tb_gateway_upstream_lb` (
  `id` varchar(32) NOT NULL DEFAULT '' COMMENT '唯一id',
  `create_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `update_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  `is_deleted` varchar(1) NOT NULL DEFAULT 'N' COMMENT '逻辑删除',
  `zone_id` varchar(32) NOT NULL DEFAULT '' COMMENT '所属的zone',
  `org_id` varchar(32) NOT NULL COMMENT '企业标识id',
  `project_id` varchar(32) NOT NULL COMMENT '项目标识id',
  `lb_name` varchar(128) NOT NULL COMMENT '负载均衡名称',
  `env` varchar(32) NOT NULL COMMENT '应用所属环境',
  `az` varchar(32) NOT NULL COMMENT '集群名',
  `kong_upstream_id` varchar(128) DEFAULT '' COMMENT 'kong的upstream_id',
  `config` blob COMMENT '负载均衡配置',
  `healthcheck_path` varchar(128) NOT NULL DEFAULT '' COMMENT 'HTTP健康检查路径',
  `last_deployment_id` int(11) NOT NULL COMMENT '最近一次target上线请求的部署id',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8 BLOCK_FORMAT=ENCRYPTED COMMENT='API 网关注册 SDK 的服务负载均衡信息';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `tb_gateway_upstream_lb_target`
--

DROP TABLE IF EXISTS `tb_gateway_upstream_lb_target`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `tb_gateway_upstream_lb_target` (
  `id` varchar(32) NOT NULL DEFAULT '' COMMENT '唯一id',
  `create_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `update_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  `is_deleted` varchar(1) NOT NULL DEFAULT 'N' COMMENT '逻辑删除',
  `lb_id` varchar(32) NOT NULL DEFAULT '' COMMENT '关联的lb id',
  `target` varchar(64) NOT NULL COMMENT '目的地址',
  `weight` int(11) NOT NULL DEFAULT '100' COMMENT '轮询权重',
  `healthy` tinyint(1) DEFAULT '1' COMMENT '是否健康',
  `kong_target_id` varchar(128) NOT NULL DEFAULT '' COMMENT 'kong的target_id',
  `deployment_id` int(11) NOT NULL COMMENT '上线时的deployment_id',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8 BLOCK_FORMAT=ENCRYPTED COMMENT='API 网关注册 SDK 的服务节点信息';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `tb_gateway_upstream_register_record`
--

DROP TABLE IF EXISTS `tb_gateway_upstream_register_record`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `tb_gateway_upstream_register_record` (
  `id` varchar(32) NOT NULL DEFAULT '' COMMENT '唯一id',
  `create_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `update_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  `is_deleted` varchar(1) NOT NULL DEFAULT 'N' COMMENT '逻辑删除',
  `upstream_id` varchar(32) NOT NULL COMMENT '应用标识id',
  `register_id` varchar(64) NOT NULL COMMENT '应用注册id',
  `upstream_apis` blob COMMENT 'api注册列表',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8 BLOCK_FORMAT=ENCRYPTED COMMENT='API 网关注册 SDK 的注册记录';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `tb_gateway_zone`
--

DROP TABLE IF EXISTS `tb_gateway_zone`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `tb_gateway_zone` (
  `id` varchar(32) NOT NULL DEFAULT '' COMMENT '唯一id',
  `create_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `update_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  `is_deleted` varchar(1) NOT NULL DEFAULT 'N' COMMENT '逻辑删除',
  `name` varchar(1024) NOT NULL DEFAULT '' COMMENT '名称',
  `type` varchar(16) NOT NULL DEFAULT '' COMMENT '类型',
  `kong_policies` blob COMMENT '包含的kong策略id',
  `ingress_policies` blob COMMENT '包含的ingress策略id',
  `bind_domain` varchar(1024) DEFAULT NULL COMMENT '绑定的域名',
  `dice_org_id` varchar(32) NOT NULL DEFAULT '' COMMENT 'dice企业标识id',
  `dice_project_id` varchar(32) DEFAULT '' COMMENT 'dice项目标识id',
  `dice_env` varchar(32) NOT NULL COMMENT 'dice应用所属环境',
  `dice_cluster_name` varchar(32) NOT NULL COMMENT 'dice集群名',
  `dice_app` varchar(128) DEFAULT '' COMMENT 'dice应用名',
  `dice_service` varchar(128) DEFAULT '' COMMENT 'dice服务名',
  `tenant_id` varchar(128) NOT NULL DEFAULT '' COMMENT '租户id',
  `package_api_id` varchar(32) NOT NULL DEFAULT '' COMMENT '流量入口中指定api的id',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8 BLOCK_FORMAT=ENCRYPTED COMMENT='API 网关控制路由和策略的作用对象';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `tb_gateway_zone_in_package`
--

DROP TABLE IF EXISTS `tb_gateway_zone_in_package`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `tb_gateway_zone_in_package` (
  `id` varchar(32) NOT NULL DEFAULT '' COMMENT '唯一主键',
  `package_id` varchar(32) DEFAULT '' COMMENT '所属的产品包id',
  `package_zone_id` varchar(32) DEFAULT '' COMMENT '产品包的zone id',
  `zone_id` varchar(32) DEFAULT '' COMMENT '依赖的zone id',
  `route_prefix` varchar(128) NOT NULL COMMENT '路由前缀',
  `create_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `update_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  `is_deleted` varchar(1) NOT NULL DEFAULT 'N' COMMENT '逻辑删除',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8 BLOCK_FORMAT=ENCRYPTED COMMENT='流量入口下的路由作用对象(已废弃)';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `tb_middle_addon`
--

DROP TABLE IF EXISTS `tb_middle_addon`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `tb_middle_addon` (
  `id` varchar(32) NOT NULL COMMENT '数据库唯一id',
  `name` varchar(32) NOT NULL COMMENT 'addon显示名称',
  `engine` varchar(32) NOT NULL COMMENT 'addon内核名称',
  `description` varchar(256) NOT NULL COMMENT 'addon描述',
  `icon_url` varchar(256) DEFAULT NULL COMMENT 'addon icon信息',
  `logo_url` varchar(256) NOT NULL COMMENT 'addon图片地址',
  `addon_type` varchar(32) NOT NULL COMMENT 'addon分类',
  `support_env` varchar(64) NOT NULL COMMENT '支持部署环境',
  `introduce` text COMMENT 'addon内容介绍',
  `images` text COMMENT '展示图片地址',
  `create_time` datetime NOT NULL DEFAULT '1970-01-01 00:00:00' COMMENT '创建时间',
  `update_time` datetime NOT NULL DEFAULT '1970-01-01 00:00:00' COMMENT '修改时间',
  `is_deleted` varchar(1) NOT NULL DEFAULT 'N' COMMENT '逻辑删除',
  `config_vars` varchar(512) DEFAULT '' COMMENT '返回内容配置约定',
  `envs` varchar(512) DEFAULT '' COMMENT '添加非第三方addon需要的环境变量',
  `requires` varchar(256) DEFAULT '' COMMENT 'addon 配置要求',
  `is_platform` tinyint(1) DEFAULT '0' COMMENT '是否为平台Addon',
  `parent_name` varchar(32) DEFAULT '' COMMENT '所属父Addon的名称',
  `menu_display` tinyint(1) DEFAULT '0' COMMENT '是否展示功能菜单',
  `platform_addon_type` varchar(32) DEFAULT 'NORMAL' COMMENT '平台addon类型',
  `default_inject` tinyint(1) DEFAULT '0' COMMENT '是否默认注入应用',
  `diff_env` tinyint(1) DEFAULT '1' COMMENT '实例创建是否区分环境',
  `front_display` tinyint(1) DEFAULT '1' COMMENT '是否在前端展示',
  `support_cluster_type` int(1) DEFAULT '5' COMMENT '支持的集群类型',
  `can_register` int(1) DEFAULT '1' COMMENT '是否需要注册到addon-platform',
  `ext_deps` varchar(512) DEFAULT '' COMMENT '依赖的外部addon列表',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8 BLOCK_FORMAT=ENCRYPTED COMMENT='addon信息表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `tb_middle_addon_class`
--

DROP TABLE IF EXISTS `tb_middle_addon_class`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `tb_middle_addon_class` (
  `id` int(11) NOT NULL AUTO_INCREMENT COMMENT '数据库自增id',
  `engine` varchar(32) NOT NULL COMMENT 'addon内核名称',
  `class_name` varchar(16) NOT NULL COMMENT '规格名称',
  `class_cn_name` varchar(16) NOT NULL COMMENT '规格中文显示名称',
  `offerings` text COMMENT '规格对应指标参数',
  `order_weight` int(11) NOT NULL COMMENT '规格排序权重',
  `config` text COMMENT 'addon规格对应的服务配置',
  `create_time` datetime NOT NULL DEFAULT '1970-01-01 00:00:00' COMMENT '创建时间',
  `update_time` datetime NOT NULL DEFAULT '1970-01-01 00:00:00' COMMENT '修改时间',
  `is_deleted` varchar(1) NOT NULL DEFAULT 'N' COMMENT '逻辑删除',
  PRIMARY KEY (`id`),
  KEY `idx_engine` (`engine`)
) ENGINE=InnoDB AUTO_INCREMENT=51 DEFAULT CHARSET=utf8 BLOCK_FORMAT=ENCRYPTED COMMENT='addon规格信息表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `tb_middle_addon_extra`
--

DROP TABLE IF EXISTS `tb_middle_addon_extra`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `tb_middle_addon_extra` (
  `id` int(11) NOT NULL AUTO_INCREMENT COMMENT '数据库自增id',
  `addon_id` varchar(32) NOT NULL DEFAULT '' COMMENT 'addon id',
  `field` varchar(32) NOT NULL DEFAULT '' COMMENT '字段名称',
  `value` varchar(32) NOT NULL DEFAULT '' COMMENT '字段value',
  `create_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `update_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '最后更新时间',
  `is_deleted` varchar(1) NOT NULL DEFAULT 'N' COMMENT '逻辑删除',
  PRIMARY KEY (`id`),
  KEY `idx_addon_field` (`addon_id`,`field`)
) ENGINE=InnoDB AUTO_INCREMENT=45 DEFAULT CHARSET=utf8 BLOCK_FORMAT=ENCRYPTED COMMENT='addon扩展信息';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `tb_middle_addon_internal`
--

DROP TABLE IF EXISTS `tb_middle_addon_internal`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `tb_middle_addon_internal` (
  `id` int(11) NOT NULL AUTO_INCREMENT COMMENT '数据库自增id',
  `engine` varchar(32) NOT NULL COMMENT '内部addon内核名称',
  `parent_engine` varchar(32) NOT NULL COMMENT '发布addon内核名称',
  `create_time` datetime NOT NULL DEFAULT '1970-01-01 00:00:00' COMMENT '创建时间',
  `update_time` datetime NOT NULL DEFAULT '1970-01-01 00:00:00' COMMENT '修改时间',
  `is_deleted` varchar(1) NOT NULL DEFAULT 'N' COMMENT '逻辑删除',
  PRIMARY KEY (`id`),
  KEY `idx_engine` (`parent_engine`)
) ENGINE=InnoDB AUTO_INCREMENT=11 DEFAULT CHARSET=utf8 BLOCK_FORMAT=ENCRYPTED COMMENT='内部addon关联表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `tb_middle_addon_version`
--

DROP TABLE IF EXISTS `tb_middle_addon_version`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `tb_middle_addon_version` (
  `id` int(11) NOT NULL AUTO_INCREMENT COMMENT '数据库自增id',
  `engine` varchar(32) NOT NULL COMMENT 'addon内核名称',
  `engine_version` varchar(16) NOT NULL COMMENT '中间件版本',
  `image_url` varchar(256) NOT NULL COMMENT 'addon镜像地址',
  `create_time` datetime NOT NULL DEFAULT '1970-01-01 00:00:00' COMMENT '创建时间',
  `update_time` datetime NOT NULL DEFAULT '1970-01-01 00:00:00' COMMENT '修改时间',
  `is_deleted` varchar(1) NOT NULL DEFAULT 'N' COMMENT '逻辑删除',
  PRIMARY KEY (`id`),
  KEY `idx_engine` (`engine`)
) ENGINE=InnoDB AUTO_INCREMENT=36 DEFAULT CHARSET=utf8 BLOCK_FORMAT=ENCRYPTED COMMENT='addon版本信息表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `tb_middle_app_instance`
--

DROP TABLE IF EXISTS `tb_middle_app_instance`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `tb_middle_app_instance` (
  `id` varchar(32) NOT NULL DEFAULT '' COMMENT '数据库唯一id',
  `engine` varchar(32) NOT NULL COMMENT 'addon内核名称',
  `app_id` varchar(32) NOT NULL DEFAULT '' COMMENT 'appID',
  `instance_id` varchar(32) NOT NULL DEFAULT '' COMMENT '实例ID',
  `create_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `update_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '修改时间',
  `is_deleted` varchar(1) NOT NULL DEFAULT 'N' COMMENT '逻辑删除',
  `instance_name` varchar(64) NOT NULL DEFAULT '' COMMENT '实例名称',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8 BLOCK_FORMAT=ENCRYPTED COMMENT='app和instance关联表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `tb_middle_ini`
--

DROP TABLE IF EXISTS `tb_middle_ini`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `tb_middle_ini` (
  `id` int(11) NOT NULL AUTO_INCREMENT COMMENT '数据库自增id',
  `ini_name` varchar(128) NOT NULL DEFAULT '' COMMENT '配置信息名称',
  `ini_desc` varchar(256) NOT NULL DEFAULT '' COMMENT '配置信息介绍',
  `ini_value` varchar(1024) NOT NULL DEFAULT '' COMMENT '配置信息参数值',
  `create_time` datetime NOT NULL DEFAULT '1970-01-01 00:00:00' COMMENT '创建时间',
  `update_time` datetime NOT NULL DEFAULT '1970-01-01 00:00:00' COMMENT '修改时间',
  `is_deleted` varchar(1) NOT NULL DEFAULT 'N' COMMENT '逻辑删除',
  PRIMARY KEY (`id`),
  KEY `idx_ini_name` (`ini_name`)
) ENGINE=InnoDB AUTO_INCREMENT=10 DEFAULT CHARSET=utf8 BLOCK_FORMAT=ENCRYPTED COMMENT='系统配置信息表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `tb_middle_instance`
--

DROP TABLE IF EXISTS `tb_middle_instance`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `tb_middle_instance` (
  `id` varchar(64) NOT NULL DEFAULT '' COMMENT '数据库唯一id',
  `instance_name` varchar(64) NOT NULL DEFAULT '' COMMENT '实例名称',
  `instance_class` varchar(32) NOT NULL COMMENT '实例规格',
  `engine` varchar(128) NOT NULL COMMENT '中间件类型，（zookeeper,redis,mysql）',
  `engine_version` varchar(16) NOT NULL COMMENT '中间件版本',
  `version` int(11) NOT NULL COMMENT '实例版本',
  `org_id` varchar(32) NOT NULL DEFAULT 'terminus' COMMENT '所属组织Id',
  `user_id` varchar(32) DEFAULT 'user_id' COMMENT '实例创建用户id',
  `env` varchar(16) NOT NULL DEFAULT 'PRO' COMMENT '实例环境',
  `addon_type` varchar(32) NOT NULL COMMENT 'addon分类',
  `is_share` tinyint(1) NOT NULL DEFAULT '1' COMMENT '实例是否共享',
  `status` varchar(16) NOT NULL COMMENT '实例当前运行状态',
  `az` varchar(128) NOT NULL DEFAULT 'MARATHONFORTERMINUS' COMMENT '实例规格',
  `create_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `update_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '修改时间',
  `is_deleted` varchar(1) NOT NULL DEFAULT 'N' COMMENT '逻辑删除',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8 BLOCK_FORMAT=ENCRYPTED COMMENT='实例信息表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `tb_middle_instance_extra`
--

DROP TABLE IF EXISTS `tb_middle_instance_extra`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `tb_middle_instance_extra` (
  `id` varchar(64) NOT NULL DEFAULT '' COMMENT '数据库唯一id',
  `instance_id` varchar(64) NOT NULL DEFAULT '' COMMENT '实例id',
  `field` varchar(32) NOT NULL DEFAULT '' COMMENT '域',
  `value` varchar(1024) DEFAULT NULL,
  `create_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `update_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  `is_deleted` varchar(1) NOT NULL DEFAULT 'N' COMMENT '逻辑删除',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8 BLOCK_FORMAT=ENCRYPTED COMMENT='addon实例额外信息';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `tb_middle_instance_inside`
--

DROP TABLE IF EXISTS `tb_middle_instance_inside`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `tb_middle_instance_inside` (
  `id` varchar(64) NOT NULL DEFAULT '' COMMENT '数据库唯一id',
  `az` varchar(128) NOT NULL COMMENT '集群域',
  `engine` varchar(128) NOT NULL COMMENT '中间件类型，（zookeeper,redis,mysql）',
  `instance_id` varchar(64) NOT NULL COMMENT '实例id',
  `create_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `update_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  `is_deleted` varchar(1) NOT NULL DEFAULT 'N' COMMENT '逻辑删除',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8 BLOCK_FORMAT=ENCRYPTED;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `tb_middle_instance_internal`
--

DROP TABLE IF EXISTS `tb_middle_instance_internal`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `tb_middle_instance_internal` (
  `id` int(11) NOT NULL AUTO_INCREMENT COMMENT '数据库自增id',
  `instance` varchar(64) NOT NULL COMMENT 'instanceId',
  `parent_instance` varchar(64) NOT NULL COMMENT '与instance列关联的instanceId',
  `create_time` datetime NOT NULL DEFAULT '1970-01-01 00:00:00' COMMENT '创建时间',
  `update_time` datetime NOT NULL DEFAULT '1970-01-01 00:00:00' COMMENT '修改时间',
  `is_deleted` varchar(1) NOT NULL DEFAULT 'N' COMMENT '逻辑删除',
  PRIMARY KEY (`id`),
  KEY `idx_instance` (`parent_instance`)
) ENGINE=InnoDB AUTO_INCREMENT=5 DEFAULT CHARSET=utf8 BLOCK_FORMAT=ENCRYPTED COMMENT='内部instance关联表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `tb_middle_instance_namespace`
--

DROP TABLE IF EXISTS `tb_middle_instance_namespace`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `tb_middle_instance_namespace` (
  `id` int(11) NOT NULL AUTO_INCREMENT COMMENT '数据库自增id',
  `instance_id` varchar(64) NOT NULL DEFAULT '实例id' COMMENT '实例id',
  `name` varchar(64) NOT NULL DEFAULT '' COMMENT 'schedule创建name',
  `namespace` varchar(128) NOT NULL COMMENT 'schedule创建namespace',
  `create_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `update_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '修改时间',
  `is_deleted` varchar(1) NOT NULL DEFAULT 'N' COMMENT '逻辑删除',
  PRIMARY KEY (`id`),
  KEY `idx_instance_id` (`instance_id`)
) ENGINE=InnoDB AUTO_INCREMENT=60 DEFAULT CHARSET=utf8 BLOCK_FORMAT=ENCRYPTED COMMENT='实例对应namespace关系标';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `tb_middle_instance_relation`
--

DROP TABLE IF EXISTS `tb_middle_instance_relation`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `tb_middle_instance_relation` (
  `id` varchar(64) NOT NULL DEFAULT '' COMMENT '数据库唯一id',
  `outside_instance_id` varchar(64) NOT NULL COMMENT '外部实例id',
  `inside_instance_id` varchar(64) NOT NULL COMMENT '内部实例id',
  `create_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `update_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  `is_deleted` varchar(1) NOT NULL DEFAULT 'N' COMMENT '逻辑删除',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8 BLOCK_FORMAT=ENCRYPTED COMMENT='addon实例依赖关系';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `tb_middle_node`
--

DROP TABLE IF EXISTS `tb_middle_node`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `tb_middle_node` (
  `id` varchar(64) NOT NULL DEFAULT '' COMMENT '数据库唯一id',
  `instance_id` varchar(64) NOT NULL DEFAULT '' COMMENT '实例id',
  `namespace` varchar(128) NOT NULL COMMENT '容器逻辑隔离空间',
  `node_name` varchar(128) NOT NULL COMMENT '节点名称',
  `cpu` double(8,2) NOT NULL COMMENT 'cpu核数',
  `mem` int(11) NOT NULL COMMENT '内存大小（M）',
  `disk_size` int(11) DEFAULT NULL COMMENT '磁盘大小（M）',
  `disk_type` varchar(32) DEFAULT '' COMMENT '磁盘类型',
  `count` int(11) DEFAULT '1' COMMENT '节点数',
  `vip` varchar(256) DEFAULT '' COMMENT '节点vip地址',
  `ports` varchar(256) DEFAULT '' COMMENT '节点服务端口',
  `create_time` datetime NOT NULL DEFAULT '1970-01-01 00:00:00' COMMENT '创建时间',
  `update_time` datetime NOT NULL DEFAULT '1970-01-01 00:00:00' COMMENT '修改时间',
  `is_deleted` varchar(1) NOT NULL DEFAULT 'N' COMMENT '逻辑删除',
  `proxy` varchar(256) DEFAULT '' COMMENT '节点proxy地址',
  `proxy_ports` varchar(256) DEFAULT '' COMMENT '节点proxy端口',
  `node_role` varchar(32) DEFAULT NULL COMMENT 'node节点角色',
  PRIMARY KEY (`id`),
  KEY `idx_instance_id` (`instance_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8 BLOCK_FORMAT=ENCRYPTED COMMENT='addon 节点信息';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `tb_middle_request_instance`
--

DROP TABLE IF EXISTS `tb_middle_request_instance`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `tb_middle_request_instance` (
  `request_instance_id` varchar(64) NOT NULL COMMENT '获取、更改、删除资源时请求的instance_id',
  `instance_id` varchar(64) NOT NULL COMMENT '实例ID',
  `create_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `update_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '修改时间',
  `is_deleted` varchar(1) NOT NULL DEFAULT 'N' COMMENT '逻辑删除',
  PRIMARY KEY (`request_instance_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 BLOCK_FORMAT=ENCRYPTED COMMENT='请求实例信息表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `tb_tmc`
--

DROP TABLE IF EXISTS `tb_tmc`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `tb_tmc` (
  `id` int(11) NOT NULL AUTO_INCREMENT COMMENT '自增id',
  `name` varchar(32) NOT NULL COMMENT '显示名称',
  `engine` varchar(32) NOT NULL COMMENT '内核名称',
  `service_type` varchar(32) NOT NULL COMMENT '分类，微服务、通用能力，addon',
  `deploy_mode` varchar(32) NOT NULL COMMENT '发布模式',
  `create_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `update_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '修改时间',
  `is_deleted` varchar(1) NOT NULL DEFAULT 'N' COMMENT '逻辑删除',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=22 DEFAULT CHARSET=utf8 BLOCK_FORMAT=ENCRYPTED COMMENT='微服务组件元信息';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `tb_tmc_ini`
--

DROP TABLE IF EXISTS `tb_tmc_ini`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `tb_tmc_ini` (
  `id` int(11) NOT NULL AUTO_INCREMENT COMMENT '自增id',
  `ini_name` varchar(128) NOT NULL COMMENT '配置信息名称',
  `ini_desc` varchar(256) NOT NULL COMMENT '配置信息介绍',
  `ini_value` varchar(4096) NOT NULL COMMENT '配置信息参数值',
  `create_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `update_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '修改时间',
  `is_deleted` varchar(1) NOT NULL DEFAULT 'N' COMMENT '逻辑删除',
  PRIMARY KEY (`id`),
  KEY `idx_ini_name` (`ini_name`)
) ENGINE=InnoDB AUTO_INCREMENT=12 DEFAULT CHARSET=utf8 BLOCK_FORMAT=ENCRYPTED COMMENT='微服务控制平台的配置信息';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `tb_tmc_instance`
--

DROP TABLE IF EXISTS `tb_tmc_instance`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `tb_tmc_instance` (
  `id` varchar(64) NOT NULL COMMENT '实例唯一id',
  `engine` varchar(128) NOT NULL COMMENT '对应内核名称',
  `version` varchar(32) NOT NULL COMMENT '版本',
  `release_id` varchar(32) DEFAULT '' COMMENT 'dicehub releaseId',
  `status` varchar(16) NOT NULL COMMENT '实例当前运行状态',
  `az` varchar(128) NOT NULL DEFAULT 'MARATHONFORTERMINUS' COMMENT '部署集群',
  `config` varchar(4096) DEFAULT NULL COMMENT '实例配置信息',
  `options` varchar(4096) DEFAULT NULL COMMENT '选填参数',
  `create_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `update_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '修改时间',
  `is_deleted` varchar(1) NOT NULL DEFAULT 'N' COMMENT '逻辑删除',
  `is_custom` varchar(1) NOT NULL DEFAULT 'N' COMMENT '是否为第三方',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8 BLOCK_FORMAT=ENCRYPTED COMMENT='微服务组件部署实例信息';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `tb_tmc_instance_tenant`
--

DROP TABLE IF EXISTS `tb_tmc_instance_tenant`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `tb_tmc_instance_tenant` (
  `id` varchar(64) NOT NULL COMMENT '租户唯一id',
  `instance_id` varchar(64) NOT NULL COMMENT '实例id',
  `config` varchar(4096) DEFAULT NULL COMMENT '租户配置信息',
  `options` varchar(4096) DEFAULT NULL COMMENT '选填参数',
  `create_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `update_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '修改时间',
  `is_deleted` varchar(1) NOT NULL DEFAULT 'N' COMMENT '逻辑删除',
  `tenant_group` varchar(64) DEFAULT NULL COMMENT '租户组',
  `engine` varchar(128) DEFAULT NULL COMMENT '内核',
  `az` varchar(128) DEFAULT NULL COMMENT '集群',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8 BLOCK_FORMAT=ENCRYPTED COMMENT='微服务组件租户信息';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `tb_tmc_request_relation`
--

DROP TABLE IF EXISTS `tb_tmc_request_relation`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `tb_tmc_request_relation` (
  `id` int(11) NOT NULL AUTO_INCREMENT COMMENT '自增id',
  `parent_request_id` varchar(64) NOT NULL COMMENT '父级请求id',
  `child_request_id` varchar(64) NOT NULL COMMENT '子级请求id',
  `create_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `update_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '修改时间',
  `is_deleted` varchar(1) NOT NULL DEFAULT 'N' COMMENT '逻辑删除',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=41 DEFAULT CHARSET=utf8 BLOCK_FORMAT=ENCRYPTED COMMENT='微服务组件部署过程依赖关系';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `tb_tmc_version`
--

DROP TABLE IF EXISTS `tb_tmc_version`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `tb_tmc_version` (
  `id` int(11) NOT NULL AUTO_INCREMENT COMMENT '自增id',
  `engine` varchar(32) NOT NULL COMMENT '内核名称',
  `version` varchar(32) NOT NULL COMMENT '版本',
  `release_id` varchar(32) DEFAULT NULL COMMENT 'dicehub releaseId',
  `create_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `update_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '修改时间',
  `is_deleted` varchar(1) NOT NULL DEFAULT 'N' COMMENT '逻辑删除',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=27 DEFAULT CHARSET=utf8 BLOCK_FORMAT=ENCRYPTED COMMENT='微服务组件版本元信息';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `uc_client_details`
--

DROP TABLE IF EXISTS `uc_client_details`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `uc_client_details` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT COMMENT '主键',
  `user_id` bigint(20) DEFAULT NULL COMMENT '创建的用户id',
  `client_id` varchar(255) DEFAULT NULL COMMENT 'clientID',
  `client_name` varchar(255) DEFAULT NULL COMMENT 'client名称',
  `client_secret` varchar(255) DEFAULT NULL COMMENT 'client密钥',
  `access_token_validity` varchar(255) DEFAULT NULL COMMENT 'access token 校验',
  `access_token_validity_seconds` int(11) DEFAULT NULL COMMENT 'access token 过期时长',
  `additional_information` varchar(255) DEFAULT NULL COMMENT '额外信息',
  `authorities` varchar(255) DEFAULT NULL COMMENT '权限',
  `authorized_grant_types` varchar(255) DEFAULT NULL COMMENT '授权方式',
  `auto_approve_scopes` varchar(255) DEFAULT NULL COMMENT '自动核准作用域',
  `refresh_token_validity` varchar(255) DEFAULT NULL COMMENT 'refresh token 校验',
  `refresh_token_validity_seconds` int(11) DEFAULT NULL COMMENT 'refresh token 过期时长',
  `registered_redirect_uris` varchar(2048) DEFAULT NULL COMMENT '注册的重定向地址',
  `resource_ids` varchar(255) DEFAULT NULL COMMENT '资源信息的id',
  `scope` varchar(255) DEFAULT NULL COMMENT '可使用资源域',
  `enabled` tinyint(1) NOT NULL DEFAULT '1' COMMENT '是否启用',
  `deleted` tinyint(1) NOT NULL DEFAULT '0' COMMENT '逻辑删除delete flag (0:not,1:yes)',
  `created_at` datetime NOT NULL COMMENT '创建时间',
  `updated_at` datetime DEFAULT NULL COMMENT '更新时间',
  `app_level` tinyint(4) DEFAULT NULL COMMENT '应用等级',
  `sensitived` tinyint(1) DEFAULT NULL COMMENT '敏感&非敏感',
  `alter_notify_url` varchar(1024) DEFAULT NULL COMMENT '数据变更通知URL',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=5 DEFAULT CHARSET=utf8mb4 ROW_FORMAT=DYNAMIC BLOCK_FORMAT=ENCRYPTED COMMENT='用户客户端详情表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `uc_client_details_old`
--

DROP TABLE IF EXISTS `uc_client_details_old`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `uc_client_details_old` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT,
  `access_token_validity` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `access_token_validity_seconds` int(11) DEFAULT NULL,
  `additional_information` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `approved` tinyint(4) DEFAULT '0',
  `authorities` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `authorized_grant_types` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `auto_approve` tinyint(4) DEFAULT '0',
  `auto_approve_scopes` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `client_id` varchar(128) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `client_logo_url` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `client_name` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `client_secret` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `refresh_token_validity` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `refresh_token_validity_seconds` int(11) DEFAULT NULL,
  `registered_redirect_uris` varchar(2048) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `resource_ids` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `scope` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `user_id` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `web_server_redirect_uri` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `un_uc_client_details_client_id` (`client_id`)
) ENGINE=InnoDB AUTO_INCREMENT=5 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci BLOCK_FORMAT=ENCRYPTED;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `uc_distribute_message`
--

DROP TABLE IF EXISTS `uc_distribute_message`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `uc_distribute_message` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT COMMENT '记录id',
  `err_message` varchar(1024) DEFAULT NULL COMMENT '错误信息',
  `extra` varchar(2048) DEFAULT NULL COMMENT '扩展字段',
  `send_num` int(11) DEFAULT NULL COMMENT '发送次数',
  `status` int(11) DEFAULT NULL COMMENT '发送状态',
  `tag` varchar(50) DEFAULT NULL COMMENT '消息标签',
  `message_id` varchar(50) DEFAULT NULL COMMENT '消息id',
  `body` varchar(4096) DEFAULT NULL COMMENT '消息内容',
  `message_type` varchar(50) DEFAULT NULL COMMENT '消息类型，用于消息的反序列化',
  `created_at` datetime(6) NOT NULL COMMENT '创建时间',
  `updated_at` datetime(6) NOT NULL COMMENT '修改时间',
  `identify` varchar(32) DEFAULT NULL COMMENT '消息标识',
  PRIMARY KEY (`id`),
  KEY `idx_dist_retry_and_type` (`tag`,`status`,`send_num`) COMMENT 'tag and sendNum字段索引'
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='分发消息表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `uc_sequences_old`
--

DROP TABLE IF EXISTS `uc_sequences_old`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `uc_sequences_old` (
  `name` varchar(40) COLLATE utf8mb4_unicode_ci NOT NULL,
  `value` bigint(20) DEFAULT NULL,
  PRIMARY KEY (`name`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci BLOCK_FORMAT=ENCRYPTED;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `uc_setting_properties`
--

DROP TABLE IF EXISTS `uc_setting_properties`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `uc_setting_properties` (
  `id` int(11) unsigned NOT NULL AUTO_INCREMENT COMMENT '主键',
  `setting_key` varchar(128) NOT NULL DEFAULT '' COMMENT '生效setting key，用于指定 properties 关联到哪个setting上面',
  `properties` varchar(2048) DEFAULT NULL COMMENT '具体属性值',
  `created_at` datetime DEFAULT NULL COMMENT '创建时间',
  `updated_at` datetime DEFAULT NULL COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `UN_INDEX_SETTING_KEY` (`setting_key`)
) ENGINE=InnoDB AUTO_INCREMENT=2 DEFAULT CHARSET=utf8mb4 BLOCK_FORMAT=ENCRYPTED COMMENT='设置属性表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `uc_user`
--

DROP TABLE IF EXISTS `uc_user`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `uc_user` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT COMMENT 'id',
  `pk` bigint(20) NOT NULL COMMENT '用户标识',
  `tenant_id` int(20) NOT NULL COMMENT '租户ID',
  `username` varchar(32) NOT NULL COMMENT '用户名',
  `nickname` varchar(120) DEFAULT NULL COMMENT '昵称',
  `avatar` varchar(255) DEFAULT NULL COMMENT '头像',
  `mobile` varchar(64) CHARACTER SET utf8 DEFAULT NULL COMMENT '手机号',
  `mobile_prefix` varchar(64) DEFAULT NULL COMMENT '手机号前缀',
  `email` varchar(128) DEFAULT NULL COMMENT '邮箱',
  `password` varchar(255) CHARACTER SET utf8 DEFAULT NULL COMMENT '密码',
  `pwd_expire_at` date DEFAULT NULL COMMENT '密码过期时间',
  `enabled` tinyint(1) NOT NULL DEFAULT '1' COMMENT '是否启用',
  `locked` tinyint(1) NOT NULL DEFAULT '0' COMMENT '是否冻结',
  `deleted` tinyint(1) NOT NULL DEFAULT '0' COMMENT '逻辑删除delete flag (0:not,1:yes)',
  `channel` varchar(255) DEFAULT NULL COMMENT '注册渠道',
  `channel_type` varchar(64) DEFAULT NULL COMMENT '渠道类型',
  `source` varchar(255) DEFAULT NULL COMMENT '用户来源',
  `source_type` varchar(64) DEFAULT NULL COMMENT '来源类型',
  `tag` varchar(255) DEFAULT NULL COMMENT '标签',
  `version` int(11) DEFAULT NULL COMMENT '版本',
  `extra` varchar(2048) DEFAULT NULL COMMENT '扩展字段',
  `updated_by` varchar(128) DEFAULT NULL COMMENT '更新人，操作日志',
  `last_login_at` datetime DEFAULT NULL COMMENT '最后登录时间',
  `created_at` datetime NOT NULL COMMENT '创建时间',
  `updated_at` datetime DEFAULT NULL COMMENT '更新时间',
  `record_create_msg` tinyint(1) DEFAULT '0' COMMENT '是否记录创建用户消息',
  `record_update_msg` tinyint(1) DEFAULT '1' COMMENT '是否记录更新用户消息',
  `invitation_code` varchar(255) DEFAULT NULL COMMENT '邀请码',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uni_username` (`username`,`tenant_id`),
  UNIQUE KEY `uni_pk` (`pk`,`tenant_id`),
  UNIQUE KEY `uni_mobile` (`mobile`,`tenant_id`),
  UNIQUE KEY `uni_email` (`email`,`tenant_id`)
) ENGINE=InnoDB AUTO_INCREMENT=1000123 DEFAULT CHARSET=utf8mb4 ROW_FORMAT=DYNAMIC BLOCK_FORMAT=ENCRYPTED COMMENT='用户表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `uc_user_detail`
--

DROP TABLE IF EXISTS `uc_user_detail`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `uc_user_detail` (
  `user_id` bigint(20) NOT NULL COMMENT '用户ID',
  `info` varchar(2048) DEFAULT NULL COMMENT '用户JSON信息',
  PRIMARY KEY (`user_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 ROW_FORMAT=DYNAMIC BLOCK_FORMAT=ENCRYPTED COMMENT='用户信息表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `uc_user_event_log`
--

DROP TABLE IF EXISTS `uc_user_event_log`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `uc_user_event_log` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT COMMENT '事件日志id',
  `user_id` bigint(20) DEFAULT NULL COMMENT '用户id',
  `event_type` varchar(32) DEFAULT NULL COMMENT '事件类型',
  `event` varchar(512) DEFAULT NULL COMMENT '事件',
  `event_time` datetime DEFAULT NULL COMMENT '事件产生时间',
  `mac_address` varchar(128) DEFAULT NULL COMMENT 'mac地址',
  `ip` varchar(64) DEFAULT NULL COMMENT '请求ip',
  `extra` varchar(2048) DEFAULT NULL COMMENT '扩展',
  `created_at` datetime DEFAULT NULL COMMENT '创建时间',
  `updated_at` datetime DEFAULT NULL COMMENT '更新时间',
  `tenant_id` int(20) NOT NULL DEFAULT '1' COMMENT '租户ID',
  `operator_id` int(20) DEFAULT NULL COMMENT '操作者ID',
  PRIMARY KEY (`id`),
  KEY `idx_user_id_event_type` (`user_id`,`event_type`)
) ENGINE=InnoDB AUTO_INCREMENT=236595 DEFAULT CHARSET=utf8mb4 ROW_FORMAT=DYNAMIC BLOCK_FORMAT=ENCRYPTED COMMENT='用户事件日志表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `uc_user_third_account`
--

DROP TABLE IF EXISTS `uc_user_third_account`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `uc_user_third_account` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT COMMENT '主键',
  `user_id` bigint(20) NOT NULL COMMENT '用户ID',
  `account_id` varchar(255) NOT NULL COMMENT '三方账户ID',
  `account_name` varchar(255) DEFAULT NULL COMMENT '第三方账户名',
  `account_type` varchar(255) NOT NULL COMMENT '第三方账户类型：QQ、WECHAT、WECHAT-MP、WEIBO',
  `app_id` varchar(255) NOT NULL COMMENT '授权APPID',
  `open_id` varchar(64) DEFAULT NULL COMMENT 'OPEN ID',
  `extra` varchar(2048) DEFAULT NULL COMMENT '扩展JSON字段',
  `deleted` tinyint(1) NOT NULL DEFAULT '0' COMMENT '逻辑删除',
  `created_at` datetime NOT NULL COMMENT '创建时间',
  `updated_at` datetime DEFAULT NULL COMMENT '更新时间',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 BLOCK_FORMAT=ENCRYPTED COMMENT='第三方账户';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `uc_users_old`
--

DROP TABLE IF EXISTS `uc_users_old`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `uc_users_old` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT,
  `authorities` varchar(1024) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `avatar_url` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `birthday` datetime(6) DEFAULT NULL,
  `channel` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `created_at` datetime(6) DEFAULT NULL,
  `email` varchar(128) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `email_verified` tinyint(1) DEFAULT '0',
  `enabled` tinyint(1) DEFAULT '1',
  `extra` varchar(2048) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `gender` varchar(10) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `invitation_code` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `last_login_at` datetime(6) DEFAULT NULL,
  `locked` tinyint(1) DEFAULT '0',
  `nickname` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `password` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `password_strength` int(11) DEFAULT NULL,
  `phone_number` varchar(20) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `phone_number_verified` tinyint(1) DEFAULT '0',
  `real_name` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `sign_in_type` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `type` tinyint(6) DEFAULT '0',
  `updated_at` datetime(6) DEFAULT NULL,
  `username` varchar(64) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `un_uc_users_phone` (`phone_number`),
  UNIQUE KEY `un_uc_users_email` (`email`),
  UNIQUE KEY `un_uc_users_username` (`username`)
) ENGINE=InnoDB AUTO_INCREMENT=32 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci BLOCK_FORMAT=ENCRYPTED;
/*!40101 SET character_set_client = @saved_cs_client */;
/*!40103 SET TIME_ZONE=@OLD_TIME_ZONE */;

/*!40101 SET SQL_MODE=@OLD_SQL_MODE */;
/*!40014 SET FOREIGN_KEY_CHECKS=@OLD_FOREIGN_KEY_CHECKS */;
/*!40014 SET UNIQUE_CHECKS=@OLD_UNIQUE_CHECKS */;
/*!40101 SET CHARACTER_SET_CLIENT=@OLD_CHARACTER_SET_CLIENT */;
/*!40101 SET CHARACTER_SET_RESULTS=@OLD_CHARACTER_SET_RESULTS */;
/*!40101 SET COLLATION_CONNECTION=@OLD_COLLATION_CONNECTION */;
/*!40111 SET SQL_NOTES=@OLD_SQL_NOTES */;

-- Dump completed on 2021-10-27 20:57:40
