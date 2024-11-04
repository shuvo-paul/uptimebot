.PHONY: build test run migrate

build:
	pnpm build
	go build -o monitor ./main.go

test:
	go test ./...

run:
	./monitor

migrate:
	go build -o migrate ./cmd/migrate/migrate.go
