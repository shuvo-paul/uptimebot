.PHONY: build test run migrate

build:
	pnpm build
	go build -o sitemonitor ./cmd/sitemonitor/sitemonitor.go

test:
	go test ./...

run:
	./sitemonitor
