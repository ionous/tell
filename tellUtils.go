package tell

import (
	"fmt"
	"unicode"

	"github.com/ionous/tell/charm"
)

// turns any unhandled states returned by the watched state into errors
func UnhandledError(watch charm.State) charm.State {
	return charm.Self("unhandled error", func(self charm.State, r rune) (ret charm.State) {
		if next := watch.NewRune(r); next == nil {
			ret = charm.Error(fmt.Errorf("unexpected character %q(%d) during %q", r, r,
				charm.StateName(watch)))
		} else {
			ret, watch = self, next // keep checking until watch returns nil
		}
		return
	})
}

// returns an state which errors on all control codes other than newlines
func FilterControlCodes() charm.State {
	return charm.Self("filter control codes", func(next charm.State, r rune) charm.State {
		if r != Newline && unicode.IsControl(r) {
			e := fmt.Errorf("invalid character %d", r)
			next = charm.Error(e)
		}
		return next
	})
}
