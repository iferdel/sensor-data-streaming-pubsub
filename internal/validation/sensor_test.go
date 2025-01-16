package validation

import "testing"

func TestSensorSerialNumberInputHasValidCharacters(t *testing.T) {
	cases := []struct {
		input string
		want  bool
	}{
		{input: "AAD-123", want: true},
		{input: "AAD12313", want: true},
		{input: "AAD123130", want: false}, // more than 8 characters
		{input: "----", want: false},
		{input: "<AAD-123>", want: false},
		{input: "$AA'-1`3>", want: false},
	}

	for _, tc := range cases {
		got := HasValidCharacters(tc.input)
		want := tc.want
		if got != want {
			t.Errorf("sensor serial number input %v: got %v, want %v", tc.input, got, want)
		}
	}
}
