INSERT INTO `erda_token` (id, code, access, refresh, data, type, created_at, updated_at, expired_at)
SELECT UUID(), code, access, refresh, data, 'Oauth2', created_at, created_at, expired_at
FROM `openapi_oauth2_tokens`;
