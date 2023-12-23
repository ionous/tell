package note

import "strings"

// provides communication between note takers:
// each in progress document decoder should have its own unique context.
// concrete instances shouldn't be copied.
type Context struct {
	buf strings.Builder
}

func (ctx *Context) writeInto(out *strings.Builder) {
	if str := ctx.buf.String(); len(str) > 0 {
		out.WriteString(str)
		ctx.buf.Reset()
	}
}

func (ctx *Context) append(str string) {
	appendLine(&ctx.buf, str)
}

func (ctx *Context) pending() bool {
	return ctx.buf.Len() > 0
}
