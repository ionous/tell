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

type ErrorPos struct {
	y, x int
	err  error
}

func (e ErrorPos) Pos() (y int, x int) {
	return y, x
}

func (e ErrorPos) Unwrap() error {
	return e.err
}

func (e ErrorPos) Error() string {
	return fmt.Sprintf("%s at (%d, %d)", e.err, e.y, e.x)
}

func ErrorAt(y, x int, err error) ErrorPos {
	return ErrorPos{y, x, err}
}
