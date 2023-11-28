package notes

import (
	"github.com/ionous/tell/charm"
	"github.com/ionous/tell/runes"
)

type collectionDecoder struct {
	*context
	keyCommentStart int
}

// flush the current buffer to the new collection
func newCollection(ctx *context) charm.State {
	ctx.newBlock()
	d := collectionDecoder{ctx, 0}
	return d.keyContents()
}

func (d *collectionDecoder) keyContents() charm.State {
	d.keyCommentStart = d.out.Len()
	return charm.Step(keyComments(d.context), d.keyValue())
}

//
// just got a key, handle whatever's next.
//
func (d *collectionDecoder) keyValue() charm.State {
	return charm.Statement("keyValue", func(q rune) (ret charm.State) {
		wroteDash := d.out.Len()-d.keyCommentStart > 0
		switch q {
		case runeTerm:
			// flush any buffer collected from keyComments
			// ( we're stepped to -- its parent -- so we'll hit here if its canceled )
			d.flush(runes.Newline)
			ret = charm.Error(nil) // there's only one buffer, so we're done.

		case runeKey: // a sub-collection
			d.newBlock()
			ret = d.keyContents()

		case runeValue: // a scalar value
			// flush the buffer (from keyComments) to the current collection
			// because there is no new collection.
			d.flush(runes.Newline)
			ret = charm.Step(readTrailing(d.context, wroteDash), d.interElement())

		default: // ex. cant pop before there's a value
			ret = invalidRune("keyValue", q)
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

		case runeTerm:
			// if the document gets terminated before the collection is closed
			// we should be writing any buffered comments to the parent container
			// ( this happens for any document with a top level or deeper sequence )
			// ex. see: TestTermHeaders, TestDocCollection
			if str := d.buf.Resolve(); len(str) > 0 {
				parent := d.stack.top()
				parent.writeTerms()
				writeBuffer(parent, str, runes.Record)
			}
			ret = charm.Error(nil) // there's only one buffer, so we're done.

		// buffer everything
		// the comments will either become footer comments for the parent container
		// or, a header for a new element
		case runes.Hash:
			ret = charm.Step(handleAll(&d.buf), self)

		case runeKey:
			// new key for current container
			// - "prev key"
			// # trailing comment
			// - "new key"
			if d.buf.Len() == 0 { // no trailing comments
				d.out.terms++
			} else {
				// there were some trailing comments
				// this requires an end marker
				d.flush(runes.Record)
			}
			ret = d.keyContents()

		case runeCollected:
			// any buffer right now is for the parent container
			// ( pop handles that )
			if d.pop(); len(d.stack) > 0 {
				ret = self
			}
		default:
			ret = invalidRune("interElement", q)
		}
		return
	})
}
