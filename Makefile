test:
	go test -race ./internal/storage/... -cover
	cd cmd/calendar && make test
build:
	docker compose -f deployments/docker-compose.yaml --env-file deployments/postgres.env build
build-backend:
	docker compose -f deployments/docker-compose.yaml build backend
build-frontend:
	docker compose -f deployments/docker-compose.yaml build frontend
up:
	docker compose -f deployments/docker-compose.yaml --env-file deployments/postgres.env up
down:
	docker compose -f deployments/docker-compose.yaml --env-file deployments/postgres.env down
lint:
	golangci-lint run ./internal/...
codegen:
	go generate ./...
certgen:
	openssl req -nodes -x509 -newkey rsa:4096 \
		-keyout cert/serverKey.pem -out cert/serverCert.pem -days 365 \
		-subj "/C=RU/L=Saint Petersburg/O=Calendar Corp./OU=Calendar/CN=calendar.com/emailAddress=v.v.vinogradovv@gmail.com"
	openssl req -nodes -x509 -newkey rsa:4096 \
		-keyout cert/clientKey.pem -out cert/clientCert.pem -days 365 \
		-subj "/C=RU/L=Saint Petersburg/O=Calendar Corp./OU=Calendar/CN=calendar.com/emailAddress=v.v.vinogradovv@gmail.com"

.PHONY: test build build-backend build-frontend up down lint codegen certgen

