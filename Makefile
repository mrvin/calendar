check-format:
	test -z $$(go fmt ./...)
build:
	$(MAKE) -C cmd/calendar build
	$(MAKE) -C cmd/scheduler build
	$(MAKE) -C cmd/sender build
test:
	$(MAKE) -C cmd/calendar test
lint:
	golangci-lint run ./...
certgen:
	openssl req -nodes -x509 -newkey rsa:4096 \
		-keyout cert/serverKey.pem -out cert/serverCert.pem -days 365 \
		-subj "/C=RU/L=Saint Petersburg/O=Calendar Corp./OU=Calendar/CN=calendar.com/emailAddress=v.v.vinogradovv@gmail.com"
	openssl req -nodes -x509 -newkey rsa:4096 \
		-keyout cert/clientKey.pem -out cert/clientCert.pem -days 365 \
		-subj "/C=RU/L=Saint Petersburg/O=Calendar Corp./OU=Calendar/CN=calendar.com/emailAddress=v.v.vinogradovv@gmail.com"

run:
	mkdir -p certs
	docker compose -f deployments/docker-compose.yaml --env-file configs/postgres.env --profile prod up --build
down:
	docker compose -f deployments/docker-compose.yaml --env-file configs/postgres.env --profile prod down

.PHONY: check-format build test lint certgen run down

