ALTER TABLE pipeline_extras MODIFY snippets MEDIUMTEXT NOT NULL DEFAULT '' COMMENT 'snippet 历史';

ALTER TABLE pipeline_tasks MODIFY result MEDIUMTEXT NOT NULL DEFAULT '' COMMENT 'result';
