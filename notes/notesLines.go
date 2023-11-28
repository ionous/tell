package notes

import (
	"strings"

	"github.com/ionous/tell/runes"
)

// lines are automatically whitespace trimmed.
type Lines struct {
	buf    strings.Builder
	spaces int // helpers to trim trailing whitespace
}

func (n *Lines) String() string {
	return n.buf.String()
}

// number of runes in the buffer.
func (n *Lines) Len() int {
	return n.buf.Len()
}

// return the buffer, then clear it.
func (n *Lines) Resolve() string {
	str := n.buf.String()
	n.buf.Reset()
	return str
}

// fix? for now, assume that s is trimmed
func (n *Lines) WriteString(s string) (ret int, err error) {
	if len(s) > 0 {
		n.writeSpaces()
		ret, err = n.buf.WriteString(s)
	}
	return
}

// receive a character from the tell document
// newlines indicate separation between hash lines
// they aren't always written into the comment block.
func (n *Lines) WriteRune(q rune) (_ int, _ error) {
	if q == runes.Space {

		n.spaces++ // helper to trim trailing spaces
	} else {
		if q == runes.Newline {
			n.spaces = 0 // drop any trailing spaces
		} else {
			n.writeSpaces()
		}
		n.buf.WriteRune(q)
	}
	return
}

func (n *Lines) writeSpaces() {
	if n.spaces > 0 {
		dupe(&n.buf, n.spaces, runes.Space)
		n.spaces = 0
	}
}

func dupe(w runeWriter, cnt int, q rune) {
	for i := 0; i < cnt; i++ {
		w.WriteRune(q)
	}
}
