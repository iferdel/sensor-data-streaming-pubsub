package sensorlogic

type Sensor struct {
	SerialNumber string
}

type SensorState struct {
	Sensor                    Sensor       // 16 bytes (string)
	SampleFrequency           float64      // 8 bytes
	SampleFrequencyChangeChan chan float64 // 8 bytes
	IsSleep                   bool         // 1 byte, at the end to avoid memory layout (7 bytes of padding)
}

func NewSensorState(serialNumber string, SampleFrequency float64) *SensorState {
	return &SensorState{
		Sensor: Sensor{
			SerialNumber: serialNumber,
		},
		SampleFrequency:           SampleFrequency,
		SampleFrequencyChangeChan: make(chan float64, 1),
		IsSleep:                   false,
	}
}
