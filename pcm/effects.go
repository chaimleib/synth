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

// Fadeout reduces the volume, linearly, until at the end of the buffer the
// volume is zero.
func (b *Buffer) Fadeout(duration time.Duration) {
	count := b.encoder.SamplesForDuration(duration)
	zero := b.encoder.ZeroValue()
	end := b.SampleLen()

	for i := 0; i < count; i++ {
		scale := float64(count-i) / float64(count)
		j := end - count + i
		for channel := 0; channel < b.encoder.Channels; channel++ {
			value := b.ReadValue(j, channel) - zero
			value = int(float64(value)*scale) + zero
			b.WriteValue(value, j, channel)
		}
	}
}

// Fadein increases the volume from zero, linearly, starting from the beginning
// of the buffer.
func (b *Buffer) Fadein(duration time.Duration) {
	count := b.encoder.SamplesForDuration(duration)
	zero := b.encoder.ZeroValue()

	for i := 0; i < count; i++ {
		scale := float64(i) / float64(count)
		for channel := 0; channel < b.encoder.Channels; channel++ {
			value := b.ReadValue(i, channel) - zero
			value = int(float64(value)*scale) + zero
			b.WriteValue(value, i, channel)
		}
	}
}
