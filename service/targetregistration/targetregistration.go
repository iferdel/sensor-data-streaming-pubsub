package targetregistration

import (
	"github.com/google/uuid"
	"github.com/iferdel/treanteyes/domain/target"
	"github.com/iferdel/treanteyes/domain/target/memory"
)

type TargetRegistrationConfiguration func(ts *TargetRegistrationService) error

type TargetRegistrationService struct {
	targets target.TargetRepository
}

func NewTargetRegistrationService(cfgs ...TargetRegistrationConfiguration) (*TargetRegistrationService, error) {

	ts := &TargetRegistrationService{}

	for _, cfg := range cfgs {
		err := cfg(ts)
		if err != nil {
			return nil, err
		}
	}
	return ts, nil
}

func WithTargetRepository(tr target.TargetRepository) TargetRegistrationConfiguration {
	return func(ts *TargetRegistrationService) error {
		ts.targets = tr
		return nil
	}
}

func WithMemoryTargetRepository() TargetRegistrationConfiguration {
	tr := memory.New()
	return WithTargetRepository(tr)
}

func (ts *TargetRegistrationService) AddTarget(customer, name, description string, mountingPointsPositions []string) (uuid.UUID, error) {
	t, err := target.NewTarget(customer, name, description, mountingPointsPositions)
	if err != nil {
		return uuid.Nil, err
	}

	// Add to repo
	err = ts.targets.Add(t)
	if err != nil {
		return uuid.Nil, err
	}

	return t.GetID(), nil
}
