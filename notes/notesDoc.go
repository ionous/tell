package notes

import (
	"github.com/ionous/tell/charm"
	"github.com/ionous/tell/runes"
)

func newDocument(ctx *context) charm.State {
	return charm.Step(newHeader(ctx), newBody(ctx))
}

func newBody(ctx *context) charm.State {
	return charm.Statement("awaitValue", func(q rune) (ret charm.State) {
		switch q {
		case runeEof:
			// flush the unused buffer as additional headers with newline
			if str := ctx.resolveBuffer(); len(str) > 0 {
				writeBuffer(&ctx.out, str, runes.Newline)
			}
			ret = charm.Error(nil) // there's only one buffer, so we're done.

		case runeValue:
			ret = charm.Step(readInline(ctx),
				charm.Statement("afterScalar", func(q rune) (ret charm.State) {
					return docEnd(ctx).NewRune(q)
				}))

		case runeKey:
			ret = charm.Step(newCollection(ctx),
				charm.Statement("afterCollection", func(q rune) (ret charm.State) {
					if q == runeCollected {
						ret = docEnd(ctx)
					}
					return
				}))
		default:
			ret = invalidRune("awaitValue", q)
		}
		return
	})
}
