.PHONY: build test run

build:
	go build -o sitemonitor ./cmd/main.go

test:
	go test ./...

run:
	./sitemonitor

