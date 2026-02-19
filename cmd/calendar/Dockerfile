## Build
FROM golang:1.26.0-alpine AS dev

LABEL maintainer="mrvin v.v.vinogradovv@gmail.com"

RUN apk add --no-cache make tzdata git gcc musl-dev

WORKDIR /app

COPY cmd/calendar cmd/calendar
COPY internal internal
COPY pkg pkg
COPY Makefile ./
COPY .git .git

# Copy and download dependency using go mod.
COPY go.mod go.sum ./
RUN go mod download

RUN cd cmd/calendar/ && CGO_ENABLED=0 make build

RUN mkdir /var/log/calendar/

COPY --from=ghcr.io/tarampampam/microcheck:1 ["/bin/httpscheck", "/bin/httpscheck"]

ENV TZ=Europe/Moscow

EXPOSE 8080 50051

CMD ["cmd/calendar/start.sh"]

## Deploy
FROM scratch AS prod

LABEL maintainer="mrvin v.v.vinogradovv@gmail.com"

WORKDIR /

COPY --from=dev ["/bin/httpscheck", "/bin/httpscheck"]
COPY --from=dev ["/var/log/calendar/", "/var/log/calendar/"]
COPY --from=dev ["/usr/share/zoneinfo", "/usr/share/zoneinfo"]
COPY --from=dev ["/app/bin/calendar", "/usr/local/bin/calendar"]

ENV TZ=Europe/Moscow

EXPOSE 8080 50051

ENTRYPOINT ["/usr/local/bin/calendar"]