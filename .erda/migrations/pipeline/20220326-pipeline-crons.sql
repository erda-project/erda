update pipeline_crons set `extra`=REPLACE(extra,'eventbox','core_services'),`extra`=REPLACE(extra,'version: \\"1.0\\"','version: \\"2.0\\"') where `extra` like '%eventbox_addr%';
