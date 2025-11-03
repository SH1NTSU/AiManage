# Database connection string
DB_URL=postgresql://postgres:postgres@localhost:5432/ai_db?sslmode=disable

# Run all pending migrations
migrate-up:
	migrate -path migrations -database "$(DB_URL)" up

# Rollback the last migration
migrate-down:
	migrate -path migrations -database "$(DB_URL)" down 1

# Create a new migration file (will prompt for name)
migrate-create:
	@read -p "Enter migration name: " name; \
	migrate create -ext sql -dir migrations -seq $$name

.PHONY: migrate-up migrate-down migrate-create
