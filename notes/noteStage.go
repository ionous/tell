package notes

import "fmt"

// a separate state machine is possible,
// but might break my brain
type blockStage int

//go:generate stringer -type=blockStage
const (
	emptyStage blockStage = iota
	startingStage
	headerStage
	subheaderStage
	paddingStage
	bufferStage
	inlineStage
	footerStage
)

type stageFlags int

func (f *stageFlags) update(c blockStage) (ret bool) {
	if (*f & (1 << c)) == 0 {
		*f |= 1 << c
		ret = true
	}
	return
}

func (c blockStage) allowNesting() (okay bool) {
	switch c {
	case headerStage, paddingStage, inlineStage:
		okay = true
	}
	return
}

func (c blockStage) allowMultiple() (okay bool) {
	switch c {
	case subheaderStage, bufferStage, footerStage:
		okay = true
	}
	return
}

func (c *blockStage) set(next blockStage) {
	switch at := *c; {
	case next < at:
		msg := fmt.Sprintf("can't revert from %s to %s", at, next)
		panic(msg)
	case next == at && !next.allowMultiple():
		msg := fmt.Sprintf("%s doesnt support multiple lines", at)
		panic(msg)
	case at != next:
		*c = next
	}
	return
}
