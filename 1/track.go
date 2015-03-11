package drum

import (
	"fmt"
)

const stepOutputFormat = "|----|----|----|----|"

type Track struct {
	id    uint32
	name  string
	steps [16]bool
}

func (track Track) String() string {
	return fmt.Sprintf("(%d) %s\t%s", track.id, track.name, track.stepString())
}

func (track Track) stepString() string {
	outputLine := []rune(stepOutputFormat) // copy of format to modify
	stepIndex := 0                         // increments only once each step is checked

	for i, char := range outputLine {
		if char == '-' { // is this an output spot?
			if track.steps[stepIndex] {
				outputLine[i] = 'x'
			}
			stepIndex += 1
		}
	}

	return string(outputLine)
}
