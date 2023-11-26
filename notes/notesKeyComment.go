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
type docKeyDecoder struct {
	*context
}

// awaits an initial paragraph, comment hash, or newline.
func docKey(ctx *context) charm.State {
	d := docKeyDecoder{ctx}
	return charm.Step(d.awaitComment(), d.alwaysBuffer())
}

// everything after an blank line gets buffered together.
func (d *docKeyDecoder) alwaysBuffer() charm.State {
	return charm.Statement("alwaysBuffer", func(q rune) (ret charm.State) {
		if q == runes.Newline {
			ret = awaitParagraph("emptyPadding", func() charm.State {
				return readAll(&d.buf)
			})
		}
		return
	})
}

// awaits the initial paragraph, comment hash
func (d *docKeyDecoder) awaitComment() (ret charm.State) {
	return charm.Statement("awaitComment", func(q rune) (ret charm.State) {
		switch q {
		case runes.Hash:
			keyLine := readLine("keyLine", d.out, d.awaitOutput)
			ret = charm.RunState(q, keyLine)
		case runeParagraph:
			ret = readLine("keyParagraph", d.out, d.awaitOutput)
		}
		return
	})
}

// nest output, or shift to buffering
func (d *docKeyDecoder) awaitOutput() (ret charm.State) {
	return charm.Statement("awaitOutput", func(q rune) (ret charm.State) {
		switch q {
		case runes.Hash:
			ret = nestLine("nestOutput", d.out, d.awaitOutput)
		case runeParagraph:
			ret = readLine("firstBuffer", &d.buf, d.awaitBuffering)
		}
		return
	})
}

// additional lines are added to the existing block in the buffer
// new paragraphs flush the existing buffer, and start buffering a new paragraph
func (d *docKeyDecoder) awaitBuffering() (ret charm.State) {
	return charm.Statement("awaitBuffering", func(q rune) (ret charm.State) {
		switch q {
		case runes.Hash:
			ret = nestLine("nestBuffer", &d.buf, d.awaitBuffering)
		case runeParagraph:
			d.flush(runes.Newline)
			ret = readLine("newBuffer", &d.buf, d.awaitBuffering)
		}
		return
	})
}
