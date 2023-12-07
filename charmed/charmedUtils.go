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
func FilterControlCodes() charm.State {
	return charm.Self("filter control codes", func(next charm.State, q rune) charm.State {
		if isInvalidRune(q) {
			e := fmt.Errorf("invalid character %q(%d)", q, q)
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
		ret = unicode.IsControl(q)
	}
	return
}
