package sensorlogic

type Sensor struct {
	SerialNumber string
}

type Log struct {
	Level   string
	Message string
}

type SensorState struct {
	Sensor                    Sensor       // 16 bytes (string)
	LogsInfo                  chan string  // 16 bytes
	LogsWarning               chan string  // 16 bytes
	LogsError                 chan string  // 16 bytes
	SampleFrequency           float64      // 8 bytes
	SampleFrequencyChangeChan chan float64 // 8 bytes
	IsSleep                   bool         // 1 byte, at the end to avoid memory layout (7 bytes of padding)
}

func NewSensorState(serialNumber string, SampleFrequency float64) *SensorState {
	return &SensorState{
		Sensor: Sensor{
			SerialNumber: serialNumber,
		},
		LogsInfo:                  make(chan string, 1),
		LogsWarning:               make(chan string, 1),
		LogsError:                 make(chan string, 1),
		SampleFrequency:           SampleFrequency,
		SampleFrequencyChangeChan: make(chan float64, 1),
		IsSleep:                   false,
	}
}
