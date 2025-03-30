package routing

import "os"

// PubSub Broker
var (
	RabbitConnString       = os.Getenv("RABBIT_AMQP_CONN_STRING")
	RabbitMQTTConnString   = os.Getenv("RABBIT_MQTT_CONN_STRING")
	RabbitStreamConnString = os.Getenv("RABBIT_STREAM_CONN_STRING")
)

// Streams
const (
	StreamConsumerName = "iot"
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

// keys are used in consumers with wildcards and in publishers with the specific value
const (
	KeySensorMeasurements   = "sensor.%s.measurements"
	KeySensorCommandsFormat = "sensor.%s.commands"
	KeySensorRegistryFormat = "sensor.%s.registry"
	KeySensorLogsFormat     = "sensor.%s.logs"
)
