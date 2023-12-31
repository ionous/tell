package token

//go:generate stringer -type=Type
type Type int

const (
	Invalid Type = iota // placeholder, not generated by the tokenizer
	Array               // the value is the open or close rune
	Bool
	Comment // a completely empty comment is a blank line
	Key     // an empty key means a sequence; otherwise a mapping
	Number
	String
)
