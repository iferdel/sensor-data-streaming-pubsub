package treanteyes

import (
	"time"

	"github.com/google/uuid"
)

// depending on how the logic expands,
// this may be changed to an Entity,
// since a Sensor may require a change over its mount point.
// But i wonder if this can be achiavable by adding another entry for this.
// I dont see any need for unique identifier, but the idea of a valueobject says its
// for structs that do not require unique identifier AND are immutable
// in this case, mounting may be mutable so this condition for valueobject is not met
type Mounting struct {
	// closely related with the asset itself, e.g. 'Bearing A from mechanical draft sheet'
	position   string
	fromSensor uuid.UUID
	onAsset    uuid.UUID
	mountedAt  time.Time
}
