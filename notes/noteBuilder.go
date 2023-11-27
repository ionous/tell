package notes

import (
	"fmt"

	"github.com/ionous/tell/charm"
	"github.com/ionous/tell/runes"
)

func build(state charm.State) Builder {
	return Builder{state}
}

// adapts the notes api to charm states
type Builder struct {
	state charm.State
}

// internal runes for the Commentator interface:
// one per Commentator method.
const (
	runePopped = '\f'
	runeValue  = '\v'
	runeKey    = '\r'
)

// helper for testing: returns b without doing anything.
func (b *Builder) Inplace() Commentator {
	return b
}

func (b *Builder) OnNestedComment() Commentator {
	b.send(runes.HTab)
	return b
}

func (b *Builder) OnKeyDecoded() Commentator {
	b.send(runeKey)
	return b
}

func (b *Builder) OnScalarValue() Commentator {
	b.send(runeValue)
	return b
}

func (b *Builder) WriteRune(q rune) (_ int, _ error) {
	b.send(q)
	return
}

func (b *Builder) GetComments() string {
	// b.state =b.send(runePopped)
	// // pop
	// return ""
	panic("fix")
}

func (b *Builder) send(q rune) {
	if next := b.state.NewRune(q); next == nil {
		// no states left to parse remaining input
		err := fmt.Errorf("unhandled rune %q in %q", q, charm.StateName(b.state))
		panic(err)
	} else if es, ok := next.(charm.Terminal); ok {
		err := fmt.Errorf("error for rune %q in %q %w", q, charm.StateName(b.state), es)
		panic(err)
	} else {
		b.state = next
	}
}
