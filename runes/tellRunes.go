package runes

const (
	ArraySeparator    = ','
	ArrayStop         = '.'
	Dash              = '-'  // values in a sequence are prefixed by a dash ( and whitespace )
	Hash              = '#'  // comment marker
	HTab              = '\t' // tab is considered invalid whitespace
	InterpretedString = '"'  // interpreted strings are bookended with double quotes
	CollectionMark    = '\r' // in comment blocks, represents the dash or key of a sequence
	Newline           = '\n'
	RawString         = '`'
	Record            = '\f' // form feed is used to separate comment entries
	WordConnector     = '_'  // valid in words between colons
	WordSep           = ':'  // keywords in a signature are separated by a colon
	Space             = ' '
)

const Nestline = "\n\t"
