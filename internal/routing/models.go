package routing

import "time"

type SensorStatus struct {
	SensorName string
	Status     string
}

type SensorMeasurement struct {
	SensorName string
	Timestamp  time.Time
	Value      float64
}

type CommandMessage struct {
	SensorName string
	Timestamp  time.Duration
	Command    string                 // intended for 'sleep' 'resume' 'changeSampleFreq'
	Params     map[string]interface{} // command specific parameters e.g. {"frequency": 1000}
}

type SensorLog struct {
	SensorName  string
	CurrentTime time.Time
	Level       string
	Message     string
}
