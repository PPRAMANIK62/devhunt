# devhunt

A job board REST API built with Go. Companies post jobs, seekers apply.

## Stack

- **Go 1.25** — chi router, pgx/v5 (postgres), golang-jwt, go-playground/validator
- **Postgres** — all persistence

## Setup

```bash
cp .env.example .env  # fill in values
go run ./cmd/api
```

### Environment variables

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `DATABASE_URL` | yes | — | Postgres connection string |
| `JWT_SECRET` | yes | — | Signing secret for JWT tokens |
| `SERVER_PORT` | no | `8080` | HTTP listen port |
| `JWT_EXPIRY_MINUTES` | no | `10` | Token lifetime |
| `ENV` | no | `development` | Environment name |

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

```bash
go test ./...           # all tests
go test ./... -v        # with output
go test ./... -count=1  # bypass cache
```

Tests are handler-level using `net/http/httptest` — no database required.
