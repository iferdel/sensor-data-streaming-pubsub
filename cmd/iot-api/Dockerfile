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
# 2) Development Stage
# ---------------------------------------
FROM golang:1.23-alpine AS dev

WORKDIR /app

# Install CompileDaemon for hot-reloading
RUN apk add --no-cache bash git && \
    go install github.com/githubnemo/CompileDaemon@latest

# Bind mounts will overlay files
