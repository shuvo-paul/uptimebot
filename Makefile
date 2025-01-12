.PHONY: all build test run

build:
	pnpm build
	go build -o ./tmp/main ./cmd/sitemonitor/main.go

test:
	go test ./...

run:
	./tmp/main

clean:
	rm -rf ./tmp