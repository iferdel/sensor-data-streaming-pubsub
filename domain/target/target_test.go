package target

import "testing"

func TestTarget_NewTarget(t *testing.T) {

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
			test:           "All fields valid",
			customer:       "ValidCustomer for a Target",
			name:           "ValidName for a Target",
			description:    "ValidDescription, more stuff here",
			mountingPoints: []string{"Bearing A", "Bearing B"},
			expectedErr:    nil,
		},
		{
			test:           "All fields are invalid",
			customer:       "",
			name:           "",
			description:    "",
			mountingPoints: []string{},
			expectedErr:    ErrInvalidTarget,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.test, func(t *testing.T) {
			_, err := NewTarget("", tc.name, "", make([]string, 0))
			if err != tc.expectedErr {
				t.Errorf("Expected error %v, got %v", tc.expectedErr, err)
			}
		})
	}
}

