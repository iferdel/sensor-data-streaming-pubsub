package sensorlogic

type SensorState struct {
	Sensor  Sensor
	IsSleep bool
}

func NewSensorState(serialNumber string) *SensorState {
	return &SensorState{
		Sensor: Sensor{
			SerialNumber: serialNumber,
		},
		IsSleep: false,
	}
}
