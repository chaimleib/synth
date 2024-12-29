package main

import (
	"bytes"
	"io"
	"log"
	"math"
	"time"

	"github.com/chaimleib/synth/samplebuffer"
	"github.com/hajimehoshi/oto"
)

const (
	sampleRate         = 48000
	channels           = 1
	byteDepth          = 2
	chunkSize          = sampleRate * channels * byteDepth / 100
	duration           = 500 * time.Millisecond
	frequency  float64 = 440.0
	amplitude  float64 = 0.05
	phase      float64 = 0
)

func main() {
	if err := beepTest(); err != nil {
		log.Fatal(err)
	}
}

func beepTest() error {
	p, err := monoPlayer()
	if err != nil {
		return err
	}

	sound, err := tone(duration, frequency, amplitude, phase)
	if err != nil {
		return err
	}

	silence, err := samplebuffer.NewSilence(sampleRate, byteDepth, channels, duration)
	if err != nil {
		return err
	}

	// Chunks only play after completely sent to the player, so finish the chunk.
	chunkFinishLen := (-11 * sound.Len()) % chunkSize
	chunkFinish := make([]byte, chunkFinishLen)

	reader := io.MultiReader(
		bytes.NewReader(sound.Bytes()),
		bytes.NewReader(silence.Bytes()),

		bytes.NewReader(sound.Bytes()),
		bytes.NewReader(silence.Bytes()),

		bytes.NewReader(sound.Bytes()),
		bytes.NewReader(silence.Bytes()),

		bytes.NewReader(sound.Bytes()),
		bytes.NewReader(silence.Bytes()),

		bytes.NewReader(sound.Bytes()),
		bytes.NewReader(silence.Bytes()),

		bytes.NewReader(sound.Bytes()),
		bytes.NewReader(chunkFinish),
	)
	_, err = io.Copy(p, reader)
	return err
}

func monoPlayer() (*oto.Player, error) {
	c, err := oto.NewContext(sampleRate, channels, byteDepth, chunkSize)
	if err != nil {
		return nil, err
	}
	return c.NewPlayer(), nil
}

func tone(duration time.Duration, frequency, amplitude, phase float64) (*samplebuffer.Buffer, error) {
	buf, err := samplebuffer.New(sampleRate, byteDepth, channels, duration)
	if err != nil {
		return nil, err
	}

	// theta0 is initial phase offset in samples
	periodSamples := float64(sampleRate) / frequency
	theta0 := phase * periodSamples / (2 * math.Pi)
	maxAmplitude := buf.MaxValue()
	iAmplitude := int(float64(maxAmplitude) * amplitude)

	for i := 0; i < buf.SamplesForDuration(duration); i++ {
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
