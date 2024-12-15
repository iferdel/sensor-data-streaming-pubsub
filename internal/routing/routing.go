package routing

// Exchange
const (
	ExchangeSensorsTopic = "sensors"
)

// Queue Names
const (
	QueueSensorData     = "sensor.data_stream"
	QueueSensorCommands = "sensor.commands"
)

// Routing Keys
const (
	// Sensor data
	SensorDataTemplate = "sensor.*.data"
	SensorDataFormat   = "sensor.%s.data"

	// Sensor command
	SensorCommandTemplate = "sensor.*.command"
	SensorCommandFormat   = "sensor.%s.command"
)
