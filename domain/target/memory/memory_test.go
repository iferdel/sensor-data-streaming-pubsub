package memory

import (
	"testing"

	"github.com/google/uuid"
	"github.com/iferdel/treanteyes/domain/target"
)

func TestMemory_GetTarget(t *testing.T) {

	type testCase struct {
		test        string
		id          uuid.UUID
		expectedErr error
	}

	// Create a fake target to add to repository
	newTarget, err := target.NewTarget("Customer", "Pump", "Vertical Pump", []string{"Bearing A", "Bearing B"})
	if err != nil {
		t.Fatal(err)
	}
	id := newTarget.GetID()
	// Create the repo to use, and add some test Data to it for testing
	// Skip Factory for this
	repo := MemoryRepository{
		targets: map[uuid.UUID]target.Target{
			id: newTarget,
		},
	}

	testCases := []testCase{
		{
			test:        "No Target by ID",
			id:          uuid.MustParse("3aa41c7c-2133-425e-aeeb-583d64cbf145"),
			expectedErr: target.ErrTargetNotFound,
		},
		{
			test:        "Target by ID",
			id:          id,
			expectedErr: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.test, func(t *testing.T) {
			_, err := repo.Get(tc.id)
			if err != tc.expectedErr {
				t.Errorf("Expected error %v, got %v", tc.expectedErr, err)
			}
		})
	}
}

func TestMemory_AddTarget(t *testing.T) {
	type testCase struct {
		test           string
		customer       string
		name           string
		description    string
		mountingPoints []string
		expectedErr    error
	}

	testCases := []testCase{
		{
			test:           "Add Target",
			customer:       "ValidCustomer for a Target",
			name:           "ValidName for a Target",
			description:    "ValidDescription, more stuff here",
			mountingPoints: []string{"Bearing A", "Bearing B"},
			expectedErr:    nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			repo := MemoryRepository{
				targets: map[uuid.UUID]target.Target{},
			}

			newTarget, err := target.NewTarget(tc.customer, tc.name, tc.description, tc.mountingPoints)
			if err != nil {
				t.Fatal(err)
			}

			err = repo.Add(newTarget)
			if err != tc.expectedErr {
				t.Errorf("Expected error %v, got %v", tc.expectedErr, err)
			}

			found, err := repo.Get(newTarget.GetID())
			if err != nil {
				t.Fatal(err)
			}
			if found.GetID() != newTarget.GetID() {
				t.Errorf("Expected %v, got %v", newTarget.GetID(), found.GetID())
			}
		})
	}
}
