CREATE TABLE `dice_aksk` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT COMMENT 'Primary Key',
  `ak` char(24) NOT NULL COMMENT 'Access Key ID',
  `sk` char(32) NOT NULL COMMENT 'Secret Key',
  `is_internal` tinyint(1) NOT NULL COMMENT 'identify weather used for internal component communication',
  `scope` varchar(255) NOT NULL COMMENT 'affect scope. eg: organization, micro_service',
  `owner` varchar(255) NOT NULL COMMENT 'owner identifier. eg: <orgID>',
  `description` varchar(255) NOT NULL DEFAULT "" COMMENT 'description',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT 'created time',
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT 'updated time',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_ak` (`ak`),
  UNIQUE KEY `uk_sk` (`sk`)
) ENGINE=InnoDB AUTO_INCREMENT=8 DEFAULT CHARSET=utf8mb4 COMMENT='store secret key pair';
