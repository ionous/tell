package decode

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
	out     strings.Builder
	pending []rune
}

// for now defined as unicode is letter, but might be useful to be more lenient
var isValidSignaturePrefix = unicode.IsLetter

// first character of the signature must be a letter
// subsequent characters of words can be letters, numbers, spaces, or "connectors" (underscore)
// colons separate word parts
func (sig *Signature) NewRune(r rune) (ret charm.State) {
	switch {
	case isWhitespace(r) && !sig.IsKeyPending():
		break // done

	case r == runes.Newline || r == runes.Eof:
		if sig.IsKeyPending() {
			e := errors.New("keys can't span lines")
			ret = charm.Error(e)
		}

	case isValidSignaturePrefix(r):
		sig.append(r)
		ret = sig

	case r == runes.WordSep: // aka, a colon
		if !sig.IsKeyPending() {
			e := errors.New("words in signatures should be separated by a single colon")
			ret = charm.Error(e)
		} else {
			sig.append(r) // the signature includes the separator
			sig.flushWord()
			ret = sig
		}

	case r == runes.Space || r == runes.WordConnector || unicode.IsDigit(r):
		if len(sig.pending) == 0 && sig.out.Len() == 0 {
			e := errors.New("signatures must start with a letter")
			ret = charm.Error(e)
		} else {
			sig.append(r)
			ret = sig
		}
	}
	return
}

// resets the signature
func (sig *Signature) GetKey() (ret string, err error) {
	if len(sig.pending) > 0 {
		err = errors.New("signature must end with a colon")
	} else {
		ret = sig.out.String()
		sig.out.Reset()
	}
	return
}

func (sig *Signature) IsKeyPending() bool {
	return len(sig.pending) > 0
}

func (sig *Signature) flushWord() {
	sig.out.WriteString(string(sig.pending))
	sig.pending = sig.pending[0:0]
}

func (sig *Signature) append(r rune) {
	sig.pending = append(sig.pending, r)
}
