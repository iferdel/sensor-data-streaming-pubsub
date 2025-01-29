package treanteyes

import (
	"time"

	"github.com/google/uuid"
)

type Mounting struct {
	// closely related with the target itself, e.g. 'Bearing A from mechanical draft sheet'
	position   string
	fromSensor uuid.UUID
	onTarget   uuid.UUID
	mountedAt  time.Time
}
