CREATE TABLE `access_key` (
  `id` VARCHAR(36) NOT NULL COMMENT 'Primary Key',
  `access_key` char(24) NOT NULL COMMENT 'Access Key ID',
  `secret_key` char(32) NOT NULL COMMENT 'Secret Key',
  `status` int NOT NULL DEFAULT 0 COMMENT 'status of access key',
  `subject_type` int NOT NULL DEFAULT 0 COMMENT 'authentication subject type. eg: SYSTEM, MICRO_SERVICE',
  `subject` varchar(256) NOT NULL COMMENT 'authentication subject identifier. eg: id, name or something',
  `description` varchar(255) NOT NULL DEFAULT "" COMMENT 'description',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT 'created time',
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT 'updated time',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_ak` (`access_key`),
  UNIQUE KEY `uk_sk` (`secret_key`),
  INDEX `idx_status` (`status`)
) ENGINE=InnoDB AUTO_INCREMENT=8 DEFAULT CHARSET=utf8mb4 COMMENT='store secret key pair';
