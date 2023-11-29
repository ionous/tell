package notes

import (
	"fmt"

	"github.com/ionous/tell/charm"
	"github.com/ionous/tell/runes"
)

// adapts the notes api to charm states
type runecast struct {
	state charm.State
}

// internal runes for the Commentator interface:
// one per Commentator method.
const (
	runeEof       = -1 // early termination; ex. eof
	runeCollected = '\f'
	runeValue     = '\v'
	runeKey       = '\r'
)

// helper for testing: returns b without doing anything.
func (b *runecast) Inplace() Commentator {
	return b
}

// note: doesnt do anything with the runewriter
// the whole statemachine requires a data sink anyways
// ( see: commentBuilder )
func (b *runecast) BeginCollection(RuneWriter) Commentator {
	b.send(runeKey) // fix? could have a different indicator...
	return b
}

func (b *runecast) OnNestedComment() Commentator {
	b.send(runes.HTab)
	return b
}

func (b *runecast) OnKeyDecoded() Commentator {
	b.send(runeKey)
	return b
}

func (b *runecast) OnScalarValue() Commentator {
	b.send(runeValue)
	return b
}

func (b *runecast) OnCollectionEnded() Commentator {
	b.send(runeCollected)
	return b
}

func (b *runecast) WriteRune(q rune) (_ int, _ error) {
	b.send(q)
	return
}

func (b *runecast) send(q rune) {
	if next := b.state.NewRune(q); next == nil && q != runeEof {
		// no states left to parse remaining input
		err := fmt.Errorf("unhandled rune %q in %q", q, charm.StateName(b.state))
		panic(err)
	} else if es, ok := next.(charm.Terminal); ok && es != charm.Error(nil) {
		err := fmt.Errorf("error for rune %q in %q %w", q, charm.StateName(b.state), es)
		panic(err)
	} else {
		b.state = next
	}
}
