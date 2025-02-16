.DEFAULT_GOAL := help

PROJECT_NAME := Platform Go Challenge
DOCKER_COMPOSE_FILE := build/docker-compose.yaml
DOCKER_COMPOSE_TEST_FILE := build/docker-compose-test.yaml
GO_FILES := $(shell find . -type f -name '*.go')
COMPOSE_BAKE := COMPOSE_BAKE=true

#-----------------------------------------------------------------------
# Help
#-----------------------------------------------------------------------
.PHONY: help
help:
	@echo "------------------------------------------------------------------------"
	@echo "${PROJECT_NAME}"
	@echo "------------------------------------------------------------------------"
	@grep -E '^[a-zA-Z0-9_/%\-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

#-----------------------------------------------------------------------
# Code Quality & Security
#-----------------------------------------------------------------------
.PHONY: format vet vulncheck check-all
format: ## Format Go code
		@echo "Formatting code..."
		@gofmt -w $(GO_FILES)
		@echo "Done formatting."

vet: ## Run go vet
		@echo "Running go vet..."
		@go vet ./...
		@echo "Code looks good!"

vulncheck: ## Check for known vulnerabilities
		@echo "Checking for vulnerabilities..."
		@go install golang.org/x/vuln/cmd/govulncheck@latest
		@govulncheck ./...

check-all: format vet vulncheck ## Run all code quality checks
		@echo "All quality checks completed!"

#-----------------------------------------------------------------------
# Database Management )
#-----------------------------------------------------------------------
.PHONY: db-up db-down
db-up: ## Start app database in container
	docker compose -f $(DOCKER_COMPOSE_FILE) up -d db --remove-orphans
	@sleep 2 # give the db time to start

db-down: ## Stop and remove app  database container and volume
	docker compose -f $(DOCKER_COMPOSE_FILE) down -v db
db-test-up: ## Start test database in container
	docker compose -f $(DOCKER_COMPOSE_TEST_FILE) up -d test_db --remove-orphans
	@sleep 2

db-test-down: ## Stop and remove test database container and volume
	docker compose -f $(DOCKER_COMPOSE_TEST_FILE) down -v test_db

#-----------------------------------------------------------------------
# Local Development
#-----------------------------------------------------------------------
.PHONY: run
run: db-up ## Run the application locally
	go run cmd/pgc/main.go cmd/pgc/setup.go

.PHONY: test-unit
test-unit: ## Run unit tests
	go test -short -v -count=1 -race -cover ./...

.PHONY: test-it
test-it: db-test-up ## Run integration tests locally
	go test -v -count=1 -race ./tests/integration/...

.PHONY: test-e2e
test-e2e: db-test-up ## Run end-to-end tests locally
	go test -v -count=1 -race ./tests/e2e/...

.PHONY: test-all
test-all: ## Run all tests locally
	@echo "Running unit tests..."
	@make test-unit
	@echo "Running integration tests..."
	@make test-it
	@echo "Running end-to-end tests..."
	@make test-e2e
	@make db-down

#-----------------------------------------------------------------------
# Container Operations
#-----------------------------------------------------------------------
.PHONY: docker-build docker-up docker-down docker-logs
docker-build: ## Build docker images if needed
	$(COMPOSE_BAKE) docker compose -f $(DOCKER_COMPOSE_FILE) build --no-cache
	$(COMPOSE_BAKE) docker compose -f $(DOCKER_COMPOSE_TEST_FILE) build --no-cache

docker-up: ## Start the application and databases in containers
	$(COMPOSE_BAKE) docker compose -f $(DOCKER_COMPOSE_FILE) up -d

docker-down: ## Stop and remove all containers and volumes
	docker compose -f $(DOCKER_COMPOSE_FILE) down -v --remove-orphans
	docker compose -f $(DOCKER_COMPOSE_TEST_FILE) down -v --remove-orphans

docker-logs: ## View application logs
	docker compose -f $(DOCKER_COMPOSE_FILE) logs -f app

#-----------------------------------------------------------------------
# Container Tests
#-----------------------------------------------------------------------
.PHONY: test-unit-docker test-it-docker test-e2e-docker test-all-docker
test-unit-docker: ## Run unit tests in Docker
	$(COMPOSE_BAKE) docker compose -f $(DOCKER_COMPOSE_TEST_FILE) run --rm test_runner go test -short -v -count=1 -race -cover ./...

test-it-docker: ## Run integration tests in Docker
	$(COMPOSE_BAKE) docker compose -f $(DOCKER_COMPOSE_TEST_FILE) up -d test_db
	$(COMPOSE_BAKE) docker compose -f $(DOCKER_COMPOSE_TEST_FILE) run --rm test_runner go test -v ./tests/integration/...
	@make docker-down

test-e2e-docker: ## Run e2e tests in Docker
	$(COMPOSE_BAKE) docker compose -f $(DOCKER_COMPOSE_TEST_FILE) up -d test_db
	$(COMPOSE_BAKE) docker compose -f $(DOCKER_COMPOSE_TEST_FILE) run --rm test_runner go test -v ./tests/e2e/...
	@make docker-down

test-all-docker: ## Run all tests in Docker
	$(COMPOSE_BAKE) docker compose -f $(DOCKER_COMPOSE_TEST_FILE) up -d test_db
	$(COMPOSE_BAKE) docker compose -f $(DOCKER_COMPOSE_TEST_FILE) run --rm test_runner go test -v -count=1 -race ./tests/integration/... ./tests/e2e/...
	@make docker-down
