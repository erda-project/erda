ALTER TABLE erda_access_key DROP COLUMN is_system;
ALTER TABLE erda_access_key MODIFY id VARCHAR(36);
ALTER TABLE erda_access_key CHANGE access_key_id access_key char(24) NOT NULL;
ALTER TABLE erda_access_key MODIFY status int NOT NULL DEFAULT 0 COMMENT 'status of access key';
ALTER TABLE erda_access_key MODIFY subject_type int NOT NULL DEFAULT 0 COMMENT 'authentication subject type. eg: SYSTEM, MICRO_SERVICE';
