ALTER TABLE dice_config_namespace_relation CONVERT TO CHARACTER SET utf8mb4;

ALTER TABLE dice_member CONVERT TO CHARACTER SET utf8mb4;

ALTER TABLE `dice_config_namespace_relation` MODIFY COLUMN `default_namespace` varchar (191) NOT NULL COMMENT '默认配置命名空间名称';
ALTER TABLE `dice_config_namespace_relation` MODIFY COLUMN `namespace` varchar (191) NOT NULL COMMENT '配置命名空间名称';
ALTER TABLE `dice_label_relations` MODIFY `ref_id` varchar (191) NOT NULL COMMENT '标签关联目标 id, eg: issue_id';
ALTER TABLE `dice_member` MODIFY `user_id` varchar(36) NOT NULL DEFAULT '' COMMENT '用户Id';