# ---------------------------------------
# Base Stage
# ---------------------------------------
FROM golang:1.24-alpine AS base

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o sensor-simulation ./cmd/sensor-simulation

# ---------------------------------------
# Development Build
# ---------------------------------------
FROM golang:1.24-alpine AS dev

WORKDIR /app

# Install Air and other necessary tools
RUN apk add --no-cache gcc musl-dev && \
    go install github.com/air-verse/air@latest

# Copy air config
COPY ./cmd/sensor-simulation/air.toml .

# Copy wait for dependencies
COPY utils/wait-for-services.sh /app/

CMD ["/bin/sh", "/app/wait-for-services.sh", "air", "-c", "./air.toml"]

# ---------------------------------------
# Staging Build
# ---------------------------------------
FROM alpine:3.16 AS staging

WORKDIR /app

# Copy the compiled binary from the base stage
COPY --from=base /app/sensor-simulation /app/sensor-simulation

CMD ["/app/sensor-simulation"]

# ---------------------------------------
# Production Build
# ---------------------------------------
# ...
