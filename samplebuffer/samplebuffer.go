package samplebuffer

import (
	"errors"
	"time"
)

type Buffer struct {
	rate     int
	depth    int
	channels int

	data []byte
}

func New(rate, depth, channels int, d time.Duration) (*Buffer, error) {
	b := &Buffer{
		rate:     rate,
		depth:    depth,
		channels: channels,
	}
	if err := b.allocate(d); err != nil {
		return nil, err
	}
	return b, nil
}

func NewSilence(rate, depth, channels int, d time.Duration) (*Buffer, error) {
	b := &Buffer{
		rate:     rate,
		depth:    depth,
		channels: channels,
	}
	l, ok := b.bytesForDuration(d)
	if !ok {
		return nil, errMaxInt
	}
	b.data = make([]byte, l)
	return b, nil
}

const (
	maxInt      = int(^uint(0) >> 1)
	minDuration = 10 * time.Millisecond
)

var (
	errMaxInt           = errors.New("exceeded maxint bytes")
	errNegativeDuration = errors.New("negative duration")
)

// Reset erases the audio, allowing reuse of a Buffer.
func (b *Buffer) Reset() {
	b.data = b.data[:0]
}

func (b *Buffer) SamplesForDuration(d time.Duration) int {
	return int(time.Duration(b.rate) * d / time.Second)
}

func (b *Buffer) samplesForBytes(n int) int {
	return n / int(b.channels*b.depth)
}

// bytesForDuration converts a time.Duration into the number of bytes required
// to encode that amount of audio, given the current Buffer settings.
func (b *Buffer) bytesForDuration(d time.Duration) (int, bool) {
	size := d * time.Duration(b.rate*b.depth*b.channels) / time.Second
	if size > time.Duration(maxInt) {
		return 0, false
	}
	return int(size), true
}

// durationForBytes converts a number of bytes into a time.Duration that that
// number of bytes can represent with the current Buffer settings.
func (b *Buffer) durationForBytes(length int) time.Duration {
	return time.Duration(length) * time.Second / time.Duration(b.rate*b.depth*b.channels)
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
	size, ok := b.bytesForDuration(d)
	if !ok {
		return errMaxInt
	}

	// make-append rounds capacity up to the next memory chunk size.
	b.data = append([]byte(nil), make([]byte, 0, size)...)[:0]
	return nil
}

// WriteChanSample encodes the given audio level for the Buffer's depth and
// writes it to the Buffer. This must be called for each audio channel to
// complete a single sample.
func (b *Buffer) WriteChanSample(x int) {
	b.data = append(b.data, byte(0xff&x))
	if b.depth == 2 {
		b.data = append(b.data, byte(0xff&(x>>8)))
	}
}

// Bytes returns the encoded audio.
func (b *Buffer) Bytes() []byte {
	return b.data
}

// SampleLen returns the number of samples of encoded audio.
func (b *Buffer) SampleLen() int { return b.samplesForBytes(b.Len()) }

// Len returns the number of bytes of encoded audio.
func (b *Buffer) Len() int { return len(b.data) }

// Duration returns the duration of the accumulated audio.
func (b *Buffer) Duration() time.Duration {
	return b.durationForBytes(b.Len())
}

// MaxValue returns the maximum signal value, based on the Buffer's depth. For
// example, a byte depth of 1 will yield 127, and a depth of 2 will yield 32767.
func (b *Buffer) MaxValue() int {
	return ^(-1 << (b.depth*8 - 1))
}
