# ---------------------------------------
# Base Stage
# ---------------------------------------
FROM golang:1.23-alpine AS base

WORKDIR /app

# Add dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy all sources (including cmd/sensor-measurements-ingester)
COPY . .

RUN go build -o sensor-measurements-ingester ./cmd/sensor-measurements-ingester

# ---------------------------------------
# Development Build
# ---------------------------------------
FROM golang:1.23-alpine AS dev

WORKDIR /app

# Install Air and other necessary tools
RUN apk add --no-cache gcc musl-dev && \
    go install github.com/air-verse/air@latest

# Copy air config
COPY ./cmd/sensor-measurements-ingester/air.toml .

# Copy wait for dependencies
COPY utils/wait-for-services.sh /app/

CMD ["/bin/sh", "/app/wait-for-services.sh", "air", "-c", "./air.toml"]

# ---------------------------------------
# Staging Build
# ---------------------------------------
FROM alpine:3.16 AS staging

WORKDIR /app

# Copy the compiled binary from the base stage
COPY --from=base /app/sensor-measurements-ingester /app/sensor-measurements-ingester

# Copy wait for dependencies
COPY utils/wait-for-services.sh /app/

CMD ["/bin/sh", "/app/wait-for-services.sh", "/app/sensor-measurements-ingester"]

# ---------------------------------------
# Production Build
# ---------------------------------------
# ...
