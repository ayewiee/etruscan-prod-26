# stolen from my 2nd stage project

FROM golang:1.24.7-alpine3.21 AS builder

# Setup base software for building an app.
RUN apk add --no-cache ca-certificates git gcc g++ libc-dev binutils

WORKDIR /opt

# Download dependencies.
COPY src/go.mod src/go.sum ./
RUN go mod download && go mod verify

RUN go install github.com/pressly/goose/v3/cmd/goose@v3.26.0

# Copy application source.
COPY ./src .

# Build the application.
RUN go build -o bin/application ./cmd/api

# Prepare executor image.
FROM alpine:3.21 AS runner

RUN apk add --no-cache ca-certificates libc6-compat openssh bash curl

WORKDIR /opt

COPY --from=builder /opt/bin/application ./
COPY --from=builder /opt/entrypoint.sh ./entrypoint.sh

COPY --from=builder /go/bin/goose /usr/local/bin/goose
COPY --from=builder /opt/db/migrations ./migrations

ENTRYPOINT ["sh", "./entrypoint.sh"]