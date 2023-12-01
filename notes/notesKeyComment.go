package notes

import (
	"github.com/ionous/tell/charm"
	"github.com/ionous/tell/runes"
)

// reads comments in the area after the dash, and before the value.
// the first and middle blocks goes to the container.
// if there's a new collection after the key,
// the last block goes to the new collection
// ( otherwise it also goes to the container. )
// everything after an blank line gets buffered together.
type keyCommentDecoder struct {
	*context
	wroteKey bool // written if there's any comment
}

// awaits an initial paragraph, comment hash, or newline.
func makeKeyComments(ctx *context) keyCommentDecoder {
	return keyCommentDecoder{ctx, false}
}

func (d *keyCommentDecoder) NewRune(q rune) charm.State {
	next := charm.Step(d.awaitFirst(), d.bufferAll())
	return next.NewRune(q)
}

func (d *keyCommentDecoder) writeKey() {
	if d.wroteKey {
		panic("invalid key state")
	}
	d.out.writeTerms()
	d.out.WriteRune(runes.KeyValue)
	d.wroteKey = true
}

// everything after any blank line gets buffered together.
func (d *keyCommentDecoder) bufferAll() charm.State {
	return charm.Statement("bufferAll", func(q rune) (ret charm.State) {
		switch q {
		case runes.Newline:
			ret = awaitParagraph("afterBlankLine", func() charm.State {
				if !d.wroteKey {
					d.writeKey()
				}
				return handleAll(&d.buf)
			})
		}
		return
	})
}

// awaits the comment or new line.
func (d *keyCommentDecoder) awaitFirst() (ret charm.State) {
	return charm.Statement("awaitFirst", func(q rune) (ret charm.State) {
		switch q {
		case runes.Hash:
			d.writeKey()
			ret = handleComment("keyLine", &d.out, d.awaitSecond)
		}
		return
	})
}

// nest or check the third line for nesting.
func (d *keyCommentDecoder) awaitSecond() (ret charm.State) {
	return charm.Statement("awaitSecond", func(q rune) (ret charm.State) {
		switch q {
		case runes.HTab: // nesting opts in *subsequent* comments to the key/header cycle
			ret = nestLine("nestOutput", &d.out, d.splitComments)
		case runes.Hash: // could be all left-aligned, or the third might use nesting.
			ret = handleComment("secondLine", &d.buf, d.awaitThird)
		}
		return
	})
}

// on the third line after two with no nesting
// we might have nesting ( opting in to the alternating key/header cycle. )
// or we might have all comments left aligned ( its all key comments )
func (d *keyCommentDecoder) awaitThird() (ret charm.State) {
	return charm.Statement("awaitThird", func(q rune) (ret charm.State) {
		switch q {
		case runes.HTab: // nesting opts-in to buffering headers
			ret = nestLine("nestOutput", &d.buf, d.splitComments)
		case runes.Hash:
			// three unnested comments means everything is a key comment
			// make sure to flush after reading so it doesnt get moved to a header.
			d.flush(runes.Newline)
			ret = charm.Step(handleAll(&d.buf), charm.OnExit("flush", func() {
				d.flush(runes.Newline)
			}))
		default:
			// when only two comment lines; flush so they dont get moved to a header.
			d.flush(runes.Newline)
		}
		return
	})
}

// additional lines are added to the existing block in the buffer
// new paragraphs flush the existing buffer, and start buffering a new paragraph
func (d *keyCommentDecoder) splitComments() (ret charm.State) {
	return charm.Statement("splitComments", func(q rune) (ret charm.State) {
		switch q {
		case runes.HTab:
			ret = nestLine("nestBuffer", &d.buf, d.splitComments)
		case runes.Hash:
			// to get here, we must have had a single key or blank line already
			d.flush(runes.Newline)
			ret = handleComment("newBuffer", &d.buf, d.splitComments)
		}
		return
	})
}
