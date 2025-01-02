package routing

import "time"

// DTOs (Data Transfer Objects) for messaging system

type Sensor struct {
	SensorName string
}

type SensorMeasurement struct {
	SensorName string
	Timestamp  time.Time
	Value      float64
}

type CommandMessage struct {
	SensorName string
	Timestamp  time.Time
	Command    string                 // intended for 'sleep' 'awake' 'changeSampleFrequency'
	Params     map[string]interface{} // command specific parameters e.g. {"sampleFrequency": 1000}
}

type SensorLog struct {
	SensorName string
	TimeStamp  time.Time
	Level      string
	Message    string
}
