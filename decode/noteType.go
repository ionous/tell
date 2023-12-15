package decode

type noteType int

const (
	NoteNone noteType = iota

	// writes to buffer
	NoteHeader
	NotePrefix
	NotePrefixInline

	// can write straight out
	NoteSuffix
	NoteSuffixInline
	NoteFooter
)

func (n noteType) Prefix() (okay bool) {
	switch n {
	case NotePrefix, NotePrefixInline:
		okay = true
	}
	return
}

func (n noteType) Suffix() (okay bool) {
	switch n {
	case NoteSuffix, NoteSuffixInline:
		okay = true
	}
	return
}
