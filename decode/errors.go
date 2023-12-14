package decode

import (
	"fmt"

	"github.com/ionous/tell/token"
)

type invalidIndent struct {
	want, got token.Pos
}

func InvalidIndent(want, got token.Pos) invalidIndent {
	return invalidIndent{want, got}
}

func (e invalidIndent) Error() string {
	return fmt.Sprintf("Invalid indent: %d", e.got.X)
}
