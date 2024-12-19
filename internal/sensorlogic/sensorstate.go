package sensorlogic

type SensorState struct {
	Sensor                    Sensor
	IsSleep                   bool
	SampleFrequency           int
	SampleFrequencyChangeChan chan int
}

func NewSensorState(serialNumber string, SampleFrequency int) *SensorState {
	return &SensorState{
		Sensor: Sensor{
			SerialNumber: serialNumber,
		},
		IsSleep:                   false,
		SampleFrequency:           SampleFrequency,
		SampleFrequencyChangeChan: make(chan int, 1),
	}
}
