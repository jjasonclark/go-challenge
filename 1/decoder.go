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
	if _, err := decodeTracks(input, &p, header.Length-36); err != nil {
		return nil, err
	}
	p.Version = convertFromZeroTerminatedString(header.Version[:])
	p.Tempo = float64(header.Tempo)
	return &p, nil
}

var spliceHeader = [6]byte{0x53, 0x50, 0x4c, 0x49, 0x43, 0x45} // SPLICE as bytes

type spliceFileHeader struct {
	Header  [6]byte
	Length  uint64
	Version [32]byte
	Tempo   float32
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

type foo func(input io.Reader, pattern *Pattern, maxLength uint64) (uint64, error)
type trackReadPartial func(io.Reader, *Track) (uint64, error)

func decodeVersion(input io.Reader, pattern *Pattern, maxLength uint64) (uint64, error) {
	var version [32]byte
	if err := readValue(input, &version); err != nil {
		return 0, err
	}
	//trim trailing 0s because string is zero terminated
	pattern.Version = string(bytes.TrimRight(version[:], "\u0000"))
	return 32, nil
}

func decodeTempo(input io.Reader, pattern *Pattern, maxLength uint64) (uint64, error) {
	var tempo float32
	if err := readValue(input, &tempo); err != nil {
		return 0, err
	}
	pattern.Tempo = float64(tempo)
	return 4, nil
}

func decodeTracks(input io.Reader, pattern *Pattern, maxLength uint64) (uint64, error) {
	output := make([]Track, 0, initialTrackCapacity)
	var err error
	var totalRead, bytesRead uint64
	for err == nil && maxLength > 0 {
		var track Track
		bytesRead, err = readTrack(input, &track, maxLength)
		totalRead += bytesRead
		maxLength -= bytesRead
		if err == nil {
			output = append(output, track)
		}
	}
	if err == nil || err == io.EOF || err == io.ErrUnexpectedEOF {
		pattern.Tracks = output
		return totalRead, nil
	}
	return totalRead, err
}

func readTrack(input io.Reader, track *Track, maxLength uint64) (uint64, error) {
	readers := []trackReadPartial{readId, readInstramentName, readSteps}
	var totalRead, bytesRead uint64
	var err error
	for i := 0; i < len(readers) && err == nil; i++ {
		if maxLength <= totalRead {
			return totalRead, io.EOF
		}
		bytesRead, err = readers[i](input, track)
		totalRead += bytesRead
	}
	return totalRead, err
}

func readId(input io.Reader, track *Track) (uint64, error) {
	if err := readValue(input, &track.Id); err != nil {
		return 0, err
	}
	return trackIdLength, nil
}

func readInstramentName(input io.Reader, track *Track) (uint64, error) {
	var length byte
	if err := readValue(input, &length); err != nil {
		return 0, err
	}
	nameBytes := make([]byte, length)
	if err := readValue(input, &nameBytes); err != nil {
		return 1, err
	}
	track.Name = string(nameBytes[:])
	return uint64(length + 1), nil
}

func readSteps(input io.Reader, track *Track) (uint64, error) {
	var notes [16]byte
	if err := readValue(input, &notes); err != nil {
		return 0, err
	}
	for i, note := range notes {
		track.Steps[i] = note != 0
	}
	return trackStepsLength, nil
}

func readValue(input io.Reader, data interface{}) error {
	return binary.Read(input, binary.LittleEndian, data)
}
