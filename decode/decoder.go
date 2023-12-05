package decode

import (
	"fmt"

	"github.com/ionous/tell/charm"
	"github.com/ionous/tell/token"
)

type decoder struct {
	state    func(token.Pos, token.Type, any) error
	out      output
	mapMaker mapMaker
}

func (d *decoder) decode() charm.State {
	d.state = d.docStart
	return token.NewTokenizer(d)
}

// implements the token thingy
func (d *decoder) Decoded(at token.Pos, tokenType token.Type, val any) error {
	return d.state(at, tokenType, val)
}

func (d *decoder) docStart(at token.Pos, tokenType token.Type, val any) (err error) {
	// fix: change these into functions?
	switch tokenType {
	case token.Comment:
		// fix
	case token.Key:
		key := val.(string)
		d.out.setPending(at.X, d.mapMaker.newCollection(key))
		d.state = d.waitForValue

	case token.Bool, token.Number, token.InterpretedString, token.RawString:
		d.out.setPending(at.X, newScalar(val))
		d.state = d.docValue

	default:
		panic("unknown token")
	}
	return
}

// the document value was written:
func (d *decoder) docValue(at token.Pos, tokenType token.Type, val any) (err error) {
	switch tokenType {
	case token.Comment:
	// fix
	default:
		err = fmt.Errorf("unexpected %s", tokenType)
	}
	return
}

// a key has just been decoded:
func (d *decoder) waitForValue(at token.Pos, tokenType token.Type, val any) (err error) {
	switch tokenType {
	case token.Comment:
	// fix
	case token.Key:
		if key := val.(string); at.X > d.out.indent {
			// a new collection
			p := d.mapMaker.newCollection(key)
			d.out.push(at.X, p)
		} else {
			// the key is for the same or an earlier collection
			// write a nil value, and go find the right collection
			if e := d.out.setValue(nil); e != nil {
				err = e
			} else if e := d.out.popToIndent(at.X); e != nil {
				err = e
			} else if e := d.out.setKey(key); e != nil {
				err = e
			}
		}

	case token.Bool, token.Number, token.InterpretedString, token.RawString:
		// the value detected is for this collection
		if at.X >= d.out.indent {
			err = d.out.setValue(val)
			d.state = d.waitForKey
		} else {
			// the value is for an earlier collection
			// write a nil value, and go find the right collection
			if e := d.out.setValue(nil); e != nil {
				err = e
			} else if e := d.out.popToIndent(at.X); e != nil {
				err = e
			} else if e := d.out.setValue(val); e != nil {
				err = e
			} else {
				d.state = d.waitForKey
			}
		}
	default:
		panic("unknown token")
	}
	return
}

// a value has just been decoded:
func (d *decoder) waitForKey(at token.Pos, tokenType token.Type, val any) (err error) {
	switch tokenType {
	case token.Comment:
		// fix
	case token.Key:
		if at.X > d.out.indent {
			err = fmt.Errorf("unexpected %s", tokenType)
		} else {
			key := val.(string) // find the collection this key is for:
			if e := d.out.popToIndent(at.X); e != nil {
				err = e
			} else if e := d.out.setKey(key); e != nil {
				err = e
			} else {
				d.state = d.waitForValue
			}
		}

	case token.Bool, token.Number, token.InterpretedString, token.RawString:
		err = fmt.Errorf("unexpected %s", tokenType)

	default:
		panic("unknown token")
	}
	return
}
