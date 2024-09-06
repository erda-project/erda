ALTER TABLE dice_config_namespace_relation CONVERT TO CHARACTER SET utf8mb4;

ALTER TABLE dice_member CONVERT TO CHARACTER SET utf8mb4;

ALTER TABLE `ci_v3_build_caches` MODIFY COLUMN `name` varchar (191) DEFAULT NULL COMMENT '缓存名';
ALTER TABLE `ci_v3_build_caches` MODIFY COLUMN `cluster_name` varchar (191) DEFAULT NULL COMMENT '集群名';
ALTER TABLE `dice_config_namespace_relation` MODIFY COLUMN `default_namespace` varchar (191) NOT NULL COMMENT '默认配置命名空间名称';
ALTER TABLE `dice_config_namespace_relation` MODIFY COLUMN `namespace` varchar (191) NOT NULL COMMENT '配置命名空间名称';
ALTER TABLE dice_label_relations MODIFY `ref_id` varchar (191) NOT NULL COMMENT '标签关联目标 id, eg: issue_id'