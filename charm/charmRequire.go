package charm

import "errors"

// the next rune returns unhandled
func UnhandledNext() State {
	return Statement("unhandled next", func(r rune) (_ State) {
		return
	})
}

// zero or more of the runes must pass the filter
func Optional(filter func(r rune) bool) State {
	return Self("optional", func(self State, r rune) (ret State) {
		if filter(r) {
			ret = self
		}
		return
	})
}

// one or more of the runes must pass the filter
func AtleastOne(filter func(r rune) bool) State {
	return Statement("require", func(r rune) (ret State) {
		if filter(r) {
			ret = Optional(filter)
		} else {
			e := errors.New("unexpected rune")
			ret = Error(e)
		}
		return
	})
}
