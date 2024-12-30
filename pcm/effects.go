package pcm

import (
	"math"
	"time"
)

// Decay exponentially reduces the volume beginning from the start point,
// reducing the volume by half for every halfLife afterwards.
func (b *Buffer) Decay(start, halfLife time.Duration) {
	hl := b.encoder.SamplesForDuration(halfLife)
	scale := float64(1)
	factor := math.Pow(0.5, 1.0/float64(hl))
	zero := b.encoder.ZeroValue()

	for i := b.encoder.SamplesForDuration(start); i < b.SampleLen(); i++ {
		scale *= factor
		for channel := 0; channel < b.encoder.Channels; channel++ {
			value := b.ReadValue(i, channel) - zero
			value = int(float64(value)*scale) + zero
			b.WriteValue(value, i, channel)
		}
	}
}
