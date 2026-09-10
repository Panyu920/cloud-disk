db_url =mysql://panyu:panyu@tcp(localhost:3306)/cloud-disk?multiStatements=true
.PHONY: run remove  setup-mysql-master-slave db_docs db_schema migrateup migrateup1 migratedown migratedown1 new_migration sqlc

run:
	go run main.go

remove:
	rm -rf ./uploads/*


setup-mysql-master-slave:
	@chmod +x ./scripts/setup_mysql_master_slave.sh && ./scripts/setup_mysql_master_slave.sh $(PASSWORD) $(MASTER_HOST)

db_docs:
	dbdocs build ./doc/db.dbml

db_schema:
	dbml2sql ./doc/db.dbml -o ./doc/schema.sql --mysql

migrateup:
	migrate -path db/migration -database "$(db_url)" -verbose up

migrateup1:
	migrate -path db/migration -database "$(db_url)" -verbose up 1

migratedown:
	migrate -path db/migration -database "$(db_url)" -verbose down

migratedown1:
	migrate -path db/migration -database "$(db_url)" -verbose down 1

new_migration:
	migrate create -ext sql -dir db/migration -seq $(name)

sqlc:
	sqlc generate

test:
	go test ./... -cover -v