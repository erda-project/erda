CREATE TABLE `erda_cmp_resource_version` (
    `id` bigint(20) NOT NULL AUTO_INCREMENT COMMENT 'Primary Key',
    `cluster_name` varchar(253) NOT NULL COMMENT 'Cluster Name',
    `namespace` varchar(253) NOT NULL COMMENT 'Namespace of resource',
    `name` varchar(253) NOT NULL COMMENT 'Name of resource',
    `resource_version` int(10) NOT NULL COMMENT 'Resource version',
    `user_id` bigint(20) NOT NULL COMMENT 'User ID',
    `annotation` varchar(255) NOT NULL COMMENT 'Describe changes of update',
    `detail` text NOT NULL COMMENT 'The complete yaml description of the resource',
    `last_version_detail` text NOT NULL COMMENT 'The complete yaml description of the previous version of the resource',
    `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT 'created time',
    `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT 'updated time',
    PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='store history resource version';
