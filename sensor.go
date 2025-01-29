package treanteyes

import (
	"github.com/google/uuid"
)

type Family int

const (
	GeoLocation Family = iota
	Acceleration
	Velocity
	Strain
	Humidity
	Temperature
)

type Axiality int

const (
	UniAxial Axiality = iota
	BiAxial
	TriAxial
)

type SensorState int

const (
	IsWaitingForTargetDefinition SensorState = iota
	IsAwake
	IsSleep
)

// sensor reprensents a core Entity in terms of DDD.
type Sensor struct {
	ID uuid.UUID
	// The reason for not using SerialNumber as unique identification of this entity is the following:
	// Even if it is remote, the possibility of two sensors from different families with same serial number exists.
	SerialNumber string
	Family       Family
	Axiality     Axiality
	// SensorState and SampleFrequency are two values that will be subjected to change in the lifetime of a Sensor
	// Thats why the a channel for each one is created, to simplify this pattern
	SensorState         SensorState
	SensorStateChan     chan SensorState
	SampleFrequency     float64
	SampleFrequencyChan chan float64
}
