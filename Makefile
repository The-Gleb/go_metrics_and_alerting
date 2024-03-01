postgres:
	docker run --name metric_db -e POSTGRES_USER=metric_db -e POSTGRES_PASSWORD=metric_db -p 5434:5432 -d postgres:alpine
postgresrm:
	docker stop metric_db
	docker rm metric_db

migrateup:
	migrate -path internal/adapter/db/postgres/migration -database "postgres://metric_db:metric_db@localhost:5434/metric_db?sslmode=disable" -verbose up

migratedown:
	migrate -path internal/adapter/db/postgres/migration -database "postgres://metric_db:metric_db@localhost:5434/metric_db?sslmode=disable" -verbose down
.PHONY: postgres postgresrm migrateup migratedown