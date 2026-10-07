.PHONY: run test build

run:
	go run ./cmd/shortener

test:
	go test ./...

build:
	go build -o bin/shortener ./cmd/shortener
