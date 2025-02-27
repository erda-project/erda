ALTER TABLE `ps_v2_project_runtimes` MODIFY COLUMN `name` varchar (191) DEFAULT NULL;
ALTER TABLE `ps_v2_project_runtimes` MODIFY COLUMN `workspace` varchar (191) DEFAULT NULL;
ALTER TABLE `ps_runtime_instances` MODIFY COLUMN `instance_id` varchar (191) NOT NULL DEFAULT '' COMMENT '实例Id';
ALTER TABLE `ps_v2_domains` MODIFY COLUMN `domain` varchar (191) DEFAULT NULL;

ALTER TABLE ps_v2_project_runtimes CONVERT TO CHARACTER SET utf8mb4;
ALTER TABLE ps_runtime_instances CONVERT TO CHARACTER SET utf8mb4;
ALTER TABLE ps_v2_domains CONVERT TO CHARACTER SET utf8mb4;