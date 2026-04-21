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
	env $(cat .env | xargs) go run *.go

.PHONY: build
build: ## Build the application
	go build -o oauth2-server *.go

.PHONY: deps
deps: ## Download and install dependencies
	go mod tidy
	go mod download

.PHONY: db
db: ## Run the database
	docker compose exec postgres psql -U $(DB_USER) -d $(DB_NAME)

.PHONY: db-sync-demo-redirects
db-sync-demo-redirects: ## 起動済み DB に init.sql を再適用（スキーマ・シード・デモ redirect を冪等に揃える）
	docker compose exec -T postgres psql -U $(DB_USER) -d $(DB_NAME) < init.sql

.PHONY: create-key
create-key: ## JWTに必要なキーを作成
	@mkdir -p ./certificate
	@openssl genrsa 4096 > ./certificate/secret.pem
	@echo "Created secret.pem"
	@openssl rsa -pubout < ./certificate/secret.pem > ./certificate/public.pem
	@echo "Created public.pem"

.PHONY: client-dotenv
client-dotenv: ## client/.env.local が無ければ env.example からコピー（既存は上書きしない）
	@if [ ! -f client/.env.local ]; then cp client/env.example client/.env.local && echo "Created client/.env.local from client/env.example"; fi

.PHONY: client-install
client-install: ## Install Next.js demo client dependencies
	cd client && npm install

.PHONY: client-dev
client-dev: client-dotenv ## Next.js 開発サーバー（:3000）。初回は .env.local を env.example から作成
	cd client && npm run dev

.PHONY: backend-dotenv
backend-dotenv: ## backend/.env が無ければ .env.example からコピー（既存は上書きしない）
	@if [ ! -f backend/.env ]; then cp backend/.env.example backend/.env && echo "Created backend/.env from backend/.env.example"; fi

.PHONY: backend-run
backend-run: backend-dotenv ## リソースサーバー（JWKS で JWT 検証）。:8080 起動後に（env + .env はルートの make run と同型）
	cd backend && \
	if grep -q '^[^#[:space:]]' .env 2>/dev/null; then \
	  env $$(grep -v '^[[:space:]]*#' .env | grep -v '^[[:space:]]*$$' | xargs) go run .; \
	else \
	  go run .; \
	fi

.PHONY: help
help: ## Show help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'
