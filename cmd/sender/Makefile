build:
	go build -o ../../bin/sender  -ldflags '-w -s'

run:
	go run -race main.go

test:
	go test -race ./... -cover

lint:
	golangci-lint run ./...

.PHONY: build run test lint
