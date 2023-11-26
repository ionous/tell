package notes

import (
	"github.com/ionous/tell/charm"
	"github.com/ionous/tell/runes"
)

// decodes inline trailing comment(s)
// ( documents don't have trailing block comments;
//   instead they have document footer comments. )
type docScalarDecoder struct {
	*context
	docEnd makeState
}

// starting immediately after a document scalar has been detected:
// sus out if there's an inline trailing comment,
// and if so: write the value maker, and decode that comment and any nesting.
// eventually, move to doc end.
func docScalar(ctx *context, docEnd makeState) charm.State {
	d := docScalarDecoder{ctx, docEnd}
	return charm.Step(d.awaitInline(), kickOff(docEnd))
}

func (d *docScalarDecoder) awaitInline() charm.State {
	return charm.Statement("awaitInline", func(q rune) (ret charm.State) {
		switch q {
		case runes.Hash:
			// decode a comment line, preceding it with a value marker.
			readInline := readLine("readInline", d.out, d.waitForNest)
			d.out.WriteRune(runes.CollectionMark)
			ret = charm.RunState(runes.Hash, readInline)
		}
		return
	})
}

// keep reading nested comments
func (d *docScalarDecoder) waitForNest() charm.State {
	return charm.Statement("waitForNest", func(q rune) (ret charm.State) {
		switch q {
		case runes.Hash:
			ret = nestLine("readAligned", d.out, d.waitForNest)
		}
		return
	})
}
