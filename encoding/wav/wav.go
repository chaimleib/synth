package wav

import (
	"bytes"
	"encoding/binary"
	"io"
)

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
	byteDepth   int
}

func NewEncoder(audioFormat, nbrChannels, byteDepth, frequency int) *encoder {
	return &encoder{
		AudioFormat: uint16(audioFormat),
		NbrChannels: uint16(nbrChannels),
		Frequency:   uint32(frequency),
		byteDepth:   byteDepth,
	}
}

func (e *encoder) Encode(r io.Reader) ([]byte, error) {
	buf, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}
	w := new(WAV)
	copy(w.FileTypeBlocID[:], "RIFF")
	copy(w.FileFormatID[:], "WAVE")
	copy(w.FormatBlocID[:], "fmt ")
	copy(w.DataBlocID[:], "data")

	w.BlocSize = 16
	w.AudioFormat = e.AudioFormat
	w.NbrChannels = e.NbrChannels
	w.Frequency = e.Frequency
	w.BytePerBloc = e.NbrChannels * uint16(e.byteDepth)
	w.BytePerSec = uint32(e.Frequency) * uint32(w.BytePerBloc)
	w.BitsPerSample = uint16(8 * e.byteDepth)
	w.DataSize = uint32(len(buf))
	// w.SampledData = buf
	w.FileSize = 16 + w.BlocSize + w.DataSize

	var out bytes.Buffer
	err = binary.Write(&out, binary.LittleEndian, w)
	if err != nil {
		return nil, err
	}
	_, err = out.Write(buf)
	if err != nil {
		return nil, err
	}
	return out.Bytes(), nil
}
