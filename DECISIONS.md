# DECISIONS — SoundFlow Feature Voting System

Status: **Proposed / awaiting sign-off**
Date: 2026-06-21
Author: Engineering (Staff review follow-up)

This document resolves the open questions, contradictions, and gaps found in the
spec review. It is the **source of truth** for implementation. Where it conflicts
with `PROJECT.md`, `DESIGN.md`, or `ARCHITECTURE.md`, **this document wins** until
those docs are updated.

Each decision is marked:

- **[LOCKED]** — decided, build to this.
- **[NEEDS YOU]** — a default is proposed but I want explicit confirmation.

---

## Part A — Resolutions to review findings

### Conflicts

| ID | Item | Decision |
|----|------|----------|
| C1 | Duplicate detection: core rule vs. stretch goal | **[LOCKED]** MVP ships an **exact-match** guard on normalized title (trim + lowercase + collapse internal whitespace). Collision → `409 DUPLICATE_FEATURE`. Fuzzy/"similar request" detection stays a stretch goal. |
| C2 | "Migrate to PostgreSQL" leftover text | **[LOCKED]** Remove the migration sentence from `ARCHITECTURE.md`. Postgres is the only datastore. The repository abstraction is justified by **testability** (mocked unit tests), not DB-swap. |
| C3 | Password example violates policy | **[LOCKED]** Update all doc examples to a policy-valid password: `"Pa$s"` (≥1 special, within 4–12). See D-AUTH for the final password rule. |
| C4 | Error-code list malformed/incomplete | **[LOCKED]** Canonical status set: `200, 201, 400, 401, 403, 404, 409, 422?, 500`. See Part B error contract. `403` = self-vote / forbidden; `409` = duplicate vote or duplicate feature. (`422` not used — we use `400` for validation; see D-ERR.) |
| C5 | "Production-minded" vs. Vite dev server | **[LOCKED]** Ship the **Vite dev server** in Docker on `:5173` (`pnpm dev --host`), backend on `3000`, Postgres on `5432`. Chosen for setup simplicity in the assessment. Consequence: we drop the literal "production-minded" claim for the frontend delivery and call it "review-friendly dev stack." No nginx stage. |
| C6 | Vote endpoint rules stated inconsistently | **[LOCKED]** Canonical vote rules: (1) auth required, (2) no self-vote, (3) one vote per user per feature. All three enforced; see D-VOTE. |

### Risks

| ID | Item | Decision |
|----|------|----------|
| R1 | Missing composite unique on votes | **[LOCKED]** DB-level `UNIQUE (user_id, feature_request_id)` on `votes`. App pre-check for a friendly error, but the constraint is the source of truth; on unique violation return `409 ALREADY_VOTED`. |
| R2 | `depends_on` doesn't wait for readiness | **[LOCKED]** Postgres gets a `healthcheck` (`pg_isready`); backend uses `depends_on: { db: { condition: service_healthy } }` **and** connection retry/backoff (5 attempts, exponential) as belt-and-suspenders. |
| R3 | `LIKE %q%` can't use an index | **[LOCKED]** Accept full-scan search for assessment scale. Document the limitation; `pg_trgm` is explicitly out of scope. Drop any claim that an index accelerates search. |
| R4 | Trending default with undefined formula | **[LOCKED]** Freeze a v1 formula (see D-TREND). Deterministic and unit-testable. "Improved trending" stays stretch. |
| R5 | Dynamic-sort pagination instability | **[LOCKED]** Accept it. UI uses **page-based** pagination (not infinite scroll) so reshuffle is visually bounded. Documented as known behavior. |
| R6 | JWT in localStorage / no server logout | **[LOCKED]** Keep as specified (assessment scope). Mitigations: rely on React default escaping for all user content, 24h expiry, never log tokens. Documented residual risk. |
| R7 | No CORS / API base URL story | **[LOCKED]** Backend enables CORS for the frontend origin (env `CORS_ALLOWED_ORIGINS`). Frontend reads `VITE_API_BASE_URL` (defaults `http://localhost:3000/api/v1`). |
| R8 | No abuse controls | **[LOCKED]** Out of scope; stated explicitly. No rate limiting in MVP. |

### Overengineering

| ID | Item | Decision |
|----|------|----------|
| O1 | Repo interfaces justified by DB-swap | **[LOCKED]** Keep repository interfaces, re-justify as **testability**. No extra abstraction layers beyond handler → service → repository. |
| O2 | `GET /features/{id}` "for integrations" | **[LOCKED]** Keep the endpoint (cheap, aids deep-linking/completeness) but **drop the "external integrations" justification**. Public, no auth. |
| O3 | "Ranking position" / "trending score" tracked | **[LOCKED]** **Derived**, never persisted. `rank` is the 1-based index within the current sorted+paginated result. `trending_score` is computed at query time and may be returned for transparency but is not stored. |
| O4 | Heavy visual spec vs. time box | **[LOCKED]** Sequence: functional correctness first, polish second. Hero background imagery is the **most droppable** item. Glass/gradient effects implemented via reusable utility classes, not per-component bespoke CSS. |

