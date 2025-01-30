package treanteyes

import "errors"

var (
	ErrInvalidMountingPointPosition = errors.New("a mounting point must have a valid position")
)

type MountingPoint struct {
	// closely related with the target itself, e.g. 'Bearing A from mechanical draft sheet'
	position string
}

func NewMountingPoint(pos string) (MountingPoint, error) {
	if pos == "" {
		return MountingPoint{}, ErrInvalidMountingPointPosition
	}
	return MountingPoint{position: pos}, nil
}
