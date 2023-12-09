package notes

import (
	"github.com/ionous/tell/charm"
	"github.com/ionous/tell/runes"
)

type collectionDecoder struct {
	*context
}

// flush the current buffer to the new collection
func newCollection(ctx *context) charm.State {
	ctx.newBlock()
	d := collectionDecoder{ctx}
	return d.keyContents()
}

// reads comments in the area after the dash, and before the value.
// if the value is a collection, the comments are treated as a header for it
// if the value is a scalar, the comments are stored in the parent container.
func (d *collectionDecoder) keyContents() charm.State {
	key := awaitParagraph("keyContents", func() charm.State {
		return handleAll(&d.buf)
	})
	return charm.Step(key, d.keyValue())
}

// just got a key, handle whatever's next
func (d *collectionDecoder) keyValue() charm.State {
	return charm.Statement("keyValue", func(q rune) (ret charm.State) {
		switch q {
		case runeKey: // a sub-collection
			d.newBlock() // passes the buffer along as header
			ret = d.keyContents()

		case runeValue: // a scalar value
			// flush any buffer collected from keyComments to the current collection
			// because there is no new collection; trailing comments write directly to "out".
			wroteKey := d.flush(runes.KeyValue)
			ret = charm.Step(readTrailing(d.context, wroteKey), d.interElement())

		case runes.Eof:
			// flush any buffer collected from keyComments to the current collection
			d.flush(runes.KeyValue)
			ret = charm.Error(nil) // there's only one buffer, so we're done.

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
		// buffer everything
		// the comments will either become footer comments for the parent container
		// or, a header for a new element
		case runes.Hash:
			ret = charm.Step(handleAll(&d.buf), self)

		case runeKey:
			// new key for current container
			// - "prev key"
			// # footer comment ( from the above interElement "buffer everything" )
			// - "new key"
			if d.buf.Len() == 0 { // no footer comments
				d.out.terms++
			} else {
				// there were some trailing comments
				// this requires an end marker
				d.flush(runes.NextRecord)
			}
			ret = d.keyContents()

		case runeCollected:
			// write any buffered comments to the parent container
			// tbd: it makes sense for documents, not sure in general.
			if d.pop(); len(d.stack) > 0 {
				ret = self
			} else {
				// ex. TestDocCollection
				ret = charm.Error(nil)
			}

		case runes.Eof:
			// write any buffered comments to the parent container
			d.pop()
			ret = charm.Error(nil) // there's only one buffer, so we're done.

		default:
			ret = invalidRune("interElement", q)
		}
		return
	})
}
