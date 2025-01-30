package target

import (
	"errors"

	"github.com/google/uuid"
	"github.com/iferdel/treanteyes"
)

var (
	ErrInvalidTarget               = errors.New("a target must have valid fields")
	ErrInvalidTargetCustomer       = errors.New("a target must have a valid customer")
	ErrInvalidTargetName           = errors.New("a target must have a valid name")
	ErrInvalidTargetDescription    = errors.New("a target must have a valid Description")
	ErrInvalidTargetMountingPoints = errors.New("a target must have valid MountingPoints")
)

type Target struct {
	asset          *treanteyes.Asset
	mountingPoints []treanteyes.MountingPoint
}

func NewTarget(customer, name, description string, mountingPointsPositions []string) (Target, error) {

	asset := &treanteyes.Asset{
		ID:          uuid.New(),
		Customer:    customer,
		Name:        name,
		Description: description,
	}

	mountingPoints := make([]treanteyes.MountingPoint, 0, len(mountingPointsPositions))

	for _, position := range mountingPointsPositions {
		mountingPoint, err := treanteyes.NewMountingPoint(position)
		if err != nil {
			return Target{}, err
		}
		mountingPoints = append(mountingPoints, mountingPoint)
	}

	return Target{
		asset:          asset,
		mountingPoints: mountingPoints,
	}, nil
}

func (t *Target) GetID() uuid.UUID {
	return t.asset.ID
}

func (t *Target) GetName() string {
	return t.asset.Name
}

func (t *Target) SetDescription(description string) {
	if t.asset == nil {
		t.asset = &treanteyes.Asset{}
	}
	t.asset.Description = description
}

func (t *Target) AddMountingPoint(mp treanteyes.MountingPoint) {
	t.mountingPoints = append(t.mountingPoints, mp)
}

func (t *Target) GetMountingPoints() []treanteyes.MountingPoint {
	return t.mountingPoints
}
