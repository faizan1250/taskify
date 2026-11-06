# Taskify

A full-stack task manager: Go (Echo) API + React (Vite) frontend, with Clerk auth, Postgres, Redis-backed background jobs, S3 attachments, and scheduled email digests.

## What it is

Taskify lets users create and organize todos with subtasks, categories, comments, file attachments, priorities, due dates, and recurrence rules. It adds power-user workflows — bulk select/update/delete, drag-style reordering, shareable read-only links, URL-shareable filters, a `Cmd/Ctrl+K` command palette with keyboard shortcuts — plus a dashboard (stats, overdue and due-this-week lanes) and cron-driven email notifications (due-date reminders, overdue alerts, weekly reports, auto-archive).

## Features

- **Todos**: create, edit, complete, delete; priorities, due/start dates, rich detail drawer.
- **Subtasks**: nested child todos; toggle, rename, delete, or promote a subtask to a top-level task.
- **Categories**: create/rename/delete; filter task lists by category.
- **Comments**: per-todo discussion thread (add/edit/delete).
- **Attachments**: S3-backed uploads (10 MB request limit) with presigned download URLs.
- **Stats**: totals plus `draft / active / completed / archived / overdue` breakdowns, surfaced on the todos page, dashboard, and settings.
- **Filters & search**: status tab, priority, overdue-only, text search, date range, sorting (`due_date`, `priority`, `updated_at`); filter state lives in the URL so views are link-shareable.
- **Bulk actions**: multi-select (persisted per page/filter) with bulk status update and bulk delete.
- **Recurring todos**: `none | daily | weekly | monthly` with `next_run_at`; a cron job materializes the next occurrence.
- **Position / reorder**: fractional `position` column with a dedicated reorder endpoint.
- **Share links**: per-todo bearer capability tokens (`POST /todos/:id/share`, `DELETE /todos/:id/share`); anyone with the token can read via public `GET /shared/:token` (no login required).
- **Command palette (`Cmd/Ctrl+K`)**: jump to pages, run actions (new task, toggle theme, copy view link), apply filter views, resume recent picks/tasks.
- **Keyboard shortcuts**: `/` search, `N`/`C` new task, `T` theme, `?` shortcuts help, `J`/`K` move active task, `X` select, `E` edit, `Esc` clear.
- **Dashboard**: stat cards linked to filtered lists, overdue lane, due-this-week lane, first-run checklist, category overview.
- **Cron emails**: due-date reminders, overdue notifications, weekly productivity reports (asynq queues + Resend).
- **Auto-archive**: completed todos older than the threshold (default 30 days) are archived on a schedule.

## Architecture

```text
                    ┌─────────────┐
                    │   Clerk     │  auth (JWT session claims)
                    └──────┬──────┘
                           │
┌──────────┐  HTTPS   ┌────▼────────┐   SQL    ┌──────────┐
│ React    │ ────────▶│ Go + Echo   │─────────▶│ Postgres │
│ Vite SPA │◀──────── │ :8080       │          │ (Neon /  │
└──────────┘   JSON   └────┬────────┘          │  local)  │
  :3010                   │                   └──────────┘
                          │ Redis (Upstash / local)
                          ├──────────────────▶│ asynq queues │
                          │                   │ critical /   │
                          │ S3 attachments    │ default / low│
                          ├──────────────────▶│              │
                          │                   └──────┬───────┘
                          │ Resend emails◀──────────┘
                          │ (welcome / reminder /
                          │  weekly report)
                          │
                    ┌─────▼──────┐
                    │ cron binary│  due-date-reminders,
                    │ (cobra)    │  overdue-notifications,
                    │            │  weekly-reports, auto-archive,
                    │            │  recurring-todos
                    └────────────┘
```

- Backend validates the Clerk session per request (`RequireAuth` middleware); the only public API route is `GET /shared/:token`, where token possession is the authorization.
- Shared contract: `packages/openapi` (typed ts-rest contract + `openapi.json`, served as docs UI at `GET /docs`) and `packages/zod` (shared validation schemas).
- Rate limiting: in-memory token bucket (~20 req/s) with `429` + structured logging.

## Monorepo layout

