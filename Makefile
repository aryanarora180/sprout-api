include .envrc

# DEVELOPMENT ----------------------------------------------------------------------------------------------------------

.PHONY run/api:
run/api:
	@echo 'Running API...'
	go run ./cmd/api -db-dsn=${SPROUT_DB_DSN} -smtp-username=${SMTP_USERNAME} -smtp-password=${SMTP_PASSWORD}

.PHONY db/migrations/new:
db/migrations/new:
	@echo 'Creating migration files for ${name}'
	migrate create -seq -ext=.sql -dir=./migrations ${name}

.PHONY db/migrations/up:
db/migrations/up:
	@echo 'Running up migrations'
	migrate -path=./migrations -database=${SPROUT_DB_DSN} up

.PHONY db/psql:
db/psql:
	psql ${SPROUT_DB_DSN}

# QUALITY CONTROL ------------------------------------------------------------------------------------------------------

.PHONY audit:
audit:
	@echo 'Tidying and verifying module dependencies'
	go mod tidy
	go mod verify
	@echo 'Formatting code'
	go fmt ./...
	@echo 'Vetting code'
	go vet ./...

# BUILD ----------------------------------------------------------------------------------------------------------------

git_description = $(shell git describe --always --dirty)
current_time = $(shell date -u +%FT%T%z)
linker_flags = '-s -X main.version=${git_description} -X main.buildTime=${current_time}'

.PHONY build/api:
build/api:
	@echo 'Building cmd/api'
	go build -ldflags=${linker_flags} -o=./bin/api ./cmd/api
