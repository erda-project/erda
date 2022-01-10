ALTER TABLE `erda_deployment_order`
    MODIFY COLUMN `workspace` VARCHAR (16) NOT NULL DEFAULT '' COMMENT 'workspace';