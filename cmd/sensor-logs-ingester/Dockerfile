# ---------------------------------------
# 1) Build stage
# ---------------------------------------
FROM golang:1.23-alpine AS build

WORKDIR /app

# Add dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy all sources (including cmd/sensor-logs-ingester)
COPY . .

# Build binary for production
RUN go build -o sensor-logs-ingester ./cmd/sensor-logs-ingester

# ---------------------------------------
# 2) Final stage
# ---------------------------------------
FROM alpine:3.16

WORKDIR /app

# Copy the compiled binary from the builder stage
COPY --from=build /app/sensor-logs-ingester /app/sensor-logs-ingester

COPY utils/wait-for-services.sh /app/
CMD ["/bin/sh", "/app/wait-for-services.sh", "/app/sensor-logs-ingester"]

