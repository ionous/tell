package notes

import (
	"github.com/ionous/tell/charm"
	"github.com/ionous/tell/runes"
)

// decode every comment encountered into a designated stream
// treats everything that isn't a paragraph, blank line, or comment line as unhandled.
type mulitBlockDecoder struct {
	w runeWriter
}

// assumes the next rune is a comment hash
func readAll(w runeWriter) charm.State {
	d := mulitBlockDecoder{w}
	return readLine("readFirst", w, d.awaitAll)
}

// assumes there's already been a comment hash, and we need to read content.
func handleAll(w runeWriter) charm.State {
	d := mulitBlockDecoder{w}
	return d.handleNext()
}

// read a line without nesting, then await the end of all things.
// any subsequent lines will nest
func (d *mulitBlockDecoder) handleNext() charm.State {
	return handleComment("handleNext", d.w, d.awaitAll)
}

// add new paragraphs, or add lines to existing ones.
func (d *mulitBlockDecoder) awaitAll() charm.State {
	return charm.Statement("awaitAll", func(q rune) (ret charm.State) {
		switch q {
		case runes.HTab:
			ret = nestLine("nestAll", d.w, d.awaitAll)

		case runes.Hash:
			writeRunes(d.w, runes.Newline)
			ret = d.handleNext()

		case runes.Newline:
			ret = awaitParagraph("eatLines", func() charm.State {
				writeRunes(d.w, runes.Newline)
				return d.handleNext()
			})
		}
		return
	})
}
