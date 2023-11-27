package notes

import (
	"github.com/ionous/tell/charm"
	"github.com/ionous/tell/runes"
)

// read an inline or trailing comment ( after a collection scalar )
func readTrailing(w RuneWriter, wroteDash bool) charm.State {
	in := trailingDecoder{w, !wroteDash}
	return in.awaitComment()
}

type trailingDecoder struct {
	w     RuneWriter
	extra bool
}

func (d *trailingDecoder) writeMark() {
	d.w.WriteRune(runes.CollectionMark)
	if d.extra {
		d.w.WriteRune(runes.CollectionMark)
	}
}

// trailing block comments start on a newline in the document
// and write their first comment as nested
func (d *trailingDecoder) awaitComment() charm.State {
	return charm.Statement("awaitComment", func(q rune) (ret charm.State) {
		switch q {
		case runes.Hash: // its an inline comment...
			d.writeMark()
			ret = handleComment("firstInline", d.w, d.awaitNested)
		case runes.Newline: // now, we see it might be a block.
			ret = d.awaitBlock()
		}
		return
	})
}

// alternate entry for doc scalar, which only has inline comments
// inline trailing comments start on the same line as their value
// ( alt: would be child(await block) and parent(await inline) with block jumping out
//   to "readBlock" after first newline... charm makes that a bit icky. )
func (d *trailingDecoder) awaitInline() charm.State {
	return charm.Statement("awaitInline", func(q rune) (ret charm.State) {
		switch q {
		case runes.Hash:
			d.writeMark()
			ret = handleComment("firstInline", d.w, d.awaitNested)
		}
		return
	})
}

// trailing block comments start after a newline
func (d *trailingDecoder) awaitBlock() charm.State {
	return charm.Self("awaitBlock", func(self charm.State, q rune) (ret charm.State) {
		switch q {
		case runes.Newline: // keep looking
			ret = self
		case runes.HTab: // after the newline, the comment should be indented:
			d.writeMark()
			ret = nestLine("firstBlock", d.w, d.awaitNested)
		}
		return
	})
}

// keep reading nested comments
func (d *trailingDecoder) awaitNested() charm.State {
	return charm.Statement("awaitNested", func(q rune) (ret charm.State) {
		switch q {
		case runes.HTab:
			ret = nestLine("readAligned", d.w, d.awaitNested)
		}
		return
	})
}
