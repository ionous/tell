package charmed

import (
	"github.com/ionous/tell/charm"
	"github.com/ionous/tell/runes"
)

// update the cursor
func DecodePos(y, x *int) charm.State {
	return charm.Self("cursor", func(self charm.State, q rune) (ret charm.State) {
		switch q {
		case runes.Eof:
			ret = nil // return unhandled; in a parallel state this stops updating
		case runes.Newline:
			(*y)++
			(*x) = 0
			ret = self
		default:
			(*x)++
			ret = self
		}
		return
	})
}
