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
//
func decodeHere(out *strings.Builder, quote rune, escape bool) charm.State {
	var endTag = []rune{quote, quote, quote}
	return charm.Step(decoodeTag(&endTag), decodeBody(out, escape, endTag))
}

func decodeBody(out *strings.Builder, escape bool, endTag []rune) charm.State {
	var lines indentedLines
	var lineBuf strings.Builder
	return decodeLines(&lineBuf, escape, endTag, func(lineType lineType, lhs, rhs int) (err error) {
		switch lineType {
		case lineText:
			lines.addLine(lhs, rhs, lineBuf.String())
			lineBuf.Reset()
		case lineClose:
			err = lines.writeLines(out, lhs, !escape)
		default:
			panic("unknown lineType")
		}
		return
	})
}

// track the indentation of each line in the heredoc
// ( to subtract the indentation of the closing marker once that's known )
type indentedLine struct {
	lhs, rhs int
	str      string
}

type indentedLines []indentedLine

func (ls *indentedLines) addLine(lhs, rhs int, str string) {
	*ls = append(*ls, indentedLine{lhs, rhs, str})
}

// a literalLine means every newline ( except the last ) is a newline.
// otherwise, it takes a fully blank line to write a newline
func (ls indentedLines) writeLines(out *strings.Builder, leftEdge int, literalLines bool) (err error) {
	for i, el := range ls {
		if str := el.str; len(str) == 0 {
			out.WriteRune(runes.Newline)
		} else if newLhs := el.lhs - leftEdge; newLhs < 0 {
			err = underIndentAt(i)
			break
		} else {
			if i > 0 {
				if literalLines {
					out.WriteRune(runes.Newline)
				} else {
					out.WriteRune(runes.Space)
				}
			}
			dupe(out, runes.Space, newLhs)
			out.WriteString(str)
			if literalLines {
				dupe(out, runes.Space, el.rhs)
			}
		}
	}
	return
}

type underIndentAt int

func (u underIndentAt) Error() string {
	return fmt.Sprintf("heredoc line %d has a smaller indent than its closing tag", u)
}
