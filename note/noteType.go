package note

// differentiates different comment types
// as described in the package README.
type Type int

//go:generate stringer -type=Type
const (
	None Type = iota
	Header
	PrefixInline
	Prefix
	SuffixInline
	Suffix
	Footer
)

// is this one of the two inline types?
func (n Type) inline() (okay bool) {
	switch n {
	case PrefixInline, SuffixInline:
		okay = true
	}
	return
}

// return the type without the inline status
func (n Type) withoutInline() Type {
	switch n {
	case PrefixInline:
		n = Prefix
	case SuffixInline:
		n = Suffix
	}
	return n
}

// number of key-value markers preceding this element within the current term
func (n Type) mark() (ret int) {
	switch n {
	case Prefix, PrefixInline:
		ret = 1
	case Suffix, SuffixInline:
		ret = 2
	}
	return
}
