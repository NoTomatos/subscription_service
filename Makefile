BINARY_NAME=subscription-service
BIN_DIR=bin

.PHONY: build run test clean docker-up docker-down migrate

build:
	go build -o $(BIN_DIR)/$(BINARY_NAME) ./cmd/api/main.go

run:
	go run ./cmd/api/main.go

test:
	go test -v ./...

clean:
	rm -rf $(BIN_DIR)/
	go clean

docker-up:
	docker-compose up --build

docker-down:
	docker-compose down

migrate-up:
	migrate -path migrations -database "postgres://postgres:postgres@localhost:5432/subscription_db?sslmode=disable" up

migrate-down:
	migrate -path migrations -database "postgres://postgres:postgres@localhost:5432/subscription_db?sslmode=disable" down

swagger:
	swag init -g cmd/api/main.go -o docs