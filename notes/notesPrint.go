package notes

func NewPrinter(c Commentator) Commentator {
	return printer{c}
}

type printer struct{ c Commentator }

func (p printer) GetComments() string {
	println("GetComments")
	return p.c.GetComments()
}

func (p printer) WriteRune(r rune) (int, error) {
	print(string(r))
	return p.c.WriteRune(r)
}

func (p printer) OnParagraph() Commentator {
	println("OnParagraph")
	p.c.OnParagraph()
	return p
}

func (p printer) OnScalarValue() Commentator {
	println("OnScalarValue")
	p.c.OnScalarValue()
	return p
}

func (p printer) OnBeginCollection() Commentator {
	println("OnBeginCollection")
	p.c.OnBeginCollection()
	return p
}

func (p printer) OnKeyDecoded() Commentator {
	println("OnKeyDecoded")
	p.c.OnKeyDecoded()
	return p
}

func (p printer) OnFootnote() Commentator {
	println("OnFootnote")
	p.c.OnFootnote()
	return p
}
