ALTER TABLE `pipeline_definition` ADD COLUMN  `remote` varchar(100) NOT NULL DEFAULT '' COMMENT 'remote';
ALTER TABLE `pipeline_definition_extra` ADD COLUMN  `remote` varchar(100) NOT NULL DEFAULT '' COMMENT 'remote';

UPDATE `pipeline_definition` pd SET `remote` = IFNULL((SELECT `remote` FROM pipeline_source ps WHERE pd.pipeline_source_id = ps.id),"");

UPDATE `pipeline_definition_extra` pde SET `remote` = IFNULL((SELECT `remote` FROM pipeline_definition pd WHERE pde.pipeline_definition_id = pd.id),"");