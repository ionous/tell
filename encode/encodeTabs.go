package encode

import (
	"io"

	"github.com/ionous/tell/runes"
)

type TabWriter struct {
	depth  int
	spaces int
	lines  int
	io.Writer
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

func (tab *TabWriter) WriteRune(q rune) (ret int, err error) {
	return runes.WriteRune(tab.Writer, q)
}

func writeSpaces(w io.Writer, cnt int) {
	b := []byte{runes.Space} // what is the fast way to do this?
	for i := 0; i < cnt; i++ {
		w.Write(b)
	}
}
