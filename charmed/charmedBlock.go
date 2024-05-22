package charmed

import (
	"fmt"
	"log"
	"strings"

	"github.com/ionous/tell/runes"
)

func KeepEnding(q rune) (okay bool) {
	if q == runes.QuoteRaw || q == runes.QuoteDouble {
		okay = true
	} else if q != runes.QuoteSingle {
		log.Panicf("unknown rune %q", q)
	}
	return
}

func EscapeHere(q rune) (okay bool) {
	if q == runes.QuoteDouble {
		okay = true
	} else if q != runes.QuoteSingle &&
		q != runes.QuoteRaw {
		log.Panicf("unknown rune %q", q)
	}
	return
}

// track the indentation of each line in the heredoc
// ( to subtract the indentation of the closing marker once that's known )
type indentedLine struct {
	indented int
	line     string
}

func (el *indentedLine) getLine(escape bool) (ret string, err error) {
	if !escape || len(el.line) == 0 {
		ret = el.line
	} else {
		var buf strings.Builder
		if e := escapeString(&buf, el.line); e != nil {
			err = e
		} else {
			ret = buf.String()
		}
	}
	return
}

// accumulates text for line
type indentedBlock struct {
	buf   strings.Builder
	lines []indentedLine
}

// record a rune on the current line
// ( left hand spaces are expected to be accumulated externally and add via flushLine )
func (ls *indentedBlock) WriteRune(r rune) (int, error) {
	return ls.buf.WriteRune(r)
}

// record a string on the current line
func (ls *indentedBlock) WriteString(str string) (int, error) {
	return ls.buf.WriteString(str)
}

func (ls *indentedBlock) addLine(lhs int, str string) {
	ls.lines = append(ls.lines, indentedLine{lhs, str})
}

func (ls *indentedBlock) flushLine(lhs, rhs int) {
	// if the line only had spaces
	if ls.buf.Len() == 0 {
		ls.addLine(-1, "")
	} else {
		dupe(&ls.buf, runes.Space, rhs) // as per yaml, trailing spaces are preserved
		ls.buf.WriteRune(runes.Newline)
		ls.addLine(lhs, ls.buf.String())
		ls.buf.Reset()
	}
}

// the leftEdge counts the number of leading spaces to eat
// if a line has less than that, its an underflow.
func (ls indentedBlock) writeBlock(out *strings.Builder, lineType rune, leftEdge int) (err error) {
	escape, keepEnding := EscapeHere(lineType), KeepEnding(lineType)
	for i, end := 0, len(ls.lines)-1; i <= end; i++ {
		el, atEnd := ls.lines[i], i == end
		// a blank line:
		if el.indented < 0 {
			if !atEnd || keepEnding {
				out.WriteRune(runes.Newline)
			}
		} else {
			// a content line:
			if newLhs := el.indented - leftEdge; newLhs < 0 {
				err = underIndentAt(i)
				break
			} else if str, e := el.getLine(escape); e != nil {
				err = e // ^ interprets the string, or returns it raw depending
				break
			} else {
				dupe(out, runes.Space, newLhs) // add leading spaces
				if atEnd && !keepEnding {
					str = str[:len(str)-1]
				}
				out.WriteString(str)
			}
		}
	}
	return
}

type underIndentAt int

func (u underIndentAt) Error() string {
	return fmt.Sprintf("heredoc line %d has a smaller indent than its closing tag", u)
}
