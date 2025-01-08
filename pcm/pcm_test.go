package pcm

import (
	"fmt"
	"testing"
	"time"
)

func Test_ForDuration(t *testing.T) {
	encodings := []*Encoder{
		{48000, 1, 1},
		{48000, 2, 1},
		{48000, 1, 2},
		{48000, 2, 2},
	}
	for _, enc := range encodings {
		enc := enc
		cases := []struct {
			duration        time.Duration
			expectedSamples int
		}{
			{0, 0},
			{1, 1},
			{2, 1},
			{3, 1},
			{20 * time.Microsecond, 1},
			{21 * time.Microsecond, 2},
			{41 * time.Microsecond, 2},
			{62 * time.Microsecond, 3},
			{62_499 * time.Nanosecond, 3},
			{62_500 * time.Nanosecond, 3},
			{62_501 * time.Nanosecond, 4},
			{63 * time.Microsecond, 4},
			{125 * time.Microsecond, 6},
			{250 * time.Microsecond, 12},
			{500 * time.Microsecond, 24},
			{time.Millisecond, 48},
		}
		for _, c := range cases {
			c := c
			name := fmt.Sprintf(
				"channels=%d&depth=%d&duration=%s",
				enc.Channels,
				enc.Depth,
				c.duration,
			)
			t.Run(name, func(t *testing.T) {

				t.Run("BytesForDuration", func(t *testing.T) {
					gotBytes, err := enc.BytesForDuration(c.duration)
					if err != nil {
						t.Errorf("unexpected error: %v", err)
					}
					expectedBytes := c.expectedSamples * enc.Depth * enc.Channels
					if expectedBytes != gotBytes {
						t.Errorf("%d (got) != %d (expected)", gotBytes, expectedBytes)
					}
				})

				t.Run("SamplesForDuration", func(t *testing.T) {
					gotSamples := enc.SamplesForDuration(c.duration)
					if c.expectedSamples != gotSamples {
						t.Errorf("%d (got) != %d (expected)", gotSamples, c.expectedSamples)
					}
				})
			})
		}
	}
}
