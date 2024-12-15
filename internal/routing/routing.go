package routing

// Exchange
const (
	ExchangeTopicIoT = "iot" // would be great to test as a direct exchange since it should be faster
)

// Queues follow pattern: entity.id.type.consumer
const (
	QueueSensorTelemetryFormat = "sensor.%s.telemetry.db_writer"    // subjected to sensor id
	QueueSensorCommandsFormat  = "sensor.%s.commands.state_handler" // subjected to sensor id
)

// Routing Keys follow pattern: entity.id.type
// Even if noun.verb is prefered, due to the domain of IoT, an exception is proposed
const (
	// Sensor data
	KeySensorDataTemplate = "sensor.*.telemetry"
	KeySensorDataFormat   = "sensor.%s.telemetry"

	// Sensor command
	KeySensorCommandTemplate = "sensor.*.commands"
	KeySensorCommandFormat   = "sensor.%s.commands"
)

const (
	SensorLogsSlug = "sensor_logs"
)
