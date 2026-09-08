.PHONY: run remove  setup-mysql-master-slave

run:
	go run main.go

remove:
	rm -rf ./uploads/*


setup-mysql-master-slave:
	@chmod +x ./scripts/setup_mysql_master_slave.sh && ./scripts/setup_mysql_master_slave.sh $(PASSWORD) $(MASTER_HOST)