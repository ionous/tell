package notes

import (
	"github.com/ionous/tell/charm"
	"github.com/ionous/tell/runes"
)

// read an inline or trailing comment ( after a collection scalar )
func readTrailing(ctx *context) charm.State {
	in := trailingDecoder{&ctx.buf}
	return charm.FirstOf("readTrailing", in.awaitInline(), in.awaitBlock())
}

type trailingDecoder struct {
	w RuneWriter
}

// inline trailing comments start on the same line as their value;
// their first line isnt nested.
func (d *trailingDecoder) awaitInline() charm.State {
	return charm.Statement("awaitInline", func(q rune) (ret charm.State) {
		switch q {
		case runes.Hash:
			// decode a comment line, preceding it with a value marker.
			readInline := readLine("readInline", d.w, d.waitForNest)
			writeRunes(d.w, runes.CollectionMark)
			ret = charm.RunState(runes.Hash, readInline)
		}
		return
	})
}

// trailing block comments start on a newline in the document
// and write their first comment as nested
func (d *trailingDecoder) awaitBlock() charm.State {
	return charm.Self("awaitBlock", func(self charm.State, q rune) (ret charm.State) {
		switch q {
		case runes.Newline:
			ret = self
		case runes.Hash:
			writeRunes(d.w, runes.CollectionMark)
			ret = nestLine("readBlock", d.w, d.waitForNest)
		}
		return
	})
}

// keep reading nested comments
func (d *trailingDecoder) waitForNest() charm.State {
	return charm.Statement("waitForNest", func(q rune) (ret charm.State) {
		switch q {
		case runes.Hash:
			ret = nestLine("readAligned", d.w, d.waitForNest)
		}
		return
	})
}
