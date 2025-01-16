# ---------------------------------------
# 1) Build Stage
# ---------------------------------------
FROM golang:1.23-alpine AS build

WORKDIR /app

# Add dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy all sources (including cmd/iot-api)
COPY . .

# Build binary for production
RUN go build -o iot-api ./cmd/iot-api

# ---------------------------------------
# Development Stage
# ---------------------------------------
FROM golang:1.23-alpine AS dev

WORKDIR /app

# Install Air and other necessary tools
RUN apk add --no-cache gcc musl-dev && \
    go install github.com/air-verse/air@latest

# Copy air config
COPY ./cmd/iot-api/air.toml .

# We'll mount our code volume in docker-compose
CMD ["air", "-c", "./air.toml"]
