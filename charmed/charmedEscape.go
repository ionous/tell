package charmed

import (
	"fmt"
	"unicode/utf8"

	"github.com/ionous/tell/charm"
	"github.com/ionous/tell/runes"
)

func escapeString(w runes.RuneWriter, str string) (err error) {
	// not going to win any awards for efficiency that's for sure.
	return charm.ParseEof(str,
		charm.Self("descape", func(self charm.State, q rune) (ret charm.State) {
			if q == runes.Escape {
				ret = charm.Step(decodeEscape(w), self)
			} else if q != runes.Eof {
				w.WriteRune(q)
				ret = self
			}
			return
		}))
}

// starting after a backslash, read an escape encoded rune.
// the subsequent rune will return unhandled.
// \xFF, \uFFFF, \UffffFFFF
func decodeEscape(w runes.RuneWriter) charm.State {
	return charm.Statement("decodeEscape", func(q rune) (ret charm.State) {
		if totalWidth := hexEncodings[q]; totalWidth > 0 {
			var v rune // build this up over multiple steps
			width := totalWidth
			ret = charm.Self("captureEscape", func(self charm.State, q rune) (ret charm.State) {
				if x, ok := unhex(q); !ok {
					e := fmt.Errorf("expected %d hex values", totalWidth)
					ret = charm.Error(e)
				} else {
					v = v<<4 | x
					if width = width - 1; width > 0 {
						ret = self // not done, keep going.
					} else {
						if !utf8.ValidRune(v) {
							ret = charm.Error(charm.InvalidRune(v))
						} else {
							w.WriteRune(v)
							ret = charm.UnhandledNext()
						}
					}
				}
				return
			})
		} else {
			// single replacement escapes:
			if v, ok := escapes[q]; !ok {
				e := fmt.Errorf("%w is not recognized after a backslash", charm.InvalidRune(q))
				ret = charm.Error(e)
			} else {
				w.WriteRune(v)
				ret = charm.UnhandledNext()
			}
		}
		return
	})
}

// from strconv/quote.go - by the go authors under a bsd-style license.
func unhex(c rune) (v rune, ok bool) {
	switch {
	case '0' <= c && c <= '9':
		return c - '0', true
	case 'a' <= c && c <= 'f':
		return c - 'a' + 10, true
	case 'A' <= c && c <= 'F':
		return c - 'A' + 10, true
	}
	return
}

// value encoded escapes
var hexEncodings = map[rune]int{
	'x': 2, // '\x80'
	'u': 4, // '\ue000'
	'U': 8, // ex. '\U0010ffff'
}

// standard escapes
var escapes = map[rune]rune{
	'a':  '\a',
	'b':  '\b',
	'f':  '\f',
	'n':  '\n',
	'r':  '\r',
	't':  '\t',
	'v':  '\v',
	'\\': '\\',
	'"':  '"',
}
