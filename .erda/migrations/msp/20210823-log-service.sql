CREATE TABLE sp_log_service_instance
(
    `id`          int           AUTO_INCREMENT PRIMARY KEY,
    `instance_id` varchar(128)  NOT NULL COMMENT 'log-instance唯一ID',
    `es_urls`     varchar(1024) NOT NULL COMMENT 'es urls, split by comma',
    `es_config`   varchar(1024) NOT NULL COMMENT 'es connection info, serialized json format',
    `created`     datetime      NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated`     datetime      NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    UNIQUE INDEX uniq_instance_id(`instance_id`)
) COMMENT '日志服务实例表';


INSERT INTO tb_tmc_ini (ini_name, ini_desc, ini_value, create_time, update_time, is_deleted)
VALUES ('MK_log-service', '', 'LogAnalyze', DEFAULT, DEFAULT, DEFAULT);