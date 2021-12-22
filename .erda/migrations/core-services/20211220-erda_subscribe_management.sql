CREATE TABLE `erda_subscribe_management` (
    `id` bigint(20) NOT NULL AUTO_INCREMENT COMMENT 'primary key',
    `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT 'created time',
    `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT 'updated time',
    `type` varchar(64) NOT NULL COMMENT 'type: like project or application',
    `type_id` bigint(20) NOT NULL COMMENT 'type_id: like project_id or application_id',
    `name` varchar(64) NOT NULL COMMENT 'name: like project_name or application_name',
    `user_id` varchar(64) NOT NULL COMMENT 'user_id: user id',
    PRIMARY KEY (`id`),
    KEY `idx_type_id_user_id` (`type_id`,`user_id`)
) ENGINE=InnoDB AUTO_INCREMENT=6 DEFAULT CHARSET=utf8mb4 ROW_FORMAT=DYNAMIC COMMENT='erda subscribe management'