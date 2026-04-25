.PHONY: run test lint migrate-up migrate-down docker-up docker-down

run:
	go run ./cmd/server

test:
	go test ./... -v -count=1

lint:
	golangci-lint run ./...

migrate-up:
	migrate -path migrations -database "$$DATABASE_URL" up

migrate-down:
	migrate -path migrations -database "$$DATABASE_URL" down

docker-up:
	docker compose up -d --build

docker-down:
	docker compose down