```text
taskify/
├── apps/
│   ├── backend/                 # Go + Echo API (:8080)
│   │   ├── cmd/taskify/         # API server entrypoint
│   │   ├── cmd/cron/            # cron job runner (cobra)
│   │   ├── internal/
│   │   │   ├── config/          # koanf env config (TASKIFY_ prefix)
│   │   │   ├── router/ + handler/ + service/ + repository/
│   │   │   ├── middleware/      # Clerk auth, CORS, rate limit, tracing
│   │   │   ├── cron/            # 5 scheduled jobs
│   │   │   ├── lib/job/         # asynq client/server, email tasks
│   │   │   ├── lib/email/       # Resend client + templates
│   │   │   ├── database/        # pgx pool + tern migrations
│   │   │   │   └── migrations/  # 001–008 (todos, attachments, subtasks,
│   │   │   │                    #   notifications, recurrence, position, share)
│   │   │   └── testing/         # testcontainers Postgres helpers
│   │   └── Taskfile.yml         # task run / migrations:new / migrations:up / tidy
│   └── frontend/                # React 19 + Vite SPA (:3010)
│       └── src/
│           ├── pages/           # landing, dashboard, todos, categories, settings
│           ├── components/      # todos, dashboard, categories, command-palette, ui
│           ├── api/hooks/       # TanStack Query hooks over @taskify/openapi
│           ├── hooks/           # URL filter state, hotkeys
│           └── lib/shortcuts.ts # keyboard shortcut registry (source of truth)
├── packages/
│   ├── openapi/                 # ts-rest contract + openapi.json (frontend dev waits on its build)
│   ├── zod/                     # shared zod schemas
│   └── emails/                  # email templates
├── .github/workflows/ci.yml    # backend (Go) + frontend & packages (Bun/turbo) jobs
└── nixpacks.toml                # frontend deploy; apps/backend/nixpacks.toml builds both Go binaries
```

## Prerequisites

- **Go** 1.24+ (backend)
- **Bun** 1.2.13 + **Node** 22+ (frontend, packages, turborepo)
- **Postgres** 15+ (local or Neon) and **Redis** (local or Upstash)
- **Clerk** account (publishable key for frontend, secret key for backend)
- **S3-compatible** bucket (attachments) and **Resend** API key (emails)
- **Docker** — only for repository/integration tests (testcontainers) and optional local Postgres/Redis
- `task` runner (optional, for `Taskfile.yml` shortcuts) and `tern` (bundled via `go run`/binary for migrations)

## Environment setup

### Backend (`apps/backend/.env`)

Copy the sample and fill in real values:

```bash
cd apps/backend
cp .env.sample .env
```

Key variables (all use the `TASKIFY_` prefix, loaded via koanf):

| Variable | Purpose | Example |
|---|---|---|
| `TASKIFY_PRIMARY.ENV` | `local` skips auto-migrate on boot; anything else runs tern migrations at startup | `local` |
| `TASKIFY_SERVER.PORT` | API listen port | `8080` |
| `TASKIFY_SERVER.CORS_ALLOWED_ORIGINS` | Comma-separated frontend origins — must include the Vite dev URL | `http://localhost:3000,http://localhost:3010` |
| `TASKIFY_DATABASE.HOST/PORT/USER/PASSWORD/NAME/SSL_MODE` | Postgres connection (Neon: host + `require` SSL mode) | — |
| `TASKIFY_DATABASE.MAX_OPEN_CONNS/MAX_IDLE_CONNS/CONN_MAX_LIFETIME/CONN_MAX_IDLE_TIME` | Pool tuning | `25/25/300/300` |
| `TASKIFY_AUTH.SECRET_KEY` | Clerk secret key (server-side session verification) | `sk_test_...` |
| `TASKIFY_REDIS.ADDRESS` / `PASSWORD` | Redis URL (Upstash or local) for asynq | `redis://localhost:6379` |
| `TASKIFY_AWS.REGION/ACCESS_KEY_ID/SECRET_ACCESS_KEY/UPLOAD_BUCKET/ENDPOINT_URL` | S3 attachments (set `ENDPOINT_URL` for S3-compatible stores) | — |
| `TASKIFY_INTEGRATION.RESEND_API_KEY` | Resend key for outgoing mail | `re_...` |
| `TASKIFY_OBSERVABILITY.*` | Logging level/format, New Relic, health checks | see `.env.sample` |
| `TASKIFY_CRON.*` (optional) | `ARCHIVE_DAYS_THRESHOLD` (30), `BATCH_SIZE` (100), `REMINDER_HOURS` (24), `MAX_TODOS_PER_USER_NOTIFICATION` (10) | — |

