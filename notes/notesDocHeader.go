package notes

import (
	"github.com/ionous/tell/charm"
	"github.com/ionous/tell/runes"
)

type docHeader struct {
	*context
}

// the starting point of every tell document.
// by default, comments at the start are "document headers".
// by opting in, authors can designate comments for the first element of the first collection.
//
// opting in happens by nesting comments or having a fully blank line.
// after nesting, it gives the last nested group to the element.
//
// if there is no such collection, treats the buffer as header.
//
func newHeader(ctx *context) charm.State {
	d := docHeader{ctx}
	return d.awaitHeader()
}

// child state: awaits to decode the first header comment;
// allow parent state to decode values
func (d *docHeader) awaitHeader() (ret charm.State) {
	return charm.Self("awaitHeader", func(self charm.State, q rune) (ret charm.State) {
		switch q {
		case runes.Hash:
			ret = handleComment("firstLine", d.out, d.extendHeader)
		case runes.Newline:
			ret = self // keep looping on fully blank lines
		default:
			// unhandled
		}
		return
	})
}

// immediately after the first comment:
// keep reading paragraphs of header, nest the header, or switch to buffering lines.
// allow parent state to decode values
func (d *docHeader) extendHeader() charm.State {
	return charm.Statement("extendHeader", func(q rune) (ret charm.State) {
		switch q {
		case runes.HTab: // nested header, switch to buffering after done
			ret = nestLine("nestHeader", d.out, d.awaitNest)
		case runes.Hash:
			d.out.WriteRune(runes.Newline)
			ret = handleComment("nextLine", d.out, d.extendHeader)
		case runes.Newline: // fully blank line
			ret = d.awaitParagraph() // buf is empty, so dont need to flush
		default:
			// unhandled
		}
		return
	})
}

// after a blank line, start looking for new paragraphs.
// other events are handled by the parent
func (d *docHeader) awaitParagraph() charm.State {
	// buffers comments to send them to the first element of the next collection ( if any )
	return awaitParagraph("docLines", func() charm.State {
		return handleComment("newParagraph", &d.buf, d.awaitBuf)
	})
}

// keep nesting the output, or start buffering.
// allow parent state to decode values
func (d *docHeader) awaitNest() charm.State {
	return charm.Statement("awaitNest", func(q rune) (ret charm.State) {
		switch q {
		case runes.HTab:
			ret = nestLine("nestHeader", d.out, d.awaitNest)
		default:
			ret = d.awaitBuf().NewRune(q)
		}
		return
	})
}

// keep the buffer filled with a maximum of one paragraph.
// allow parent state to decode values
func (d *docHeader) awaitBuf() charm.State {
	return charm.Statement("awaitBuf", func(q rune) (ret charm.State) {
		switch q {
		case runes.HTab: // nest into the current buffered paragraph
			ret = nestLine("nestBuffer", &d.buf, d.awaitBuf)
		case runes.Hash:
			d.flush(runes.Newline) // flush, and begin buffering a new paragraph
			ret = handleComment("bufferLine", &d.buf, d.awaitBuf)
		case runes.Newline:
			d.flush(runes.Newline)   // flush
			ret = d.awaitParagraph() // and begin waiting for a new paragraph
		}
		return
	})
}
