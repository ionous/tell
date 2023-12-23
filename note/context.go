package note

import (
	"strings"

	"github.com/ionous/tell/runes"
)

// provides communication between note takers:
// each in progress document decoder should have its own unique context.
// concrete instances shouldn't be copied.
type Context []string

func (ctx *Context) pending() bool {
	return len(*ctx) > 0
}

func (ctx *Context) append(str string) {
	(*ctx) = append(*ctx, str)
}

func (ctx *Context) writeInto(out *strings.Builder) {
	if ctx.pending() {
		for i, el := range *ctx {
			if i > 0 {
				out.WriteRune(runes.Newline)
			}
			out.WriteString(el)
		}
		*ctx = (*ctx)[:0]
	}
}
