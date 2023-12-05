package token

import (
	"errors"
	"strings"
	"unicode"

	"github.com/ionous/tell/charm"
	"github.com/ionous/tell/runes"
)

// parses a dictionary key of ascii words separated by, and terminating with, a colon.
// the words must start with a letter, but can contain spaces and underscores.
// ex. `a:`, `a:b:`, `and:more complex:keys_like_this:`
type Signature struct {
	strings.Builder
	lastSep int
}

// for now defined as unicode is letter, but might be useful to be more lenient
var isValidSignaturePrefix = unicode.IsLetter

func (sig *Signature) Pending() bool {
	return sig.lastSep == 0 || (sig.lastSep < sig.Len())
}

// first character of the signature must be a letter
// subsequent characters of words can be letters, numbers, spaces, or "connectors" (underscore)
// colons separate word parts
func (sig *Signature) Decoder() charm.State {
	decode := sig.lede
	return charm.Self("signature", func(self charm.State, q rune) (ret charm.State) {
		if done, e := decode(q); e != nil {
			ret = charm.Error(e)
		} else if !done {
			ret, decode = self, sig.body
		}
		return
	})
}

func (sig *Signature) lede(q rune) (done bool, err error) {
	if !isValidSignaturePrefix(q) {
		err = errors.New("keys must start with a letter")
	} else {
		sig.WriteRune(q)
	}
	return
}

func (sig *Signature) body(q rune) (done bool, err error) {
	switch {
	case runes.IsWhitespace(q) && !sig.Pending():
		done = true

	case q == runes.Newline || q == runes.Eof:
		if sig.Pending() {
			err = errors.New("keys can't span lines")
		}

	case q == runes.Colon: // aka, a colon
		if !sig.Pending() {
			err = errors.New("words in signatures should be separated by a single colon")
		} else {
			sig.WriteRune(q)        // the signature includes the separator
			sig.lastSep = sig.Len() // makes it not pending till next valid rune
		}

	case q == runes.Space || q == runes.Underscore || q == runes.Dash || unicode.IsDigit(q):
		if !sig.Pending() {
			err = errors.New("words in a signature should start with a letter")
		} else {
			sig.WriteRune(q)
		}

	case unicode.IsLetter(q):
		sig.WriteRune(q)

	default:
		err = errors.New("invalid rune")
	}
	return
}
