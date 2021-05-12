# MIGRATION_BASE

CREATE TABLE `ai_environment` (
  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT,
  `name` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `description` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `owner_name` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `owner_id` bigint(20) unsigned DEFAULT NULL,
  `organization_id` bigint(20) unsigned DEFAULT NULL,
  `organization_name` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `workspace` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `created_at` datetime DEFAULT NULL,
  `updated_at` datetime DEFAULT NULL,
  `requires` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `labels` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='AI 依赖集合配置信息';

CREATE TABLE `ai_mod` (
  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT,
  `type` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `name` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `version` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=10 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='AI 依赖配置信息';


INSERT INTO `ai_mod` (`id`, `type`, `name`, `version`) VALUES (1,'PyPI','Flask','1.0.0,1.1.2'),(2,'R','ggplot2','3.3.1,3.3.2'),(3,'PyPI','virtualenv','20.0.31'),(4,'Conda','Flask','1.1.2'),(5,'PyPI','pandas','1.1.2,1.0.5,0.25.3,0.24.2,0.23.4'),(6,'PyPI','numpy','1.19.2,1.18.5,1.17.5,1.16.6,1.15.4'),(7,'PyPI','scipy','1.5.2,1.4.1,1.3.3,1.2.3,1.1.0,1.0.1'),(8,'PyPI','seaborn','0.11.0,0.10.1,0.9.1,0.8.1,0.7.1'),(9,'PyPI','scikit-learn','0.23.2,0.22.2,0.21.3,0.20.4,0.19.2,0.18.2');

CREATE TABLE `ai_notebook` (
  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT,
  `name` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `description` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `owner_name` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `owner_id` bigint(20) unsigned DEFAULT NULL,
  `organization_id` bigint(20) unsigned DEFAULT NULL,
  `organization_name` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `workspace` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `created_at` datetime DEFAULT NULL,
  `updated_at` datetime DEFAULT NULL,
  `cluster_name` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `project_name` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `application_name` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `envs` text COLLATE utf8mb4_unicode_ci,
  `image` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `requirement_env_id` bigint(20) unsigned DEFAULT NULL,
  `data_source_id` bigint(20) unsigned DEFAULT NULL,
  `generic_domain` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `cluster_domain` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `resource_cpu` double DEFAULT NULL,
  `resource_memory` int(11) DEFAULT NULL,
  `status_started_at` datetime DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='AI Jupyter IDE 配置信息';

