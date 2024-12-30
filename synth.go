package synth

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/chaimleib/synth/encoding/wav"
	"github.com/chaimleib/synth/pcm"
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

func ExampleTones() (io.Reader, error) {
	enc := pcm.New(sampleRate, byteDepth, channels)

	squareSound, err := enc.Square(duration, frequency, amplitude, phase)
	if err != nil {
		return nil, err
	}

	sawtoothSound, err := enc.Sawtooth(duration, frequency, amplitude, phase)
	if err != nil {
		return nil, err
	}

	triangleSound, err := enc.Triangle(duration, frequency, amplitude, phase)
	if err != nil {
		return nil, err
	}

	sineSound, err := enc.Sine(duration, frequency, amplitude, phase)
	if err != nil {
		return nil, err
	}

	silence, err := enc.NewSilence(duration)
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

func Play(r io.Reader) error {
	p, err := monoPlayer()
	if err != nil {
		return err
	}

	_, err = io.Copy(p, r)
	return err
}

func Save(r io.Reader, fpath string) error {
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
