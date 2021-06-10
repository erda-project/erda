CREATE TABLE `erda_issue_subscriber` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT COMMENT '主键',
  `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `issue_id` bigint(20) NOT NULL COMMENT '事件ID',
  `user_id` varchar(255) NOT NULL COMMENT '订阅人',
  PRIMARY KEY (`id`),
  KEY `idx_issue_id_user_id` (`issue_id`,`user_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='issue订阅表';
