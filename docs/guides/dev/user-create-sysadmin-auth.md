# User creation requires SysAdmin

`POST /api/users` creates platform users. OpenAPI authentication only proves
that the caller is logged in or presents a valid token, so the user service
also checks the caller's SysAdmin membership before invoking IAM or UC.

The authorization flow is:

1. Read the authenticated `User-ID`.
2. Query `dice_member` for an active system-management role.
3. Continue to the configured IAM or UC user provider only for a SysAdmin.

Missing user identity returns `401`. A non-SysAdmin returns `403`. A failed
permission lookup fails closed and returns `500`.
