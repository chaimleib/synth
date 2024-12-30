package pcm

import (
	"errors"
	"math"
	"time"
)

const (
	maxInt      = int(^uint(0) >> 1)
	minDuration = 100 * time.Millisecond
)

// Encoder specifies how to encode audio into a buffer.
type Encoder struct {
	Rate     int
	Depth    int
	Channels int
}

// New creates a new Encoder.
func New(rate, depth, channels int) *Encoder {
	enc := &Encoder{
		Rate:     rate,
		Depth:    depth,
		Channels: channels,
	}

	return enc
}

// Buffer contains audio data and how it was encoded.
type Buffer struct {
	encoder *Encoder

	data []byte
}

// NewBuffer creates an empty buffer with capacity to store the given duration
// of audio.
func (enc *Encoder) NewBuffer(d time.Duration) (*Buffer, error) {
	b := &Buffer{
		encoder: enc,
	}
	if err := b.allocate(d); err != nil {
		return nil, err
	}
	return b, nil
}

// NewSilence creates a zeroed-out buffer lasting for the given duration of
// audio.
func (enc *Encoder) NewSilence(d time.Duration) (*Buffer, error) {
	b := &Buffer{
		encoder: enc,
	}
	l, ok := enc.bytesForDuration(d)
	if !ok {
		return nil, errMaxInt
	}
	b.data = make([]byte, l)
	return b, nil
}

var (
	errMaxInt           = errors.New("exceeded maxint bytes")
	errNegativeDuration = errors.New("negative duration")
)

// Reset erases the audio, allowing reuse of a Buffer.
func (b *Buffer) Reset() {
	b.data = b.data[:0]
}

// SamplesForDuration returns the number of waveform points which span the
// given duration. In multichannel audio, a point for each channel counts as a
// single sample.
func (enc *Encoder) SamplesForDuration(d time.Duration) int {
	return int(time.Duration(enc.Rate) * d / time.Second)
}

func (enc *Encoder) samplesForBytes(n int) int {
	return n / int(enc.Channels*enc.Depth)
}

// bytesForDuration converts a time.Duration into the number of bytes required
// to encode that amount of audio, given the current Buffer settings.
func (enc *Encoder) bytesForDuration(d time.Duration) (int, bool) {
	size := d * time.Duration(enc.Rate*enc.Depth*enc.Channels) / time.Second
	if size > time.Duration(maxInt) {
		return 0, false
	}
	return int(size), true
}

// durationForBytes converts a number of bytes into a time.Duration that that
// number of bytes can represent with the current Buffer settings.
func (enc *Encoder) durationForBytes(length int) time.Duration {
	return time.Duration(length) * time.Second / time.Duration(enc.Rate*enc.Depth*enc.Channels)
}

// allocate initializes the buffer with enough storage for the given duration of
// audio.
func (b *Buffer) allocate(d time.Duration) error {
	if d < 0 {
		return errNegativeDuration
	}
	// Ensure a minimum buffer duration.
	if d < minDuration {
		d = minDuration
	}

	// Convert d to the number of bytes being requested.
	size, ok := b.encoder.bytesForDuration(d)
	if !ok {
		return errMaxInt
	}

	// make-append rounds capacity up to the next memory chunk size.
	b.data = append([]byte(nil), make([]byte, size)...)[:0]
	return nil
}

// WriteChanSample encodes the given audio level for the Buffer's depth and
// writes it to the Buffer. This must be called for each audio channel to
// complete a single sample.
func (b *Buffer) WriteChanSample(x int) {
	b.data = append(b.data, byte(0xff&x))
	if b.encoder.Depth == 2 {
		b.data = append(b.data, byte(0xff&(x>>8)))
	}
}

// Bytes returns the encoded audio.
func (b *Buffer) Bytes() []byte {
	return b.data
}

// SampleLen returns the number of samples of encoded audio.
func (b *Buffer) SampleLen() int { return b.encoder.samplesForBytes(b.Len()) }

// Len returns the number of bytes of encoded audio.
func (b *Buffer) Len() int { return len(b.data) }

// Duration returns the duration of the accumulated audio.
func (b *Buffer) Duration() time.Duration {
	return b.encoder.durationForBytes(b.Len())
}

