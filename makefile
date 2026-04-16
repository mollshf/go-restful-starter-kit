include	.env
export

run:
	PORT=$(PORT) go run ./cmd/app

migrate-up:
	goose -dir ./migrations postgres $(DATABASE_URL) up

migrate-down:
	goose -dir ./migrations postgres $(DATABASE_URL) down

migrate-status:
	goose -dir ./migrations postgres $(DATABASE_URL) status

migrate-create:
	goose -dir ./migrations create -s $(name) sql

sqlgen:
	sqlc generate