package sensorlogic

type SensorState struct {
	Sensor Sensor
	Paused bool
}

func NewSensorState(serialNumber string) *SensorState {
	return &SensorState{
		Sensor: Sensor{
			SerialNumber: serialNumber,
		},
		Paused: false,
	}
}
