package routing

import "time"

// DTOs (Data Transfer Objects) for messaging system

// registry service
type Sensor struct {
	SerialNumber string
}

// measurements-ingester service
type SensorMeasurement struct {
	SerialNumber string
	Timestamp    time.Time
	Value        float64
}

// iotctl service
type SensorCommandMessage struct {
	SerialNumber string
	Timestamp    time.Time
	Command      string                 // intended for 'sleep' 'awake' 'changeSampleFrequency'
	Params       map[string]interface{} // command specific parameters e.g. {"sampleFrequency": 1000}
}

// logs-ingester service
type SensorLog struct {
	SerialNumber string
	TimeStamp    time.Time
	Level        string
	Message      string
}
