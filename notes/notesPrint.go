package notes

// wrap commentator calls with print/ln(s)
func NewPrinter(c Commentator) Commentator {
	return printer{c}
}

type printer struct{ c Commentator }

func (p printer) BeginCollection(w RuneWriter) Commentator {
	println("BeginCollection")
	p.c.BeginCollection(w)
	return p
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

func (p printer) OnCollectionEnded() Commentator {
	println("OnCollectionEnded")
	p.c.OnCollectionEnded()
	return p
}

func (p printer) WriteRune(r rune) (int, error) {
	print(string(r))
	return p.c.WriteRune(r)
}
