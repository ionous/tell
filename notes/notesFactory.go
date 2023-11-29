package notes

import "github.com/ionous/tell/charm"

func DiscardComments() Commentator {
	return Nothing{}
}

// passing w will discard all contents
func NewCommentator(w RuneWriter) (ret Commentator) {
	if w != nil {
		ret = newNotes(w)
	} else {
		ret = DiscardComments()
	}
	return
}

func newNotes(w RuneWriter) *commentBuilder {
	ctx := newContext(w)
	return newCommentBuilder(ctx, newDocument(ctx))
}

func newCommentBuilder(ctx *context, state charm.State) *commentBuilder {
	return &commentBuilder{ctx, makeRunecast(state)}
}

func makeRunecast(state charm.State) runecast {
	return runecast{state}
}

// binds the state machine api to the data used to build comments
// because go doesnt have true vtables,
// to properly wrap the runecast, we have to implement all its methods to return our own pointer
type commentBuilder struct {
	ctx  *context
	cast runecast
}

// helper for testing: returns b without doing anything.
func (p *commentBuilder) Inplace() Commentator {
	return p
}

// tell will pop all its pending collections triggering the final flush
// for testing, sometimes that's a bit annoying
func (p *commentBuilder) OnEof() {
	p.cast.send(runeEof)
}

func (p *commentBuilder) BeginCollection(w RuneWriter) Commentator {
	p.ctx.nextCollection = w
	p.cast.BeginCollection(w)
	return p
}

func (p *commentBuilder) OnNestedComment() Commentator {
	p.cast.OnNestedComment()
	return p
}

func (p *commentBuilder) OnScalarValue() Commentator {
	p.cast.OnScalarValue()
	return p
}

func (p *commentBuilder) OnKeyDecoded() Commentator {
	p.cast.OnKeyDecoded()
	return p
}

func (p *commentBuilder) OnCollectionEnded() Commentator {
	if len(p.ctx.stack) == 0 {
		p.cast.send(runeEof)
	} else {
		p.cast.OnCollectionEnded()
	}
	return p
}

func (p *commentBuilder) WriteRune(r rune) (int, error) {
	return p.cast.WriteRune(r)
}
