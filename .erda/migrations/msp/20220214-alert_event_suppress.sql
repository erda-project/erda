ALTER TABLE `sp_alert_event_suppress`
    ADD COLUMN `org_id` bigint NOT NULL COMMENT '组织ID';

ALTER TABLE `sp_alert_event_suppress`
    ADD COLUMN `scope_id` varchar(128) NOT NULL COMMENT '域ID';

ALTER TABLE `sp_alert_event_suppress`
    ADD COLUMN `scope` varchar(20) NOT NULL COMMENT '域';