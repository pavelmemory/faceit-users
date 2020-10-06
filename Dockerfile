FROM golang:1.14-alpine3.12 as builder

RUN apk add --no-cache \
    gcc \
    git \
    make

WORKDIR /faceit-users

COPY cmd/ cmd/
COPY internal/ internal/
COPY go.mod go.mod
COPY go.sum go.sum
COPY Makefile Makefile

RUN make build

FROM alpine:3.12

COPY --from=builder /faceit-users/build/bin/ /usr/local/bin/

RUN addgroup -g 1000 faceit && \
    adduser -h /faceit -D -u 1000 -G faceit faceit && \
    chown faceit:faceit /faceit
USER faceit

ENTRYPOINT faceit-users
