package sensorlogic

type SensorState struct {
	Sensor                    Sensor
	IsSleep                   bool
	SampleFrequency           float64
	SampleFrequencyChangeChan chan int
}

func NewSensorState(serialNumber string, SampleFrequency float64) *SensorState {
	return &SensorState{
		Sensor: Sensor{
			SerialNumber: serialNumber,
		},
		IsSleep:                   false,
		SampleFrequency:           SampleFrequency,
		SampleFrequencyChangeChan: make(chan int, 1),
	}
}
