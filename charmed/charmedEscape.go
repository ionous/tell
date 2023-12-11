package charmed

import (
	"errors"
	"fmt"

	"github.com/ionous/tell/charm"
	"github.com/ionous/tell/runes"
)

func decodeEscape(w runes.RuneWriter, decoded charm.State) charm.State {
	return charm.Statement("decodeEscape", func(q rune) (ret charm.State) {
		if width := hexEncodings[q]; width > 0 {
			var v rune
			ret = charm.Self("captureEscape", func(self charm.State, q rune) (ret charm.State) {
				if x, ok := unhex(q); !ok {
					e := errors.New("syntax error")
					ret = charm.Error(e)
				} else {
					v = v<<4 | x
					if width = width - 1; width == 0 {
						w.WriteRune(v)
						ret = decoded
					} else {
						ret = self
					}
				}
				return
			})
		} else if x, ok := escapes[q]; !ok {
			e := fmt.Errorf("unknown escape %s", runes.RuneName(q))
			ret = charm.Error(e)
		} else {
			w.WriteRune(x)
			ret = decoded // loop...
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
