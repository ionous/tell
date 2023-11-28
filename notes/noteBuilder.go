package notes

import (
	"fmt"
	"slices"

	"github.com/ionous/tell/charm"
	"github.com/ionous/tell/runes"
)

func build(state charm.State) runecast {
	return runecast{state}
}

// adapts the notes api to charm states
type runecast struct {
	state charm.State
}

// the builder and context have to work together to get all the comments properly
func (b *runecast) GetAllComments(ctx *context) (ret []string) {
	b.send(runeTerm)
	//
	if ctx.buf.Len() > 0 {
		panic("buffer should be empty")
	}
	//
	ret = append(ret, ctx.out.Resolve())
	for len(ctx.stack) > 0 {
		prev := ctx.stack.pop()
		ret = append(ret, prev.String())
	}
	slices.Reverse(ret)
	return
}

// internal runes for the Commentator interface:
// one per Commentator method.
const (
	runeTerm      = -1 // early termination; ex. eof
	runeCollected = '\f'
	runeValue     = '\v'
	runeKey       = '\r'
)

// helper for testing: returns b without doing anything.
func (b *runecast) Inplace() Events {
	return b
}

func (b *runecast) OnNestedComment() Events {
	b.send(runes.HTab)
	return b
}

func (b *runecast) OnKeyDecoded() Events {
	b.send(runeKey)
	return b
}

func (b *runecast) OnScalarValue() Events {
	b.send(runeValue)
	return b
}

func (b *runecast) OnCollectionEnded() Events {
	b.send(runeCollected)
	return b
}

func (b *runecast) WriteRune(q rune) (_ int, _ error) {
	b.send(q)
	return
}

func (b *runecast) send(q rune) {
	if next := b.state.NewRune(q); next == nil && q != runeTerm {
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
