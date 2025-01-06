# ---------------------------------------
# 1) Build stage
# ---------------------------------------
FROM golang:1.23-alpine AS build

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o sensor-registry ./cmd/sensor-registry

# ---------------------------------------
# 2) Final stage
# ---------------------------------------
FROM alpine:3.16

WORKDIR /app

# Copy the compiled binary from the builder stage
COPY --from=build /app/sensor-registry /app/sensor-registry

COPY utils/wait-for-services.sh /app/
CMD ["/bin/sh", "/app/wait-for-services.sh", "/app/sensor-registry"]

