build:
	go build -o ../../bin/calendar -ldflags '-w -s'

run:
	go run -race main.go

test:
	go test -race ../../test/http/server_test.go -cover -v

lint:
	golangci-lint run ./...

.PHONY: build run test lint
