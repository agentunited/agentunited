# 2026-03-06 — User profile API

## Added data model fields
- `users.display_name VARCHAR(100)`
- `users.avatar_url TEXT`

Migration: `apps/api/migrations/010_user_profile.sql`

## Added authenticated endpoints
All endpoints require `Authorization: Bearer <token>` and are mounted on both `/api/v1` and `/v1`:

- `GET /me`
  - Returns current user profile:
    - `id`, `email`, `display_name`, `avatar_url`, `created_at`, `updated_at`

- `PUT /me`
  - Updates profile fields:
    - body: `{ "display_name"?: string, "avatar_url"?: string }`

- `POST /me/password`
  - Changes password:
    - body: `{ "current_password": string, "new_password": string }`
  - Validates current password and enforces existing password strength rules.

## Implementation notes
- Repository: added `UpdateProfile(ctx, id, displayName, avatarURL)`.
- Service: added `GetCurrentUser`, `UpdateProfile`, `ChangePassword`.
- Handler: new `MeHandler` with auth-context user lookup.
- URL avatars are currently accepted via `avatar_url`; multipart upload can be added later if needed.
