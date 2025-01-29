package treanteyes

// Not quite sure if this is the best implementation for this value object
type Signal struct {
	// Generic fields that all signals may share
	amplitude *float64
	frequency *float64
	phase     *float64
	latitude  *float64
	longitude *float64
}
