CREATE TABLE `erda_cmp_kube_config` (
    `id` bigint(20) NOT NULL AUTO_INCREMENT COMMENT 'Primary Key',
    `name` varchar(32) NOT NULL COMMENT 'Name of kube config, used to form primary key',
    `cluster_name` varchar(253) NOT NULL COMMENT 'Cluster Name',
    `user_id` bigint(20) NOT NULL COMMENT 'User ID',
    `token` varchar(255) NOT NULL COMMENT 'User token in kube config',
    `create_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT 'Created time',
    PRIMARY KEY (`id`),
    UNIQUE KEY (`name`, `cluster_name`, `user_id`),
    UNIQUE KEY (`token`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='store user kube config';

