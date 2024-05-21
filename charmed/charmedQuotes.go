package charmed

import (
	"strings"

	"github.com/ionous/tell/charm"
	"github.com/ionous/tell/runes"
)

// assuming q is a rune that starts a string scalar or heredoc
// return an appropriate decoder for decoding the rest of the string.
// otherwise, returns false.
func DecodeQuote(q rune, out *strings.Builder) (ret charm.State, okay bool) {
	switch q {
	case runes.QuoteDouble:
		ret, okay = DecodeDouble(out), true
	case runes.QuoteSingle:
		ret, okay = DecodeSingle(out), true
	case runes.QuoteRaw:
		ret, okay = DecodeRaw(out), true
	case runes.QuotePipe:
		ret, okay = DecodePipe(out), true
	}
	return
}

// read until a heredoc ending marker is found.
func DecodePipe(out *strings.Builder) charm.State {
	return charm.Self("pipe whitespace", func(self charm.State, q rune) (ret charm.State) {
		switch q {
		case runes.Space: // ignore spaces
			ret = self
		case runes.Newline: // we expect to see a newline after the pipe
			ret = decodeUntilTriple(out, runes.QuoteRaw, runes.QuoteDouble, runes.QuoteSingle)
		default:
			ret = charm.Error(charm.InvalidRune(q))
		}
		return
	})
}

// read until a (new) double-quote (") marker is found.
func DecodeDouble(out *strings.Builder) charm.State {
	return scanRemainingString(out, runes.QuoteDouble, AllowHere|AllowEscapes|FoldLines)
}

// read until a (new) single-quote (') marker is found.
func DecodeSingle(out *strings.Builder) charm.State {
	return scanRemainingString(out, runes.QuoteSingle, AllowHere|FoldLines)
}

// read until a (new) back-tick (`) marker is found.
func DecodeRaw(out *strings.Builder) charm.State {
	return scanRemainingString(out, runes.QuoteRaw, AllowHere)
}

// these control how inline strings are processed
type QuoteOptions int

const (
	AllowHere    QuoteOptions = 1 << iota
	FoldLines                 // otherwise, keep all line feeds and leading spaces.
	AllowEscapes              // otherwise, backslashes are backslashes.
)

func (opt QuoteOptions) Is(flag QuoteOptions) bool {
	return opt&flag != 0
}

type pendingSpace bool

func (p *pendingSpace) writeSpace(out *strings.Builder) {
	if *p {
		out.WriteRune(runes.Space)
		*p = false
	}
}

// a leading quote has already been processed
// returns unhandled after the closing quote, or error if finished incorrectly.
func scanRemainingString(out *strings.Builder, match rune, opt QuoteOptions) charm.State {
	var padding pendingSpace
	var lineStarted bool // the start of the string isnt the start of a line
	return charm.Self("scanQuote", func(self charm.State, q rune) (ret charm.State) {
		allowHere := opt.Is(AllowHere)
		opt &= ^AllowHere // cant ever be a heredoc after the start
		ret = self        // provisionally, loop.
		switch {
		case q == runes.Eof:
			// invalid because the string wasn't closed
			ret = charm.Error(charm.InvalidRune(runes.Eof))

		case q == runes.Newline:
			if opt.Is(FoldLines) && !lineStarted {
				padding = true
				lineStarted = true
			} else {
				out.WriteRune(runes.Newline)
				padding = false
			}

		case lineStarted && q == runes.Space:
			// quiet leading spaces when folding, otherwise record them.
			if !opt.Is(FoldLines) {
				out.WriteRune(runes.Space)
			}

		case q != match:
			padding.writeSpace(out)
			lineStarted = false
			if q == runes.Escape && opt.Is(AllowEscapes) {
				ret = charm.Step(decodeEscape(out), self)
			} else {
				out.WriteRune(q)
			}

		default:
			padding.writeSpace(out)
			// a matching quote was detected either we're at the end of the string,
			// or we're still at the start... so it might be a heredoc.
			// if at the end: we eat the matching quote, and return unhandled for the next rune.
			// for a heredoc: we eat the matching quote, *and* the next rune, then parse the doc.
			ret = charm.Statement("quoted", func(secondRune rune) (ret charm.State) {
				if allowHere && secondRune == match {
					ret = decodeHereAfter(out, match)
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
