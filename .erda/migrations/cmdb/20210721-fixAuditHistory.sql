ALTER TABLE dice_audit_history ADD COLUMN `fdp_project_id` varchar(128) DEFAULT "" NOT NULL COMMENT 'fdp project id';
ALTER TABLE dice_audit MODIFY COLUMN `fdp_project_id` varchar(128) DEFAULT "" NOT NULL COMMENT 'fdp project id';
