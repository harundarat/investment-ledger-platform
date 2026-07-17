-include .env
export

.PHONY: dev-up migrate-up run test vet

dev-up:
	docker compose up -d postgres

migrate-up:
	goose -dir migrations postgres "$(DATABASE_URL)" up

run:
	go run ./cmd/api

test:
	go test ./...

vet:
	go vet ./...

