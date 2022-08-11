-- because dop is bootstrapped on erda-server, some snippet config clients need to change the address to erda-server.
-- direct update does not determine whether it exists, because pipeline-base.sql is data created with a fixed id.
UPDATE `dice_pipeline_snippet_clients` SET `host` = 'erda-server:9095' where `id` in (1, 2);
