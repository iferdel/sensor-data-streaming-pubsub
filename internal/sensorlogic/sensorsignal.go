package sensorlogic

import (
	"math"
	"math/rand/v2"
)

type SineWave struct {
	Amplitude float64
	Frequency float64
	Phase     float64
}

func (sw *SineWave) Generate(dt float64) float64 {
	return sw.Amplitude * math.Sin(2.0*math.Pi*sw.Frequency*dt+sw.Phase)
}

func SimulateSignal(waves []SineWave, t float64) float64 {

	// Generate the superimposed signal
	var value float64
	for _, wave := range waves {
		value += wave.Generate(t)
	}

	noiseAmplitude := rand.Float64()
	value += (noiseAmplitude * rand.NormFloat64())

	return value
}
