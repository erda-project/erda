CREATE TABLE `dice_code_coverage_exec_record`
(
    `id`             bigint(20) unsigned NOT NULL AUTO_INCREMENT COMMENT '主键',
    `project_id`     bigint(20) NOT NULL COMMENT '项目ID',
    `status`         varchar(128) NOT NULL COMMENT 'running,ready,ending,success,fail',
    `coverage`       decimal(65, 2) DEFAULT NULL COMMENT '行覆盖率',
    `report_url`     varchar(255)   DEFAULT NULL COMMENT '报告下载地址',
    `report_content` text COMMENT '报告分析内容',
    `start_executor` varchar(255) NOT NULL COMMENT '开始执行者',
    `end_executor`   varchar(255) NOT NULL COMMENT '结束执行者',
    `time_begin`     datetime     NOT NULL COMMENT '开始时间',
    `time_end`       datetime       DEFAULT NULL COMMENT '结束时间',
    `created_at`     datetime     NOT NULL COMMENT '创建时间',
    `updated_at`     datetime       DEFAULT NULL COMMENT '更新时间',
    PRIMARY KEY (`id`),
    KEY              `project_id` (`project_id`) USING BTREE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='代码覆盖率执行记录';