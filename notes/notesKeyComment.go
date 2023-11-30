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
}

// awaits an initial paragraph, comment hash, or newline.
func keyComments(ctx *context) charm.State {
	d := keyCommentDecoder{ctx}
	return d.awaitFirst()
}

// everything after any blank line gets buffered together.
func (d *keyCommentDecoder) bufferAll(mark bool) charm.State {
	return awaitParagraph("bufferAll", func() charm.State {
		if mark {
			d.out.WriteRune(runes.KeyValue)
		}
		return handleAll(&d.buf)
	})
}

// awaits the comment or new line.
func (d *keyCommentDecoder) awaitFirst() (ret charm.State) {
	return charm.Statement("awaitFirst", func(q rune) (ret charm.State) {
		switch q {
		case runes.Hash:
			d.out.WriteRune(runes.KeyValue)
			ret = handleComment("keyLine", &d.out, d.awaitNest)
		case runes.Newline:
			ret = d.bufferAll(true)
		}
		return
	})
}

// nest output, or shift to buffering
func (d *keyCommentDecoder) awaitNest() (ret charm.State) {
	return charm.Statement("awaitNest", func(q rune) (ret charm.State) {
		switch q {
		case runes.HTab:
			ret = nestLine("nestOutput", &d.out, d.awaitNest)
		case runes.Hash:
			ret = handleComment("firstBuffer", &d.buf, d.awaitBuffering)
		case runes.Newline:
			ret = d.bufferAll(false)
		}
		return
	})
}

// additional lines are added to the existing block in the buffer
// new paragraphs flush the existing buffer, and start buffering a new paragraph
func (d *keyCommentDecoder) awaitBuffering() (ret charm.State) {
	return charm.Statement("awaitBuffering", func(q rune) (ret charm.State) {
		switch q {
		case runes.HTab:
			ret = nestLine("nestBuffer", &d.buf, d.awaitBuffering)
		case runes.Hash:
			// to get here, we must have had a single key or blank line already
			d.flush(runes.Newline)
			ret = handleComment("newBuffer", &d.buf, d.awaitBuffering)
		case runes.Newline:
			ret = d.bufferAll(false)
		}
		return
	})
}
