package note

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

func (n Type) Prefix() (okay bool) {
	switch n {
	case Prefix, PrefixInline:
		okay = true
	}
	return
}

func (n Type) Suffix() (okay bool) {
	switch n {
	case Suffix, SuffixInline:
		okay = true
	}
	return
}
