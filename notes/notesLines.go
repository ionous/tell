package notes

import (
	"github.com/ionous/tell/runes"
)

// lines are automatically whitespace trimmed.
type Lines struct {
	out           RuneWriter
	spaces, total int // helpers to trim trailing whitespace
}

// approximate count of runes in the buffer.
func (n *Lines) Len() int {
	return n.total
}

// for now assume that s is trimmed
func (n *Lines) WriteString(str string) (ret int, err error) {
	ret, err = writeString(n.out, str)
	n.total = ret + n.writeSpaces()
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
		n.out.WriteRune(q)
		n.total++
	}
	return
}

func (n *Lines) writeSpaces() (ret int) {
	if ret = n.spaces; ret > 0 {
		dupe(n.out, n.spaces, runes.Space)
		n.spaces = 0
	}
	return
}

func writeString(w RuneWriter, str string) (ret int, _ error) {
	if out, ok := w.(stringWriter); ok {
		ret, _ = out.WriteString(str)
	} else {
		for _, q := range str {
			n, _ := w.WriteRune(q)
			ret += n
		}
	}
	return
}

func dupe(w RuneWriter, cnt int, q rune) {
	for i := 0; i < cnt; i++ {
		w.WriteRune(q)
	}
}
