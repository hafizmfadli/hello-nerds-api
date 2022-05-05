.PHONY: run
run:
	go run ./cmd/api -db-dsn "root:debezium@tcp(localhost:3306)/inventory?parseTime=true"