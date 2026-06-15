# Makefile for subian Go Project

.PHONY: build run migrate migrate-dev migrate-prod migrate-sql migrate-sql-prod \
        clean create-migration db-stats migrate-fresh-dev migrate-fresh-dev-sql \
        migrate-fresh-prod migrate-fresh-prod-sql seed seed-prod seed-fresh \
        migrate-seed migrate-fresh-seed gen-jwt-dev gen-jwt-prod test test-auth \
        test-users test-rbac swagger-install swagger-gen swagger swagger-fmt \
        run-api gen-module help

# ─── Build & Run ───────────────────────────────────────────────────────────────

build: ## Build the API binary
	go build -o bin/api ./cmd/api

run: ## Run the API locally
	go run ./cmd/api/main.go

# ─── Migrations ────────────────────────────────────────────────────────────────

migrate-dev: ## Run GORM auto-migration for DEV
	@echo "Running GORM auto-migration (development)..."
	go run ./cmd/migrate/main.go -env=DEV -type=gorm

migrate-prod: ## Run GORM auto-migration for PROD
	@echo "Running GORM auto-migration (production)..."
	go run ./cmd/migrate/main.go -env=PROD -type=gorm

migrate-sql: ## Run SQL-based migrations for DEV
	@echo "Running SQL-based migrations (development)..."
	go run ./cmd/migrate/main.go -env=DEV -type=sql

migrate-sql-prod: ## Run SQL-based migrations for PROD
	@echo "Running SQL-based migrations (production)..."
	go run ./cmd/migrate/main.go -env=PROD -type=sql

create-migration: ## Create a new migration file
	@echo "Creating new migration file..."
	@read -p "Enter migration name: " name; \
	touch internal/module/$$name/migrations/$$(date +%Y%m%d%H%M%S)_$$name.go

db-stats: ## Show database connection stats
	go run -c 'package main; import ("fmt"; "subian_go/config"); func main() { cfg := config.LoadConfig("DEV"); db, _ := cfg.ConnectDB(); defer config.CloseDB(db); stats, _ := config.GetDBStats(db); fmt.Printf("%+v\n", stats) }'

# ─── Fresh Migrations (Drop All + Re-migrate) ─────────────────────────────────

migrate-fresh-dev: ## Drop all tables and re-migrate (DEV)
	@echo "⚠️  WARNING: This will DROP ALL TABLES on DEV and re-migrate!"
	@read -p "Are you sure? (yes/no): " confirm; \
	if [ "$$confirm" = "yes" ]; then \
		echo "🔄 Running fresh migration (DEV)..."; \
		go run ./cmd/migrate/main.go -env=DEV -type=gorm -fresh=true; \
	else \
		echo "❌ Aborted."; \
	fi

migrate-fresh-dev-sql: ## Drop all tables and re-migrate SQL (DEV)
	@echo "⚠️  WARNING: This will DROP ALL TABLES on DEV and re-migrate!"
	@read -p "Are you sure? (yes/no): " confirm; \
	if [ "$$confirm" = "yes" ]; then \
		echo "🔄 Running fresh SQL migration (DEV)..."; \
		go run ./cmd/migrate/main.go -env=DEV -type=sql -fresh=true; \
	else \
		echo "❌ Aborted."; \
	fi

migrate-fresh-prod: ## Drop all tables and re-migrate (PROD) - DANGER
	@echo "🚨 DANGER: This will DROP ALL TABLES on PRODUCTION!"
	@echo "🚨 This action is IRREVERSIBLE!"
	@read -p "Type 'PRODUCTION' to confirm: " confirm; \
	if [ "$$confirm" = "PRODUCTION" ]; then \
		read -p "Are you absolutely sure? (yes/no): " confirm2; \
		if [ "$$confirm2" = "yes" ]; then \
			echo "🔄 Running fresh migration (PROD)..."; \
			go run ./cmd/migrate/main.go -env=PROD -type=gorm -fresh=true -force=true; \
		else \
			echo "❌ Aborted."; \
		fi \
	else \
		echo "❌ Aborted. You must type 'PRODUCTION' exactly."; \
	fi

migrate-fresh-prod-sql: ## Drop all tables and re-migrate SQL (PROD) - DANGER
	@echo "🚨 DANGER: This will DROP ALL TABLES on PRODUCTION!"
	@echo "🚨 This action is IRREVERSIBLE!"
	@read -p "Type 'PRODUCTION' to confirm: " confirm; \
	if [ "$$confirm" = "PRODUCTION" ]; then \
		read -p "Are you absolutely sure? (yes/no): " confirm2; \
		if [ "$$confirm2" = "yes" ]; then \
			echo "🔄 Running fresh SQL migration (PROD)..."; \
			go run ./cmd/migrate/main.go -env=PROD -type=sql -fresh=true -force=true; \
		else \
			echo "❌ Aborted."; \
		fi \
	else \
		echo "❌ Aborted. You must type 'PRODUCTION' exactly."; \
	fi

# ─── Seeder ────────────────────────────────────────────────────────────────────

seed: ## Run seeder (DEV)
	@echo "🌱 Menjalankan seeder..."
	go run ./cmd/seed/main.go -env=DEV

seed-prod: ## Run seeder (PROD)
	@echo "🌱 Menjalankan seeder (PROD)..."
	go run ./cmd/seed/main.go -env=PROD

