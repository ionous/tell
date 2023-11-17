package decode

import (
	"github.com/ionous/tell/charm"
	"github.com/ionous/tell/runes"
)

type Cursor struct {
	Col, Row int // x,y
}

// update the cursor; errors on all control characters except Newline.
func (c *Cursor) NewRune(r rune) charm.State {
	switch {
	case r == runes.Newline:
		c.Row++
		c.Col = 0
	default:
		c.Col++
	}
	return c
}
