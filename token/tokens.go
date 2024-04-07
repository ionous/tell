package token

import (
	"errors"
	"strings"
	"unicode"

	"github.com/ionous/tell/charm"
	"github.com/ionous/tell/charmed"
	"github.com/ionous/tell/runes"
)

// tbd: maybe a channel instead?
type Notifier interface {
	Decoded(Pos, Type, any) error
}

type Tokenizer struct {
	Notifier Notifier
	// configure the upcoming Decode to produce only floating point numbers.
	// otherwise it will produce int for integers, and unit for hex specifications.
	UseFloats bool // controls number decoding
}

// return a state to parse a stream of runes and notify as they are detected.
func (cfg Tokenizer) Decode() charm.State {
	n := tokenizer{Tokenizer: cfg}
	return charm.Parallel("tokenizer", n.decode(false), charmed.DecodePos(&n.curr.Y, &n.curr.X))
}

func NewTokenizer(n Notifier) charm.State {
	cfg := Tokenizer{Notifier: n}
	return cfg.Decode()
}

type tokenizer struct {
	Tokenizer
	curr, start Pos
}

func (n *tokenizer) decode(afterIndent bool) charm.State {
	return charm.Step(n.whitespace(afterIndent), n.tokenize())
}

func (n *tokenizer) notifyRune(q rune, t Type, v any) (ret charm.State) {
	if e := n.Notifier.Decoded(n.start, t, v); e != nil {
		ret = charm.Error(e)
	} else {
		ret = send(n.decode(true), q)
	}
	return
}

// eat whitespace between tokens;
// previously, would error if it didnt detect whitespace between tokens
// however that doesnt work well for arrays. ex: `5,`
func (n *tokenizer) whitespace(afterIndent bool) charm.State {
	var spaces int
	return charm.Self("whitespace", func(self charm.State, q rune) (ret charm.State) {
		switch q {
		case runes.Space:
			spaces++
			ret = self
		case runes.Newline:
			spaces = 0
			afterIndent = false
			ret = self
		case runes.Eof:
			ret = charm.Error(nil)
		}
		return
	})
}

func (n *tokenizer) tokenize() charm.State {
	return charm.Statement("tokenize", func(q rune) (ret charm.State) {
		n.start = n.curr
		switch q {
		case runes.HTab:
			ret = charm.Error(errors.New("tabs are invalid whitespace"))
		case runes.Hash:
			next := n.commentDecoder()
			ret = send(next, q)

		case runes.InterpretQuote:
			ret = n.interpretDecoding()

		case runes.RawQuote:
			ret = n.rawDecoding()

		case runes.Dash: // negative numbers or sequences
			ret = n.dashDecoding()

		case runes.ArrayOpen, runes.ArrayClose, runes.ArraySeparator:
			if e := n.Notifier.Decoded(n.start, Array, q); e != nil {
				ret = charm.Error(e)
			} else {
				ret = n.decode(true)
			}

		default:
			switch {
			case runes.IsNumber(q) || q == '+': // a leading negative gets handled by dashDecoding.
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
				// sign succeeds and turns nil on whitespace after a colon;
				// boolean on the rune after its last letter.
				if sign = sign.NewRune(q); sign == nil {
					ret = n.notifyRune(q, Key, sig.String())
				} else if boolean = boolean.NewRune(q); boolean == nil {
					// boolean shouldnt match: ex. "falsey"
					if !runes.IsWhitespace(q) {
						boolean = charm.Error(nil)
					} else {
						// note: this means a key "true true:" will be interpreted as
						// a bool (true) followed by a key (true:)
						ret = n.notifyRune(q, Bool, b == boolTrue)
					}
				} else if terminal(sign) {
					// sign is mostly superset of bool; (except for the eof/eol cases)
					// if it dies and boolean didnt just succeed; they're both dead.
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
			ret = n.notifyRune(q, Comment, b.String())
		}
		return
	})
}

func (n *tokenizer) interpretDecoding() charm.State {
	var d charmed.QuoteDecoder
	return charm.Step(d.Interpret(), charm.Statement("interpreted", func(q rune) charm.State {
		return n.notifyRune(q, String, d.String())
	}))
}

func (n *tokenizer) rawDecoding() charm.State {
	var d charmed.QuoteDecoder
	return charm.Step(d.Record(), charm.Statement("recorded", func(q rune) charm.State {
		return n.notifyRune(q, String, d.String())
	}))
}

// fix? returns float64 because json does
// could also return int64 when its int like
func (n *tokenizer) numDecoder() charm.State {
	var d charmed.NumParser
	return charm.Step(d.Decode(), charm.Statement("numDecoder", func(q rune) (ret charm.State) {
		if n.UseFloats {
			if v, e := d.GetFloat(); e != nil {
				ret = charm.Error(e)
			} else {
				ret = n.notifyRune(q, Number, v)
			}
		} else {
			if v, e := d.GetNumber(); e != nil {
				ret = charm.Error(e)
			} else {
				ret = n.notifyRune(q, Number, v)
			}
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
