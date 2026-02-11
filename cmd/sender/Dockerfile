## Build
FROM golang:1.23.1-alpine AS build

LABEL maintainer="mrvin v.v.vinogradovv@gmail.com"

RUN apk update && apk add make && apk add tzdata

WORKDIR  /app

# Copy the code into the container.
COPY cmd/sender cmd/sender
COPY internal internal

# Copy and download dependency using go mod.
COPY go.mod go.sum ./
RUN go mod download

RUN cd cmd/sender/ && make build

## Deploy
FROM scratch

WORKDIR /

COPY --from=build ["/usr/share/zoneinfo", "/usr/share/zoneinfo"]
COPY --from=build ["/app/bin/sender", "/usr/local/bin/sender"]

ENV TZ=Europe/Moscow

ENTRYPOINT ["//usr/local/bin/sender"]
