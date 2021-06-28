CREATE TABLE `erda_aksk` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT COMMENT 'Primary Key',
  `ak` char(24) NOT NULL COMMENT 'Access Key ID',
  `sk` char(32) NOT NULL COMMENT 'Secret Key',
  `is_system` tinyint(1) NOT NULL COMMENT 'identify weather used for system component communication',
  `subject_type` varchar(64) NOT NULL COMMENT 'authentication subject type. eg: organization, micro_service',
  `subject` varchar(256) NOT NULL COMMENT 'authentication subject identifier. eg: id, name or something',
  `description` varchar(255) NOT NULL DEFAULT "" COMMENT 'description',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT 'created time',
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT 'updated time',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_ak` (`ak`),
  UNIQUE KEY `uk_sk` (`sk`),
  UNIQUE KEY `uk_subject_type_subject` (`subject_type`, `subject`(32))
) ENGINE=InnoDB AUTO_INCREMENT=8 DEFAULT CHARSET=utf8mb4 COMMENT='store secret key pair';
