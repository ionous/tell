package notes

import (
	"github.com/ionous/tell/charm"
	"github.com/ionous/tell/runes"
)

// decode every comment encountered into a designated stream
// treats everything that isn't a paragraph, blank line, or comment line as unhandled.
type mulitBlockDecoder struct {
	w RuneWriter
}

// assumes the next rune is a comment hash
func readAll(w RuneWriter) charm.State {
	d := mulitBlockDecoder{w}
	return d.readAll()
}

// read a line without nesting, then await the end of all things.
// any subsequent lines will nest
func (d *mulitBlockDecoder) readAll() charm.State {
	return readLine("readAll", d.w, d.awaitAll)
}

// add new paragraphs, or add lines to existing ones.
func (d *mulitBlockDecoder) awaitAll() charm.State {
	return charm.Statement("awaitAll", func(q rune) (ret charm.State) {
		switch q {
		case runes.Hash:
			ret = nestLine("nestAll", d.w, d.awaitAll)
		case runeParagraph:
			writeRunes(d.w, runes.Newline)
			ret = d.readAll()
		case runes.Newline:
			ret = awaitParagraph("eatLines", func() charm.State {
				writeRunes(d.w, runes.Newline)
				return d.readAll()
			})
		}
		return
	})
}
