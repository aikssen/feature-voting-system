# SoundFlow — Feature Voting System

A public feature-voting portal for the fictional **SoundFlow** music platform.
Users submit feature ideas, discover what the community wants, and vote on what
matters most — so the product team can prioritise with confidence.

- **Backend** — Go (Chi), PostgreSQL, JWT auth, vertical-slice architecture
- **Frontend** — React 19 + TypeScript + Vite + Tailwind v4 (dark-mode, single-page)
- **Runs with one command** — `docker compose up --build`

> Planning docs live alongside this README: product scope (`PROJECT.md`), UX/visual
> spec (`DESIGN.md`), system design (`ARCHITECTURE.md`), and the **frozen API/data
> contract + resolved decisions** (`DECISIONS.md`). When docs disagree, `DECISIONS.md`
> wins.

---

## Quick start (Docker)

Requirements: Docker with Compose v2.

```bash
docker compose up --build
```

Then open:

| Service | URL |
|---------|-----|
| Frontend | http://localhost:5173 |
| API | http://localhost:3000/api/v1 |
| Health | http://localhost:3000/api/v1/health |
| Postgres | `localhost:5432` (user/pass/db: `soundflow`) |

The backend waits for Postgres to become healthy, applies the schema
(idempotent, `IF NOT EXISTS`), and seeds demo data **once** (only when the
database is empty). Data persists across restarts in the `pgdata` named volume.

### Seeded demo accounts

All seeded users share the password **`Pa$s`**. Log in to vote and submit:

| Email | Name |
|-------|------|
| `ever@example.com` | Ever |
| `mia@example.com` | Mia |
| `leo@example.com` | Leo |
| `nina@example.com` | Nina |
| `sam@example.com` | Sam |

(You can also sign up a fresh account from the UI.)

To wipe the database and reseed from scratch:

```bash
docker compose down -v && docker compose up --build
```

---

## Local development (without Docker)

You need Go 1.26+, Node 22+, pnpm, and a PostgreSQL 16 instance.

**Backend**

```bash
cd backend
cp .env.example .env          # set DATABASE_URL + JWT_SECRET
export $(grep -v '^#' .env | xargs)
go run ./cmd/api              # serves on :3000, auto-migrates + seeds
go test ./...                 # unit tests (mock-based, no DB needed)
```

**Frontend**

```bash
cd frontend
cp .env.example .env          # VITE_API_BASE_URL (defaults to localhost:3000)
pnpm install
pnpm dev                      # serves on :5173
pnpm test                     # Vitest
pnpm build                    # type-check + production build
```

---

## Configuration

**Backend** (`backend/.env.example`)

| Variable | Required | Default | Notes |
|----------|----------|---------|-------|
| `DATABASE_URL` | yes | — | Postgres DSN |
| `JWT_SECRET` | yes | — | App refuses to boot if empty |
| `JWT_TTL_HOURS` | no | `24` | Token lifetime (no refresh token) |
| `PORT` | no | `3000` | API port |
| `CORS_ALLOWED_ORIGINS` | no | `http://localhost:5173` | Comma-separated |

**Frontend** (`frontend/.env.example`)

| Variable | Default | Notes |
|----------|---------|-------|
| `VITE_API_BASE_URL` | `http://localhost:3000/api/v1` | Backend base URL |

---

## API reference

Base path: `/api/v1`. All bodies are JSON; all timestamps are ISO-8601 UTC.
Authenticated requests send `Authorization: Bearer <token>`.

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| GET | `/health` | — | Liveness check |
| POST | `/auth/signup` | — | Create account, returns a token (auto-login) |
| POST | `/auth/login` | — | Log in, returns a token |
| POST | `/features` | ✅ | Create a feature request |
| GET | `/features` | optional | List with `search`, `sort`, `page`, `limit` |
| GET | `/features/{id}` | optional | Fetch one feature request |
| POST | `/features/{id}/vote` | ✅ | Vote on a feature request |

`sort` ∈ `trending` (default) · `most_voted` · `newest`. `limit` defaults to 20,
clamped to 1–50. Unknown sort falls back to `trending`. On the list/get
endpoints, an optional token enriches each item with `has_voted` / `is_author`.

### Error envelope

Every non-2xx response uses a single shape; clients branch on `code`, not message:

```json
{ "error": { "code": "VALIDATION_ERROR", "message": "…", "details": [ { "field": "title", "issue": "…" } ] } }
```

