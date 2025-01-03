package pcm

import (
	"errors"
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

// SamplesForDuration returns the number of waveform points which span the
// given duration. In multichannel audio, a point for each channel counts as a
// single sample.
func (enc *Encoder) SamplesForDuration(d time.Duration) int {
	return int(time.Duration(enc.Rate) * d / time.Second)
}

func (enc *Encoder) samplesForBytes(n int) int {
	return n / int(enc.Channels*enc.Depth)
}

// BytesForDuration converts a time.Duration into the number of bytes required
// to encode that amount of audio, given the current Encoder settings.
func (enc *Encoder) BytesForDuration(d time.Duration) (int, error) {
	size := d * time.Duration(enc.Rate*enc.Depth*enc.Channels) / time.Second
	if size > time.Duration(maxInt) {
		return 0, errMaxInt
	}
	return int(size), nil
}

// durationForBytes converts a number of bytes into a time.Duration that that
// number of bytes can represent with the current Encoder settings.
func (enc *Encoder) durationForBytes(length int) time.Duration {
	return time.Duration(length) * time.Second / time.Duration(enc.Rate*enc.Depth*enc.Channels)
}

// MaxAmplitude returns the maximum signal value, based on the Encoder's
// depth. For example, a byte depth of 1 will yield 127, and a depth
// of 2 will yield 32767.
func (enc *Encoder) MaxAmplitude() int {
	return ^(-1 << (enc.Depth*8 - 1))
}

// ZeroValue returns the "at-rest" value. For signed types, this is 0; for
// unsigned types, this is a uint with the greatest bit set, which is
// half of the maximum amplitude, rounded up.
func (enc *Encoder) ZeroValue() int {
	if enc.Depth == 1 {
		return 0x80
	}
	return 0
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

var (
	errMaxInt           = errors.New("exceeded maxint bytes")
	errNegativeDuration = errors.New("negative duration")
)

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
	size, err := b.encoder.BytesForDuration(d)
	if err != nil {
		return err
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

// Reset erases the audio, allowing reuse of a Buffer.
func (b *Buffer) Reset() {
	b.data = b.data[:0]
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

// ReadValue returns the value of sample number i for the given channel.
func (b *Buffer) ReadValue(i, channel int) int {
	if i < 0 || channel < 0 || channel >= b.encoder.Channels {
		return 0 // no such value
	}
	i0 := i*b.encoder.Depth*b.encoder.Channels + b.encoder.Depth*channel
	var result int
	// concatenate the bytes
	for shift := 0; shift < b.encoder.Depth; shift++ {
		result += int(b.data[i0+shift]) << (8 * shift)
	}

	// if representing a signed int, sign-extend
	if b.encoder.Depth != 1 {
		mask := 1 << (b.encoder.Depth*8 - 1) // sign bit mask
		result = (result ^ mask) - mask
	}
	return result
}

// WriteValue changes the value of sample number i for the given channel.
// Assumes that i already exists. If it doesn't, use WriteChanSample instead.
func (b *Buffer) WriteValue(value, i, channel int) {
	if i < 0 || channel < 0 || channel >= b.encoder.Channels {
		return
	}
	i0 := i*b.encoder.Depth*b.encoder.Channels + b.encoder.Depth*channel
	for shift := 0; shift < b.encoder.Depth; shift++ {
		b.data[i0+shift] = byte(0xff & (value >> (8 * shift)))
	}
}
