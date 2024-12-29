package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"math"
	"os"
	"time"

	"github.com/chaimleib/synth/encoding/wav"
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
	amplitude  float64 = 0.1
	phase      float64 = 0
)

func main() {
	reader, err := beepTest()
	if err != nil {
		log.Fatal(err)
	}
	if err := save(reader, "beep.wav"); err != nil {
		log.Fatal(err)
	}
}

func beepTest() (io.Reader, error) {
	squareSound, err := square(duration, frequency, amplitude, phase)
	if err != nil {
		return nil, err
	}

	sawtoothSound, err := sawtooth(duration, frequency, amplitude, phase)
	if err != nil {
		return nil, err
	}

	triangleSound, err := triangle(duration, frequency, amplitude, phase)
	if err != nil {
		return nil, err
	}

	sineSound, err := sine(duration, frequency, amplitude, phase)
	if err != nil {
		return nil, err
	}

	silence, err := samplebuffer.NewSilence(sampleRate, byteDepth, channels, duration)
	if err != nil {
		return nil, err
	}

	// Chunks only play after completely sent to the player, so finish the chunk.
	chunkFinishLen := (-15 * squareSound.Len()) % chunkSize
	chunkFinish := make([]byte, chunkFinishLen)

	reader := io.MultiReader(
		&printReader{name: "square", buf: squareSound.Bytes()},
		bytes.NewReader(silence.Bytes()),

		&printReader{name: "sawtooth", buf: sawtoothSound.Bytes()},
		bytes.NewReader(silence.Bytes()),

		&printReader{name: "triangle", buf: triangleSound.Bytes()},
		bytes.NewReader(silence.Bytes()),

		&printReader{name: "sine", buf: sineSound.Bytes()},
		bytes.NewReader(silence.Bytes()),

		&printReader{name: "square", buf: squareSound.Bytes()},
		bytes.NewReader(silence.Bytes()),

		&printReader{name: "sawtooth", buf: sawtoothSound.Bytes()},
		bytes.NewReader(silence.Bytes()),

		&printReader{name: "triangle", buf: triangleSound.Bytes()},
		bytes.NewReader(silence.Bytes()),

		&printReader{name: "sine", buf: sineSound.Bytes()},
		bytes.NewReader(chunkFinish),
	)
	return reader, nil
}

func play(r io.Reader) error {
	p, err := monoPlayer()
	if err != nil {
		return err
	}

	_, err = io.Copy(p, r)
	return err
}

func save(r io.Reader, fpath string) error {
	buf, err := wav.NewEncoder(wav.AudioFormatPCM, channels, byteDepth, sampleRate).Encode(r)
	if err != nil {
		return err
	}

	err = os.WriteFile(fpath, buf, 0644)
	if err != nil {
		return err
	}

	return nil
}

type printReader struct {
	name string
	buf  []byte
	i    int
}

var _ io.Reader = (*printReader)(nil)

func (p *printReader) Read(out []byte) (n int, err error) {
	if p.i == 0 {
		fmt.Println(p.name)
	}
	end := p.i + len(out)
	if end > len(p.buf) {
		end = len(p.buf)
	}
	n = copy(out, p.buf[p.i:end])
	p.i += n
	if p.i >= len(p.buf) {
		return n, io.EOF
	}
	return n, nil
}

func monoPlayer() (*oto.Player, error) {
	c, err := oto.NewContext(sampleRate, channels, byteDepth, chunkSize)
	if err != nil {
		return nil, err
	}
	return c.NewPlayer(), nil
}

func square(duration time.Duration, frequency, amplitude, phase float64) (*samplebuffer.Buffer, error) {
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

func sawtooth(duration time.Duration, frequency, amplitude, phase float64) (*samplebuffer.Buffer, error) {
	buf, err := samplebuffer.New(sampleRate, byteDepth, channels, duration)
	if err != nil {
		return nil, err
	}

	// theta0 is initial phase offset in samples
	periodSamples := float64(sampleRate) / frequency
	theta0 := phase * periodSamples / (2 * math.Pi)
	maxAmplitude := buf.MaxValue()
	iAmplitude := float64(maxAmplitude) * amplitude

	for i := 0; i < buf.SamplesForDuration(duration); i++ {
		// waveFraction is how far we have progressed into the waveform's repeating
		// period. It ranges from -0.5 to 0.5.
		waveFraction := math.Remainder(float64(i)+theta0, periodSamples) / periodSamples
		buf.WriteChanSample(int(2 * iAmplitude * waveFraction))
	}

	return buf, nil
}

func triangle(duration time.Duration, frequency, amplitude, phase float64) (*samplebuffer.Buffer, error) {
	buf, err := samplebuffer.New(sampleRate, byteDepth, channels, duration)
	if err != nil {
		return nil, err
	}

	// theta0 is initial phase offset in samples
	periodSamples := float64(sampleRate) / frequency
	theta0 := (phase + math.Pi/2) * periodSamples / (2 * math.Pi)
	maxAmplitude := buf.MaxValue()
	iAmplitude := float64(maxAmplitude) * amplitude

	for i := 0; i < buf.SamplesForDuration(duration); i++ {
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

func sine(duration time.Duration, frequency, amplitude, phase float64) (*samplebuffer.Buffer, error) {
	buf, err := samplebuffer.New(sampleRate, byteDepth, channels, duration)
	if err != nil {
		return nil, err
	}

	// theta0 is initial phase offset in samples
	periodSamples := float64(sampleRate) / frequency
	maxAmplitude := buf.MaxValue()
	iAmplitude := float64(maxAmplitude) * amplitude

	for i := 0; i < buf.SamplesForDuration(duration); i++ {
		theta := float64(i) * 2 * math.Pi / periodSamples
		buf.WriteChanSample(int(iAmplitude * math.Sin(theta+phase)))
	}

	return buf, nil
}
