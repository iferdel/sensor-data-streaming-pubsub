package memory

import (
	"fmt"
	"sync"

	"github.com/google/uuid"
	"github.com/iferdel/treanteyes/domain/target"
)

type MemoryRepository struct {
	targets map[uuid.UUID]target.Target
	sync.Mutex
}

func New() *MemoryRepository {
	return &MemoryRepository{
		targets: make(map[uuid.UUID]target.Target),
	}
}

func (mr *MemoryRepository) Get(id uuid.UUID) (target.Target, error) {
	if target, ok := mr.targets[id]; ok {
		return target, nil
	}
	return target.Target{}, target.ErrInvalidTarget
}

func (mr *MemoryRepository) Add(t target.Target) error {
	if mr.targets == nil {
		// Saftey check if target is not created, shouldn't happen if using the Factory, but you never know
		mr.Lock()
		mr.targets = make(map[uuid.UUID]target.Target)
		mr.Unlock()
	}

	if _, ok := mr.targets[t.GetID()]; ok {
		return fmt.Errorf("target already exists: %w", target.ErrFailedToAddTarget)
	}

	mr.Lock()
	mr.targets[t.GetID()] = t
	mr.Unlock()

	return nil
}

func (mr *MemoryRepository) Update(t target.Target) error {
	if _, ok := mr.targets[t.GetID()]; !ok {
		return fmt.Errorf("target does not exists: %w", target.ErrUpdateTarget)
	}

	mr.Lock()
	mr.targets[t.GetID()] = t
	mr.Unlock()

	return nil
}
