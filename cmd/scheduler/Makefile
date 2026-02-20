build:
	go build -o ../../bin/scheduler -ldflags '-w -s'
lint:
	golangci-lint run 
	golangci-lint run ../../internal/config
	golangci-lint run ../../internal/logger
	golangci-lint run ../../internal/storage/...
#	golangci-lint run  ../../internal/queue/...
	golangci-lint run  ../../internal/scheduler/...

.PHONY: build lint