// MaxValue returns the maximum signal value, based on the Buffer's depth. For
// example, a byte depth of 1 will yield 127, and a depth of 2 will yield 32767.
func (b *Buffer) MaxValue() int {
	return ^(-1 << (b.encoder.Depth*8 - 1))
}

func (enc *Encoder) Square(duration time.Duration, frequency, amplitude, phase float64) (*Buffer, error) {
	buf, err := enc.NewBuffer(duration)
	if err != nil {
		return nil, err
	}

	// theta0 is initial phase offset in samples
	periodSamples := float64(enc.Rate) / frequency
	theta0 := phase * periodSamples / (2 * math.Pi)
	maxAmplitude := buf.MaxValue()
	iAmplitude := int(float64(maxAmplitude) * amplitude)

	for i := 0; i < enc.SamplesForDuration(duration); i++ {
		// waveSample is which sample within the waveform's repeating period. It
		// ranges from -periodSamples/2 to periodSamples/2.
		waveSample := math.Remainder(float64(i)+theta0, periodSamples)
		if waveSample > 0 {
			buf.WriteChanSample(iAmplitude)
		} else {
			buf.WriteChanSample(-iAmplitude)
		}
	}

	return buf, nil
}

func (enc *Encoder) Sawtooth(duration time.Duration, frequency, amplitude, phase float64) (*Buffer, error) {
	buf, err := enc.NewBuffer(duration)
	if err != nil {
		return nil, err
	}

	// theta0 is initial phase offset in samples
	periodSamples := float64(enc.Rate) / frequency
	theta0 := phase * periodSamples / (2 * math.Pi)
	maxAmplitude := buf.MaxValue()
	iAmplitude := float64(maxAmplitude) * amplitude

	for i := 0; i < enc.SamplesForDuration(duration); i++ {
		// waveFraction is how far we have progressed into the waveform's repeating
		// period. It ranges from -0.5 to 0.5.
		waveFraction := math.Remainder(float64(i)+theta0, periodSamples) / periodSamples
		buf.WriteChanSample(int(2 * iAmplitude * waveFraction))
	}

	return buf, nil
}

func (enc *Encoder) Triangle(duration time.Duration, frequency, amplitude, phase float64) (*Buffer, error) {
	buf, err := enc.NewBuffer(duration)
	if err != nil {
		return nil, err
	}

	// theta0 is initial phase offset in samples
	periodSamples := float64(enc.Rate) / frequency
	theta0 := (phase + math.Pi/2) * periodSamples / (2 * math.Pi)
	maxAmplitude := buf.MaxValue()
	iAmplitude := float64(maxAmplitude) * amplitude

	for i := 0; i < enc.SamplesForDuration(duration); i++ {
		// waveFraction is how far we have progressed into the waveform's repeating
		// period. It ranges from -0.5 to 0.5.
		waveFraction := math.Remainder(float64(i)+theta0, periodSamples) / periodSamples
		// output varies linearly from -iAmplitude at 0 to iAmplitude at pi, and
		// back down to -iAmplitude at 2pi. Note that this means output is not 0 at
		// phase 0, so theta0 includes an offset to provide for this.
		if waveFraction >= 0.0 {
			buf.WriteChanSample(int((4*waveFraction - 1.0) * iAmplitude))
		} else {
			buf.WriteChanSample(int((-4*waveFraction - 1.0) * iAmplitude))
		}
	}

	return buf, nil
}

func (enc *Encoder) Sine(duration time.Duration, frequency, amplitude, phase float64) (*Buffer, error) {
	buf, err := enc.NewBuffer(duration)
	if err != nil {
		return nil, err
	}

	// theta0 is initial phase offset in samples
	periodSamples := float64(enc.Rate) / frequency
	maxAmplitude := buf.MaxValue()
	iAmplitude := float64(maxAmplitude) * amplitude

	for i := 0; i < enc.SamplesForDuration(duration); i++ {
		theta := float64(i) * 2 * math.Pi / periodSamples
		buf.WriteChanSample(int(iAmplitude * math.Sin(theta+phase)))
	}

	return buf, nil
}
