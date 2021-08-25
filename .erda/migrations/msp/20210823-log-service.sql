CREATE TABLE sp_log_service_instance
(
    `id`          bigint        AUTO_INCREMENT PRIMARY KEY COMMENT 'Id',
    `instance_id` varchar(128)  NOT NULL COMMENT 'log-instance唯一ID',
    `es_urls`     varchar(1024) NOT NULL COMMENT 'es urls, split by comma',
    `es_config`   varchar(1024) NOT NULL COMMENT 'es connection info, serialized json format',
    `created_at`  datetime      NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT 'created time',
    `updated_at`  datetime      NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT 'updated time',
    UNIQUE INDEX uk_instance_id(`instance_id`)
) DEFAULT CHARSET=utf8mb4 COMMENT '日志服务实例表';


INSERT INTO tb_tmc_ini (ini_name, ini_desc, ini_value, create_time, update_time, is_deleted)
VALUES ('MK_log-service', '', 'LogAnalyze', DEFAULT, DEFAULT, DEFAULT);