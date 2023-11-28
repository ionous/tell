package notes

import (
	"github.com/ionous/tell/charm"
	"github.com/ionous/tell/runes"
)

func doNothing() charm.State {
	return nil
}

// for testing: write a comment and a newline
// to write a fully blank line, pass the empty string
func WriteLine(w RuneWriter, str string) {
	if len(str) > 0 {
		w.WriteRune(runes.Hash)
		w.WriteRune(runes.Space)
		for _, r := range str {
			w.WriteRune(r)
		}
	}
	w.WriteRune(runes.Newline)
}
