package notes

import (
	"strings"

	"github.com/ionous/tell/runes"
)

// buffers multiple comment lines.
// by default, the lines are nested;
// lines are automatically whitespace trimmed.
type Lines struct {
	buf     strings.Builder
	writing bool // tracks whether a line is in progress
	newline bool // force a newline instead of nesting ;/
	spaces  int  // helpers to trim trailing whitespace
	lines   int  // number of newlines in the buffer.
}

// number of newlines in the buffer.
func (n *Lines) NumLines() int {
	return n.lines
}

// return line(s) that have been written and reset the buffer.
// automatically ends any line in progress ( by writing a newline )
func (n *Lines) GetComments() (ret string) {
	ret = n.buf.String()
	n.buf.Reset()
	n.writing = false
	n.newline = false
	n.spaces = 0
	n.lines = 0
	return
}

func (n *Lines) WriteRune(r rune) (_ int, _ error) {
	if !n.writing {
		if n.newline {
			n.buf.WriteRune(runes.Newline)
			n.newline = false
		} else if n.lines > 0 {
			n.buf.WriteString(nestIndent)
		}
		n.writing = true
	}
	switch r {
	case runes.Newline:
		n.writing = false
		n.spaces = 0 // drop any trailing spaces
		n.lines++
	case runes.Space:
		n.spaces++ // helper to trim trailing spaces
	default:
		if n.spaces > 0 {
			dupe(&n.buf, n.spaces, runes.Space)
			n.spaces = 0
		}
		n.buf.WriteRune(r)
	}
	return
}

const (
	nestIndent = string(runes.Newline) + string(runes.HTab)
)
