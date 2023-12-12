package charmed

import (
	"fmt"
	"unicode"

	"github.com/ionous/tell/charm"
	"github.com/ionous/tell/runes"
)

// turns any unhandled states returned by the watched state into errors
func UnhandledError(watch charm.State) charm.State {
	return charm.Self("unhandled error", func(self charm.State, q rune) (ret charm.State) {
		if next := watch.NewRune(q); next == nil {
			ret = charm.Error(fmt.Errorf("unexpected character %q(%d) during %q", q, q,
				charm.StateName(watch)))
		} else {
			ret, watch = self, next // keep checking until watch returns nil
		}
		return
	})
}

// returns an state which errors on all control codes other than newlines
func FilterInvalidRunes() charm.State {
	return charm.Self("filter control codes", func(next charm.State, q rune) charm.State {
		if isInvalidRune(q) {
			e := charm.InvalidRune(q)
			next = charm.Error(e)
		}
		return next
	})
}

func isInvalidRune(q rune) (ret bool) {
	switch q {
	case runes.HTab, runes.Space, runes.Newline, runes.Eof:
		ret = false
	default:
		// this allows through lots of things:
		// half-width spaces, symbols, and the like
		// not sure what's best.
		ret = unicode.IsControl(q)
	}
	return
}
