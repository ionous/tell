package charmed

import (
	"fmt"
	"strings"

	"github.com/ionous/tell/charm"
	"github.com/ionous/tell/runes"
)

// three opening quotes have been found:
// 1. read the custom closing tag ( if any )
// 2. read here doc lines until the closing tag
func decodeHereAfter(out *strings.Builder, quote rune, escape bool) charm.State {
	var endTag = []rune{quote, quote, quote}
	return charm.Step(decoodeTag(&endTag), charm.Statement("capture", func(q rune) (ret charm.State) {
		// can't call directly, or it wont see the (possibly new) slice from decode tag
		// and anyway, need to ensure the last rune was a newline
		if q != runes.Newline {
			ret = charm.Error(charm.InvalidRune(q))
		} else {
			ret = decodeBody(out, escape, endTag)
		}
		return
	}))
}

func decodeBody(out *strings.Builder, escape bool, endTag []rune) charm.State {
	var lines indentedLines
	return decodeCustomTag(&lines, endTag, func(_ rune, depth int) error {
		return lines.writeLines(out, depth, escape)
	})
}

// track the indentation of each line in the heredoc
// ( to subtract the indentation of the closing marker once that's known )
type indentedLine struct {
	lhs, rhs int
	str      string
}

func (el *indentedLine) getLine(escape bool) (ret string, err error) {
	if !escape || len(el.str) == 0 {
		ret = el.str
	} else {
		var buf strings.Builder
		if e := escapeString(&buf, el.str); e != nil {
			err = e
		} else {
			ret = buf.String()
		}
	}
	return
}

type indentedLines struct {
	lines []indentedLine
	buf   strings.Builder
}

func (ls *indentedLines) WriteRune(r rune) (int, error) {
	return ls.buf.WriteRune(r)
}

func (ls *indentedLines) WriteString(str string) (int, error) {
	return ls.buf.WriteString(str)
}

func (ls *indentedLines) addLine(lhs, rhs int, str string) {
	ls.lines = append(ls.lines, indentedLine{lhs, rhs, str})
}

func (ls *indentedLines) nextLine(lhs, rhs int) {
	ls.addLine(lhs, rhs, ls.buf.String())
	ls.buf.Reset()
}

// a literalLine means every newline ( except the last ) is a newline.
// otherwise, it takes a fully blank line to write a newline
func (ls indentedLines) writeLines(out *strings.Builder, leftEdge int, escape bool) (err error) {
	literalLines := !escape // these could be tied to different states
	var afterNewLine bool   // when writing interpreted lines, we want only a space OR a newline.
	for i, el := range ls.lines {
		if str, e := el.getLine(escape); e != nil {
			err = e
			break
		} else if len(str) == 0 {
			out.WriteRune(runes.Newline)
			afterNewLine = true
		} else if newLhs := el.lhs - leftEdge; newLhs < 0 {
			err = underIndentAt(i)
			break
		} else {
			if i > 0 {
				if literalLines {
					out.WriteRune(runes.Newline)
				} else if !afterNewLine {
					out.WriteRune(runes.Space)
				}
			}
			dupe(out, runes.Space, newLhs)
			out.WriteString(str)
			if literalLines {
				dupe(out, runes.Space, el.rhs)
			} else if cnt := len(str); cnt > 0 {
				// obscure: if the author ends the line with a manual \n
				// then absorb the space on the next line.
				afterNewLine = str[cnt-1] == runes.Newline
			}
		}
	}
	return
}

type underIndentAt int

func (u underIndentAt) Error() string {
	return fmt.Sprintf("heredoc line %d has a smaller indent than its closing tag", u)
}
