all: test vet fmt lint build

test:
	go test ./...

vet:
	go vet ./...

fmt:
	go list -f '{{.Dir}}' ./... | grep -v /vendor/ | xargs -L1 gofmt -l
	test -z $(go list -f '{{.Dir}}' ./... | grep -v /vendor/ | xargs -L1 gofmt -l)

lint:
	staticcheck ./...

build:
	go build -o bin/iot-sensor-simulation ./cmd/sensor-simulation
	go build -o bin/iot-sensor-registry ./cmd/sensor-registry
	go build -o bin/iot-logs-ingester ./cmd/sensor-logs-ingester
	go build -o bin/iot-measurements-ingester ./cmd/sensor-measurements-ingester
	go build -o bin/iot-api ./cmd/iot-api
