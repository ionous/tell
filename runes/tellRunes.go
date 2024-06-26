package runes

const (
	ArrayClose     = ']'
	ArrayOpen      = '['
	ArraySeparator = ','
	Colon          = ':' // keywords in a signature are separated by a colon
	Dash           = '-' // values in a sequence are prefixed by a dash ( and whitespace )
	Eof            = -1
	Escape         = '\\'
	Hash           = '#'  // comment marker
	HTab           = '\t' // invalid outside of strings or comments.
	KeyValue       = '\r' // in comment blocks, replaces both the  key and the value.
	Newline        = '\n'
	NextTerm       = '\f' // form feed is used to separate comment entries
	QuoteDouble    = '"'  // double quotes
	QuotePipe      = '|'  // a nod towards yaml compatibility
	QuoteRaw       = '`'  // back tick
	QuoteSingle    = '\'' // single quotes
	Redirect       = '<'  // for closing tags
	Space          = ' '
	Underscore     = '_' // valid in words between colons
)

// https://golang.org/ref/spec#decimal_digit
func IsNumber(r rune) bool {
	return r >= '0' && r <= '9'
}

// https://golang.org/ref/spec#hex_digit
func IsHex(r rune) bool {
	return (r >= 'a' && r <= 'f') || (r >= 'A' && r <= 'F') || IsNumber(r)
}

func IsWhitespace(q rune) (ret bool) {
	switch q {
	case Space, Newline, Eof:
		ret = true
	}
	return
}
