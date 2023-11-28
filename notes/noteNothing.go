package notes

// a Commentator implementation which takes no action
type Nothing struct{}

// helper to see whether the implementation of Commentator discards comments
func IsNothing(c Events) (okay bool) {
	_, okay = c.(Nothing)
	return
}

func (n Nothing) OnNestedComment() Events   { return n }
func (n Nothing) OnKeyDecoded() Events      { return n }
func (n Nothing) OnScalarValue() Events     { return n }
func (n Nothing) OnCollectionEnded() Events { return n }
func (n Nothing) GetComments() (_ string)   { return }
func (n Nothing) WriteRune(rune) (_ int, _ error) {
	return
}
