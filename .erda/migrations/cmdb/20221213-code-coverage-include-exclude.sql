ALTER TABLE erda_code_coverage_setting MODIFY COLUMN `includes` text COMMENT '包含的package';
ALTER TABLE erda_code_coverage_setting MODIFY COLUMN `excludes` text COMMENT '排除的package';