### Missing requirements (now specified)

| ID | Item | Decision |
|----|------|----------|
| M1 | Signup response shape | **[LOCKED]** Signup **auto-logs-in**: returns `201` with `{ token, user }`. Avoids a second round-trip; matches "low friction." |
| M2 | Error envelope | **[LOCKED]** See D-ERR. Single envelope everywhere. |
| M3 | Vote response + codes | **[LOCKED]** `201` with updated `{ feature_id, total_votes, has_voted: true }`. Duplicate → `409`, self → `403`. |
| M4 | Pagination metadata | **[LOCKED]** Responses wrap items in `{ items, page, limit, total, total_pages, has_next }`. |
| M5 | `name` rules | **[LOCKED]** Required, trimmed, 2–40 chars, **not** required to be unique. |
| M6 | Email format | **[LOCKED]** Required, trimmed, lowercased, RFC-5322-lite regex, unique. |
| M7 | Invalid sort/page/limit | **[LOCKED]** Unknown `sort` → fall back to `trending`. `page<1`→1. `limit` clamped to `1..50`, default `20`. No 400 for these (lenient query params). |
| M8 | Env vars | **[LOCKED]** Enumerated in D-ENV. JWT secret is mandatory; app refuses to boot without it. Ports are configurable from `.env` on the host side only — container ports are fixed (see D-ENV, port convention). |
| M9 | Auth per endpoint | **[LOCKED]** See the auth column in Part B. `GET` endpoints public; create/vote require auth. |
| M10 | Sanitization/trimming | **[LOCKED]** All string inputs trimmed before validation. Length checks apply post-trim. Reject whitespace-only. No HTML stripping needed (React escapes on render; API stores raw). |
| M11 | E2E tooling | **[LOCKED]** **Manual E2E** for the assessment. Automated tests are unit-only: Go `testing` + mocks (service/business rules) and Vitest (frontend). Core flow (signup → create → vote → rank) verified by manual click-through. |
| M12 | Timezone / relative time | **[LOCKED]** Store all timestamps as UTC `timestamptz`. API returns ISO-8601 UTC. Frontend renders relative ("2 days ago") in the browser's local tz. |

---

## Part B — Frozen contract

Base path: `/api/v1`. All bodies are JSON. All timestamps ISO-8601 UTC.

### D-ENV — Environment variables

Backend:

| Var | Required | Default | Notes |
|-----|----------|---------|-------|
| `DATABASE_URL` | yes | — | `postgres://user:pass@db:5432/soundflow?sslmode=disable` |
| `JWT_SECRET` | yes | — | App refuses to boot if empty |
| `JWT_TTL_HOURS` | no | `24` | Token expiry |
| `BACKEND_PORT` | no | `3000` | Port the API binds to (was `PORT`) |
| `LOG_LEVEL` | no | `info` | `debug` \| `info` \| `warn` \| `error`; an invalid value aborts boot |
| `CORS_ALLOWED_ORIGINS` | no | `http://localhost:5173` | Comma-separated |

The app reads only the vars above (all from the process environment). Config lives in
a single gitignored `.env` at the repo root (`.env.example` is the template); Docker
Compose auto-loads it. `POSTGRES_USER` / `POSTGRES_PASSWORD` / `POSTGRES_DB` are
Compose-only inputs that configure the `db` container and build the backend
`DATABASE_URL` (host `db`) — they are not consumed by the application directly.

**Port convention.** Container ports are **fixed** — API `3000`, Postgres `5432`,
Vite `5173` — and services address each other on those inside the Compose network.
The `*_PORT` vars only choose the **published host port**, so several stacks can run
side by side without colliding:

| Var | Default | Publishes |
|-----|---------|-----------|
| `BACKEND_PORT` | `3000` | host → API container `3000` |
| `DATABASE_PORT` | `5432` | host → db container `5432` |
| `FRONTEND_PORT` | `5173` | host → Vite container `5173` |

Because `env_file` hands the whole `.env` to every container, Compose **pins**
`BACKEND_PORT: "3000"` on the backend service. Without that override the app would
bind the host-facing value and Compose would publish to a port nothing listens on.
Running the backend natively (no Docker) there is no override, so `BACKEND_PORT` is
the real bind port — the one knob works in both modes.

Two host-facing values must agree with the published ports or the browser breaks:
`VITE_API_BASE_URL` must target `BACKEND_PORT`, and `CORS_ALLOWED_ORIGINS` must list
the origin `FRONTEND_PORT` serves.

