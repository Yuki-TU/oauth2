.DEFAULT_GOAL := help

DB_USER := oauth2_user
DB_PASSWORD := oauth2_password
DB_NAME := oauth2_db

.PHONY: up
up: ## Start the services
	docker compose up -d

.PHONY: down
down: ## Stop the services
	docker compose down

.PHONY: logs
logs: ## View logs
	docker compose logs -f

.PHONY: test
test: ## Run tests
	go test -v ./...

.PHONY: run
run: ## Run the application
	go run main.go

.PHONY: db
db: ## Run the database
	docker compose exec postgres psql -U $(DB_USER) -d $(DB_NAME)

.PHONY: create-key
create-key: ## JWTに必要なキーを作成
	@mkdir -p ./certificate
	@openssl genrsa 4096 > ./certificate/secret.pem
	@echo "Created secret.pem"
	@openssl rsa -pubout < ./certificate/secret.pem > ./certificate/public.pem
	@echo "Created public.pem"

.PHONY: help
help: ## Show help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'
