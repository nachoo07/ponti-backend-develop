SHELL := /bin/bash

# Variables base
ROOT_DIR           := $(shell pwd)
VERSION            := 1.0
BUILD_DIR          := $(ROOT_DIR)/bin
DOCKER_COMPOSE_YML := $(ROOT_DIR)/docker-compose.yml

# Recomiendo usar variables de entorno para la base de datos
DB_URL             := postgres://${DB_USER}:${DB_PASSWORD}@${DB_HOST}:${DB_PORT}/${DB_NAME}?sslmode=${DB_SSL_MODE}
MIGRATIONS_DIR     := ./migrations
MIGRATIONS_NAME    := $(NAME)  # pasar NAME=nombre al crear

.PHONY: all bin-build run test bin-clean lint \
        build up down logs reset rebuild clean \
        run-api seed \
        migrate-up migrate-down migrate-force migrate-force-dc migrate-version migrate-create

# --------------------------------------------------
# Migraciones
# --------------------------------------------------
migrate-up:
	@echo "Running migrations..."
	@$(MIGRATE) up

migrate-down:
	@echo "Running migrations down..."
	@$(MIGRATE) down

# Crea una nueva migración usando la variable NAME
migrate-create:
	@echo "Creating migration $(MIGRATIONS_NAME)..."
	@migrate create -ext sql -dir $(MIGRATIONS_DIR) -seq $(MIGRATIONS_NAME)

# Recargar la última migración (bajar y subir en un acto, sin borrar volúmenes)
migrate-reload:
	@echo "Reloading last migration (down + up without losing data)..."
	@docker compose run --rm migrate \
		-path /migrations \
		-database "postgres://admin:admin@ponti-db:5432/ponti_api_db?sslmode=disable" \
		down 1
	@docker compose run --rm migrate \
		-path /migrations \
		-database "postgres://admin:admin@ponti-db:5432/ponti_api_db?sslmode=disable" \
		up 1
	@echo "Migration reloaded successfully!"
migrate-force:
	@echo "Forcing migration to -1..."
	@$(MIGRATE) force -1

migrate-version:
	@echo "Current migration version:"
	@$(MIGRATE) version

# Forzar migración en docker-compose pasando el número como argumento
migrate-force-dc:
	@if [ -z "$(filter-out $@,$(MAKECMDGOALS))" ]; then \
		echo "Error: You must pass the version number. Example: make migrate-force-dc 31"; \
		exit 1; \
	fi
	@echo "Forcing migration to version $(filter-out $@,$(MAKECMDGOALS))..."
	@docker compose run --rm migrate \
		-path /migrations \
		-database "postgres://admin:admin@ponti-db:5432/ponti_api_db?sslmode=disable" \
		force $(filter-out $@,$(MAKECMDGOALS))

%:
	@:

# Baja todo, borra volúmenes, recompila, levanta, limpia esquema y corre migraciones desde cero
migrate-reset-dc:
	@echo "Resetting everything: down -v, build, up -d, drop schema and run migrations from scratch..."
	@docker compose down -v --remove-orphans
	@docker compose build
	@docker compose up -d
	@echo "Waiting for DB (ponti-db) to be ready..."
	@until docker compose exec -T ponti-db pg_isready -U admin -d ponti_api_db >/dev/null 2>&1; do sleep 1; done
	@echo "Dropping schema with migrate drop -f..."
	@docker compose run --rm migrate \
		-path /migrations \
		-database "postgres://admin:admin@ponti-db:5432/ponti_api_db?sslmode=disable" \
		drop -f
	@echo "Running migrations from scratch..."
	@docker compose run --rm migrate \
		-path /migrations \
		-database "postgres://admin:admin@ponti-db:5432/ponti_api_db?sslmode=disable" \
		up
	@echo "Reset completed."

# --------------------------------------------------
# Compilación y ejecución
# --------------------------------------------------
all: build run

bin-build:
	@echo "Building the project..."
	@mkdir -p $(BUILD_DIR)
	APP_NAME=ponti-api go build -gcflags "all=-N -l" \
		-o $(BUILD_DIR)/ponti-api \
		-ldflags "-X main.Version=$(VERSION)" \
		$(ROOT_DIR)/cmd/

run:
	@echo "Running the project (local)..."
	@go run $(ROOT_DIR)/cmd/

test:
	@echo "Running tests..."
	@go test ./...

bin-clean:
	@echo "Cleaning up build artifacts..."
	@rm -rf $(BUILD_DIR)

lint:
	@echo "Linting the project..."
	@golangci-lint run --config .golangci.yml --verbose

run-api:
	@echo "Starting API server..."
	@go run ./cmd/api/

seed:
	@echo "Seeding database..."
	@go run ./cmd/seed/main.go

seed-dashboard:
	@echo "Seeding dashboard test data..."
	@go run ./cmd/api/main.go seed

# --------------------------------------------------
# Docker Compose
# --------------------------------------------------
up:
	@echo "Starting services (compose up)..."
	docker compose -f $(DOCKER_COMPOSE_YML) up

down:
	@echo "Stopping services (compose down)..."
	docker compose -f $(DOCKER_COMPOSE_YML) down --remove-orphans

reset: down up
	@echo "Reset done (down & up)."

build:
	@echo "Building services (compose up --build)..."
	docker compose -f $(DOCKER_COMPOSE_YML) up --build

clean:
	@echo "Cleaning: stopping services, removing containers, volumes, networks and build artifacts..."
	@docker compose -f $(DOCKER_COMPOSE_YML) down -v --remove-orphans
	@docker system prune -f --volumes
	@rm -rf $(BUILD_DIR)
	@echo "Clean completed."

rebuild: clean build
	@echo "Rebuild completed (clean + build)."

logs:
	@echo "Tailing logs (compose logs -f)..."
	docker compose -f $(DOCKER_COMPOSE_YML) logs -f