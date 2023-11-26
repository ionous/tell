package notes

import (
	"github.com/ionous/tell/charm"
	"github.com/ionous/tell/runes"
)

type docEndDecoder struct {
	*context
}

// begins decoding with a paragraph or newline.
// adds a formfeed to separate comments at the start
// of a document from comments at the end.
func docEnd(ctx *context) charm.State {
	d := docEndDecoder{ctx}
	return d.awaitFooter()
}

// wait to see the first ending document comment
// then write the form feed to separate it from all that came before.
func (d *docEndDecoder) awaitFooter() charm.State {
	return awaitParagraph("awaitFooter", func() charm.State {
		writeRunes(d.out, runes.Record)
		return readAll(d.out)
	})
}
