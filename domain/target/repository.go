package target

import (
	"errors"

	"github.com/google/uuid"
)

var (
	// this errors are returned by each one of the methods defined in the repository interface
	ErrTargetNotFound    = errors.New("The target was not found in repository")
	ErrFailedToAddTarget = errors.New("Failed to add Target to repository")
	ErrUpdateTarget      = errors.New("Failer to update Target in repository")
)

type TargetRepository interface {
	Get(uuid.UUID) (Target, error)
	Add(Target) error
	Update(Target) error
}
