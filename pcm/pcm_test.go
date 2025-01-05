package pcm

import (
	"testing"
	"time"
)

var enc = &Encoder{
	Rate:     48000,
	Depth:    2,
	Channels: 1,
}

func TestBytesForDuration(t *testing.T) {
	cases := []struct {
		duration time.Duration
		expected int
	}{
		{0, 0},
		{1, 2},
		{2, 2},
		{3, 2},
		{20 * time.Microsecond, 2},
		{21 * time.Microsecond, 4},
		{41 * time.Microsecond, 4},
		{62 * time.Microsecond, 6},
		{62_499 * time.Nanosecond, 6},
		{62_500 * time.Nanosecond, 6},
		{62_501 * time.Nanosecond, 8},
		{63 * time.Microsecond, 8},
		{125 * time.Microsecond, 12},
		{250 * time.Microsecond, 24},
		{500 * time.Microsecond, 48},
		{time.Millisecond, 96},
	}
	for _, c := range cases {
		t.Run(c.duration.String(), func(t *testing.T) {
			got, err := enc.BytesForDuration(c.duration)
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if c.expected != got {
				t.Errorf("%d (got) != %d (expected)", got, c.expected)
			}
		})
	}
}
