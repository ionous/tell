package charm

import (
	"fmt"
	"strconv"
)

// implements error
type InvalidRune rune

func (e InvalidRune) Error() string {
	var name string
	switch q := rune(e); q {
	case ' ':
		name = "<space>"
	case '\n':
		name = "<newline>"
	case '\t':
		name = "<tab>"
	case Eof:
		name = "<end of file>"
	default:
		if strconv.IsPrint(q) {
			name = strconv.QuoteRune(q)
		} else {
			name = "0x" + strconv.FormatInt(int64(q), 16)
		}
	}
	return fmt.Sprintf("invalid rune: %s", name)
}

// the next rune will return unhandled
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

// ensure the next rune passes the filter
func Require(filter func(r rune) bool) State {
	return Statement("require", func(r rune) (ret State) {
		if !filter(r) {
			ret = Error(InvalidRune(r))
		}
		return
	})
}

// one or more of the runes must pass the filter
func AtleastOne(filter func(r rune) bool) State {
	return Statement("several", func(r rune) (ret State) {
		if filter(r) {
			ret = Optional(filter)
		} else {
			ret = Error(InvalidRune(r))
		}
		return
	})
}
