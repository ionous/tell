package charmed

import (
	"errors"
	"strconv"
	"strings"

	"github.com/ionous/tell/charm"
	"github.com/ionous/tell/runes"
)

// heredoc headers can produce two kinds of tokens:
// a word composed of one or more printable characters, or
// a stream token equal to the string `<<<`.
// it also recognizes ( but does not report ) any spaces between them.
type headerToken int

//go:generate stringer -type=headerToken
const (
	headerSpaces headerToken = iota
	headerWord
	headerStream
)
const ostream rune = '<'

type headerNotifier func(headerToken) error

// report the end of every word and stream marker
// writing each word into the passed builder.
// ( the caller can reset the builder whenever it wants, this never does. )
func decodeHeaderHere(out *strings.Builder, report headerNotifier) charm.State {
	var curr headCount
	return charm.Self("decodeHeaderHere", func(self charm.State, q rune) (ret charm.State) {
		if e := curr.update(q, report); e != nil {
			ret = charm.Error(e)
		} else {
			if curr.token == headerWord {
				out.WriteRune(q)
			}
			if q != runes.Newline {
				ret = self
			}
		}
		return
	})
}

// creates tokens out of a series of runes
type headCount struct {
	token headerToken
	width int
}

// see if the passed rune extends the existing token
// if not, report the end of that token, and start a new one.
func (h *headCount) update(q rune, report headerNotifier) (err error) {
	if t, ok := classify(q); !ok {
		err = InvalidRune(q)
	} else if prev, width := h.token, h.width; t == prev {
		h.width++
	} else if prev == headerStream && width != 3 {
		err = customTagError
	} else {
		h.token, h.width = t, 1
		if prev != headerSpaces {
			if e := report(prev); e != nil {
				err = e
			}
		}
	}
	return
}

var customTagError = errors.New("custom closing tags require exactly three stream markers ('<<<')")

// determine which header type, if any, the passed rune belongs to
// ( false if its some classifiable rune )
func classify(q rune) (ret headerToken, okay bool) {
	switch q {
	case runes.Space, runes.Newline:
		ret, okay = headerSpaces, true
	case ostream:
		ret, okay = headerStream, true
	default:
		if strconv.IsPrint(q) {
			ret, okay = headerWord, true
		}
	}
	return
}
