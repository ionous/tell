package note

import "strings"

// record comment block
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
func (p *Book) NextKey() {
	if p.book.buf != nil {
		p.book.NextKey()
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
