package treanteyes

import "time"

type Measurement struct {
	timestamp time.Time // values in lowercase since they are immutable
	// values is an array since it depends on the sensor it belongs to.
	// For instance:
	// geolocation may have lat and lon;
	// humidity only one value per measurement;
	// acceleration depends on sensor's axiality.
	values []float64
}