Frontend (build-time):

| Var | Default | Notes |
|-----|---------|-------|
| `VITE_API_BASE_URL` | `http://localhost:3000/api/v1` | Must point at the *published* backend port |
| `VITE_LOG_LEVEL` | `info` | `debug` \| `info` \| `warn` \| `error` \| `silent`. Vite only exposes `VITE_`-prefixed vars to the browser, so Compose mirrors the root `LOG_LEVEL` into it — `.env` keeps a single knob |

### D-LOG — Logging & request correlation

**[LOCKED]** Both tiers log structurally and share one trace id.

- **Backend** — `log/slog` with a JSON handler on stdout at `LOG_LEVEL`. Every
  request gets a scoped logger carrying `correlation_id`, `method`, `path`, plus
  `user_id` once authenticated; handlers and services pull it off the context with
  `logging.FromContext(ctx)`. One access-log line closes each request with `status`,
  `bytes` and `duration_ms`.
- **Frontend** — `src/lib/logger.ts`, scoped per module (`[api]`, `[auth]`,
  `[features]`), filtered by `VITE_LOG_LEVEL`.
- **Correlation id** — header `X-Correlation-ID`. The frontend mints one per API
  call; the backend adopts it when it is ≤128 chars of `[A-Za-z0-9_.:-]` and mints a
  replacement otherwise (a hostile header must not be able to inject structure into
  a log line). It is always echoed on the response, listed in the CORS
  `AllowedHeaders` **and** `ExposedHeaders`, and attached to `ApiError.correlationId`
  so a failure surfaced in the UI can be traced to its server-side lines.

