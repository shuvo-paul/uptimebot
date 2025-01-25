.PHONY: build test run clean

build:
	pnpm build
	go build -o ./tmp/main ./cmd/main.go

test:
	go test ./...

run:
	./tmp/main

clean:
	rm -rf ./tmp