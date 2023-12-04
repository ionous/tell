package token

import (
	"errors"
	"strings"
	"unicode"

	"github.com/ionous/tell/charm"
	"github.com/ionous/tell/charmed"
	"github.com/ionous/tell/runes"
)

//go:generate stringer -type=Type
type Type int

const (
	Invalid Type = iota // placeholder, not generated by the tokenizer
	Bool
	Number
	InterpretedString
	RawString
	Comment // a completely empty comment is a blank line
	Key     // an empty key means a sequence; otherwise a mapping

	// heredoc?
	// array?
)

func (t Type) Scalar() (ret bool) {
	switch t {
	case Bool, Number, InterpretedString, RawString:
		ret = true
	}
	return
}

type Notifier interface {
	Decoded(Type, any) error
}

func MakeTokenizer(notify Notifier) charm.State {
	t := tokenizer{notifier: notify}
	return t.decode()
}

type tokenizer struct {
	notifier    Notifier
	indent      int  // left aligned whitespace
	spaces      int  // token separated whitespace
	afterIndent bool // specifically, are we *after* the indent
}

func (n *tokenizer) decode() charm.State {
	return charm.Step(n.whitespace(), n.tokenize())
}

func (n *tokenizer) whitespace() charm.State {
	return charm.Self("ledespace", func(self charm.State, q rune) (ret charm.State) {
		switch q {
		case runes.Space:
			n.spaces++
			if !n.afterIndent {
				n.indent++
			}
			ret = self
		case runes.Newline:
			if !n.afterIndent {
				n.notifier.Decoded(Comment, "")
			}
			n.spaces = 0
			n.indent = 0
			n.afterIndent = false
			ret = self
		case runes.Eof:
			ret = charm.Error(nil)
		}
		return
	})
}

func (n *tokenizer) tokenize() charm.State {
	return charm.Statement("tokenize", func(q rune) (ret charm.State) {
		if n.afterIndent && n.spaces == 0 {
			e := errors.New("expected whitespace between tokens")
			ret = charm.Error(e)
		} else {
			n.afterIndent = true
			n.spaces = 0
			//
			switch {
			case q == runes.Hash:
				next := n.commentDecoder()
				ret = send(next, q)

			case q == runes.InterpretedString:
				ret = n.interpretDecoding()

			case q == runes.RawString:
				ret = n.rawDecoding()

			case q == runes.Dash: // negative numbers or sequences
				ret = n.dashDecoding()

			case runes.IsNumber(q) || q == '+':
				next := n.numDecoder()
				ret = send(next, q)

			case unicode.IsLetter(q): // maps and bools
				next := n.wordDecoder()
				ret = send(next, q)
			}
		}
		return
	})
}

func (n *tokenizer) notifyRune(q rune, t Type, v any) (ret charm.State) {
	if e := n.notifier.Decoded(t, v); e != nil {
		ret = charm.Error(e)
	} else {
		ret = send(n.decode(), q)
	}
	return
}

// if the passed rune might be start a bool value
// for example, `trouble:` would match `true` temporarily
// and `false:` would match `false` until the colon.
func (n *tokenizer) wordDecoder() charm.State {
	return charm.Statement("wordDecoder", func(q rune) (ret charm.State) {
		var b boolValue
		if q == 't' {
			b = boolTrue
		} else if q == 'f' {
			b = boolFalse
		}
		if b == boolInvalid {
			ret = n.decodeSignature()
		} else {
			var sig Signature
			sign := sig.Decoder()
			boolean := charmed.StringMatch(b.String())
			ret = charm.Self("parallel", func(self charm.State, q rune) (ret charm.State) {
				ret = self
				// the boolean and sign states return nil on success
				if boolean = boolean.NewRune(q); boolean == nil {
					ret = n.notifyRune(q, Bool, b == boolTrue)
				} else if sign = sign.NewRune(q); sign == nil {
					// signature ends on whitespace ( so pass that on )
					ret = n.notifyRune(q, Key, sig.String())
				} else if terminal(boolean) && terminal(sign) {
					// if they both have error'd; we're done.
					// ( if only one has error'd, it will keep returning the same error. )
					ret = charm.Error(wordyError)
				}
				return
			})
		}
		return send(ret, q)
	})
}

// negative numbers or sequences
func (n *tokenizer) decodeSignature() charm.State {
	var sig Signature
	sign := sig.Decoder() // use self, instead of step to customize the error response
	return charm.Self("signature", func(self charm.State, q rune) (ret charm.State) {
		ret = self // provisionally
		if sign = sign.NewRune(q); sign == nil {
			ret = n.notifyRune(q, Key, sig.String())
		} else if terminal(sign) {
			ret = charm.Error(wordyError)
		}
		return
	})
}

// negative numbers or sequences
func (n *tokenizer) dashDecoding() charm.State {
	return charm.Statement("dashing", func(q rune) (ret charm.State) {
		if runes.IsWhitespace(q) { // a space indicates a sequence `- 5`
			ret = n.notifyRune(q, Key, "") // a blank key means a dash
		} else { // no space indicates a number ( or an error ) `-5`
			next := n.numDecoder()
			ret = send(next, runes.Dash, q)
		}
		return
	})
}

func (n *tokenizer) commentDecoder() charm.State {
	var b strings.Builder
	return charm.Self("comments", func(self charm.State, q rune) (ret charm.State) {
		switch q {
		default:
			b.WriteRune(q)
			ret = self
		case runes.Newline, runes.Eof:
			// tbd: using .indenting could send "trailing" vs. "full line comment"
			ret = n.notifyRune(q, Comment, b.String())
		}
		return
	})
}

func (n *tokenizer) interpretDecoding() charm.State {
	var d charmed.QuoteDecoder
	return charm.Step(d.Interpret(), charm.Statement("interpreted", func(q rune) charm.State {
		return n.notifyRune(q, InterpretedString, d.String())
	}))
}

func (n *tokenizer) rawDecoding() charm.State {
	var d charmed.QuoteDecoder
	return charm.Step(d.Record(), charm.Statement("recorded", func(q rune) charm.State {
		return n.notifyRune(q, RawString, d.String())
	}))
}

// fix? returns float64 because json does
// could also return int64 when its int like
func (n *tokenizer) numDecoder() charm.State {
	var d charmed.NumberDecoder
	return charm.Step(d.Decode(), charm.Statement("numDecoder", func(q rune) (ret charm.State) {
		if v, e := d.GetNumber(); e != nil {
			ret = charm.Error(e)
		} else {
			ret = n.notifyRune(q, Number, v)
		}
		return
	}))
}

// a tri-boolean: 0 is invalid, not false.
type boolValue int

//go:generate stringer -type=boolValue -linecomment
const (
	boolInvalid boolValue = iota
	boolFalse             // false
	boolTrue              // true
)

var wordyError = errors.New("couldn't read words. strings should be quoted, booleans should be 'true' or 'false', and map keys should start with a letter and end with a colon.")

func terminal(next charm.State) (okay bool) {
	_, okay = next.(charm.Terminal)
	return
}

func send(next charm.State, qs ...rune) charm.State {
	for _, q := range qs {
		if next = next.NewRune(q); next == nil {
			break
		}
	}
	return next
}
