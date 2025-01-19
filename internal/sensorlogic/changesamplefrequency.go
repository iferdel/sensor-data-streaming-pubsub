package sensorlogic

import (
	"fmt"
)

func (sensorState *SensorState) HandleChangeSampleFrequency(params map[string]interface{}) (float64, error) {
	var sampleFrequency float64

	if sf, ok := params["sampleFrequency"]; ok {
		sampleFrequency = sf.(float64)

		sensorState.SampleFrequency = sampleFrequency
		// signal the channel of the change of sample frequency
		sensorState.SampleFrequencyChangeChan <- sampleFrequency

		if sensorState.IsSleep {
			// TODO: using channel for logs may be better to spread logs over the handle
			fmt.Println("changes of sample frequency applied, but sensor is currently in a sleep state")
			return sampleFrequency, nil
		}
		fmt.Println("changes of sample frequency applied")
		return sampleFrequency, nil
	} else {
		fmt.Println("SampleFrequency is not a number")
	}
	return sampleFrequency, nil
}
