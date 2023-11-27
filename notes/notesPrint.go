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

func (p printer) OnNestedComment() Commentator {
	println("OnNestedComment")
	p.c.OnNestedComment()
	return p
}

func (p printer) OnScalarValue() Commentator {
	println("OnScalarValue")
	p.c.OnScalarValue()
	return p
}

func (p printer) OnKeyDecoded() Commentator {
	println("OnKeyDecoded")
	p.c.OnKeyDecoded()
	return p
}
