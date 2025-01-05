package sensorlogic

type Sensor struct {
	SerialNumber string
}

type SensorState struct {
	Sensor                    Sensor
	IsSleep                   bool
	SampleFrequency           float64
	SampleFrequencyChangeChan chan float64
}

func NewSensorState(serialNumber string, SampleFrequency float64) *SensorState {
	return &SensorState{
		Sensor: Sensor{
			SerialNumber: serialNumber,
		},
		IsSleep:                   false,
		SampleFrequency:           SampleFrequency,
		SampleFrequencyChangeChan: make(chan float64, 1),
	}
}
