package decode

import (
	"github.com/ionous/tell/charm"
	"github.com/ionous/tell/runes"
)

// find the next indent, and use the callback to determine the next state.
// if the callback is null or returns a null state, this pops to find an appropriate state.
func NextIndent(doc *Document, onIndent func(at int) charm.State) charm.State {
	return charm.Self("next indent", func(nextIndent charm.State, r rune) (ret charm.State) {
		if r == runes.Space {
			ret = nextIndent
		} else if r == runes.Newline {
			ret = nextIndent
		} else {
			var next charm.State
			if onIndent != nil {
				next = onIndent(doc.Col)
			}
			if next == nil {
				next = doc.popToIndent()
			}
			if isDone(next) {
				ret = next
			} else {
				ret = next.NewRune(r)
			}
		}
		return
	})
}

// return true if the passed state is unhandled or in error
func isDone(c charm.State) (okay bool) {
	switch c.(type) {
	case nil, charm.Terminal:
		okay = true
	}
	return
}

func MaintainIndent(doc *Document, loop charm.State, depth int) charm.State {
	return NextIndent(doc, func(at int) (ret charm.State) {
		if at == depth {
			ret = loop
		}
		return
	})
}