### Frontend (`apps/frontend/.env.local`)

```bash
cd apps/frontend
cp .env.sample .env.local
```

| Variable | Purpose | Example |
|---|---|---|
| `VITE_CLERK_PUBLISHABLE_KEY` | Clerk publishable key (required — app throws without it) | `pk_test_...` |
| `VITE_API_URL` | Backend base URL | `http://localhost:8080` |
| `VITE_ENV` | `local \| development \| production` | `local` |

## Run locally

Install workspace deps once from the repo root, and build the shared contract packages (the frontend dev server waits on `@taskify/openapi` dist):

```bash
bun install
bun run build --filter=@taskify/openapi --filter=@taskify/zod
```

### Backend (`:8080`)

```bash
cd apps/backend
go run ./cmd/taskify
```

Or via the task runner: `task run`. Health check: `GET http://localhost:8080/status`. API docs UI: `GET http://localhost:8080/docs`.

### Frontend (`:3010`)

```bash
cd apps/frontend
bun run dev
# → http://localhost:3010 (strict port)
```

### Ports + CORS note

- Backend defaults to `:8080` (`TASKIFY_SERVER.PORT`), frontend dev server is pinned to `:3010` (`strictPort: true` in `vite.config.ts`).
- `TASKIFY_SERVER.CORS_ALLOWED_ORIGINS` **must** contain `http://localhost:3010` for browser calls to succeed (the sample already includes both `:3000` and `:3010`).

## Migrations (tern)

Migrations live in `apps/backend/internal/database/migrations/` (`001`–`008`). When `TASKIFY_PRIMARY.ENV != "local"`, the API server applies them automatically at boot. For manual control:

```bash
cd apps/backend
task migrations:new name=my_migration   # tern new -m ./internal/database/migrations
task migrations:up                      # needs TASKIFY_DB_DSN, prompts for confirmation
```

`TASKIFY_DB_DSN` is a standard Postgres DSN, e.g. `postgres://user:pass@localhost:5432/taskify?sslmode=disable`.

## Tests

### Backend

```bash
cd apps/backend
go test ./...
```

- Handler, cron orchestration, config, and lib unit tests are **Docker-free** (fake services, stub materializers, `httptest`-backed Echo contexts).
- Repository/integration tests use **testcontainers** Postgres (`internal/testing/SetupTestDB`) and require a working Docker daemon; without Docker those packages fail at container startup.
- CI runs `go mod verify`, `go vet ./...`, `go test -v ./...`, builds both binaries, and runs `golangci-lint`.

### Frontend & packages

```bash
bun x turbo run lint        # eslint across workspaces
bun x turbo run typecheck   # tsc
bun x turbo run build       # openapi + zod, then frontend vite build
```

Note: the frontend `test` script (`vite test`) is a placeholder — there is no configured unit-test runner for the SPA; verification is via `lint`, `typecheck`, and `build` in CI.

## Lint

```bash
# backend (uses .golangci.yml: goimports, govet, staticcheck, gosec, revive, …)
cd apps/backend && golangci-lint run --timeout=5m

# everything JS/TS from the root
bun run lint            # turbo run lint
bun run lint:fix
bun run format:check || true   # where configured
cd apps/backend && go vet ./... && go fmt ./...
```

## API overview

Base path: `/api/v1`. Auth: Clerk session header on everything except `GET /shared/:token`. Full contract in `packages/openapi/openapi.json`, browsable at `GET /docs`.

