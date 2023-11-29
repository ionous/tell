package notes

// a Commentator implementation which takes no action
type Nothing struct{}

// helper to see whether the implementation of Commentator discards comments
func IsNothing(c Commentator) (okay bool) {
	_, okay = c.(Nothing)
	return
}

func (n Nothing) BeginCollection(RuneWriter) Commentator { return n }
func (n Nothing) OnNestedComment() Commentator           { return n }
func (n Nothing) OnKeyDecoded() Commentator              { return n }
func (n Nothing) OnScalarValue() Commentator             { return n }
func (n Nothing) OnCollectionEnded() Commentator         { return n }
func (n Nothing) WriteRune(rune) (_ int, _ error) {
	return
}
