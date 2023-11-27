package notes

import (
	"github.com/ionous/tell/charm"
)

// starting immediately after a document scalar has been detected:
// decodes any inline trailing comment(s).
// documents don't have trailing block comments;
// instead they have document footer comments ( handled by docEnd )
func docScalar(ctx *context, docEnd makeState) charm.State {
	d := trailingDecoder{ctx.out}
	return charm.Step(d.awaitInline(), kickOff(docEnd))
}
