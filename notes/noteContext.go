package notes

import (
	"io"
	"slices"
	"strings"

	"github.com/ionous/tell/runes"
)

type context struct {
	out   *pendingBlock
	stack stack
	buf   strings.Builder
}

func newContext() *context {
	return &context{out: new(pendingBlock)}
}

func (ctx *context) newBlock() {
	prev, next := ctx.out, new(pendingBlock)
	ctx.stack.push(prev) // remember the former block
	ctx.out = next
	ctx.flush() // write the current buffer to out ( the new collection comment )
}

func (ctx *context) GetComments() string {
	// hrm.... the correct thing might be sending pop() to everyone...
	if ctx.buf.Len() > 0 {
		ctx.flush(runes.Newline)
	}
	str := ctx.out.String()
	ctx.out = nil
	return str
}

func (ctx *context) GetAllComments() (ret []string) {
	ret = append(ret, ctx.GetComments())
	for len(ctx.stack) > 0 {
		prev := ctx.stack.pop()
		ret = append(ret, prev.String())
	}
	slices.Reverse(ret)
	return
}

// write passed runes, and then the buffer, to out
func (ctx *context) flush(qs ...rune) {
	if cnt := ctx.out.terms; cnt > 0 {
		for i := 0; i < cnt; i++ {
			ctx.out.WriteRune(runes.Record)
		}
		ctx.out.terms = 0
	}
	if str := ctx.buf.String(); len(str) > 0 {
		writeRunes(ctx.out, qs...)
		io.WriteString(ctx.out, str)
		ctx.buf.Reset()
	}
}
