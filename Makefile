postgres:
	docker run --name postgres12 --network bank-network -p 5432:5432 -e POSTGRES_USER=root -e POSTGRES_PASSWORD=8520 -d postgres:12-alpine
terminate:
	docker exec -it postgres12 psql -U root -c "SELECT pg_terminate_backend(pg_stat_activity.pid) FROM pg_stat_activity WHERE pg_stat_activity.datname = 'simple_bank' AND pid <> pg_backend_pid();"
start:
	docker start postgres12
createdb:
	docker exec -it postgres12 createdb --username=root --owner=root simple_bank
dropdb:
	docker exec -it postgres12 dropdb simple_bank

migrateup:
	migrate -path db/migration -database "postgresql://root:8520@localhost:5432/simple_bank?sslmode=disable" -verbose up
migrateup1:
	migrate -path db/migration -database "postgresql://root:8520@localhost:5432/simple_bank?sslmode=disable" -verbose up 1

migratedown:
	migrate -path db/migration -database "postgresql://root:8520@localhost:5432/simple_bank?sslmode=disable" -verbose down
migratedown1:
	migrate -path db/migration -database "postgresql://root:8520@localhost:5432/simple_bank?sslmode=disable" -verbose down 1

sqlc:
	sqlc generate
test:
	go test -v -cover ./...
psql:
	docker exec -it postgres12 psql -U root -d simple_bank
server:
	go run main.go
mock:
	mockgen -package mockdb -destination db/mock/store.go github.com/kjasn/simple-bank/db/sqlc Store
build:
	docker build --network host -t simplebank:latest .
dcrun:
	docker run --name simplebank --network bank-network -p 8080:8080 -e DSN="postgresql://root:8520@postgres12:5432/simple_bank?sslmode=disable" simplebank:latest

.PHONY: postgres terminate createdb dropdb migrateup migrateup1 migratedown migratedown1 sqlc test start psql server build dcrun