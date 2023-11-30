package runes

const (
	ArraySeparator    = ','
	ArrayStop         = '.'
	Dash              = '-' // values in a sequence are prefixed by a dash ( and whitespace )
	Eof               = -1
	Hash              = '#'  // comment marker
	HTab              = '\t' // in documents invalid outside of strings or comments, used for nesting in comment blocks.
	InterpretedString = '"'  // interpreted strings are bookended with double quotes
	KeyValue          = '\r' // in comment blocks, replaces both the  key and the value.
	Newline           = '\n'
	NextRecord        = '\f' // form feed is used to separate comment entries
	RawString         = '`'
	Space             = ' '
	WordConnector     = '_' // valid in words between colons
	WordSep           = ':' // keywords in a signature are separated by a colon
)
