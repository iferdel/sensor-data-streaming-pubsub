package routing

import "time"

// DTOs (Data Transfer Objects) for messaging system

// registry service
type Sensor struct {
	SerialNumber    string
	SampleFrequency float64
}

// measurements-ingester service
type SensorMeasurement struct {
	SerialNumber string
	Timestamp    time.Time
	Value        float64
}

// measurements-ingester service
type SensorLocation struct {
	SerialNumber string
	Timestamp    time.Time
	Lat          float64
	Lon          float64
}

// iotctl service
type SensorCommandMessage struct {
	SerialNumber string
	Timestamp    time.Time
	Command      string         // intended for 'sleep' 'awake' 'changeSampleFrequency'
	Params       map[string]any // command specific parameters e.g. {"sampleFrequency": 1000}
}

// logs-ingester service
type SensorLog struct {
	SerialNumber string
	Timestamp    time.Time
	Level        string
	Message      string
}
