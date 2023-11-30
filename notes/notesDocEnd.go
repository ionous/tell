package notes

import (
	"github.com/ionous/tell/charm"
	"github.com/ionous/tell/runes"
)

// parsing after a document scalar
// ( or, via an explicit end collection
//   which, tbd, might happen if a top level document collection is indented.
//   b/c any footer would be to the left of the collection keys.
//   document collections normally wind up parsing the end of a document as
//   a header for a "missing" final element.re: interElement )
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
		writeRunes(&d.out, runes.NextRecord)
		return handleAll(&d.out)
	})
}
