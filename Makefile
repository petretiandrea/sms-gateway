APP_NAME ?= sms-gateway
BINARY ?= bin/$(APP_NAME)
POSTGRES_DSN ?= postgres://sms_gateway:sms_gateway@localhost:5432/sms_gateway?sslmode=disable
RABBITMQ_DSN ?= amqp://sms_gateway:sms_gateway@localhost:5672/
MONGODB_URI ?= mongodb://localhost:27017/smsgateway?directConnection=true
MONGODB_DATABASE ?= smsgateway
CHART_DIR ?= chart
VALUES_FILE ?=

.PHONY: help
help:
	@echo "Available targets:"
	@echo "  deps                  Download Go dependencies"
	@echo "  install-delve         Install or update Delve debugger"
	@echo "  tidy                  Tidy Go modules"
	@echo "  fmt                   Format Go files"
	@echo "  test                  Run Go tests"
	@echo "  build                 Build the application binary"
	@echo "  run                   Run the HTTP service locally"
	@echo "  postgres-up           Start local Postgres with Docker Compose"
	@echo "  postgres-down         Stop local Postgres"
	@echo "  postgres-logs         Tail local Postgres logs"
	@echo "  rabbitmq-up           Start local RabbitMQ with Docker Compose"
	@echo "  rabbitmq-logs         Tail local RabbitMQ logs"
	@echo "  migrate-service       Run database migrations in Docker Compose"
	@echo "  outbox-up             Start local outbox worker with Docker Compose"
	@echo "  outbox-logs           Tail local outbox logs"
	@echo "  infra-up              Start local Postgres, RabbitMQ, and outbox"
	@echo "  infra-down            Stop local infrastructure"
	@echo "  migrate-up            Apply database migrations"
	@echo "  migrate-down          Roll back all database migrations"
	@echo "  migrate-version       Print current migration version"
	@echo "  rabbitmq-topology     Apply RabbitMQ topology"
	@echo "  mongodb-copy-to-postgres Copy legacy MongoDB data into Postgres"
	@echo "  chart-template        Render the Helm chart"
	@echo "  chart-lint            Lint the Helm chart"
	@echo "  verify                Run fmt, tidy, tests, and chart lint"

.PHONY: deps
deps:
	go mod download

.PHONY: install-delve
install-delve:
	go install github.com/go-delve/delve/cmd/dlv@latest

.PHONY: tidy
tidy:
	go mod tidy

.PHONY: fmt
fmt:
	go fmt ./...

.PHONY: test
test:
	go test ./...

.PHONY: build
build:
	go build -o $(BINARY) ./cmd

.PHONY: run
run:
	go run ./cmd start

.PHONY: migrate-service
migrate-service:
	docker compose up migrate

.PHONY: migrate-up
migrate-up:
	go run ./cmd migrate --database "$(POSTGRES_DSN)" up

.PHONY: migrate-down
migrate-down:
	go run ./cmd migrate --database "$(POSTGRES_DSN)" down

.PHONY: migrate-version
migrate-version:
	go run ./cmd migrate --database "$(POSTGRES_DSN)" version

.PHONY: rabbitmq-topology
rabbitmq-topology:
	go run ./cmd rabbitmq topology apply --dsn "$(RABBITMQ_DSN)"

.PHONY: mongodb-copy-to-postgres
mongodb-copy-to-postgres:
	go run ./cmd mongodb copy-to-postgres --mongo-uri "$(MONGODB_URI)" --mongo-database "$(MONGODB_DATABASE)" --postgres-dsn "$(POSTGRES_DSN)"

.PHONY: chart-template
chart-template:
	helm template $(APP_NAME) ./$(CHART_DIR) $(if $(VALUES_FILE),-f $(VALUES_FILE),)

.PHONY: chart-lint
chart-lint:
	helm lint ./$(CHART_DIR)

.PHONY: verify
verify: fmt tidy test chart-lint
