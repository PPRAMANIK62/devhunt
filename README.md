# devhunt

A job board built with Go (primary) and a React frontend. Companies post jobs, seekers apply.

## Stack

**Backend (primary)**

- **Go 1.25** — chi router, pgx/v5 (postgres), golang-jwt, go-playground/validator
- **Postgres** — all persistence

**Frontend (`web/`)**

- **React 19** + TypeScript + Vite + Tailwind v4
- **shadcn/ui** (Radix UI primitives) — thin wrapper over the Go API, no client-side business logic

## Setup

```bash
cp .env.example .env  # fill in values
go run ./cmd/api      # API on :8080
```

### Frontend

```bash
cd web
bun install
bun dev              # dev server on :5173, proxies /api/v1 → :8080
bun run build        # production build → web/dist/
```

### Seed data

Populate the database with 7 companies, 50+ jobs, and a seeker account:

```bash
go run ./cmd/seed
```

Safe to re-run — wipes and recreates seed rows each time.

Seed accounts (password: `password123`):

- **seeker** — `seeker@example.com`
- **company** — `acme@example.com`, `devstudio@example.com`, `finledger@example.com`, `neural-labs@example.com`, `forge-tools@example.com`, `pixel-agency@example.com`, `cloudnine@example.com`

### Environment variables

| Variable             | Required | Default       | Description                   |
| -------------------- | -------- | ------------- | ----------------------------- |
| `DATABASE_URL`       | yes      | —             | Postgres connection string    |
| `JWT_SECRET`         | yes      | —             | Signing secret for JWT tokens |
| `SERVER_PORT`        | no       | `8080`        | HTTP listen port              |
| `JWT_EXPIRY_MINUTES` | no       | `10`          | Token lifetime                |
| `ENV`                | no       | `development` | Environment name              |

## API

All routes are under `/api/v1`.

### Auth

```
POST /auth/register   body: { email, password, role: "seeker"|"company" }
POST /auth/login      body: { email, password }  → { token, user }
```

### Jobs

```
GET    /jobs               public, paginated (?page=&page_size=&q=&location=&tag=&min_salary=)
GET    /jobs/filters       public — distinct locations + tags for open jobs
GET    /jobs/{id}          public
POST   /jobs               company only
PATCH  /jobs/{id}          company only, ownership enforced
DELETE /jobs/{id}          company only, ownership enforced
```

### Companies

```
GET    /companies/{id}     public
POST   /companies          company only, one per user
GET    /companies/me       company only
PATCH  /companies/me       company only
DELETE /companies/me       company only
GET    /companies/me/jobs  company only (?status=open|draft|closed)
```

### Applications

```
POST   /jobs/{jobID}/applications    auth required (seeker applies)
GET    /jobs/{jobID}/applications    company only, ownership enforced
GET    /applications                 auth required (own applications)
PATCH  /applications/{id}/status     company only, ownership enforced
```

Application statuses: `pending` → `reviewed` → `accepted` | `rejected`

## Testing

Tests are organised in three layers:

| Layer      | Package               | What it tests                                  | Needs Docker?          |
| ---------- | --------------------- | ---------------------------------------------- | ---------------------- |
| Handler    | `internal/handler`    | HTTP routing, middleware, request/response     | integration tests only |
| Service    | `internal/service`    | Business logic, ownership rules, error mapping | no                     |
| Repository | `internal/repository` | SQL queries against a real schema              | yes                    |

Repository and handler integration tests spin up a **Postgres 15 container automatically** via [dockertest](https://github.com/ory/dockertest) — no manual setup required, just Docker running.

```bash
# All tests (requires Docker for integration tests)
go test ./... -count=1 -timeout 120s

# Unit tests only (no Docker)
go test ./internal/service/... -v

# Integration tests
go test ./internal/repository/... -v -timeout 120s
go test ./internal/handler/... -v -timeout 120s

# Single test
go test ./internal/service/... -run TestAuthService_Login_WrongPassword -v
```
