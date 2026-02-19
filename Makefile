build:
	$(MAKE) -C cmd/calendar build
	$(MAKE) -C cmd/scheduler build
	$(MAKE) -C cmd/sender build
test:
	$(MAKE) -C cmd/calendar test
lint:
	golangci-lint run ./internal/...
certgen:
	openssl req -nodes -x509 -newkey rsa:4096 \
		-keyout cert/serverKey.pem -out cert/serverCert.pem -days 365 \
		-subj "/C=RU/L=Saint Petersburg/O=Calendar Corp./OU=Calendar/CN=calendar.com/emailAddress=v.v.vinogradovv@gmail.com"
	openssl req -nodes -x509 -newkey rsa:4096 \
		-keyout cert/clientKey.pem -out cert/clientCert.pem -days 365 \
		-subj "/C=RU/L=Saint Petersburg/O=Calendar Corp./OU=Calendar/CN=calendar.com/emailAddress=v.v.vinogradovv@gmail.com"

run:
	mkdir -p certs
	docker compose -f deployments/docker-compose.yaml --env-file deployments/postgres.env --profile prod up --build
down:
	docker compose -f deployments/docker-compose.yaml --env-file deployments/postgres.env --profile prod down

.PHONY: build test certgen run down

