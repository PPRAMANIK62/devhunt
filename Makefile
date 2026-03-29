-include .env
export

.PHONY: migrate-up migrate-down migrate-status migrate-create up down down-volumes logs ps seed test

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

up:
	docker compose up --build

down:
	docker compose down

down-volumes:
	docker compose down -v   # also deletes data volumes — fresh start

logs:
	docker compose logs -f app

ps:
	docker compose ps

seed:
	go run ./cmd/seed

test:
	docker compose -f docker-compose.test.yml up -d --wait
	go test ./... -count=1; docker compose -f docker-compose.test.yml down

docs:
	swag init -g cmd/api/main.go -o docs/
	@echo "Docs generated. Visit http://localhost:8080/docs/index.html"
