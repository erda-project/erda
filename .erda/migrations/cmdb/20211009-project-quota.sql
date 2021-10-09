ALTER TABLE erda.`ps_group_projects`
    ADD COLUMN prod_cpu_quota    DECIMAL(65, 2) NOT NULL DEFAULT 0.0 COMMENT '生产环境 CPU 配额',
    ADD COLUMN prod_mem_quota    DECIMAL(65, 2) NOT NULL DEFAULT 0.0 COMMENT '生产环境 Mem 配额',
    ADD COLUMN staging_cpu_quota DECIMAL(65, 2) NOT NULL DEFAULT 0.0 COMMENT '预发环境 CPU 配额',
    ADD COLUMN staging_mem_quota DECIMAL(65, 2) NOT NULL DEFAULT 0.0 COMMENT '预发环境 Mem 配额',
    ADD COLUMN test_cpu_quota    DECIMAL(65, 2) NOT NULL DEFAULT 0.0 COMMENT '测试环境 CPU 配额',
    ADD COLUMN test_mem_quota    DECIMAL(65, 2) NOT NULL DEFAULT 0.0 COMMENT '测试环境 Mem 配额',
    ADD COLUMN dev_cpu_quota     DECIMAL(65, 2) NOT NULL DEFAULT 0.0 COMMENT '开发环境 CPU 配额',
    ADD COLUMN dev_mem_quota     DECIMAL(65, 2) NOT NULL DEFAULT 0.0 COMMENT '发开环境 Mem 配额';
