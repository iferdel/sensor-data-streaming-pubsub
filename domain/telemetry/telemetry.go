package target

import (
	"github.com/iferdel/treanteyes"
)

type Telemetry struct {
	// Sensor is the root entity of the Aggregate in Domain Driven Design.
	sensor *treanteyes.Sensor
	// A sensor must be assigned to a target to collect telemetry data.
	target *treanteyes.Target
	// Each sensor is mounted on a single mounting point of a target, regardless of its axiality.
	mounting treanteyes.Mounting
	// The signal, representing a phenomenon, belongs to the target and mounting point.
	// The sensor measures the signal using its defined precision and parameters.
	signals []treanteyes.Signal
	// Signals represent the phenomenon (e.g., combined sinewaves for accelerometers).
	// Measurements are the data captured from the signals.
	measurements []treanteyes.Measurement
	// Each sensor records it own logs.
	logs []treanteyes.Log
}