seed-fresh: ## Delete all data and re-seed (DEV)
	@echo "⚠️  WARNING: Ini akan menghapus semua data dan seed ulang!"
	@read -p "Are you sure? (yes/no): " confirm; \
	if [ "$$confirm" = "yes" ]; then \
		echo "🔄 Running fresh seed..."; \
		go run ./cmd/seed/main.go -env=DEV -fresh=true; \
	else \
		echo "❌ Aborted."; \
	fi

migrate-seed: ## Migrate and seed (DEV)
	@echo "🔄 Migrate + seed (DEV)..."
	go run ./cmd/migrate/main.go -env=DEV -type=gorm
	go run ./cmd/seed/main.go -env=DEV

migrate-fresh-seed: ## Fresh migrate and seed (DEV)
	@echo "⚠️  WARNING: Drop semua tabel, migrate ulang, dan seed!"
	@read -p "Are you sure? (yes/no): " confirm; \
	if [ "$$confirm" = "yes" ]; then \
		echo "🔄 Fresh migrate + seed..."; \
		go run ./cmd/migrate/main.go -env=DEV -type=gorm -fresh=true; \
		go run ./cmd/seed/main.go -env=DEV; \
	else \
		echo "❌ Aborted."; \
	fi

# ─── JWT Secret Generation ─────────────────────────────────────────────────────

gen-jwt-dev: ## Generate and update JWT_SECRET for DEV
	@echo "Generating JWT Secret for DEV..."
	@NEW_SECRET=$$(openssl rand -base64 32 | head -c 32); \
	if [ -f config/.env.dev ]; then \
		sed -i "s|^JWT_SECRET=.*|JWT_SECRET=$$NEW_SECRET|" config/.env.dev; \
		echo "✅ JWT_SECRET berhasil diupdate di config/.env.dev"; \
	else \
		echo "❌ File config/.env.dev tidak ditemukan!"; \
		echo "Secret baru Anda: $$NEW_SECRET"; \
	fi

gen-jwt-prod: ## Generate and update JWT_SECRET for PROD
	@echo "Generating JWT Secret for PROD..."
	@NEW_SECRET=$$(openssl rand -base64 48 | head -c 48); \
	if [ -f config/.env.prod ]; then \
		sed -i "s|^JWT_SECRET=.*|JWT_SECRET=$$NEW_SECRET|" config/.env.prod; \
		echo "✅ JWT_SECRET berhasil diupdate di config/.env.prod"; \
	else \
		echo "❌ File config/.env.prod tidak ditemukan!"; \
		echo "Secret baru Anda: $$NEW_SECRET"; \
	fi

# ─── Swagger ───────────────────────────────────────────────────────────────────

swagger-install: ## Install swag CLI
	@echo "📦 Installing swag CLI..."
	go install github.com/swaggo/swag/cmd/swag@latest
	@echo "✅ swag installed. Pastikan $(go env GOPATH)/bin ada di PATH."

GO_BIN := $(shell go env GOPATH)/bin

swagger-gen: ## Generate swagger docs from annotations
	@echo "📝 Generating Swagger docs..."
	$(GO_BIN)/swag init \
		--generalInfo cmd/api/main.go \
		--output docs \
		--parseDependency \
		--parseInternal
	@echo "✅ Swagger docs generated di folder docs/"

swagger: ## Generate swagger docs and run API
	@make swagger-gen
	@make run

swagger-fmt: ## Format swagger comments
	swag fmt

run-api: ## Generate Swagger docs and run the API
	@echo "📝 Generating Swagger docs..."
	$(GO_BIN)/swag init \
		--generalInfo cmd/api/main.go \
		--output docs \
		--parseDependency \
		--parseInternal
	@echo "✅ Swagger docs generated di folder docs/"
	go run ./cmd/api/main.go

# ─── Code Generator ────────────────────────────────────────────────────────────

gen-module: ## Generate a new module (usage: make gen-module name=artikel add=kategori)
	@echo "🚀 Generating new module..."
	@if [ -z "$(name)" ]; then \
		echo "❌ Error: 'name' is required. Usage: make gen-module name=artikel [add=kategori]"; \
		exit 1; \
	fi
	@if [ -z "$(add)" ]; then \
		echo "📦 Creating main module: $(name)"; \
		go run generator.go -name=$(name); \
	else \
		echo "📦 Creating sub-module: $(name)/$(add)"; \
		go run generator.go -name=$(name) -add=$(add); \
	fi

# ─── Testing ───────────────────────────────────────────────────────────────────

test-auth: ## Run tests for auth module
	@go test -json ./internal/modules/auth/tests | gotestfmt

test-users: ## Run tests for users module
	@go test -json ./internal/modules/users/tests | gotestfmt

test-rbac: ## Run tests for rbac module
	@go test -json ./internal/modules/rbac/tests | gotestfmt

test: ## Run all tests
	@go test -json ./internal/modules/.../tests | gotestfmt
	@go test -json ./internal/modules/.../.../tests | gotestfmt

# ─── Utilities ─────────────────────────────────────────────────────────────────

clean: ## Clean build artifacts
	rm -rf bin/

# ─── Help ──────────────────────────────────────────────────────────────────────

help: ## Display this help message
	@echo "📚 Available commands in Makefile:"
	@echo ""
	@grep -E '^[a-zA-Z_-]+: ## ' $(MAKEFILE_LIST) | awk -F': ## ' '{printf "\033[36m%-22s\033[0m %s\n", $$1, $$2}' | sort
	@echo ""
	@echo "💡 Examples:"
	@echo "   make gen-module name=artikel"
	@echo "   make gen-module name=artikel add=kategori"
	@echo "   make help"