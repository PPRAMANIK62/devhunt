-include .env
export

.PHONY: migrate-up migrate-down migrate-status migrate-create

migrate-up:
	goose -dir internal/database/migrations postgres "$(DATABASE_URL)" up

migrate-down:
	goose -dir internal/database/migrations postgres "$(DATABASE_URL)" down

migrate-status:
	goose -dir internal/database/migrations postgres "$(DATABASE_URL)" status

migrate-create:
	goose -dir internal/database/migrations create $(name) sql
	# Usage: make migrate-create name=add_column_to_jobs
	# NOTE: goose generates a timestamp-prefixed filename (e.g. 20260329_add_column_to_jobs.sql).
	# Rename it to the sequential NNNNN_ format (e.g. 00003_add_column_to_jobs.sql) to stay consistent.
