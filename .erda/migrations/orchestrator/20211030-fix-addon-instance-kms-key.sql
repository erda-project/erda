update tb_addon_instance set kms_key = '' where kms_key is null;
alter table tb_addon_instance modify  column  `kms_key` varchar(64) NOT NULL DEFAULT '' COMMENT 'kms key id';
