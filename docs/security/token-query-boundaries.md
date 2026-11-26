# Token Query Boundaries

## Behavior

Token queries now require an explicit boundary:

- Exact access-key lookups used for authentication are allowed and return the stored value.
- Internal callers must provide at least one of `access`, `creatorId`, or `scope` + `scopeId`.
- External users are forced to query their own `creatorId`; attempts to query another user's tokens are rejected.
- External token list/detail responses mask `accessKey` and omit `secretKey`.

`GET /api/applications/{id}` only includes the caller's PAT when a real user identity is present and the existing application permission check passes. Requests with no user identity keep `gitRepoNew`/`gitRepoAbbrev` but return an empty `token`.

## MSP Access Keys

`QueryAccessKeys`, `GetAccessKey`, `DownloadAccessKeyFile`, `CreateAccessKey`, and `DeleteAccessKey` now require project Owner or Lead permission, matching the existing UI permission gate.

## Internal Callers

Internal callers that need the real access key must pass an explicit filter. Callers that previously relied on an unbounded token query must add `creatorId`, `access`, or `scope`/`scopeId`.

## Rollback

Revert commits `0f09a40e8` and `a9d858689`. No database schema changes are involved.
