package drum

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io"
	"os"
	"strings"
)

const (
	errorHeader = "Input file is not a splice file"
	errorTrack  = "There is an error reading the track data"
)

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

var spliceHeader = []byte{0x53, 0x50, 0x4c, 0x49, 0x43, 0x45} // SPLICE as bytes

type spliceFileHeader struct {
	Header  [6]byte
	Length  uint64
	Version [32]byte
	Tempo   float32
}

func decodeHeader(input io.Reader, p *Pattern) error {
	var header spliceFileHeader
	if err := binary.Read(input, binary.LittleEndian, &header); err != nil {
		return errors.New(errorHeader)
	}
	if !bytes.Equal(header.Header[:], spliceHeader) {
		return errors.New(errorHeader)
	}
	p.version = zeroTerminatedString(header.Version[:])
	p.tempo = float64(header.Tempo)
	return nil
}

func zeroTerminatedString(str []byte) string {
	//trim trailing 0s because string is zero terminated
	return strings.TrimRight(string(str), "\u0000")
}

type spliceFileStep struct {
	Id         uint32
	NameLength byte
	Name       []byte
	Notes      [16]byte
}

func decodeTracks(input io.Reader, tracks *[]Track) error {
	output := make([]Track, 0, 10) // TODO: are these the right values?
	var err error
	for err == nil {
		var step spliceFileStep
		if err = binary.Read(input, binary.LittleEndian, &step.Id); err != nil {
			continue
		}
		if err = binary.Read(input, binary.LittleEndian, &step.NameLength); err != nil {
			continue
		}
		step.Name = make([]byte, step.NameLength)
		if err = binary.Read(input, binary.LittleEndian, step.Name); err != nil {
			continue
		}
		if err = binary.Read(input, binary.LittleEndian, &step.Notes); err != nil {
			continue
		}

		nextTrack := Track{id: step.Id, name: string(step.Name[:])}
		for i, note := range step.Notes {
			nextTrack.steps[i] = note != 0
		}
		output = append(output, nextTrack)
	}
	if err == io.EOF || err == io.ErrUnexpectedEOF {
		*tracks = output
		return nil
	}
	return err
}
