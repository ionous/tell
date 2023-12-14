package decode

type noteType int

const (
	_Invalid noteType = iota
	NoteHeader
	NoteFooter

	NotePrefix
	NotePrefixInline

	NoteSuffix
	NoteSuffixInline

	NoteInterKey
)

// requires Key/Value markers
func (n noteType) Padding() (ret int) {
	switch n {
	case NotePrefix, NotePrefixInline:
		ret = 1
	case NoteSuffix, NoteSuffixInline:
		ret = 2
	}
	return
}

// requires a newline separator
func (n noteType) Newline() (okay bool) {
	switch n {
	case NotePrefix, NoteSuffix:
		okay = true
	}
	return
}
