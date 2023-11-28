package notes

// wrap commentator calls with print/ln(s)
func NewPrinter(c Commentator) Commentator {
	return printer{c}
}

type printer struct{ c Commentator }

func (p printer) GetComments() string {
	return p.c.GetComments()
}

func (p printer) WriteRune(r rune) (int, error) {
	print(string(r))
	return p.c.WriteRune(r)
}

func (p printer) OnNestedComment() Events {
	println("OnNestedComment")
	p.c.OnNestedComment()
	return p
}

func (p printer) OnScalarValue() Events {
	println("OnScalarValue")
	p.c.OnScalarValue()
	return p
}

func (p printer) OnKeyDecoded() Events {
	println("OnKeyDecoded")
	p.c.OnKeyDecoded()
	return p
}

func (p printer) OnCollectionEnded() Events {
	println("OnCollectionEnded")
	p.c.OnCollectionEnded()
	return p
}
