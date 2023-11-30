package runes

const (
	ArraySeparator    = ','
	ArrayStop         = '.'
	CollectionMark    = '\r' // in comment blocks, represents the dash or key of a sequence
	Dash              = '-'  // values in a sequence are prefixed by a dash ( and whitespace )
	Eof               = -1
	Hash              = '#'  // comment marker
	HTab              = '\t' // tab is considered invalid whitespace
	InterpretedString = '"'  // interpreted strings are bookended with double quotes
	Newline           = '\n'
	RawString         = '`'
	Record            = '\f' // form feed is used to separate comment entries
	Space             = ' '
	WordConnector     = '_' // valid in words between colons
	WordSep           = ':' // keywords in a signature are separated by a colon
)
