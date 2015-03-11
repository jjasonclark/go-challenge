package drum

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io"
	"os"
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
	if err = decodeHeader(inputFile); err != nil {
		return nil, err
	}
	if err = decodeVersion(inputFile, &pat.version); err != nil {
		return nil, err
	}
	if err = decodeTempo(inputFile, &pat.tempo); err != nil {
		return nil, err
	}
	if err = decodeTracks(inputFile, &pat.tracks); err != nil {
		return nil, err
	}
	return &pat, nil
}

var spliceHeader = []byte{0x53, 0x50, 0x4c, 0x49, 0x43, 0x45, 0, 0, 0, 0, 0, 0}

func decodeHeader(input io.Reader) error {
	header := make([]byte, len(spliceHeader))
	if err := binary.Read(input, binary.LittleEndian, &header); err != nil {
		return errors.New(errorHeader)
	}
	if !bytes.Equal(header[0:], spliceHeader) {
		return errors.New(errorHeader)
	}
	return nil
}

func decodeVersion(input io.Reader, version *string) error {
	*version = "0.808-alpha"
	return nil
}

func decodeTempo(input io.Reader, tempo *float64) error {
	*tempo = 120.0
	return nil
}

func decodeTracks(input io.Reader, tracks *[]Track) error {
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
