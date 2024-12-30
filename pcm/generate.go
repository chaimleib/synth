package pcm

import (
	"math"
	"time"
)

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

// Square generates a square wave.
func (enc *Encoder) Square(duration time.Duration, frequency, amplitude, phase float64) (*Buffer, error) {
	buf, err := enc.NewBuffer(duration)
	if err != nil {
		return nil, err
	}

	// theta0 is initial phase offset in samples
	periodSamples := float64(enc.Rate) / frequency
	theta0 := phase * periodSamples / (2 * math.Pi)
	maxAmplitude := enc.MaxValue()
	iAmplitude := int(float64(maxAmplitude) * amplitude)

	for i := 0; i < enc.SamplesForDuration(duration); i++ {
		// waveSample is which sample within the waveform's repeating period. It
		// ranges from -periodSamples/2 to periodSamples/2.
		waveSample := math.Remainder(float64(i)+theta0, periodSamples)
		x := iAmplitude
		if waveSample < 0 {
			x = -x
		}
		for c := 0; c < enc.Channels; c++ {
			buf.WriteChanSample(x)
		}
	}

	return buf, nil
}

// Sawtooth generates an ascending sawtooth wave.
func (enc *Encoder) Sawtooth(duration time.Duration, frequency, amplitude, phase float64) (*Buffer, error) {
	buf, err := enc.NewBuffer(duration)
	if err != nil {
		return nil, err
	}

	// theta0 is initial phase offset in samples
	periodSamples := float64(enc.Rate) / frequency
	theta0 := phase * periodSamples / (2 * math.Pi)
	maxAmplitude := enc.MaxValue()
	iAmplitude := float64(maxAmplitude) * amplitude

	for i := 0; i < enc.SamplesForDuration(duration); i++ {
		// waveFraction is how far we have progressed into the waveform's repeating
		// period. It ranges from -0.5 to 0.5.
		waveFraction := math.Remainder(float64(i)+theta0, periodSamples) / periodSamples
		x := int(2 * iAmplitude * waveFraction)
		for c := 0; c < enc.Channels; c++ {
			buf.WriteChanSample(x)
		}
	}

	return buf, nil
}

// Triangle generates a triangle wave.
func (enc *Encoder) Triangle(duration time.Duration, frequency, amplitude, phase float64) (*Buffer, error) {
	buf, err := enc.NewBuffer(duration)
	if err != nil {
		return nil, err
	}

	periodSamples := float64(enc.Rate) / frequency
	// theta0 is initial phase offset in samples. Includes an offset so that we
	// start at 0.
	theta0 := (phase + math.Pi/2) * periodSamples / (2 * math.Pi)
	maxAmplitude := enc.MaxValue()
	iAmplitude := float64(maxAmplitude) * amplitude

	for i := 0; i < enc.SamplesForDuration(duration); i++ {
		// waveFraction is how far we have progressed into the waveform's repeating
		// period. It ranges from -0.5 to 0.5.
		waveFraction := math.Remainder(float64(i)+theta0, periodSamples) / periodSamples

		// output varies linearly from -iAmplitude at 0 to iAmplitude at pi, and
		// back down to -iAmplitude at 2pi. Note that this means output is not 0 at
		// phase 0, so theta0 includes an offset to provide for this.
		var x int
		if waveFraction >= 0.0 {
			x = int((4*waveFraction - 1.0) * iAmplitude)
		} else {
			x = int((-4*waveFraction - 1.0) * iAmplitude)
		}
		for c := 0; c < enc.Channels; c++ {
			buf.WriteChanSample(x)
		}
	}

	return buf, nil
}

// Sine generates a sine wave.
func (enc *Encoder) Sine(duration time.Duration, frequency, amplitude, phase float64) (*Buffer, error) {
	buf, err := enc.NewBuffer(duration)
	if err != nil {
		return nil, err
	}

	// theta0 is initial phase offset in samples
	periodSamples := float64(enc.Rate) / frequency
	maxAmplitude := enc.MaxValue()
	iAmplitude := float64(maxAmplitude) * amplitude

	for i := 0; i < enc.SamplesForDuration(duration); i++ {
		theta := float64(i) * 2 * math.Pi / periodSamples
		x := int(iAmplitude * math.Sin(theta+phase))
		for c := 0; c < enc.Channels; c++ {
			buf.WriteChanSample(x)
		}
	}

	return buf, nil
}
