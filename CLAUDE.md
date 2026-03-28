# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Commands

```bash
go build ./...
go run ./cmd/api
go test ./...
go test ./internal/handler/... -run TestCompanyCreate        # single function
go test ./internal/handler/... -run TestCompanyCreate/success # single sub-test
go test ./... -count=1                                        # bypass cache
```

## Architecture

Handler → Service → Repository → Postgres. DI is manual — `main.go` constructs every dependency by hand. Adding a feature: model → repository → service → handler → wire in `main.go` → register route in `routes.go`.

## Patterns to follow

**Errors:** Repositories return `apperr.*` types (`NotFound`, `Internal`, `Conflict`, etc.). Never return raw errors from a repository. `writeError` in the handler maps these to HTTP status via `apperr.HTTPStatus`.

**PATCH requests:** Use pointer fields (`*string`, `*int`) in request structs so `nil` means "not provided". Service builds a `map[string]any` of non-nil fields and passes to `repo.Update`. See `UpdateJobRequest`.

**Slug fields:** Use `validate:"...,slug"` — not `alphanum`. The custom `slug` validator (registered in `handler/validate.go init()`) allows lowercase alphanumeric and hyphens, rejecting consecutive/leading/trailing hyphens.

**Handler testability:** Each handler holds a service interface (e.g. `companyServicer`) not a concrete type. Public constructors (`NewCompanyHandler`) still accept the concrete service. `export_test.go` exports `NewXyzHandlerWithService` variants for injecting stubs. All tests use `httptest` — no database required.

**Roles:** `NewRoleMiddleware("company")` must chain after `NewAuthMiddleware`. Extract claims with `middleware.GetUserID(ctx)` / `middleware.GetRole(ctx)`.

**Duplicate detection:** Use `pgconn.PgError` code `"23505"` to detect unique constraint violations and return `apperr.Conflict`. See `ApplicationService.Apply`.

**Ownership checks that depend on a lookup:** When a service method must look up a dependency (e.g. company) before checking ownership, only convert a `NotFound` result to `Forbidden` — don't swallow DB errors. Use `errors.As` to inspect the type: `if errors.As(err, &appErr) && appErr.Type == apperr.TypeNotFound { return apperr.Forbidden(...) }`. See `JobService.ListMine`.
