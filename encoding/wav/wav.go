// Package wav allows writing PCM or float audio data to a WAV file.
package wav

import (
	"bytes"
	"encoding/binary"
	"io"
)

// WAV is the file header layout. Source: https://en.wikipedia.org/wiki/WAV.
type WAV struct {
	// Master RIFF chunk
	FileTypeBlocID [4]byte // value:"RIFF"
	FileSize       uint32  // 4 + 8 + BlocSize + 4 + DataSize
	FileFormatID   [4]byte // value:"WAVE"

	// Format chunk
	FormatBlocID  [4]byte // value:"fmt "
	BlocSize      uint32  // chunk size - 8 bytes, which is 16 here
	AudioFormat   uint16
	NbrChannels   uint16
	Frequency     uint32
	BytePerSec    uint32 // Frequency * BytePerBloc
	BytePerBloc   uint16 // NbrChannels * BitsPerSample / 8
	BitsPerSample uint16
	DataBlocID    [4]byte // value:"data"

	// Data chunk
	DataSize uint32
	// SampledData has to be appended manually, since encoding/binary only works
	// with fixed-length fields.
	// SampledData []byte
}

const (
	AudioFormatPCM   = 1
	AudioFormatFloat = 3
)

type encoder struct {
	AudioFormat uint16
	NbrChannels uint16
	Frequency   uint32
	ByteDepth   int
}

// NewEncoder creates a new encoder, which describes the format of the audio
// samples to be encoded.
func NewEncoder(audioFormat, nbrChannels, byteDepth, frequency int) *encoder {
	return &encoder{
		AudioFormat: uint16(audioFormat),
		NbrChannels: uint16(nbrChannels),
		Frequency:   uint32(frequency),
		ByteDepth:   byteDepth,
	}
}

// Encode takes an io.Reader and returns a buffer containing a WAV file.
func (e *encoder) Encode(r io.Reader) ([]byte, error) {
	buf, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}
	w := new(WAV)

	// Set all the fixed string fields.
	copy(w.FileTypeBlocID[:], "RIFF")
	copy(w.FileFormatID[:], "WAVE")
	copy(w.FormatBlocID[:], "fmt ")
	copy(w.DataBlocID[:], "data")

	// Base description fields.
	w.BlocSize = 16
	w.AudioFormat = e.AudioFormat
	w.NbrChannels = e.NbrChannels
	w.Frequency = e.Frequency

	// Derived description fields.
	w.BytePerBloc = e.NbrChannels * uint16(e.ByteDepth)
	w.BytePerSec = e.Frequency * uint32(w.BytePerBloc)
	w.BitsPerSample = 8 * uint16(e.ByteDepth)
	w.DataSize = uint32(len(buf))
	w.FileSize = 16 + w.BlocSize + w.DataSize

	var out bytes.Buffer

	// Write the fixed-length struct fields.
	err = binary.Write(&out, binary.LittleEndian, w)
	if err != nil {
		return nil, err
	}

	// Write the audio samples.
	_, err = out.Write(buf)
	if err != nil {
		return nil, err
	}

	return out.Bytes(), nil
}
