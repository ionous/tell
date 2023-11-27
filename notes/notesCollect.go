package notes

import (
	"github.com/ionous/tell/charm"
	"github.com/ionous/tell/runes"
)

type collectionDecoder struct {
	*context
	endOfDocument   makeState
	keyCommentStart int
}

// flush the current buffer to the new collection
func newCollection(ctx *context, endOfDocument makeState) charm.State {
	ctx.newBlock()
	d := collectionDecoder{ctx, endOfDocument, 0}
	return d.keyContents()
}

func (d *collectionDecoder) keyContents() charm.State {
	d.keyCommentStart = d.out.Len()
	return charm.Step(keyComments(d.context), d.keyValue())
}

// just got a key rune, handle whatever's next.
func (d *collectionDecoder) keyValue() charm.State {
	return charm.Statement("keyValue", func(q rune) (ret charm.State) {
		wroteDash := d.out.Len()-d.keyCommentStart > 0
		switch q {
		case runeKey: // a sub-collection
			d.newBlock()
			ret = d.keyContents()

		case runeValue: // a scalar value
			// the buffer cant be used as a header for a collection...
			// because there is no collection. so flush the buffer to out
			d.flush(runes.Newline)
			ret = charm.Step(readTrailing(d.out, wroteDash), d.interElement())

		default: // ex. cant pop before there's a value
			ret = charm.Error(invalidRune("keyValue", q))
		}
		return
	})
}

// after a value:
// handle any new keys in the collection,
// handle pops out to a higher collection,
// and any inter elements comments
func (d *collectionDecoder) interElement() charm.State {
	return charm.Self("interElement", func(self charm.State, q rune) (ret charm.State) {
		switch q {
		// buffer everything
		// the comments will either become footer comments for the parent container
		// or, a header for a new element
		case runeParagraph:
			ret = charm.Step(readAll(&d.buf), self)

		case runeKey:
			// new key for current container
			if d.buf.Len() == 0 {
				d.out.terms++
			} else {
				d.flush(runes.Record)
			}
			ret = d.keyContents()

		case runePopped:
			if d.stack.pop(); len(d.stack) == 0 {
				ret = d.endOfDocument()
			} else {
				ret = self
			}
			d.flush(runes.Newline) // use the buffer in the parent container
		default:
			ret = charm.Error(invalidRune("interElement", q))
		}
		return
	})
}
