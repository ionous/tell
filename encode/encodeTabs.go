package encode

import (
	"fmt"
	"io"
	"unicode/utf8"

	"github.com/ionous/tell/runes"
)

type TabWriter struct {
	depth  int
	spaces int
	lines  int
	io.Writer
}

type RuneWriter interface {
	WriteRune(r rune) (int, error)
}

// a soft space -- eaten if theres a newline
func (tab *TabWriter) Space() {
	tab.spaces++
}

func (tab *TabWriter) Newline() {
	tab.lines++
	tab.spaces = 0
}

// todo: track the parent type [ a stack ]
// to determine whether to write inline "- - -"
// maybe also a line width
func (tab *TabWriter) Indent(inc bool, line bool) {
	if inc {
		tab.depth++
	} else {
		tab.depth--
	}
	if line {
		tab.lines++
		tab.spaces = 0
	}
}

func (tab *TabWriter) Flush() *TabWriter {
	if tab.lines > 0 {
		tab.WriteRune(runes.Newline)
		writeSpaces(tab.Writer, tab.depth*2)
		tab.lines = 0
	}
	//
	tab.writeSpaces()
	return tab
}

func (tab *TabWriter) writeSpaces() {
	if tab.spaces > 0 {
		writeSpaces(tab.Writer, tab.spaces)
		tab.spaces = 0
	}
}

func (tab *TabWriter) WriteString(s string) (int, error) {
	tab.writeSpaces()
	return io.WriteString(tab.Writer, s)
}

func (tab *TabWriter) WriteRune(r rune) (ret int, err error) {
	if rw, ok := tab.Writer.(RuneWriter); ok {
		ret, err = rw.WriteRune(r)
	} else if !utf8.ValidRune(r) {
		err = fmt.Errorf("rune %d out of range", r)
	} else {
		var scratch [utf8.UTFMax]byte
		cnt := utf8.EncodeRune(scratch[:], r)
		ret, err = tab.Write(scratch[:cnt])
	}
	return
}

func writeSpaces(w io.Writer, cnt int) {
	b := []byte{runes.Space} // what is the fast way to do this?
	for i := 0; i < cnt; i++ {
		w.Write(b)
	}
}
