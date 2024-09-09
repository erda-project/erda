ALTER TABLE pipeline_extras MODIFY snippets MEDIUMTEXT COMMENT 'snippet 历史';

ALTER TABLE pipeline_tasks MODIFY result MEDIUMTEXT COMMENT 'result';

ALTER TABLE ci_v3_build_caches CONVERT TO CHARACTER SET utf8mb4;

ALTER TABLE `ci_v3_build_caches` MODIFY COLUMN `name` varchar (191) DEFAULT NULL COMMENT '缓存名';
ALTER TABLE `ci_v3_build_caches` MODIFY COLUMN `cluster_name` varchar (191) DEFAULT NULL COMMENT '集群名';
