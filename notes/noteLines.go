package notes

import (
	"strings"

	"github.com/ionous/tell/runes"
)

// buffers multiple comment lines.
// by default, the lines are nested;
// lines are automatically whitespace trimmed.
type Lines struct {
	buf      strings.Builder
	spaces   int  // helpers to trim trailing whitespace
	writing  bool // tracks whether a line is in progress
	special  bool // track if the last character in the buffer is an escape
	skipNest bool // suppress nesting when writing new comment lines
}

// number of runes in the buffer.
func (n *Lines) Len() int {
	return n.buf.Len()
}

// return line(s) that have been written and reset the buffer.
// automatically ends any line in progress ( by writing a newline )
func (n *Lines) GetComments() (ret string) {
	ret = n.buf.String()
	n.buf.Reset()
	n.spaces = 0
	n.writing = false
	n.special = false
	n.skipNest = false
	return
}

// write a literal newline into the comment block
func (n *Lines) Break() {
	n.buf.WriteRune(runes.Newline)
	n.special = true
}

// receive a character from the tell document
// newlines indicate separation between hash lines
// they aren't always written into the comment block.
func (n *Lines) WriteRune(r rune) (_ int, _ error) {
	// writing after not having written?
	// separate the new content from previous content
	// ( unless its already separated. ex. \r \f )
	if !n.writing && n.buf.Len() > 0 && !n.special {
		n.Break()
		if !n.skipNest { // nest after newline
			n.buf.WriteRune(runes.HTab)
		}
	}
	n.writing = true
	n.skipNest = false
	switch r {
	case runes.Newline:
		n.writing = false
		n.spaces = 0 // drop any trailing spaces
	case runes.Space:
		n.spaces++ // helper to trim trailing spaces
	default:
		if n.spaces > 0 {
			dupe(&n.buf, n.spaces, runes.Space)
			n.spaces = 0
		}
		n.buf.WriteRune(r)
		n.special = isSpecial(r)
	}
	return
}

const (
	nestIndent = string(runes.Newline) + string(runes.HTab)
)

func isSpecial(r rune) bool {
	return r < runes.Space
}
