CREATE TABLE `erda_issue_subscriber` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT COMMENT '主键',
  `created_at` datetime DEFAULT NULL COMMENT '表记录创建时间',
  `updated_at` datetime DEFAULT NULL COMMENT '表记录更新时间',
  `issue_id` bigint(20) NOT NULL COMMENT '事件ID',
  `user_id` varchar(255) NOT NULL COMMENT '订阅人',
  PRIMARY KEY (`id`),
  KEY `idx_issue_id_user_id` (`issue_id`,`user_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='issue订阅表';
