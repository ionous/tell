package notes

import (
	"fmt"

	"github.com/ionous/tell/charm"
	"github.com/ionous/tell/runes"
)

type makeState func() charm.State

// a state which creates the passed state to handle a rune
func kickOff(m makeState) charm.State {
	return charm.Statement("kickOff", func(q rune) charm.State {
		return charm.RunState(q, m())
	})
}

func invalidRune(name string, q rune) error {
	return fmt.Errorf("unexpected rune %q during %s", q, name)
}

// these runes can be used by authors in comments
// includes htab because authors should be permitted to comment out literals
// and literals can include actual tabs.
// author escape sequences in a comment, ex. an escaped tab \t,
// are two separate and individually permitted runes.
func friendly(q rune) bool {
	return q == runes.HTab || q >= runes.Space
}

func writeRunes(w RuneWriter, qs ...rune) {
	for _, q := range qs {
		w.WriteRune(q)
	}
}

// writes a nest header to the passed writer, and the then reads the rest of the line
func nestLine(name string, w RuneWriter, onEol func() charm.State) (ret charm.State) {
	writeRunes(w, runes.Newline, runes.HTab)
	return readLine(name, w, onEol)
}

// errors if the next rune is not a hash,
// then reads till the end of the comment line.
func readLine(name string, w RuneWriter, onEol func() charm.State) charm.State {
	return charm.Statement(name, func(q rune) (ret charm.State) {
		if q != runes.Hash {
			ret = charm.Error(invalidRune(name, q))
		} else {
			w.WriteRune(runes.Hash)
			ret = innerLine(name, w, onEol)
		}
		return
	})
}

// assumes a comment hash has already been detected, write it and read till the end of the line.
func handleComment(name string, w RuneWriter, onEol func() charm.State) charm.State {
	writeRunes(w, runes.Hash)
	return innerLine(name, w, onEol)
}

// assumes a comment hash has already been read, read till the end of the line.
func innerLine(name string, w RuneWriter, onEol func() charm.State) charm.State {
	return charm.Self(name, func(self charm.State, q rune) (ret charm.State) {
		switch {
		case q == runes.Newline:
			ret = onEol()
		case friendly(q):
			w.WriteRune(q)
			ret = self
		default:
			ret = charm.Error(invalidRune(name, q))
		}
		return
	})
}

// assumes there was just a blank line.
// keep looping until there's a new comment hash
// nesting is not expected ( because you can't nest after a blank line )
func awaitParagraph(name string, onPara func() charm.State) (ret charm.State) {
	return charm.Self(name, func(self charm.State, q rune) (ret charm.State) {
		switch q {
		case runes.Hash:
			ret = onPara()
		case runes.Newline: // keep looping on fully blank lines.
			ret = self
		}
		return
	})
}