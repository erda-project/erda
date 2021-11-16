alter table `tb_gateway_domain`
      add column `org_id` varchar(32)  NOT NULL DEFAULT '' COMMENT '企业id',
      add key `idx_org` (`org_id`, `is_deleted`);




