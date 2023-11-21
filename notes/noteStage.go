package notes

import "fmt"

// a separate state machine is possible,
// but might break my brain
type blockStage int

//go:generate stringer -type=blockStage
const (
	startStage  blockStage = iota
	keyStage               // key decoded
	valueStage             // collection or scalar received
	footerStage            // received a footer
)

type stageFlags int

func (f *stageFlags) set(c blockStage) (okay bool) {
	if (*f & (1 << c)) == 0 {
		*f |= 1 << c
		okay = true
	}
	return
}

// validates the passed next; panics on error; returns previous stage.
func (c *blockStage) set(next blockStage) (ret blockStage) {
	// cant skip over the value stage
	if at := *c; at < valueStage && ((next <= at) || (next > valueStage)) {
		msg := fmt.Sprintf("skipped from %s to %s with no value given", at, next)
		panic(msg)
	} else {
		ret, *c = at, next
	}
	return
}

func (c blockStage) buffers() bool {
	return c < valueStage
}

func (c blockStage) allowNesting() bool {
	return c != footerStage
}