| HTTP | Codes | When |
|------|-------|------|
| 400 | `VALIDATION_ERROR` | Bad body / failed validation |
| 401 | `UNAUTHENTICATED`, `INVALID_CREDENTIALS` | Missing/invalid token, bad login |
| 403 | `SELF_VOTE_FORBIDDEN` | Voting on your own request |
| 404 | `NOT_FOUND` | Unknown feature id |
| 409 | `ALREADY_VOTED`, `DUPLICATE_FEATURE` | Unique-constraint conflicts |
| 500 | `INTERNAL` | Unexpected |

### Example

```bash
# Sign up (returns a token)
TOKEN=$(curl -s -X POST localhost:3000/api/v1/auth/signup \
  -H 'Content-Type: application/json' \
  -d '{"name":"Ada","email":"ada@example.com","password":"Pa$s"}' \
  | python3 -c "import sys,json;print(json.load(sys.stdin)['token'])")

# Create a feature request
curl -s -X POST localhost:3000/api/v1/features \
  -H "Authorization: Bearer $TOKEN" -H 'Content-Type: application/json' \
  -d '{"title":"Smart Shuffle","description":"Shuffle that learns your taste."}'

# Discover (trending) and vote
curl -s "localhost:3000/api/v1/features?sort=trending"
curl -s -X POST localhost:3000/api/v1/features/<id>/vote -H "Authorization: Bearer $TOKEN"
```

---

## Business rules

- Authentication required to create requests and to vote.
- **One vote per user per feature** — enforced by a DB unique constraint, not just
  app logic.
- **No self-voting** (`403`).
- Feature requests need a title (2–100) and description (2–200); near-duplicate
  titles are rejected (`409`) via a normalised-title unique index.
- Votes and feature requests are create-only (no unvote / edit / delete).
- **Trending** ranks by `votes / (age_hours + 2) ^ 1.5` — favouring recent
  activity while still weighting total votes.

---

## Testing

```bash
# Backend unit tests (service-layer business rules; mock repositories, no DB)
cd backend && go test ./...

# Frontend unit/component tests
cd frontend && pnpm test

# End-to-end HTTP smoke test against a running stack (29 checks)
API_BASE=http://localhost:3000/api/v1 ./scripts/smoke.sh
```

`scripts/smoke.sh` exercises the full flow — signup, login, discovery (all sorts +
search + ranking), create + duplicate guard, voting with the self-vote / duplicate
/ auth / not-found rejections, and CORS.

The frontend suite includes automated accessibility checks (axe). See
**`QUALITY.md`** for the full quality report — test coverage plus the
accessibility, responsive, and mobile-usability validation.

---

## Project structure

```
.
├── backend/                 Go API (vertical-slice architecture)
│   ├── cmd/api/             entrypoint: config → db → migrate → seed → serve
│   └── internal/
│       ├── auth/            signup · login (handler·service·repository·model)
│       ├── feature/         create · list · get · discovery
│       ├── vote/            vote + business rules
│       ├── shared/          apperr · httpx · token (JWT) · middleware
│       └── infrastructure/  config · db pool · migrate · seed · router
├── frontend/                React + TS + Vite + Tailwind SPA
│   └── src/
│       ├── api/             typed client + DTOs (mirrors the frozen contract)
│       ├── auth/            context, login/signup modal
│       ├── components/      header, hero, cards, vote, toolbar, toasts, states
│       └── features/        feature board hook + submission form
├── scripts/smoke.sh         end-to-end HTTP test
├── docker-compose.yml       db + backend + frontend
├── PROJECT.md · DESIGN.md · ARCHITECTURE.md · DECISIONS.md · TASKS.md
└── CLAUDE.md                working rules for AI agents on this repo
```

---

## Design decisions & scope

Key engineering decisions and the frozen API/error/data contract are documented in
**`DECISIONS.md`** (e.g. JWT in localStorage for assessment scope, page-based
pagination to bound dynamic-sort reshuffle, DB-level vote uniqueness, the trending
formula, and shipping the Vite dev server rather than an nginx build).

Intentionally **out of scope**: OAuth/social login, email verification, password
recovery, RBAC/admin, moderation, notifications, comments, unvote/edit/delete, and
microservices/CQRS/event-sourcing — see `PROJECT.md` and `ARCHITECTURE.md`.
