# OpenDI API Server

## Endpoints

| Method | Path | Description |
|---|---|---|
| **Auth** | | |
| `GET` | `/v0/auth/me` | Get current user profile |
| `POST` | `/v0/auth/logout` | Log out (invalidate session / tokens) |
| `POST` | `/v0/auth/token/refresh` | Exchange refresh token for new access token |
| `GET` | `/v0/auth/login/google/start` | Start web Google login (redirect) |
| `GET` | `/v0/auth/login/google/callback` | Google OAuth callback |
| `POST` | `/v0/auth/cli/login` | Start CLI device login (returns login URL + device code) |
| `POST` | `/v0/auth/cli/poll` | Poll for CLI login completion (device code) |
| **Repositories (list)** | | |
| `GET` | `/v0/repositories` | List repos; query `scope` (mine, shared-with-me, all), optional `q`, `owner` |
| `GET` | `/v0/repositories/:owner` | List repos for that owner; optional `?scope=...` |
| `POST` | `/v0/repositories` | Create a new repository |
| **Repositories (owner/slug, alias `/v0/repo/:id`)** | | |
| `GET` | `/v0/repositories/:owner/:slug` | Get repository details |
| `PATCH` | `/v0/repositories/:owner/:slug` | Update repo metadata (name, description, visibility) |
| `DELETE` | `/v0/repositories/:owner/:slug` | Delete a repository |
| `GET` | `/v0/repositories/:owner/:slug/tags` | List tags (metadata only) |
| `GET` | `/v0/repositories/:owner/:slug/tags/:tag` | Get tag metadata only (digest, size, timestamps) |
| `GET` | `/v0/repositories/:owner/:slug/tags/:tag/model` | Get full CDM JSON at tag (ETag for cache) |
| `PUT` | `/v0/repositories/:owner/:slug/tags/:tag` | Create/overwrite tag (body: upload model or retag from existing tag/digest) |
| `DELETE` | `/v0/repositories/:owner/:slug/tags/:tag` | Delete a tag |
| `GET` | `/v0/repositories/:owner/:slug/collaborators` | List collaborators |
| `PUT` | `/v0/repositories/:owner/:slug/collaborators/:username` | Add/update collaborator |
| `DELETE` | `/v0/repositories/:owner/:slug/collaborators/:username` | Remove collaborator |
| `POST` | `/v0/repositories/:owner/:slug/transfer` | Transfer ownership (body: `newOwner`) |
| `GET` | `/v0/repositories/:owner/:slug/lineage` | Get lineage (parent + children) |
| `POST` | `/v0/repositories/:owner/:slug/fork` | Fork a repository |
| **Search** | | |
| `GET` | `/v0/search?q=...` | Search repositories (full-text) |

## Notes

- **List:** `GET /v0/repositories` (no path segment) and `GET /v0/repositories/:owner` use query param `scope` (`mine`, `shared-with-me`, `all`) and optional `q`, `owner`. Unauthenticated callers see only public repos.
- **Tag:** `GET .../tags/:tag` returns metadata only; `GET .../tags/:tag/model` returns full CDM JSON with `ETag` (e.g. digest). `PUT .../tags/:tag` accepts either an upload (CDM JSON) or a retag (reference to existing tag or digest).
- **By ID / CLI alias:** Every endpoint under **Repositories (owner/slug)** has an ID-based alias for stable references (e.g. CLI). Replace the `/v0/repositories/:owner/:slug` prefix with `/v0/repo/:id` while keeping the trailing path the same. For example: `GET /v0/repositories/:owner/:slug` ↔ `GET /v0/repo/:id`, `GET /v0/repositories/:owner/:slug/tags` ↔ `GET /v0/repo/:id/tags`.
