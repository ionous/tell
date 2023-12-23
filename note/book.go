package note

// collects comments to generate a comment block
// ( during document decoding )
type Book struct {
	book content
}

func (p *Book) BeginCollection(ctx *Context) {
	if ctx != nil {
		p.book.BeginCollection(ctx)
	}
}
func (p *Book) EndCollection() {
	if p.book.ctx != nil {
		p.book.EndCollection()
	}
}
func (p *Book) NextTerm() {
	if p.book.ctx != nil {
		p.book.NextTerm()
	}
}
func (p *Book) Comment(kind Type, str string) (err error) {
	if p.book.ctx != nil {
		err = p.book.Comment(kind, str)
	}
	return
}
func (p *Book) Resolve() (ret string, okay bool) {
	if p.book.ctx != nil {
		ret, okay = p.book.Resolve()
	}
	return
}
