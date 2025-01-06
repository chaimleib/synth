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

// ExampleTones generates test audio. The chunkDuration causes
// silence to be added, if needed, to the end of the audio so that
// the duration is a multiple of chunkDuration. The resulting number of bytes
// per chunk is returned as chunkSize.
func ExampleTones(chunkDuration time.Duration) (r io.Reader, enc *pcm.Encoder, chunkSize int, err error) {
	const (
		sampleRate         = 48000
		channels           = 2
		byteDepth          = 2
		duration           = 500 * time.Millisecond
		frequency  float64 = 440.0
		amplitude  float64 = 0.3
		phase      float64 = 0
		decay              = 250 * time.Millisecond
		fadeout            = 20 * time.Millisecond
	)

	enc = pcm.New(sampleRate, byteDepth, channels)
	square, err := enc.Square(duration, frequency, amplitude, phase)
	if err != nil {
		return nil, nil, 0, err
	}
	square.Fadein(fadeout)
	square.Decay(0, decay)
	square.Fadeout(fadeout)

	sawtooth, err := enc.Sawtooth(duration, frequency, amplitude, phase)
	if err != nil {
		return nil, nil, 0, err
	}
	sawtooth.Fadein(fadeout)
	sawtooth.Decay(0, decay)
	sawtooth.Fadeout(fadeout)

	triangle, err := enc.Triangle(duration, frequency, amplitude, phase)
	if err != nil {
		return nil, nil, 0, err
	}
	triangle.Fadein(fadeout)
	triangle.Decay(0, decay)
	triangle.Fadeout(fadeout)

	sine, err := enc.Sine(duration, frequency, amplitude, phase)
	if err != nil {
		return nil, nil, 0, err
	}
	sine.Fadein(fadeout)
	sine.Decay(0, decay)
	sine.Fadeout(fadeout)

	noise, err := enc.WhiteNoise(duration, amplitude)
	if err != nil {
		return nil, nil, 0, err
	}
	noise.Fadein(fadeout)
	noise.Decay(0, decay)
	noise.Fadeout(fadeout)

	silence, err := enc.NewSilence(duration)
	if err != nil {
		return nil, nil, 0, err
	}

	sequence := []io.Reader{
		&printReader{name: "square", buf: square.Bytes()},
		bytes.NewReader(silence.Bytes()),

		&printReader{name: "sawtooth", buf: sawtooth.Bytes()},
		bytes.NewReader(silence.Bytes()),

		&printReader{name: "triangle", buf: triangle.Bytes()},
		bytes.NewReader(silence.Bytes()),

		&printReader{name: "sine", buf: sine.Bytes()},
		bytes.NewReader(silence.Bytes()),

		&printReader{name: "white noise", buf: noise.Bytes()},
		bytes.NewReader(silence.Bytes()),

		&printReader{name: "square", buf: square.Bytes()},
		bytes.NewReader(silence.Bytes()),

		&printReader{name: "sawtooth", buf: sawtooth.Bytes()},
		bytes.NewReader(silence.Bytes()),

		&printReader{name: "triangle", buf: triangle.Bytes()},
		bytes.NewReader(silence.Bytes()),

		&printReader{name: "sine", buf: sine.Bytes()},
		bytes.NewReader(silence.Bytes()),

		&printReader{name: "white noise", buf: noise.Bytes()},
	}

	if chunkDuration != 0 {
		// On playback, adding silence at the end avoids premature cutoff.
		chunkFinish, err := enc.NewSilence(chunkDuration)
		if err != nil {
			return nil, nil, 0, err
		}
		sequence = append(sequence, bytes.NewReader(chunkFinish.Bytes()))
	}

	return io.MultiReader(sequence...), enc, chunkSize, nil
}

func Play(r io.Reader, enc *pcm.Encoder, chunkSize int) error {
	p, err := monoPlayer(enc, chunkSize)
	if err != nil {
		return err
	}

	_, err = io.Copy(p, r)
	return err
}

func Save(r io.Reader, enc *pcm.Encoder, fpath string) error {
	buf, err := wav.NewEncoder(
		wav.AudioFormatPCM,
		enc.Channels,
		enc.Depth,
		enc.Rate,
	).Encode(r)
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

func monoPlayer(enc *pcm.Encoder, chunkSize int) (*oto.Player, error) {
	c, err := oto.NewContext(enc.Rate, enc.Channels, enc.Depth, chunkSize)
	if err != nil {
		return nil, err
	}
	return c.NewPlayer(), nil
}
