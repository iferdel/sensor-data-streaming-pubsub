package routing

// Relational Database
const (
	PostgresConnString = "postgres://postgres:postgres@localhost:5432/iot"
)

// PubSub Broker
const (
	RabbitConnString = "amqp://guest:guest@localhost:5672/"
)

// Exchange
const (
	ExchangeTopicIoT = "iot" // would be great to test as a direct exchange since it should be faster
)

// Queues follow pattern: entity.id.type.consumer
const (
	QueueSensorCommandsFormat  = "sensor.%s.commands"            // subjected to sensor id
	QueueSensorRegistry        = "sensor.registry"               // subjected to sensor id
	QueueSensorTelemetryFormat = "sensor.%s.telemetry.db_writer" // subjected to sensor id
	QueueIoTLogs               = "logs"
)

const (
	KeySensorCommandsFormat = "sensor.%s.commands"
	KeySensorRegistry       = "registry"
	KeyTelemetry            = "telemetry"
	KeyLogs                 = "logs"
)
