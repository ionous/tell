package encode

import (
	"io"
	"strconv"

	"github.com/ionous/tell/runes"
)

type TabWriter struct {
	depth    int // requested leading spaces on each new line
	spaces   int // trailing spaces
	newLines int // requested newlines, not written until pad()
	Writer   io.Writer
	xpos     int
}

// a soft space -- eaten if theres a newline
func (tab *TabWriter) Space() {
	tab.spaces++
}
func (tab *TabWriter) Tab() {
	tab.spaces += 2
}

func (tab *TabWriter) Softline() {
	if tab.newLines == 0 {
		tab.newLines++
	}
	tab.spaces = 0
}

func (tab *TabWriter) Nextline() {
	tab.newLines++
	tab.spaces = 0
}

// inc: increases the current indent
// line: increases the current line
func (tab *TabWriter) Indent(inc bool) {
	if inc {
		tab.depth += 2
	} else {
		tab.depth -= 2
	}
}

// call before writing runes or strings
// advances the line and pads the indent
func (tab *TabWriter) pad() {
	if n := tab.newLines; n > 0 {
		tab.newLines = 0
		for i := 0; i < n; i++ {
			runes.WriteRune(tab.Writer, runes.Newline)
		}
		writeSpaces(tab.Writer, tab.depth)
		tab.xpos = tab.depth
	}
	//
	tab.writeSpaces()
}

func (tab *TabWriter) writeSpaces() {
	if tab.spaces > 0 {
		writeSpaces(tab.Writer, tab.spaces)
		tab.xpos += tab.spaces
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

// escape the contents of a string
// returns true if there were any escapes
func (tab *TabWriter) Escape(s string) bool {
	str := strconv.Quote(s) // fix? strconv doesnt have a writer api
	cnt, _ := tab.WriteString(str[1 : len(str)-1])
	return cnt > len(s)
}

func (tab *TabWriter) WriteString(s string) (int, error) {
	tab.pad()
	tab.xpos += len(s) // approximate
	return io.WriteString(tab.Writer, s)
}

func (tab *TabWriter) WriteRune(q rune) (ret int, err error) {
	tab.pad()
	tab.xpos += 1 // approximate
	return runes.WriteRune(tab.Writer, q)
}

func (tab *TabWriter) writeLine(str string) {
	tab.WriteString(str)
	tab.Softline()
}

func (tab *TabWriter) writeLines(lines []string) {
	for _, line := range lines {
		tab.writeLine(line)
	}
}

func writeSpaces(w io.Writer, cnt int) {
	b := []byte{runes.Space} // what is the fast way to do this?
	for i := 0; i < cnt; i++ {
		w.Write(b)
	}
}
