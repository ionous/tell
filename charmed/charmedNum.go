package charmed

import (
	"fmt"
	"math/bits"
	"strconv"
	"strings"

	"github.com/ionous/tell/charm"
	"github.com/ionous/tell/runes"
)

type modeType int

const (
	modePending modeType = iota
	modeInt
	modeHex
	modeFloat
)

// return a state which reads until the end of string, returns error if finished incorrectly
type NumParser struct {
	runes strings.Builder
	mode  modeType
}

func (*NumParser) String() string {
	return "Numbers"
}

// fix: this is currently less of a number parser, and more a number validator.
func (p *NumParser) accept(q rune, s charm.State) charm.State {
	p.runes.WriteRune(q)
	return s
}

// returns int64 or float64
func (p *NumParser) GetNumber() (ret any, err error) {
	switch s := p.runes.String(); p.mode {
	case modeInt:
		ret = fromInt(s)
	case modeHex:
		ret = fromHex(s)
	case modeFloat:
		ret = fromFloat(s)
	default:
		err = fmt.Errorf("unknown number: '%v' is %v", s, p.mode)
	}
	return
}

// helper to turn a string into a value
func (p *NumParser) GetFloat() (ret float64, err error) {
	switch s := p.runes.String(); p.mode {
	case modeInt:
		ret = float64(fromInt(s))
	case modeHex:
		ret = float64(fromHex(s))
	case modeFloat:
		ret = fromFloat(s)
	default:
		err = fmt.Errorf("unknown number: '%v' is %v", s, p.mode)
	}
	return
}

// return a state capable of digit parsing.
// note: this doesn't support leading with just a "."
func (p *NumParser) Decode() charm.State {
	return charm.Statement("numberDecoder", func(r rune) (ret charm.State) {
		switch r {
		case '-', '+':
			ret = p.accept(r, charm.Statement("after lead plus", func(r rune) (ret charm.State) {
				if runes.IsNumber(r) {
					p.mode = modeInt
					ret = p.accept(r, charm.Statement("num plus", p.leadingDigit))
				}
				return
			}))
		case '0':
			// 0 can standalone; but, it might be followed by a hex qualifier.
			p.mode = modeInt
			ret = p.accept(r, charm.Statement("hex check", func(r rune) (ret charm.State) {
				// https://golang.org/ref/spec#hex_literal
				switch {
				case r == 'x' || r == 'X':
					p.mode = modePending
					ret = p.accept(r, charm.Statement("hex parse", func(r rune) (ret charm.State) {
						if runes.IsHex(r) {
							p.mode = modeHex
							ret = p.accept(r, charm.Statement("num hex", p.hexDigits))
						}
						return
					}))
				default:
					// delegate to number and dot checking...
					// in a statecharmed, it would be a super-state, and
					// x (above) would jump to a sibling of that super-state.
					ret = p.leadingDigit(r)
				}
				return
			}))
		default:
			if runes.IsNumber(r) {
				// https://golang.org/ref/spec#float_lit
				p.mode = modeInt
				ret = p.accept(r, charm.Statement("num digits", p.leadingDigit))
			}
		}
		return
	})
}

// a string of numbers, possibly followed by a decimal or exponent separator.
// note: golang numbers can end in a pure ".", this does not allow that.
func (p *NumParser) leadingDigit(r rune) (ret charm.State) {
	switch {
	case runes.IsNumber(r):
		ret = p.accept(r, charm.Statement("leading dig", p.leadingDigit))
	case r == '.':
		p.mode = modePending
		ret = p.accept(r, charm.Statement("decimal", func(r rune) (ret charm.State) {
			if runes.IsNumber(r) {
				p.mode = modeFloat
				ret = p.accept(r, charm.Statement("decimal digits", p.leadingDigit))
			} else {
				ret = p.tryExponent(r) // delegate to exponent checking,,,
			}
			return
		}))
	default:
		ret = p.tryExponent(r) // delegate to exponent checking,,,
	}
	return
}

// https://golang.org/ref/spec#exponent
// exponent  = ( "e" | "E" ) [ "+" | "-" ] decimals
func (p *NumParser) tryExponent(r rune) (ret charm.State) {
	switch {
	case r == 'e' || r == 'E':
		p.mode = modePending
		ret = p.accept(r, charm.Statement("exp", func(r rune) (ret charm.State) {
			switch {
			case runes.IsNumber(r):
				p.mode = modeFloat
				ret = p.accept(r, charm.Statement("exp decimal", p.decimals))
			case r == '+' || r == '-':
				ret = p.accept(r, charm.Statement("exp power", func(r rune) (ret charm.State) {
					if runes.IsNumber(r) {
						p.mode = modeFloat
						ret = p.accept(r, charm.Statement("exp num", p.decimals))
					}
					return
				}))
			}
			return
		}))
	}
	return
}

// a chain of decimal digits 0-9
func (p *NumParser) decimals(r rune) (ret charm.State) {
	if runes.IsNumber(r) {
		ret = p.accept(r, charm.Statement("decimals", p.decimals))
	}
	return
}

// a chain of hex digits 0-9, a-f
func (p *NumParser) hexDigits(r rune) (ret charm.State) {
	if runes.IsHex(r) {
		ret = p.accept(r, charm.Statement("hexDigits", p.hexDigits))
	}
	return
}

func fromInt(s string) (ret int) {
	s, negate := unary(s)
	if i, e := strconv.ParseInt(s, 10, bits.UintSize); e != nil {
		panic(e)
	} else if negate {
		ret = -int(i)
	} else {
		ret = int(i)
	}
	return
}

func fromHex(s string) (ret uint) {
	// hex string - chops out the 0x qualifier
	if i, e := strconv.ParseUint(s[2:], 16, bits.UintSize); e != nil {
		panic(e)
	} else {
		ret = uint(i) // no negative for hex.
	}
	return
}

func fromFloat(s string) (ret float64) {
	s, negate := unary(s)
	if f, e := strconv.ParseFloat(s, 64); e != nil {
		panic(e)
	} else if negate {
		ret = -f
	} else {
		ret = f
	}
	return
}

// in golang, leading +/- are unary operators;
// here, they are considered optional parts decimal numbers.
// note: strconv's base 10 parser doesnt handle leading signs.
// we therefore leave them out of our result, and just flag the negative ones.
func unary(s string) (ret string, negate bool) {
	switch s[0] {
	case '-':
		ret = s[1:]
		negate = true
	case '+':
		ret = s[1:]
	default:
		ret = s
	}
	return
}
