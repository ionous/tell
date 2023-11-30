package decode

import (
	"github.com/ionous/tell/charm"
	"github.com/ionous/tell/runes"
)

// find the next indent, and use the callback to determine the next state.
// if the callback is null or returns a null state, this pops to find an appropriate state.
func NextIndent(doc *Document, onIndent func(at int) charm.State) charm.State {
	return charm.Self("next indent", func(self charm.State, r rune) (ret charm.State) {
		switch r {
		case runes.Eof:
			ret = charm.Error(nil)
		case runes.Newline:
			doc.notes.WriteRune(runes.Newline)
			ret = self
		case runes.Space, commentLine:
			ret = self
		default:
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

func MaintainIndent(doc *Document, depth int, self charm.State) charm.State {
	return NextIndent(doc, func(at int) (ret charm.State) {
		if at == depth {
			ret = self
		}
		return
	})
}
