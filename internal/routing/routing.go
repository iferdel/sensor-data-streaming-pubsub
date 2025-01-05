package routing

// PubSub Broker
const (
	RabbitConnString = "amqp://guest:guest@localhost:5672/"
)

// Exchange
const (
	ExchangeTopicIoT = "iot" // would be great to test as a direct exchange since it should be faster
)

// Queues follow pattern: entity.id.consumer.type
const (
	QueueSensorMeasurements   = "sensor.all.measurements.db_writer" // could be subjected to a sensor id though
	QueueSensorCommandsFormat = "sensor.%s.commands"                // subjected to sensor id
	QueueSensorRegistry       = "sensor.all.registry.created"       // could scale up to sensor.all.registry.notifier ??
	QueueSensorLogs           = "sensor.all.logs"
)

const (
	KeySensorMeasurements   = "sensor.%s.measurements"
	KeySensorCommandsFormat = "sensor.%s.commands"
	KeySensorRegistryFormat = "sensor.%s.registry"
	KeySensorLogsFormat     = "sensor.%s.logs"
)
