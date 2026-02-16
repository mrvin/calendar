TAG := $(shell git describe --abbrev=0 --tags --always)
HASH := $(shell git rev-parse HEAD)
DATE := $(shell date +%Y-%m-%d.%H:%M:%S)
LDFLAGS := -w -s \
				-X github.com/mrvin/calendar/internal/httpserver/handlers.hash=$(HASH) \
				-X github.com/mrvin/calendar/internal/httpserver/handlers.tag=$(TAG) \
				-X github.com/mrvin/calendar/internal/httpserver/handlers.date=$(DATE)
build:
	go build -ldflags "$(LDFLAGS)" -o ../../bin/calendar
test:
	mkdir -p ../../reports
	go test -race ../../internal/storage/memory -coverprofile=../../reports/memory_coverage.out
	go test ../../internal/calendar/httpserver/handlers -coverprofile=../../reports/handlers_coverage.out
report:
	go tool cover -html=../../reports/memory_coverage.out -o ../../reports/memory_cover.html
	go tool cover -html=../../reports/handlers_coverage.out -o ../../reports/handlers_cover.html
lint:
	golangci-lint run ./...
	golangci-lint run ./internal/calendar/...
	golangci-lint run ./internal/storage/...

.PHONY: build test report lint
