package routing

const (
	RabbitConnString = "amqp://guest:guest@localhost:5672/"
)

// Exchange
const (
	ExchangeTopicIoT = "iot" // would be great to test as a direct exchange since it should be faster
)

// Queues follow pattern: entity.id.type.consumer
const (
	QueueSensorTelemetryFormat = "sensor.%s.telemetry.db_writer"    // subjected to sensor id
	QueueSensorCommandsFormat  = "sensor.%s.commands.state_handler" // subjected to sensor id
	QueueIoTLogs               = "logs"
)

// Routing Bind Keys follow pattern: entity.id.type
// Even if noun.verb is prefered, due to the domain of IoT, an exception is proposed
const (
	BindKeySensorDataFormat    = "sensor.%s.telemetry.#"
	BindKeySensorCommandFormat = "sensor.%s.commands.#"
	BindKeyIoTLogs             = "logs"
)
