# MIGRATION_BASE

CREATE TABLE `uc_client_details` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT COMMENT '主键',
  `user_id` bigint(20) DEFAULT NULL COMMENT '创建的用户id',
  `client_id` varchar(255) DEFAULT NULL COMMENT 'clientID',
  `client_name` varchar(255) DEFAULT NULL COMMENT 'client名称',
  `client_secret` varchar(255) DEFAULT NULL COMMENT 'client密钥',
  `access_token_validity` varchar(255) DEFAULT NULL COMMENT 'access token 校验',
  `access_token_validity_seconds` int(11) DEFAULT NULL COMMENT 'access token 过期时长',
  `additional_information` varchar(255) DEFAULT NULL COMMENT '额外信息',
  `authorities` varchar(255) DEFAULT NULL COMMENT '权限',
  `authorized_grant_types` varchar(255) DEFAULT NULL COMMENT '授权方式',
  `auto_approve_scopes` varchar(255) DEFAULT NULL COMMENT '自动核准作用域',
  `refresh_token_validity` varchar(255) DEFAULT NULL COMMENT 'refresh token 校验',
  `refresh_token_validity_seconds` int(11) DEFAULT NULL COMMENT 'refresh token 过期时长',
  `registered_redirect_uris` varchar(2048) DEFAULT NULL COMMENT '注册的重定向地址',
  `resource_ids` varchar(255) DEFAULT NULL COMMENT '资源信息的id',
  `scope` varchar(255) DEFAULT NULL COMMENT '可使用资源域',
  `enabled` tinyint(1) NOT NULL DEFAULT '1' COMMENT '是否启用',
  `deleted` tinyint(1) NOT NULL DEFAULT '0' COMMENT '逻辑删除delete flag (0:not,1:yes)',
  `created_at` datetime NOT NULL COMMENT '创建时间',
  `updated_at` datetime DEFAULT NULL COMMENT '更新时间',
  `app_level` tinyint(4) DEFAULT NULL COMMENT '应用等级',
  `sensitived` tinyint(1) DEFAULT NULL COMMENT '敏感&非敏感',
  `alter_notify_url` varchar(1024) DEFAULT NULL COMMENT '数据变更通知URL',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=5 DEFAULT CHARSET=utf8mb4 ROW_FORMAT=DYNAMIC COMMENT='用户客户端详情表';


INSERT INTO `uc_client_details` (`id`, `user_id`, `client_id`, `client_name`, `client_secret`, `access_token_validity`, `access_token_validity_seconds`, `additional_information`, `authorities`, `authorized_grant_types`, `auto_approve_scopes`, `refresh_token_validity`, `refresh_token_validity_seconds`, `registered_redirect_uris`, `resource_ids`, `scope`, `enabled`, `deleted`, `created_at`, `updated_at`, `app_level`, `sensitived`, `alter_notify_url`) VALUES (1,NULL,'dice','dice','083b@a5c8b29673e931adef4f',NULL,1296000,'{}','ROLE_CLIENT,ROLE_TRUSTED_CLIENT,ROLE_ADMIN_CLIENT','password,client_credentials,authorization_code,refresh_token,implicit,client_implicit','public_profile',NULL,1296000,NULL,NULL,'public_profile',1,0,'2019-09-09 20:19:29','2019-09-09 20:19:29',NULL,NULL,NULL),(2,NULL,'pipeline','pipeline','b9a3@98cb0dd261ad80e8a13d',NULL,315360000,'{}','ROLE_CLIENT','client_credentials','public_profile',NULL,315360000,NULL,NULL,'public_profile',1,0,'2019-09-09 20:19:29','2019-09-09 20:19:29',NULL,NULL,NULL),(3,NULL,'gittar','gittar','d994@1770308419cfed73cd87',NULL,1296000,'{}','ROLE_CLIENT,ROLE_TRUSTED_CLIENT','client_credentials,password,authorization_code,refresh_token','public_profile',NULL,433200,NULL,NULL,'public_profile',1,0,'2019-09-09 20:19:29','2019-09-09 20:19:29',NULL,NULL,NULL),(4,NULL,'soldier','soldier','6435@e1d8a7b16a218bfec833',NULL,433200,'{}','ROLE_CLIENT','client_credentials','public_profile',NULL,433200,NULL,NULL,'public_profile',1,0,'2019-09-09 20:19:29','2019-09-09 20:19:29',NULL,NULL,NULL);

