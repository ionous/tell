package charmed

import (
	"strings"

	"github.com/ionous/tell/charm"
	"github.com/ionous/tell/runes"
)

// wraps a string builder to read a quoted string or heredoc.
type QuoteDecoder struct {
	strings.Builder
}

func (d *QuoteDecoder) DecodeQuote(q rune) (ret charm.State, okay bool) {
	if n, ok := d.DecodePipe(q); ok {
		ret, okay = n, true
	} else if n, ok := d.DecodeDouble(q); ok {
		ret, okay = n, true
	} else if n, ok := d.DecodeSingle(q); ok {
		ret, okay = n, true
	} else if n, ok := d.DecodeRaw(q); ok {
		ret, okay = n, true
	}
	return
}

// assumes q is a pipe rune
// read until a heredoc ending marker is found
func (d *QuoteDecoder) DecodePipe(q rune) (ret charm.State, okay bool) {
	if okay = q == runes.QuotePipe; okay {
		ret = charm.Self("pipe whitespace", func(self charm.State, q rune) (ret charm.State) {
			switch q {
			case runes.Space: // ignore spaces
				ret = self
			case runes.Newline: // we expect to see a newline after the pipe
				ret = decodeUntilTriple(&d.Builder, runes.QuoteRaw, runes.QuoteDouble, runes.QuoteSingle)
			default:
				ret = charm.Error(charm.InvalidRune(q))
			}
			return
		})
	}

	return
}

// read until an double-quote (") end marker is found
func (d *QuoteDecoder) DecodeDouble(q rune) (ret charm.State, okay bool) {
	if okay = q == runes.QuoteDouble; okay {
		ret = d.scanRemainingString(q, AllowHere|AllowEscapes)
	}
	return
}

// read until a single-quote (') end marker is found
func (d *QuoteDecoder) DecodeSingle(q rune) (ret charm.State, okay bool) {
	if okay = q == runes.QuoteSingle; okay {
		ret = d.scanRemainingString(q, AllowHere)
	}
	return
}

// read until an back-tick (`) end marker is found
func (d *QuoteDecoder) DecodeRaw(q rune) (ret charm.State, okay bool) {
	if okay = q == runes.QuoteRaw; okay {
		ret = d.scanRemainingString(runes.QuoteRaw, AllowHere|KeepIndent|KeepLines)
	}
	return
}

// these control how inline strings are processed
type QuoteOptions int

const (
	AllowHere    QuoteOptions = 1 << iota
	KeepIndent                // otherwise, eat leading spaces.
	KeepLines                 // otherwise, use semantic line folding.
	AllowEscapes              // otherwise, backslashes are backslashes.
)

func (opt QuoteOptions) Is(flag QuoteOptions) bool {
	return opt&flag != 0
}

// a leading quote has already been processed.
// returns unhandled after the closing quote, or error if finished incorrectly.
func (d *QuoteDecoder) scanRemainingString(match rune, opt QuoteOptions) charm.State {
	startOfLine := false // the start of the string isnt the start of a line
	return charm.Self("scanQuote", func(self charm.State, q rune) (ret charm.State) {
		allowHere := opt.Is(AllowHere)
		opt &= ^AllowHere // cant ever be a heredoc after the start
		ret = self        // provisionally, loop.
		switch {
		case q == runes.Eof:
			// return invalid because the string wasn't closed.
			ret = charm.Error(charm.InvalidRune(q))

		case q == runes.Newline:
			if opt.Is(KeepLines) || startOfLine {
				d.WriteRune(q)
			} else {
				d.WriteRune(runes.Space)
				startOfLine = true
			}

		case startOfLine && q == runes.Space:
			// if not keeping whitespace, eat all leading indentation as per yaml.
			if opt.Is(KeepIndent) {
				d.WriteRune(q)
			}

		case q != match:
			startOfLine = false
			if q == runes.Escape && opt.Is(AllowEscapes) {
				ret = charm.Step(decodeEscape(d), self)
			} else {
				d.WriteRune(q)
			}

		default:
			// a matching quote was detected. either we're at the end of the string,
			// or we're still at the start... so it might be a heredoc.
			// if at the end: we eat the matching quote, and return unhandled for the next rune.
			// for a heredoc: we eat the matching quote, *and* the next rune, then parse the doc.
			ret = charm.Statement("quoted", func(secondRune rune) (ret charm.State) {
				if allowHere && secondRune == match {
					ret = decodeHereAfter(&d.Builder, match)
				}
				return
			})
		}
		return
	})
}

// three opening quotes have been found:
// 1. read the custom closing tag ( if any )
// 2. read here doc lines until the closing tag
func decodeHereAfter(out *strings.Builder, quote rune) charm.State {
	// start with this as the closing tag
	var endTag = []rune{quote, quote, quote}
	// the tag expands to fit whatever override the user specified (if anything)
	tagReader := decoodeTag(&endTag)
	// after determining the customTag
	return charm.Step(tagReader, charm.Statement("capture", func(q rune) (ret charm.State) {
		// can't call directly, or it wont see the (possibly new) slice from decode tag
		// and anyway, need to ensure the last rune was a newline
		if q != runes.Newline {
			ret = charm.Error(charm.InvalidRune(q))
		} else {
			ret = decodeUntilCustom(out, quote, endTag)
		}
		return
	}))
}
