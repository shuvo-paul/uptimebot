.PHONY: build test run migrate

build:
	go build -o monitor ./cmd/monitor/monitor.go

test:
	go test ./...

run:
	./monitor

migrate:
	go build -o migrate ./cmd/migrate/migrate.go
