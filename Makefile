APP_NAME=gym-management-api
DB_URL=postgres://postgres:postgres@postgres:5432/gym_management?sslmode=disable

.PHONY: dev up down logs test fmt vet migrate-up migrate-down migrate-force migrate-create

dev:
	docker compose up app

up:
	docker compose up -d

down:
	docker compose down

logs:
	docker compose logs -f app

test:
	go test ./...

fmt:
	go fmt ./...

vet:
	go vet ./...

migrate-up:
	docker compose exec app migrate -path migrations -database "$(DB_URL)" up

migrate-down:
	docker compose exec app migrate -path migrations -database "$(DB_URL)" down 1

migrate-force:
	docker compose exec app migrate -path migrations -database "$(DB_URL)" force $(version)

migrate-create:
	docker compose exec app migrate create -ext sql -dir migrations -seq $(name)
