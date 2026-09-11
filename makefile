include .env
export

migrate-up:
	migrate -path migrations -database "postgresql://${POSTGRES_USER}:${POSTGRES_PASSWORD}@localhost:5432/${POSTGRES_DB}?sslmode=disable" -verbose up

migrate-down:
	migrate -path migrations -database "postgresql://${POSTGRES_USER}:${POSTGRES_PASSWORD}@localhost:5432/${POSTGRES_DB}?sslmode=disable" -verbose down

migrate-up-one:
	migrate -path migrations -database "postgresql://${POSTGRES_USER}:${POSTGRES_PASSWORD}@localhost:5432/${POSTGRES_DB}?sslmode=disable" -verbose up 1

migrate-down-one:
	migrate -path migrations -database "postgresql://${POSTGRES_USER}:${POSTGRES_PASSWORD}@localhost:5432/${POSTGRES_DB}?sslmode=disable" -verbose down 1

migrate-version:
	migrate -path migrations -database "postgresql://${POSTGRES_USER}:${POSTGRES_PASSWORD}@localhost:5432/${POSTGRES_DB}?sslmode=disable" version

create-migration:
	migrate create -ext sql -dir migrations -seq $(name)

test:
	set -a; . ./.env; set +a; go test ./... -v