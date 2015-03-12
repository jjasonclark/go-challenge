package drum

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io"
	"os"
)

var FileError = errors.New("Input file is not a splice file")
var spliceHeader = [6]byte{0x53, 0x50, 0x4c, 0x49, 0x43, 0x45} // SPLICE as bytes
var binaryDecoders = []patternReadPartial{decodeVersion, decodeTempo, decodeTracks}
var trackDataReaders = []trackReadPartial{readTrackId, readTrackName, readTrackSteps}

const InitialTrackCapacity = 10

type spliceFileHeader struct {
	Header [6]byte
	_      uint32
	Length uint32
}

type patternReadPartial func(io.Reader, *Pattern) error
type trackReadPartial func(io.Reader, *Track) error

// DecodeFile decodes the drum machine file found at the provided path
// and returns a pointer to a parsed pattern which is the entry point to the
// rest of the data.
// TODO: implement
func DecodeFile(path string) (*Pattern, error) {
	inputFile, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer inputFile.Close() // Close when function exits
	return Decode(inputFile)
}

func Decode(input io.Reader) (*Pattern, error) {
	header, err := decodeFileHeader(input)
	if err != nil {
		return nil, err
	}

	limitedReader := io.LimitReader(input, int64(header.Length))
	var p Pattern
	for i := 0; i < len(binaryDecoders) && err == nil; i++ {
		err = binaryDecoders[i](limitedReader, &p)
	}
	if err == nil || err == io.EOF {
		return &p, nil
	}
	return nil, err
}

func decodeFileHeader(input io.Reader) (*spliceFileHeader, error) {
	var header spliceFileHeader
	if err := readValue(input, &header); err != nil {
		return nil, FileError
	}
	if header.Header != spliceHeader {
		return nil, FileError
	}
	return &header, nil
}

func convertFromZeroTerminatedString(str []byte) string {
	//trim trailing 0s because string is zero terminated
	return string(bytes.TrimRight(str, "\u0000"))
}

func decodeVersion(input io.Reader, pattern *Pattern) error {
	var version [32]byte
	if err := readValue(input, &version); err != nil {
		return err
	}
	//trim trailing 0s because string is zero terminated
	pattern.Version = string(bytes.TrimRight(version[:], "\u0000"))
	return nil
}

func decodeTempo(input io.Reader, pattern *Pattern) error {
	var tempo float32
	if err := readValue(input, &tempo); err != nil {
		return err
	}
	pattern.Tempo = float64(tempo)
	return nil
}

func decodeTracks(input io.Reader, pattern *Pattern) error {
	output := make([]Track, 0, InitialTrackCapacity)
	var err error
	for err == nil {
		var track Track
		err = readTrack(input, &track)
		if err == nil {
			output = append(output, track)
		}
	}
	if err == nil || err == io.EOF || err == io.ErrUnexpectedEOF {
		pattern.Tracks = output
		return nil
	}
	return err
}

func readTrack(input io.Reader, track *Track) error {
	var err error
	for i := 0; i < len(trackDataReaders) && err == nil; i++ {
		err = trackDataReaders[i](input, track)
	}
	return err
}

func readTrackId(input io.Reader, track *Track) error {
	if err := readValue(input, &track.Id); err != nil {
		return err
	}
	return nil
}

func readTrackName(input io.Reader, track *Track) error {
	var length byte
	if err := readValue(input, &length); err != nil {
		return err
	}
	nameBytes := make([]byte, length)
	if err := readValue(input, &nameBytes); err != nil {
		return err
	}
	track.Name = string(nameBytes[:])
	return nil
}

func readTrackSteps(input io.Reader, track *Track) error {
	var notes [16]byte
	if err := readValue(input, &notes); err != nil {
		return err
	}
	for i, note := range notes {
		track.Steps[i] = note != 0
	}
	return nil
}

func readValue(input io.Reader, data interface{}) error {
	return binary.Read(input, binary.LittleEndian, data)
}
