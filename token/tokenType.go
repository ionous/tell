package token

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
