package encode

import (
	"io"
	"strconv"

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
func (tab *TabWriter) Tab() {
	tab.spaces += 2
}

func (tab *TabWriter) Newline() {
	tab.lines++
	tab.spaces = 0
}

// inc: increases the current indent
// line: increases the current line
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

// call before writing runes or strings
// advances the line and pads the indent
func (tab *TabWriter) pad() {
	if tab.lines > 0 {
		runes.WriteRune(tab.Writer, runes.Newline)
		writeSpaces(tab.Writer, tab.depth*2)
		tab.lines = 0
	}
	//
	tab.writeSpaces()
}

func (tab *TabWriter) writeSpaces() {
	if tab.spaces > 0 {
		writeSpaces(tab.Writer, tab.spaces)
		tab.spaces = 0
	}
}

// quotes and escapes the passed string.
// fix? uses strconv, and strconv produces \x, \u, and \U escapes
// ( plus it requires multiple traversals over the string )
func (tab *TabWriter) Quote(s string) {
	str := strconv.Quote(s)
	tab.WriteString(str)
}

// write a non-quoted escaped string
func (tab *TabWriter) Escape(s string) {
	str := strconv.Quote(s) // fix? strconv doesnt have a writer api
	tab.WriteString(str[1 : len(str)-1])
}

func (tab *TabWriter) WriteString(s string) (int, error) {
	tab.pad()
	return io.WriteString(tab.Writer, s)
}

func (tab *TabWriter) WriteRune(q rune) (ret int, err error) {
	tab.pad()
	return runes.WriteRune(tab.Writer, q)
}

func writeSpaces(w io.Writer, cnt int) {
	b := []byte{runes.Space} // what is the fast way to do this?
	for i := 0; i < cnt; i++ {
		w.Write(b)
	}
}
