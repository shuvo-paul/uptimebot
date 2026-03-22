.PHONY: build test run clean build_linux

build:
	pnpm build
	go build -o ./tmp/main ./cmd/main.go

build_linux:
	pnpm build
	GOOS=linux GOARCH=amd64 go build -o ./tmp/main ./cmd/main.go

test:
	go test ./...

run:
	./tmp/main

clean:
	rm -rf ./tmp