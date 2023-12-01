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

type RuneWriter interface {
	WriteRune(r rune) (int, error)
}

// a soft space -- eaten if theres a newline
func (tab *TabWriter) Space() {
	tab.spaces++
	tab.WriteRune(runes.Space)
}

func (tab *TabWriter) Newline() {
	tab.lines++
	tab.spaces = 0
}

// todo: track the parent type [ a stack ]
// to determine whether to write inline "- - -"
// maybe also a line width
func (tab *TabWriter) Indent(inc bool) {
	if inc {
		tab.depth++
	} else {
		tab.depth--
	}
	tab.lines++
	tab.spaces = 0
}

func (tab *TabWriter) Flush() *TabWriter {
	if tab.lines > 0 {
		tab.WriteRune(runes.Newline)
		tab.writeSpaces(tab.depth * 2)
		tab.lines = 0
	}
	//
	if tab.spaces > 0 {
		tab.writeSpaces(tab.spaces)
		tab.spaces = 0
	}
	return tab
}

func (tab *TabWriter) writeSpaces(cnt int) {
	b := []byte{runes.Space} // what is the fast way to do this?
	for i := 0; i < cnt; i++ {
		tab.Write(b)
	}
}

func (tab *TabWriter) WriteString(s string) (int, error) {
	return io.WriteString(tab.Writer, s)
}

func (tab *TabWriter) WriteRune(r rune) (ret int, err error) {
	if rw, ok := tab.Writer.(RuneWriter); ok {
		ret, err = rw.WriteRune(r)
	} else {
		// var scratch byte[4]
		// utf8.EncodeRune(scratch, r)
		panic("measure rune, write bytes, advance")
	}
	return
}
