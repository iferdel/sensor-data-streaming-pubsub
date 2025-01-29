package treanteyes

type Level int

const (
	Debug Level = iota
	Info
	Warning
	Critical
)

// should a channel be attached to this valueobject in order to send logs through the lifetime of each Sensor?
type Log struct {
	level Level
	msg   string
}