**Level semantics.** `error` = needs an operator (5xx, unreachable dependency);
`warn` = suspicious but handled (rejected token, 4xx from the client's view);
`info` = business events worth keeping (signup, login, feature created, vote
recorded, 4xx rejections with a business reason); `debug` = per-request tracing.

**Never logged:** JWTs, raw passwords, password hashes. Emails are logged as
identifiers, at `info` and below.

### D-ERR — Error envelope

Every non-2xx response:

```json
{ "error": { "code": "MACHINE_CODE", "message": "Human readable", "details": [] } }
```

- `code` — stable SCREAMING_SNAKE machine code (clients branch on this, not on `message`).
- `details` — optional array of `{ field, issue }` for validation errors.

| HTTP | code (examples) | When |
|------|-----------------|------|
| 400 | `VALIDATION_ERROR` | Bad/malformed body, failed field validation |
| 401 | `UNAUTHENTICATED` | Missing/invalid/expired JWT on a protected route |
| 403 | `SELF_VOTE_FORBIDDEN` | Voting on own feature |
| 404 | `NOT_FOUND` | Unknown feature id |
| 409 | `ALREADY_VOTED` / `DUPLICATE_FEATURE` | Unique-constraint conflicts |
| 500 | `INTERNAL` | Unexpected; never leak internals |

### D-AUTH — Auth rules

- Password: **min 4, max 12, ≥1 special char** `[!@#$%^&*()_+\-=\[\]{};:'",.<>/?\\|`~]`. Validated post-trim.
- Email: required, trimmed, lowercased, unique, format-checked.
- Name: required, trimmed, 2–40 chars, non-unique.
- JWT: HS256, claims `{ sub: user_id, name, iat, exp }`, 24h TTL, no refresh token.
- Logout: client deletes token; server is stateless.

### D-VOTE — Voting rules

1. Auth required (`401` if not).
2. Feature must exist (`404`).
3. Not the author (`403 SELF_VOTE_FORBIDDEN`).
4. One vote per `(user, feature)` — enforced by DB unique; repeat → `409 ALREADY_VOTED`.
5. Unvote/edit/delete out of scope.

### D-TREND — Trending v1 formula (frozen)

```
trending_score = total_votes / pow((age_hours + 2), 1.5)
age_hours = now() - feature.created_at, in hours
```

- Hacker-News-style gravity (`1.5`). Recency-favoring per the spec; deterministic; unit-testable with a fixed clock.
- `most_voted` = order by `total_votes desc, created_at desc`.
- `newest` = order by `created_at desc`.
- `trending` (default) = order by `trending_score desc, created_at desc`.

### Endpoint table

| Method | Path | Auth | Success | Body in | Body out |
|--------|------|------|---------|---------|----------|
| GET | `/health` | no | 200 | — | `{ status: "ok" }` |
| POST | `/auth/signup` | no | 201 | `{ name, email, password }` | `{ token, user }` |
| POST | `/auth/login` | no | 200 | `{ email, password }` | `{ token, user }` |
| POST | `/features` | yes | 201 | `{ title, description }` | `FeatureView` |
| GET | `/features` | no | 200 | — (query: `search,sort,page,limit`) | `Page<FeatureView>` |
| GET | `/features/{id}` | no | 200 | — | `FeatureView` |
| POST | `/features/{id}/vote` | yes | 201 | — | `{ feature_id, total_votes, has_voted }` |

`user` shape: `{ id, name, email, created_at }` (never `password_hash`).

`FeatureView` shape:

```json
{
  "id": "uuid",
  "title": "string",
  "description": "string",
  "author": { "id": "uuid", "name": "string" },
  "created_at": "2026-06-21T10:00:00Z",
  "total_votes": 124,
  "trending_score": 3.21,
  "has_voted": false,
  "is_author": false,
  "rank": 1
}
```

- `has_voted` / `is_author` are computed **for the requesting user** (false/absent when anonymous).
- `rank` is the 1-based position within the current sorted+paginated page.

`Page<T>` wrapper:

```json
{ "items": [], "page": 1, "limit": 20, "total": 0, "total_pages": 0, "has_next": false }
```

### Validation summary (post-trim)

| Field | Rule |
|-------|------|
| name | required, 2–40 |
| email | required, format, unique, lowercased |
| password | required, 4–12, ≥1 special |
| title | required, 2–100, reject whitespace-only |
| description | required, 2–200, reject whitespace-only |

### Data model deltas vs. ARCHITECTURE.md

- `users`: `id uuid pk`, `name`, `email citext unique` (or `text` + unique lower index), `password_hash`, `created_at timestamptz`.
- `feature_requests`: add `normalized_title` (generated/maintained) with a **unique index** for C1 duplicate guard.
- `votes`: **`UNIQUE (user_id, feature_request_id)`** (R1). Indexes: `votes(feature_request_id)`, `votes(user_id)`.
- Indexes: `feature_requests(created_at)`, `users(email)` (unique).

### Docker / compose (per C5: Vite dev server)

- Services: `db` (postgres:16, volume `pgdata`, healthcheck `pg_isready`), `backend` (depends_on db healthy + retry, port `3000`), `frontend` (Vite dev server, `pnpm dev --host`, port `5173`).
- One command: `docker compose up`.
- Seeds: idempotent SQL run on backend boot **only if `feature_requests` is empty**; schema created via `IF NOT EXISTS` migration script.

---

## Open items needing your call

_None — both open items resolved 2026-06-21:_

1. **C5** — Vite dev server in Docker on `:5173` (not nginx). ✅
2. **M11** — Manual E2E; automated unit tests only. ✅

All items are now **[LOCKED]**. This contract is frozen and ready to build to, unless you object.

---

## Part C — Amendments after freeze

The contract above was frozen on 2026-06-21. Entries here amend it. Each one names the
decision it touches, what changed, and what is deliberately left alone; nothing above
is edited in place, so the original reasoning stays readable.

| #   | Amends | Decision                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                    |
| --- | ------ | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| A1  | C5     | **[LOCKED]** Add a production frontend target — built SPA served by **Caddy**, which also reverse-proxies `/api` to the backend on the same origin. C5 stands for local review: `docker compose up` still runs the Vite dev server on `:5173`, and the default `Dockerfile` is untouched. The production path is `frontend/Dockerfile.prod` + `frontend/Caddyfile` + `docker-compose.prod.yml`. C5 said "no nginx stage"; this is not one — the reason C5 gave (setup simplicity for review) is preserved, and Caddy answers a requirement C5 never faced: exposure beyond the LAN. |

### Why A1 exists

C5 was decided for an assessment read on a laptop. Publishing the same stack to the
Internet breaks on two independent counts:

- **A dev server refuses unknown hosts.** Vite rejects requests whose `Host` header is
  not in `server.allowedHosts`, so a public hostname returns "Blocked request" and the
  application never loads.
- **A dev server is not a public artifact.** It serves the module graph and project
  files, transforms on demand without minification, and has no business handling
  untrusted traffic.

Consequences worth stating:

- `VITE_API_BASE_URL` defaults to the relative **`/api/v1`** in the production image.
  The bundle is therefore hostname-agnostic; moving it to another domain needs no
  rebuild. The previous LAN-absolute value (`http://<host>:3001/api/v1`) would have
  failed from outside the network *and* as mixed content on an HTTPS page.
- `CORS_ALLOWED_ORIGINS` is empty in production. Same-origin requests never trigger a
  preflight, so an allowlist would only be a stale permission waiting to be misused.
- Nothing publishes a host port in the production compose. The reverse proxy becomes
  the only entry point, and the frontend joins `dokploy-network` so Traefik can reach
  it.
- Caddy's site address is `:80`, not a hostname. Given a hostname it would attempt
  ACME, which is wrong here: TLS terminates at the edge and this hop is internal.
