package note

import "strings"

// collects comments to generate a comment block
// ( during document decoding )
type Book struct {
	book content
}

func (p *Book) BeginCollection(buf *strings.Builder) {
	if buf != nil {
		p.book.BeginCollection(buf)
	}
}
func (p *Book) EndCollection() {
	if p.book.buf != nil {
		p.book.EndCollection()
	}
}
func (p *Book) NextTerm() {
	if p.book.buf != nil {
		p.book.NextTerm()
	}
}
func (p *Book) Comment(kind Type, str string) {
	if p.book.buf != nil {
		p.book.Comment(kind, str)
	}
}
func (p *Book) Resolve() (ret string, okay bool) {
	if p.book.buf != nil {
		ret, okay = p.book.Resolve()
	}
	return
}
