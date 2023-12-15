package decode

import (
	"errors"
	"fmt"

	"github.com/ionous/tell/runes"
	"github.com/ionous/tell/token"
)

// waiting for an array separator, or close.
// [ 1, 2 .... <-ex. here ]
func (d *Decoder) waitForSep(at token.Pos, tokenType token.Type, val any) (err error) {
	if q, ok := val.(rune); !ok {
		err = fmt.Errorf("%s unexpected", tokenType)
	} else {
		switch q {
		case runes.ArrayClose:
			d.endArray()
		case runes.ArraySeparator:
			if e := d.out.setKey(at.Y, ""); e != nil {
				err = e
			} else {
				d.state = d.waitForEl
			}
		default:
			err = errors.New("expected an array separator, or array close.")
		}
	}
	return
}

func (d *Decoder) waitForFirstEl(at token.Pos, tokenType token.Type, val any) (err error) {
	if tokenType == token.Array && val.(rune) == runes.ArrayClose {
		err = d.endArray()
	} else {
		err = d.waitForEl(at, tokenType, val)
	}
	return
}

// wait for the next array element.
// a separator here, or close, generates an implicit nil.
// [ 1, 2, .... <- ex. here ]
func (d *Decoder) waitForEl(at token.Pos, tokenType token.Type, val any) (err error) {
	switch tokenType {
	case token.Comment, token.Key:
		// fix: after cleaning up package notes, then revisit comments in arrays.
		err = fmt.Errorf("%s not allowed inside arrays", tokenType)

	case token.Bool, token.Number, token.String:
		err = d.newArrayValue(val)

	case token.Array:
		switch q := val.(rune); q {
		case runes.ArraySeparator:
			if e := d.newArrayValue(nil); e != nil {
				err = e
			} else {
				err = d.out.setKey(at.Y, "")
			}
		case runes.ArrayClose:
			if e := d.newArrayValue(nil); e != nil {
				err = e
			} else if e := d.endArray(); e != nil {
				err = e
			}
		case runes.ArrayOpen:
			// this wouldnt really be too terrible to support...
			// would have to sus out the kind of the top item on the stack
			// during close to determine what the next state should be.
			err = errors.New("nested arrays not allowed")

		default:
			panic("unknown array type")
		}

	default:
		panic("unknown token")
	}
	return
}

func (d *Decoder) newArrayValue(val any) (err error) {
	if e := d.out.setValue(val); e != nil {
		err = e
	} else {
		d.state = d.waitForSep
	}
	return
}

func (d *Decoder) endArray() (err error) {
	if e := d.out.popTop(); e != nil {
		err = e
	} else {
		if len(d.out.stack) == 0 {
			d.state = d.docFooter
		} else {
			d.state = d.waitForKey
		}
	}
	return
}
