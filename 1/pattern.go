package drum

import (
	"fmt"
	"math"
	"strings"
)

const (
	versionHeader = "Saved with HW Version: "
	tempoHeader   = "Tempo: "
)

// Pattern is the high level representation of the
// drum pattern contained in a .splice file.
// TODO: implement
type Pattern struct {
	version string
	tempo   float64
	tracks  []Track
}

func (pattern Pattern) String() string {
	output := []string{pattern.versionString(), pattern.tempoString()}
	output = append(output, pattern.trackStrings()...)
	output = append(output, []string{""}...) // make the Join below end with a \n
	return strings.Join(output, "\n")
}

func (pattern Pattern) versionString() string {
	return versionHeader + pattern.version
}

func (pattern Pattern) tempoString() string {
	numberFormat := "%f"
	if isWholeNumber(pattern.tempo) {
		numberFormat = "%0.0f"
	}
	return fmt.Sprintf(tempoHeader+numberFormat, pattern.tempo)
}

func isWholeNumber(number float64) bool {
	return math.Floor(number) == number
}

func (pattern Pattern) trackStrings() []string {
	trackLines := make([]string, len(pattern.tracks))
	for i, track := range pattern.tracks {
		trackLines[i] = track.String()
	}
	return trackLines
}
