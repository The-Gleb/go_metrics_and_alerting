server_ldflags:
	go run -ldflags "-X main.BuildVersion=v1.0.1 -X 'main.BuildDate=$(date +'%Y/%m/%d')' -X 'main.BuildCommit=$(make cur_commit)'" ./cmd/server
postgres:
	docker run --name metric_db -e POSTGRES_USER=metric_db -e POSTGRES_PASSWORD=metric_db -p 5434:5432 -d postgres:alpine
postgresrm:
	docker stop metric_db
	docker rm metric_db

migrateup:
	migrate -path internal/adapter/db/postgres/migration -database "postgres://metric_db:metric_db@localhost:5434/metric_db?sslmode=disable" -verbose up

migratedown:
	migrate -path internal/adapter/db/postgres/migration -database "postgres://metric_db:metric_db@localhost:5434/metric_db?sslmode=disable" -verbose down

golangci-lint-run:
	docker run --rm -v $(shell pwd):/app \
	-v ~/.cache/golangci-lint/v1.56.1:/root/.cache \
	-w /app golangci/golangci-lint:v1.56.2 \
	golangci-lint run --fix -v \
	> ./golangci-lint/report-unformatted.json

format-lint-output:
	cat ./golangci-lint/report-unformatted.json | jq > ./golangci-lint/report.json \

.PHONY: postgres postgresrm migrateup migratedown server_ldflags