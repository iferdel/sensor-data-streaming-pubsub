package pubsub

import "testing"

func TestQueueTypeString(t *testing.T) {
	tests := map[string]struct {
		input QueueType
		want  string
	}{
		"stream queue": {
			input: 2,
			want:  "stream",
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			got := tc.input.String()
			if tc.want != got {
				t.Fatalf("got %v, want %v", got, tc.want)
			}
		})
	}
}