CREATE TABLE `uc_distribute_message` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT COMMENT '记录id',
  `err_message` varchar(1024) DEFAULT NULL COMMENT '错误信息',
  `extra` varchar(2048) DEFAULT NULL COMMENT '扩展字段',
  `send_num` int(11) DEFAULT NULL COMMENT '发送次数',
  `status` int(11) DEFAULT NULL COMMENT '发送状态',
  `tag` varchar(50) DEFAULT NULL COMMENT '消息标签',
  `message_id` varchar(50) DEFAULT NULL COMMENT '消息id',
  `body` varchar(4096) DEFAULT NULL COMMENT '消息内容',
  `message_type` varchar(50) DEFAULT NULL COMMENT '消息类型，用于消息的反序列化',
  `created_at` datetime(6) NOT NULL COMMENT '创建时间',
  `updated_at` datetime(6) NOT NULL COMMENT '修改时间',
  `identify` varchar(32) DEFAULT NULL COMMENT '消息标识',
  PRIMARY KEY (`id`),
  KEY `idx_dist_retry_and_type` (`tag`,`status`,`send_num`) COMMENT 'tag and sendNum字段索引'
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='分发消息表';

CREATE TABLE `uc_setting_properties` (
  `id` int(11) unsigned NOT NULL AUTO_INCREMENT COMMENT '主键',
  `setting_key` varchar(128) NOT NULL DEFAULT '' COMMENT '生效setting key，用于指定 properties 关联到哪个setting上面',
  `properties` varchar(2048) DEFAULT NULL COMMENT '具体属性值',
  `created_at` datetime DEFAULT NULL COMMENT '创建时间',
  `updated_at` datetime DEFAULT NULL COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `UN_INDEX_SETTING_KEY` (`setting_key`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='设置属性表';

CREATE TABLE `uc_user` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT COMMENT 'id',
  `pk` bigint(20) NOT NULL COMMENT '用户标识',
  `tenant_id` int(20) NOT NULL COMMENT '租户ID',
  `username` varchar(32) NOT NULL COMMENT '用户名',
  `nickname` varchar(120) DEFAULT NULL COMMENT '昵称',
  `avatar` varchar(255) DEFAULT NULL COMMENT '头像',
  `mobile` varchar(64) CHARACTER SET utf8 DEFAULT NULL COMMENT '手机号',
  `mobile_prefix` varchar(64) DEFAULT NULL COMMENT '手机号前缀',
  `email` varchar(128) DEFAULT NULL COMMENT '邮箱',
  `password` varchar(255) CHARACTER SET utf8 DEFAULT NULL COMMENT '密码',
  `pwd_expire_at` date DEFAULT NULL COMMENT '密码过期时间',
  `enabled` tinyint(1) NOT NULL DEFAULT '1' COMMENT '是否启用',
  `locked` tinyint(1) NOT NULL DEFAULT '0' COMMENT '是否冻结',
  `deleted` tinyint(1) NOT NULL DEFAULT '0' COMMENT '逻辑删除delete flag (0:not,1:yes)',
  `channel` varchar(255) DEFAULT NULL COMMENT '注册渠道',
  `channel_type` varchar(64) DEFAULT NULL COMMENT '渠道类型',
  `source` varchar(255) DEFAULT NULL COMMENT '用户来源',
  `source_type` varchar(64) DEFAULT NULL COMMENT '来源类型',
  `tag` varchar(255) DEFAULT NULL COMMENT '标签',
  `version` int(11) DEFAULT NULL COMMENT '版本',
  `extra` varchar(2048) DEFAULT NULL COMMENT '扩展字段',
  `updated_by` varchar(128) DEFAULT NULL COMMENT '更新人，操作日志',
  `last_login_at` datetime DEFAULT NULL COMMENT '最后登录时间',
  `created_at` datetime NOT NULL COMMENT '创建时间',
  `updated_at` datetime DEFAULT NULL COMMENT '更新时间',
  `record_create_msg` tinyint(1) DEFAULT '0' COMMENT '是否记录创建用户消息',
  `record_update_msg` tinyint(1) DEFAULT '1' COMMENT '是否记录更新用户消息',
  `invitation_code` varchar(255) DEFAULT NULL COMMENT '邀请码',
  PRIMARY KEY (`id`) USING BTREE,
  UNIQUE KEY `uni_username` (`username`,`tenant_id`),
  UNIQUE KEY `uni_pk` (`pk`,`tenant_id`),
  UNIQUE KEY `uni_mobile` (`mobile`,`tenant_id`),
  UNIQUE KEY `uni_email` (`email`,`tenant_id`)
) ENGINE=InnoDB AUTO_INCREMENT=1000001 DEFAULT CHARSET=utf8mb4 ROW_FORMAT=DYNAMIC COMMENT='用户表';


INSERT INTO `uc_user` (`id`, `pk`, `tenant_id`, `username`, `nickname`, `avatar`, `mobile`, `mobile_prefix`, `email`, `password`, `pwd_expire_at`, `enabled`, `locked`, `deleted`, `channel`, `channel_type`, `source`, `source_type`, `tag`, `version`, `extra`, `updated_by`, `last_login_at`, `created_at`, `updated_at`, `record_create_msg`, `record_update_msg`, `invitation_code`) VALUES (1,1,1,'admin','admin',NULL,NULL,'86','admin@dice.terminus.io','d7f6@3676e5c1116a755e4c91','2120-01-01',1,0,0,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'2019-08-22 15:15:15',NULL,0,1,NULL),(2,2,1,'dice','dice',NULL,NULL,'86','dice@dice.terminus.io','ffba@eea808d1152732e87422','2120-01-01',1,0,0,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'2019-08-22 15:15:15',NULL,0,1,NULL),(3,3,1,'gittar','gittar',NULL,'18000000000','86','gittar','c844@455d84fe523093be3b25','2021-08-12',1,0,0,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'2019-08-22 15:15:15',NULL,0,1,NULL),(1100,1100,1,'tmc','TMC',NULL,NULL,'86','tmc@dice.terminus.io','no pass','2021-08-12',1,0,0,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'2019-08-22 15:15:15',NULL,0,1,NULL),(1101,1101,1,'eventbox','eventbox',NULL,NULL,'86','eventbox@dice.terminus.io','no pass',NULL,1,0,0,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'2019-12-31 11:00:00',NULL,0,1,NULL),(1102,1102,1,'cdp','cdp',NULL,NULL,'86','cdp@dice.terminus.io','no pass',NULL,1,0,0,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'2019-12-31 11:00:00',NULL,0,1,NULL),(1103,1103,1,'pipeline','pipeline',NULL,NULL,'86','pipeline@dice.terminus.io','no pass',NULL,1,0,0,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'2020-01-09 11:00:00',NULL,0,1,NULL),(1108,1108,1,'fdp','fdp',NULL,NULL,'86','fdp@dice.terminus.io','ffba@eea808d1152732e87422',NULL,1,0,0,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'2019-08-22 15:15:15',NULL,0,1,NULL),(1110,1110,1,'system','system',NULL,NULL,'86','system@dice.terminus.io','no pass',NULL,1,0,0,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'2020-03-18 11:00:00',NULL,0,1,NULL),(2020,2020,1,'support','Support',NULL,NULL,'86',NULL,'4144@3c4243e771a5d6b75d77','2120-01-01',1,0,0,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,'2021-05-12 17:46:20','2021-05-12 17:46:20','2021-05-12 17:46:20',0,1,NULL);

CREATE TABLE `uc_user_detail` (
  `user_id` bigint(20) NOT NULL COMMENT '用户ID',
  `info` varchar(2048) DEFAULT NULL COMMENT '用户JSON信息',
  PRIMARY KEY (`user_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 ROW_FORMAT=DYNAMIC COMMENT='用户信息表';

CREATE TABLE `uc_user_event_log` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT COMMENT '事件日志id',
  `user_id` bigint(20) DEFAULT NULL COMMENT '用户id',
  `event_type` varchar(32) DEFAULT NULL COMMENT '事件类型',
  `event` varchar(512) DEFAULT NULL COMMENT '事件',
  `event_time` datetime DEFAULT NULL COMMENT '事件产生时间',
  `mac_address` varchar(128) DEFAULT NULL COMMENT 'mac地址',
  `ip` varchar(64) DEFAULT NULL COMMENT '请求ip',
  `extra` varchar(2048) DEFAULT NULL COMMENT '扩展',
  `created_at` datetime DEFAULT NULL COMMENT '创建时间',
  `updated_at` datetime DEFAULT NULL COMMENT '更新时间',
  `tenant_id` int(20) NOT NULL DEFAULT '1' COMMENT '租户ID',
  `operator_id` int(20) DEFAULT NULL COMMENT '操作者ID',
  PRIMARY KEY (`id`),
  KEY `idx_user_id_event_type` (`user_id`,`event_type`) USING BTREE
) ENGINE=InnoDB AUTO_INCREMENT=16713 DEFAULT CHARSET=utf8mb4 ROW_FORMAT=DYNAMIC COMMENT='用户事件日志表';

CREATE TABLE `uc_user_third_account` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT COMMENT '主键',
  `user_id` bigint(20) NOT NULL COMMENT '用户ID',
  `account_id` varchar(255) NOT NULL COMMENT '三方账户ID',
  `account_name` varchar(255) DEFAULT NULL COMMENT '第三方账户名',
  `account_type` varchar(255) NOT NULL COMMENT '第三方账户类型：QQ、WECHAT、WECHAT-MP、WEIBO',
  `app_id` varchar(255) NOT NULL COMMENT '授权APPID',
  `open_id` varchar(64) DEFAULT NULL COMMENT 'OPEN ID',
  `extra` varchar(2048) DEFAULT NULL COMMENT '扩展JSON字段',
  `deleted` tinyint(1) NOT NULL DEFAULT '0' COMMENT '逻辑删除',
  `created_at` datetime NOT NULL COMMENT '创建时间',
  `updated_at` datetime DEFAULT NULL COMMENT '更新时间',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='第三方账户';

