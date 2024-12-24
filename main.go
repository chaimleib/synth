package main

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"math"
	"time"

	"github.com/hajimehoshi/oto"
)

const (
	sampleRate         = 48000
	channels           = 1
	byteDepth          = 2
	bufSize            = sampleRate * channels * byteDepth / 100
	duration           = 2000 * time.Millisecond
	frequency  float64 = 440.0
	amplitude  float64 = 0.01
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

	ctx := context.Background()
	sound, err := tone(ctx, duration, frequency, amplitude, phase)
	_, err = io.Copy(p, sound)
	return err
}

func monoPlayer() (*oto.Player, error) {
	c, err := oto.NewContext(sampleRate, channels, byteDepth, sampleRate)
	if err != nil {
		return nil, err
	}
	return c.NewPlayer(), nil
}

func tone(ctx context.Context, duration time.Duration, frequency, amplitude, phase float64) (io.Reader, error) {
	var (
		buf bytes.Buffer
		i   uint64
	)

	// theta0 is initial phase offset in samples
	var periodSamples float64 = sampleRate / frequency
	fmt.Println("periodSamples:", periodSamples)
	var theta0 float64 = phase * periodSamples / (2 * math.Pi)
	var maxAmplitude int = math.MaxUint16
	if byteDepth == 1 {
		maxAmplitude = math.MaxUint8
	}
	var iAmplitude float64 = float64(maxAmplitude) * amplitude
	fmt.Println("iAmplitude:", iAmplitude)
	for i = 0; time.Duration(i)*time.Second/sampleRate < duration; i++ {
		// waveSample is which sample within the waveform's repeating period
		waveSample := math.Remainder(float64(i)+theta0, periodSamples)
		// fmt.Println("waveSample:", waveSample)
		x := uint(0)
		if waveSample > 0 {
			x = uint(iAmplitude)
		}
		buf.WriteByte(byte(x & 0xff))
		if byteDepth > 1 {
			buf.WriteByte(byte((x >> 8) & 0xff))
		}
	}
	fmt.Println("samples:", i)
	for ; i%bufSize != 0; i++ {
		buf.WriteByte(0)
	}
	fmt.Println("buffer samples:", i)
	fmt.Println("buffer length:", buf.Len())
	return &buf, nil
}
