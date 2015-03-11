package drum

import "errors"

const (
	errorVersion = "There is an error reading the version identifier"
	errorTempo   = "There is an error reading the tempo value"
	errorTrack   = "There is an error reading the track data"
)

// DecodeFile decodes the drum machine file found at the provided path
// and returns a pointer to a parsed pattern which is the entry point to the
// rest of the data.
// TODO: implement
func DecodeFile(path string) (*Pattern, error) {
	// open file for reading
	// extract file header
	// extract version
	// extract tempo
	// extract tracks
	var pat Pattern

	input := []byte{0}
	if err := decodeVersion(input, &pat.version); err != nil {
		return nil, errors.New(errorVersion)
	}
	if err := decodeTempo(input, &pat.tempo); err != nil {
		return nil, errors.New(errorTempo)
	}
	if err := decodeTracks(input, &pat.tracks); err != nil {
		return nil, errors.New(errorTrack)
	}
	return &pat, nil
}

func decodeVersion(input []byte, version *string) error {
	*version = "0.808-alpha"
	return nil
}

func decodeTempo(input []byte, tempo *float64) error {
	*tempo = 120.0
	return nil
}

func decodeTracks(input []byte, tracks *[]Track) error {
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
