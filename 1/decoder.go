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
	var (
		pat       Pattern
		inputFile *os.File
		err       error
	)

	if inputFile, err = os.Open(path); err != nil {
		return nil, err
	}
	defer inputFile.Close() // Close when function exits
	if err = decodeHeader(inputFile, &pat); err != nil {
		return nil, err
	}
	if err = decodeTracks(inputFile, &pat.tracks); err != nil {
		return nil, err
	}
	return &pat, nil
}

var spliceHeader = [6]byte{0x53, 0x50, 0x4c, 0x49, 0x43, 0x45} // SPLICE as bytes

type spliceFileHeader struct {
	Header  [6]byte
	Length  uint64
	Version [32]byte
	Tempo   float32
}

// type spliceFileStep struct {
// 	Id    uint32
// 	Name  []byte
// 	Notes [16]byte
// }

func decodeHeader(input io.Reader, p *Pattern) error {
	var header spliceFileHeader
	if err := readValue(input, &header); err != nil {
		return FileError
	}
	if header.Header != spliceHeader {
		return FileError
	}
	p.version = zeroTerminatedString(header.Version[:])
	p.tempo = float64(header.Tempo)
	return nil
}

func zeroTerminatedString(str []byte) string {
	//trim trailing 0s because string is zero terminated
	return string(bytes.TrimRight(str, "\u0000"))
}

func decodeTracks(input io.Reader, tracks *[]Track) error {
	output := make([]Track, 0, initialTrackCapacity)
	var err error
	for err == nil {
		var track Track
		if err = readValue(input, &track.id); err != nil {
			continue
		}
		if err = readInstramentName(input, &track.name); err != nil {
			continue
		}
		if err = readNotes(input, &track.steps); err != nil {
			continue
		}
		output = append(output, track)
	}
	if err == io.EOF || err == io.ErrUnexpectedEOF {
		*tracks = output
		return nil
	}
	return err
}

func readInstramentName(input io.Reader, name *string) error {
	var length byte
	if err := readValue(input, &length); err != nil {
		return err
	}
	nameBytes := make([]byte, length)
	if err := readValue(input, nameBytes); err != nil {
		return err
	}
	*name = string(nameBytes[:])
	return nil
}

func readNotes(input io.Reader, steps *[16]bool) error {
	var notes [16]byte
	if err := readValue(input, &notes); err != nil {
		return err
	}
	for i, note := range notes {
		steps[i] = note != 0
	}
	return nil
}

func readValue(input io.Reader, data interface{}) error {
	return binary.Read(input, binary.LittleEndian, data)
}
