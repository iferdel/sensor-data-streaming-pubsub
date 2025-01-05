# ---------------------------------------
# 1) Build stage
# ---------------------------------------
FROM golang:1.23-alpine AS build

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o sensor-simulation ./cmd/sensor-simulation

# ---------------------------------------
# 2) Final stage
# ---------------------------------------
FROM alpine:3.16

WORKDIR /app

# Copy the compiled binary from the builder stage
COPY --from=build /app/sensor-simulation /app/sensor-simulation

COPY cmd/sensor-simulation/wait-for-rabbitmq.sh /app/
CMD ["/bin/sh", "/app/wait-for-rabbitmq.sh", "/app/sensor-simulation"]

