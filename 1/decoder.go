package drum

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io"
	"math"
	"os"
	"strings"
)

const (
	errorHeader  = "Input file is not a splice file"
	errorVersion = "There is an error reading the version identifier"
	errorTempo   = "There is an error reading the tempo value"
	errorTrack   = "There is an error reading the track data"
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

type spliceFileStep struct {
	Name  [12]byte
	Notes [16]byte
}

type spliceFile struct {
	header spliceFileHeader
	body   []spliceFileStep
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
	p.tempo = roundOff(header.Tempo)
	return nil
}

func zeroTerminatedString(str []byte) string {
	//trim trailing 0s because string is zero terminated
	return strings.TrimRight(string(str), "\u0000")
}

func roundOff(num float32) float64 {
	shift := float64(100.0)
	return math.Floor((float64(num)*shift)+0.5) / shift
}

func decodeTracks(input io.Reader, tracks *[]Track) error {
	// id: 1 byte
	// name length: 2 bytes
	// name: name length in bytes
	// steps: 8 bytes
	*tracks = []Track{
		{id: 0, name: "kick", steps: [16]bool{true, false, false, false, true, false, false, false, true, false, false, false, true, false, false, false}},
		{id: 1, name: "snare", steps: [16]bool{false, false, false, false, true, false, false, false, false, false, false, false, true, false, false, false}},
		{id: 2, name: "clap", steps: [16]bool{false, false, false, false, true, false, true, false, false, false, false, false, false, false, false, false}},
		{id: 3, name: "hh-open", steps: [16]bool{false, false, true, false, false, false, true, false, true, false, true, false, false, false, true, false}},
		{id: 4, name: "hh-close", steps: [16]bool{true, false, false, false, true, false, false, false, false, false, false, false, true, false, false, true}},
		{id: 5, name: "cowbell", steps: [16]bool{false, false, false, false, false, false, false, false, false, false, true, false, false, false, false, false}},
	}
	return nil
}
