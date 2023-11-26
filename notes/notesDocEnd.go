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
	return charm.Self("awaitFooter", func(self charm.State, q rune) (ret charm.State) {
		switch q {
		case runeParagraph:
			writeRunes(d.out, runes.Record)
			ret = d.footer()
		case runes.Newline:
			ret = self // keep looping on fully blank lines
		}
		return
	})
}

// read a comment line, then await the end of all things.
func (d *docEndDecoder) footer() charm.State {
	return readLine("footer", d.out, d.awaitTheEnd)
}

// add new paragraphs, or add lines to existing ones.
// everything else errors.
func (d *docEndDecoder) awaitTheEnd() (ret charm.State) {
	return charm.Self("awaitTheEnd", func(self charm.State, q rune) (ret charm.State) {
		switch q {
		case runeParagraph:
			writeRunes(d.out, runes.Newline)
			ret = d.footer()
		case runes.Hash:
			nest(d.out)
			ret = d.footer()
		case runes.Newline:
			ret = self // keep looping on fully blank lines
		default:
			ret = charm.Error(invalidRune("awaitTheEnd", q))
		}
		return
	})
}
