.PHONY: run remove  setup-mysql-master-slave

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