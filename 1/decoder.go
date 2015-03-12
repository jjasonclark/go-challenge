package drum

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io"
	"os"
)

var FileError = errors.New("Input file is not a splice file")

const initialTrackCapacity = 10

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
	var p Pattern

	header, err := decodeFileHeader(input)
	if err != nil {
		return nil, err
	}
	err = readData(io.LimitReader(input, int64(header.Length)), &p)
	if err == nil || err == io.EOF {
		return &p, nil
	}
	return nil, err
}

func readData(input io.Reader, pattern *Pattern) error {
	readers := []patternReadPartial{decodeVersion, decodeTempo, decodeTracks}
	var err error
	for i := 0; i < len(readers) && err == nil; i++ {
		err = readers[i](input, pattern)
	}
	return err
}

var spliceHeader = [6]byte{0x53, 0x50, 0x4c, 0x49, 0x43, 0x45} // SPLICE as bytes

type spliceFileHeader struct {
	Header [6]byte
	_      uint32
	Length uint32
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

// type spliceFileStep struct {
// 	Id    uint32
//  NameLength byte  as length of name
// 	Name  []byte as ascii string
// 	Notes [16]byte as bools
// }

const (
	trackIdLength    = 4
	trackStepsLength = 16
)

type patternReadPartial func(input io.Reader, pattern *Pattern) error
type trackReadPartial func(io.Reader, *Track) error

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
	output := make([]Track, 0, initialTrackCapacity)
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
	readers := []trackReadPartial{readId, readInstramentName, readSteps}
	var err error
	for i := 0; i < len(readers) && err == nil; i++ {
		err = readers[i](input, track)
	}
	return err
}

func readId(input io.Reader, track *Track) error {
	if err := readValue(input, &track.Id); err != nil {
		return err
	}
	return nil
}

func readInstramentName(input io.Reader, track *Track) error {
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

func readSteps(input io.Reader, track *Track) error {
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