| Method & path | Description |
|---|---|
| `POST /api/v1/todos` | Create todo |
| `GET /api/v1/todos` | List todos (filters: status/tab, priority, `overdue`, search, dates, category; sort + pagination) |
| `GET /api/v1/todos/stats` | Todo counts (total/draft/active/completed/archived/overdue) |
| `PATCH /api/v1/todos/bulk` | Bulk update (status, …) |
| `DELETE /api/v1/todos/bulk` | Bulk delete |
| `PATCH /api/v1/todos/reorder` | Reorder (position) |
| `GET /api/v1/todos/:id` | Get one todo (with subtasks/attachments) |
| `PATCH /api/v1/todos/:id` | Update todo |
| `DELETE /api/v1/todos/:id` | Delete todo |
| `POST /api/v1/todos/:id/share` | Mint share token |
| `DELETE /api/v1/todos/:id/share` | Revoke share token |
| `GET /api/v1/shared/:token` | Public shared-todo read (no auth) |
| `POST /api/v1/todos/:id/comments` / `GET /api/v1/todos/:id/comments` | Add / list comments |
| `PATCH /api/v1/comments/:id` / `DELETE /api/v1/comments/:id` | Edit / delete comment |
| `POST /api/v1/categories` / `GET /api/v1/categories` | Create / list categories |
| `PATCH /api/v1/categories/:id` / `DELETE /api/v1/categories/:id` | Update / delete category |
| `POST /api/v1/todos/:id/attachments` | Upload attachment (10 MB body limit) |
| `DELETE /api/v1/todos/:id/attachments/:attachmentId` | Delete attachment |
| `GET /api/v1/todos/:id/attachments/:attachmentId/download` | Presigned download URL |
| `GET /status` | Health check |
| `GET /docs` | OpenAPI docs UI |

### Cron runner

```bash
cd apps/backend
go run ./cmd/cron list              # list jobs
go run ./cmd/cron due-date-reminders
go run ./cmd/cron overdue-notifications
go run ./cmd/cron weekly-reports
go run ./cmd/cron auto-archive
go run ./cmd/cron recurring-todos
```

Jobs enqueue asynq email tasks (dedupe-aware) or run DB passes directly (recurring materialization, auto-archive).

## Scripts

| Location | Command | What it does |
|---|---|---|
| root | `bun run dev` | `turbo run dev` (all persistent dev tasks) |
| root | `bun run build` / `lint` / `lint:fix` / `typecheck` | turbo fan-out to workspaces |
| backend | `task run` / `go run ./cmd/taskify` | start API |
| backend | `task migrations:new name=x` / `task migrations:up` | tern migrations |
| backend | `task tidy` | `go fmt` + `go mod tidy` + `go mod verify` |
| backend | `go test ./...` | backend test suite |
| frontend | `bun run dev` | wait for openapi dist, then Vite on `:3010` |
| frontend | `bun run build` / `start` / `preview` | build / serve `dist` |
| frontend | `bun run lint` / `lint:fix` / `typecheck` | eslint / tsc |
| cron | `go run ./cmd/cron <job>` | run one scheduled job |

## Deployment notes

- **Backend** (`apps/backend/nixpacks.toml`): installs Go, `go mod download`, builds `taskify` (`cmd/taskify`) and `cron` (`cmd/cron`) binaries with `CGO_ENABLED=0`; starts `./taskify`. Provide all `TASKIFY_*` env vars in the host environment; non-`local` `ENV` auto-migrates on boot.
- **Frontend** (root `nixpacks.toml`): installs with Bun, builds `@taskify/openapi` + `@taskify/zod`, then `apps/frontend` Vite build; serves `dist` via `bun run start` on `${PORT:-3000}`.
- **Infra expectations**: managed Postgres (e.g. Neon, `sslmode=require`), managed Redis (e.g. Upstash URL), S3 bucket + credentials, Resend key, Clerk keys, and a scheduler invoking the `cron` binary per job cadence (due reminders / overdue hourly–daily, weekly reports weekly, auto-archive daily, recurring-todos frequently).
- **Observability**: zerolog structured logs, New Relic APM + distributed tracing (optional license key), `/status` health endpoint with DB/Redis checks.

## Contributing

1. Branch from `main` (`feat/…`, `fix/…`, `chore/…`).
2. Keep diffs minimal and repo-idiomatic; run the relevant checks (`go test ./...`, `golangci-lint`, `turbo run lint typecheck build`).
3. Open a PR against `main` — CI must be green before merge.

## License

MIT — see [LICENSE](./LICENSE).
