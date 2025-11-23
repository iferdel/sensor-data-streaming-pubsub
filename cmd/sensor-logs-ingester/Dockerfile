# ---------------------------------------
# Base Stage
# ---------------------------------------
FROM golang:1.25-alpine AS base

WORKDIR /app

# Add dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy all sources (including cmd/sensor-logs-ingester)
COPY . .

RUN go build -o sensor-logs-ingester ./cmd/sensor-logs-ingester

# ---------------------------------------
# Development Build
# ---------------------------------------
FROM golang:1.25-alpine AS dev

WORKDIR /app

# Install Air and other necessary tools
RUN apk add --no-cache gcc musl-dev && \
    go install github.com/air-verse/air@v1.63.1

# Copy air config
COPY ./cmd/sensor-logs-ingester/air.toml .

# Copy wait for dependencies
COPY utils/wait-for-services.sh /app/

CMD ["/bin/sh", "/app/wait-for-services.sh", "air", "-c", "./air.toml"]

# ---------------------------------------
# Staging Build
# ---------------------------------------
FROM alpine:3.16 AS staging

WORKDIR /app

# Copy the compiled binary from the base stage
COPY --from=base /app/sensor-logs-ingester /app/sensor-logs-ingester

RUN mkdir -p /app/log && touch /app/log/iot.log

CMD ["/app/sensor-logs-ingester"]

# ---------------------------------------
# Production Build
# ---------------------------------------
# ...
