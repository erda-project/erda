CREATE TABLE `dice_aksk` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT COMMENT 'Primary Key',
  `ak` varchar(24) DEFAULT NULL COMMENT 'Access Key ID',
  `sk` varchar(32) DEFAULT NULL COMMENT 'Secret Key',
  `internal` tinyint(1) DEFAULT NULL COMMENT 'identify weather used for internal component communication',
  `scope` varchar(255) DEFAULT NULL COMMENT 'affect scope. eg: organization, micro_service',
  `owner` varchar(255) DEFAULT NULL COMMENT 'owner identifier. eg: <orgID>',
  `description` varchar(255) DEFAULT NULL COMMENT 'description',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT 'created time',
  `deleted_at` datetime DEFAULT NULL COMMENT 'deleted time, marked for deletion',
  PRIMARY KEY (`id`),
  UNIQUE KEY `ak` (`ak`),
  UNIQUE KEY `sk` (`sk`),
  KEY `idx_dice_aksks_deleted_at` (`deleted_at`)
) ENGINE=InnoDB AUTO_INCREMENT=8 DEFAULT CHARSET=utf8mb4 COMMENT='store secret key pair';
