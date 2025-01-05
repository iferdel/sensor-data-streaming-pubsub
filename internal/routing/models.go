package routing

import "time"

// DTOs (Data Transfer Objects) for messaging system

type Sensor struct {
	SerialNumber string
}

type SensorMeasurement struct {
	SerialNumber string
	Timestamp    time.Time
	Value        float64
}

type SensorCommandMessage struct {
	SerialNumber string
	Timestamp    time.Time
	Command      string                 // intended for 'sleep' 'awake' 'changeSampleFrequency'
	Params       map[string]interface{} // command specific parameters e.g. {"sampleFrequency": 1000}
}

type SensorLog struct {
	SerialNumber string
	TimeStamp    time.Time
	Level        string
	Message      string
}
