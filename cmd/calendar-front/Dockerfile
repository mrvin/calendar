## Build
FROM golang:1.23.1-alpine AS build

LABEL maintainer="mrvin v.v.vinogradovv@gmail.com"

RUN apk update && apk add make && apk add tzdata

WORKDIR  /app

# Copy the code into the container.
COPY cmd/calendar-front cmd/calendar-front
COPY internal internal
COPY pkg pkg

# Copy and download dependency using go mod.
COPY go.mod go.sum ./
RUN go mod download

RUN cd cmd/calendar-front/ && make build

## Deploy
FROM scratch

WORKDIR /

COPY --from=build ["/usr/share/zoneinfo", "/usr/share/zoneinfo"]
COPY --from=build ["/app/bin/calendar-front", "/usr/local/bin/calendar-front"]

ENV TZ=Europe/Moscow

EXPOSE 8080

ENTRYPOINT ["/usr/local/bin/calendar-front"]
