package notes

// a Commentator implementation which takes no action
type Nothing struct{}

// helper to see whether the implementation of Commentator discards comments
func IsNothing(c Commentator) bool {
	t, ok := c.(interface{ IsNothing() bool })
	isNothing := ok && t.IsNothing()
	return isNothing
}

func (n Nothing) IsNothing() bool                        { return true }
func (n Nothing) BeginCollection(RuneWriter) Commentator { return n }
func (n Nothing) OnNestedComment() Commentator           { return n }
func (n Nothing) OnKeyDecoded() Commentator              { return n }
func (n Nothing) OnScalarValue() Commentator             { return n }
func (n Nothing) OnCollectionEnded() Commentator         { return n }
func (n Nothing) OnEof()                                 {}
func (n Nothing) WriteRune(rune) (_ int, _ error) {
	return
}
