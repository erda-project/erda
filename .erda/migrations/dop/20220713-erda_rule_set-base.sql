CREATE TABLE IF NOT EXISTS `erda_rule` (
  `id` varchar(36) NOT NULL COMMENT 'primary key',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT 'created time',
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT 'updated time',
  `name` varchar(191) NOT NULL DEFAULT '' COMMENT 'name',
  `scope` varchar(191) NOT NULL DEFAULT '' COMMENT 'scope',
  `scope_id` varchar(191) NOT NULL DEFAULT '' COMMENT 'scope id',
  `event_type` varchar(191) NOT NULL DEFAULT '' COMMENT 'event type',
  `code`  varchar(1024) NOT NULL DEFAULT '' COMMENT 'code',
  `params` varchar(2048) NOT NULL DEFAULT '' COMMENT 'actions',
  `enabled`  tinyint(1) NOT NULL DEFAULT 0 COMMENT 'enabled',
  `updator` varchar(191) NOT NULL DEFAULT '' COMMENT 'updator',
  `soft_deleted_at` bigint(20) NOT NULL DEFAULT '0' COMMENT 'deleted at',
  PRIMARY KEY (`id`),
  KEY `idx_scope` (`scope`,`scope_id`, `event_type`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='erda rule';

CREATE TABLE IF NOT EXISTS `erda_rule_exec_history` (
  `id` varchar(36) NOT NULL COMMENT 'primary key',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT 'created time',
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT 'updated time',
  `scope` varchar(191) NOT NULL DEFAULT '' COMMENT 'scope',
  `scope_id` varchar(191) NOT NULL DEFAULT '' COMMENT 'scope id',
  `rule_id` varchar(36) NOT NULL DEFAULT '' COMMENT 'rule_id',
  `code` varchar(1024) NOT NULL DEFAULT '' COMMENT 'code',
  `env` varchar(2048) NOT NULL DEFAULT '' COMMENT 'env',
  `succeed`  tinyint(1) NOT NULL DEFAULT 0 COMMENT 'succeed',
  `action_output` varchar(2048) NOT NULL DEFAULT '' COMMENT 'action output info',
  `soft_deleted_at` bigint(20) NOT NULL DEFAULT '0' COMMENT 'deleted at',
  PRIMARY KEY (`id`),
  KEY `idx_rule_id` (`rule_id`),
  KEY `idx_scope` (`scope`,`scope_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='erda rule execution history';
