package charmed

import (
	"fmt"
	"unicode/utf8"

	"github.com/ionous/tell/charm"
)

// returns error if failed to match, or unhandled on the rune after the matched string.
// the empty string will return unmatched immediately.
func StringMatch(str string) charm.State {
	var idx int // index in str
	return charm.Self("match", func(self charm.State, q rune) (ret charm.State) {
		if cnt := len(str); idx < cnt {
			if match, size := utf8.DecodeRuneInString(str[idx:]); match != q {
				ret = charm.Error(mismatchedString{q, idx})
			} else {
				ret, idx = self, idx+size // loop
			}
		}
		return
	})
}

type mismatchedString struct {
	q  rune
	at int
}

func (m mismatchedString) Error() string {
	return fmt.Sprintf("mismatched on %q at %d", m.q, m.at)
}
