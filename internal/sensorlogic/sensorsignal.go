package sensorlogic

import "math"

type SineWave struct {
	Amplitude float64
	Frequency float64
	Phase     float64
}

func (sw *SineWave) Generate(dt float64) float64 {
	return sw.Amplitude * math.Sin(2*math.Pi*sw.Frequency*dt+sw.Phase)
}
