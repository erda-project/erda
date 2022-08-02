CREATE TABLE `erda_dev_flow`
(
    `id`                      VARCHAR(36)  NOT NULL DEFAULT '' COMMENT 'id',
    `org_id`                  BIGINT (20) NOT NULL DEFAULT 0 COMMENT 'orgID',
    `org_name`                VARCHAR(50)  NOT NULL DEFAULT '' COMMENT 'orgName',
    `app_id`                  BIGINT (20) NOT NULL DEFAULT 0 COMMENT 'projectID',
    `app_name`                VARCHAR(50)  NOT NULL DEFAULT '' COMMENT 'projectName',
    `creator`                 VARCHAR(36)  NOT NULL DEFAULT '' COMMENT 'creator',
    `created_at`              datetime     NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT 'CREATED AT',
    `updated_at`              datetime     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT 'UPDATED AT',
    `deleted_at`              BIGINT (20) NOT NULL DEFAULT '0' COMMENT 'deleted_at',
    `branch`                  VARCHAR(50)  NOT NULL DEFAULT '' COMMENT 'branch',
    `flow_rule_name`          VARCHAR(50)  NOT NULL DEFAULT '' COMMENT 'flowRuleName',
    `issue_id`                BIGINT (20) NOT NULL DEFAULT 0 COMMENT 'issueID',
    `is_join_temp_branch`     TINYINT (1) NOT NULL DEFAULT 0 COMMENT 'is join temp branch',
    `join_temp_branch_status` VARCHAR(255) NOT NULL DEFAULT '' COMMENT 'join temp branch status',
    PRIMARY KEY (`id`),
    KEY                       `idx_issue_id` ( `issue_id` ),
    KEY                       `idx_flow_rule_name` ( `flow_rule_name` )
) ENGINE = INNODB DEFAULT CHARSET = utf8mb4 COMMENT = 'erda_dev_flow';