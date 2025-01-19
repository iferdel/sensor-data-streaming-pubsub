# ---------------------------------------
# Base Stage
# ---------------------------------------
FROM golang:1.23-alpine AS base

WORKDIR /app

# Add dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy all sources (including cmd/iot-api)
COPY . .

# Build binary for production
RUN go build -o iot-api ./cmd/iot-api

# ---------------------------------------
# Development Build
# ---------------------------------------
FROM golang:1.23-alpine AS dev

WORKDIR /app

# Install Air and other necessary tools
RUN apk add --no-cache gcc musl-dev && \
    go install github.com/air-verse/air@latest

# Copy air config
COPY ./cmd/iot-api/air.toml .

# Copy wait for dependencies
COPY utils/wait-for-services.sh /app/

CMD ["/bin/sh", "/app/wait-for-services.sh", "air", "-c", "./air.toml"]

# ---------------------------------------
# Staging Build
# ---------------------------------------
FROM alpine:3.16 AS staging

WORKDIR /app

# Copy the compiled binary from the base stage
COPY --from=base /app/iot-api /app/iot-api

# Copy wait for dependencies
COPY utils/wait-for-services.sh /app/

CMD ["/bin/sh", "/app/wait-for-services.sh", "/app/iot-api"]

# ---------------------------------------
# Production Build
# ---------------------------------------
# ...
