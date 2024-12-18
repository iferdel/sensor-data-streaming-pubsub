package routing

import "time"

// DTOs (Data Transfer Objects) for messaging system

type SensorMeasurement struct {
	SensorName string
	Timestamp  time.Time
	Value      float64
}

type CommandMessage struct {
	SensorName string
	Timestamp  time.Time
	Command    string                 // intended for 'sleep' 'awake' 'changeSampleFreq'
	Params     map[string]interface{} // command specific parameters e.g. {"frequency": 1000}
}

type SensorLog struct {
	SensorName  string
	CurrentTime time.Time
	Level       string
	Message     string
}
