package note

// differentiates different comment types
// as described in the package README.
type Type int

const (
	None Type = iota

	// writes to buffer
	Header
	Prefix
	PrefixInline

	// can write straight out
	Suffix
	SuffixInline
	Footer
)

// is this one of the two prefix types?
func (n Type) Prefix() (okay bool) {
	switch n {
	case Prefix, PrefixInline:
		okay = true
	}
	return
}

// is this one of the two suffix types?
func (n Type) Suffix() (okay bool) {
	switch n {
	case Suffix, SuffixInline:
		okay = true
	}
	return
}
