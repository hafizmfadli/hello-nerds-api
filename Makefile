# Include variables from the .envrc file
include .envrc

.PHONY: confirm
confirm:
	@echo -n 'Are you sure? [y/N]' && read ans && [ $${ans:-N} = y ]

## db/migrations/up: apply all up database migrations
.PHONY: db/migrations/up
db/migrations/up: confirm
	@echo 'Running up migrations...'
	migrate -path ./migrations -database ${HELLO_NERDS_DB_DSN} up

## db/migrations/down: apply all down database migrations
.PHONY: db/migrations/down
db/migrations/down: confirm
	@echo 'Running up migrations...'
	migrate -path ./migrations -database ${HELLO_NERDS_DB_DSN} down

## db/migrations/up: apply all up database migrations
.PHONY: db/migrations/clean
db/migrations/clean: confirm
	@echo 'Running up migrations...'
	migrate -path ./migrations -database ${HELLO_NERDS_DB_DSN} force 1


.PHONY: run
run:
	go run ./cmd/api -db-dsn "root:debezium@tcp(localhost:3306)/periplus_dev?parseTime=true"