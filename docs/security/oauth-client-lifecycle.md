# OAuth2 Client Lifecycle

Erda's internal OAuth2 clients are provisioned by the installer and stored in `erda_token_client`.

## Clients and environment variables

| Client | Consumer | Environment |
|---|---|---|
| `pipeline` | `pipeline` | `OPENAPI_OAUTH2_TOKEN_CLIENT_ID`, `OPENAPI_OAUTH2_TOKEN_CLIENT_SECRET` |
| `action-runner` | `pipeline` action-runner scheduler | `CLIENT_ID`, `CLIENT_SECRET` |
| `orchestrator` | `orchestrator` | `TOKEN_CLIENT_ID`, `TOKEN_CLIENT_SECRET` |
| `openapi` | password login token issuer | database only |
| `elf` | optional ELF component | `OPENAPI_CLIENT_ID`, `OPENAPI_CLIENT_SECRET` |
| `fdp` | optional FDP workflow/DLC/metrics/agent | `OPEN_API_CLIENT_ID`, `OPEN_API_CLIENT_SECRET` |

The installer creates or reuses these clients after database migration and before platform components start. Missing secrets fail the installation/update with a message naming the client and the expected environment variable.

The tools admin process checks and adds the rotation columns immediately after `dbmigration`.
This keeps schema preparation and random secret generation out of the migration image.

## Rotation

`erda_token_client` stores the current secret plus `previous_secret` and `previous_secret_expires_at`. The token endpoint accepts the previous secret until it expires, then rewrites the request to the current secret.

Use the tools command:

```text
admin rotate-oauth-clients --clients pipeline,orchestrator,action-runner
```

The command updates the database and the Dice CR environment values, then waits for installed components to roll out. Optional components that are not installed are skipped.

## Security rules

- `allowAccessAllApIs` is only permitted for the `openapi` login client.
- Service client metadata is allowlisted per client.
- `pipeline` may use `User-ID`, `Org-ID`, `Internal-Client`, `Internal-Action-Client`, `pipelineID`, and `taskID`.
- `orchestrator` may use `Internal-Client`, `Org-ID`, and `RuntimeID`.
- `action-runner` may use `User-ID`.
- `fdp` may use `uuid`, `user-id`, `org-id`, and `internal-client`.
