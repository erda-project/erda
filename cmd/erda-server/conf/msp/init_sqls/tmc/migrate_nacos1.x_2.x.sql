ALTER TABLE `config_info`
    MODIFY `gmt_create` DATETIME DEFAULT CURRENT_TIMESTAMP NOT NULL COMMENT '创建时间';

ALTER TABLE `config_info`
    MODIFY `gmt_modified` DATETIME DEFAULT CURRENT_TIMESTAMP NOT NULL COMMENT '修改时间';

ALTER TABLE `config_info`
    MODIFY `src_ip` VARCHAR(50) DEFAULT NULL COMMENT 'source ip';

ALTER TABLE `config_info`
    ADD `encrypted_data_key` TEXT NOT NULL COMMENT '密钥';

ALTER TABLE `config_info_beta`
    MODIFY `gmt_create` DATETIME DEFAULT CURRENT_TIMESTAMP NOT NULL COMMENT '创建时间';

ALTER TABLE `config_info_beta`
    MODIFY `gmt_modified` DATETIME DEFAULT CURRENT_TIMESTAMP NOT NULL COMMENT '修改时间';

ALTER TABLE `config_info_beta`
    ADD `encrypted_data_key` TEXT NOT NULL COMMENT '密钥';

ALTER TABLE `config_info_tag`
    MODIFY `gmt_create` DATETIME DEFAULT CURRENT_TIMESTAMP NOT NULL COMMENT '创建时间';

ALTER TABLE `config_info_tag`
    MODIFY `gmt_modified` DATETIME DEFAULT CURRENT_TIMESTAMP NOT NULL COMMENT '修改时间';

ALTER TABLE `config_info_tag`
    MODIFY `src_ip` VARCHAR(50) DEFAULT NULL COMMENT 'source ip';

ALTER TABLE `group_capacity`
    MODIFY `gmt_create` DATETIME DEFAULT CURRENT_TIMESTAMP NOT NULL COMMENT '创建时间';

ALTER TABLE `group_capacity`
    MODIFY `gmt_modified` DATETIME DEFAULT CURRENT_TIMESTAMP NOT NULL COMMENT '修改时间';

ALTER TABLE `his_config_info`
    MODIFY `gmt_create` DATETIME DEFAULT CURRENT_TIMESTAMP NOT NULL COMMENT '创建时间';

ALTER TABLE his_config_info
    MODIFY gmt_modified DATETIME DEFAULT CURRENT_TIMESTAMP NOT NULL COMMENT '修改时间';

ALTER TABLE `his_config_info`
    ADD `encrypted_data_key` TEXT NOT NULL COMMENT '密钥';

ALTER TABLE `tenant_capacity`
    MODIFY `gmt_create` DATETIME DEFAULT CURRENT_TIMESTAMP NOT NULL COMMENT '创建时间';

ALTER TABLE `tenant_capacity`
    MODIFY `gmt_modified` DATETIME DEFAULT CURRENT_TIMESTAMP NOT NULL COMMENT '修改时间';

ALTER TABLE roles
    ADD UNIQUE INDEX `idx_user_role` (`username` ASC, `role` ASC) USING BTREE;

CREATE TABLE IF NOT EXISTS `permissions`
(
    `role`     VARCHAR(50)  NOT NULL,
    `resource` VARCHAR(255) NOT NULL,
    `action`   VARCHAR(8)   NOT NULL,
    UNIQUE INDEX `uk_role_permission` (`role`, `resource`, `action`) USING BTREE
);

