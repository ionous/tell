package charmed

import (
	"errors"

	"github.com/ionous/tell/charm"
)

// determine whether the first line of the heredoc has a custom closing tag
func decoodeTag(tag *[]rune) charm.State {
	type headerStage int
	const (
		waitingForRedirect headerStage = iota
		waitingForTag
		haveTag
	)
	var buf runeSlice
	var stage headerStage

	return decodeHeaderHere(&buf, func(header headerToken) (err error) {
		switch header {
		case headerWord:
			if stage == haveTag {
				err = errors.New("expected an end tag marker with one word")
			} else if stage == waitingForTag {
				*tag, buf = buf, nil
				stage = haveTag
			} else {
				buf = buf[:0] // reset
			}

		case headerRedirect:
			if stage != waitingForRedirect {
				err = errors.New("expected at most once end tag maker")
			} else {
				stage = waitingForTag
			}
			buf = buf[:0] // reset
		}
		return
	})
}

// accumulate runes without converting to a string
// ( b/c the heredoc decoder searches for the closing tag rune by rune )
type runeSlice []rune

func (rs *runeSlice) WriteRune(q rune) (_ int, _ error) {
	(*rs) = append(*rs, q)
	return
}
