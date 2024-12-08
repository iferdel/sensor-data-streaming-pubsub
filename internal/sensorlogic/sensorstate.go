package sensorlogic

type SensorState struct {
    Sensor Sensor
    Paused bool
}


func NewSensorState(brand string) *SensorState {
    return &SensorState{
        Sensor: Sensor{
            Brand: brand,
        },
        Paused: false,
    }
}
