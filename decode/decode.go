package decode

import (
	"errors"
	"fmt"
	"io"
	"log"

	"github.com/ionous/tell/charm"
	"github.com/ionous/tell/charmed"
	"github.com/ionous/tell/collect"
	"github.com/ionous/tell/notes"
	"github.com/ionous/tell/runes"
	"github.com/ionous/tell/token"
)

// configure the production of sequences
func (d *Decoder) SetSequencer(seq collect.SequenceFactory) {
	d.collector.seqs = seq
}

// configure the production of mappings
func (d *Decoder) SetMapper(maps collect.MapFactory) {
	d.collector.maps = maps
}

// configure the production of comment blocks
func (d *Decoder) UseNotes(comments notes.Commentator) {
	d.memo = makeMemo(comments)
}

// read a tell document from the passed stream
func (d *Decoder) Decode(src io.RuneReader) (ret any, err error) {
	var x, y int
	run := charm.Parallel("parallel",
		charmed.FilterInvalidRunes(),
		d.decodeDoc(), // tbd: wrap with charmed.UnhandledError()? why/why not.
		charmed.DecodePos(&y, &x),
	)
	if e := charm.Read(src, run); e != nil {
		log.Println("error at", y, x)
		err = e
	} else if next := charm.RunState(runes.Eof, run); next != nil {
		if es, ok := next.(charm.Terminal); ok && es != charm.Error(nil) {
			log.Println("error at", y, x)
			err = es
		}
		d.memo.OnEof() // fix; can this be removed?
		if err == nil {
			ret, err = d.out.finalizeAll()
		}
	}
	return
}

// the decoder is *not* ready to use by default
// the mapper, sequencer, and notes need to be set.
type Decoder struct {
	out       output
	collector collector
	memo      memo
	state     func(token.Pos, token.Type, any) error
	// configure the tokenizer for the next decode
	UseFloats bool
}

func (d *Decoder) Position() (x int, y int) {
	pos := d.out.pos
	return pos.X, pos.Y
}

// implements token dispatch, hiding it from the public interface
type dispatcher struct{ *Decoder }

// implements the token thingy
func (dispatch dispatcher) Decoded(at token.Pos, tokenType token.Type, val any) error {
	return dispatch.state(at, tokenType, val)
}

func (d *Decoder) decodeDoc() charm.State {
	d.state = d.docStart
	t := token.Tokenizer{
		Notifier:  dispatcher{d},
		UseFloats: d.UseFloats,
	}
	return t.Decode()
}

func (d *Decoder) docStart(at token.Pos, tokenType token.Type, val any) (err error) {
	switch tokenType {
	case token.Comment:
		var zeroPos token.Pos
		d.memo.noteAt(zeroPos, at, val.(string))

	case token.Key:
		key := val.(string)
		d.out.setPending(at, d.collector.newCollection(key, d.memo.newComments()))
		d.state = d.waitForValue

	case token.Array:
		if q := val.(rune); q != runes.ArrayOpen {
			err = charm.InvalidRune(q)
		} else {
			d.out.setPending(at, d.collector.newArray(d.memo.newComments()))
			d.state = d.waitForFirstEl
		}

	case token.Bool, token.Number, token.String:
		d.out.setPending(at, newScalar(val))
		d.memo.OnDocScalar()
		d.state = d.docFooter

	default:
		panic("unknown token")
	}
	return
}

// the document value was written:
func (d *Decoder) docFooter(at token.Pos, tokenType token.Type, val any) (err error) {
	switch tokenType {
	case token.Comment:
		d.memo.noteAt(d.out.pos, at, val.(string))
	default:
		err = fmt.Errorf("unexpected %s", tokenType)
	}
	return
}

// a value has just been decoded:
func (d *Decoder) waitForKey(at token.Pos, tokenType token.Type, val any) (err error) {
	switch tokenType {
	case token.Key:
		if at.X > d.out.pos.X {
			err = fmt.Errorf("unexpected %s", tokenType)
		} else {
			key := val.(string) // find the collection this key is for:
			if ends, e := d.out.popToIndent(at.X); e != nil {
				err = e
			} else if e := d.out.setKey(at.Y, key); e != nil {
				err = e
			} else {
				d.memo.popped(ends)
				d.memo.OnKeyDecoded()
				d.state = d.waitForValue
			}
		}
	case token.Array, token.Bool, token.Number, token.String:
		err = fmt.Errorf("unexpected %s", tokenType)
	case token.Comment:
		err = d.onComment(at, tokenType, val.(string))
	default:
		panic("unknown token")
	}
	return
}

// a key has just been decoded:
func (d *Decoder) waitForValue(at token.Pos, tokenType token.Type, val any) (err error) {
	switch tokenType {
	case token.Comment:
		// greater will be a a comment ( maybe nested )
		// less will be a pop to an earlier collection.
		str := val.(string)
		if at.X != d.out.pos.X || len(str) == 0 {
			err = d.onComment(at, tokenType, str)
		} else {
			// a valid comment aligned with the key, indicates an implicit nil.
			if e := d.out.setValue(nil); e != nil {
				err = e
			} else {
				d.memo.OnScalarValue()
				d.memo.noteAt = d.memo.memoInterKey()
				d.memo.noteAt(d.out.pos, at, str)
				d.state = d.waitForKey
			}
		}

	case token.Key:
		// a new collection, or a key for some earlier one:
		if key := val.(string); at.X > d.out.pos.X {
			// a new collection
			p := d.collector.newCollection(key, d.memo.newComments())
			d.out.push(at, p)
		} else {
			// the key is for the same or an earlier collection
			// write a nil value, and go find the right collection
			if e := d.out.setValue(nil); e != nil {
				err = e
			} else if ends, e := d.out.popToIndent(at.X); e != nil {
				err = e
			} else if e := d.out.setKey(at.Y, key); e != nil {
				err = e
			} else {
				d.memo.OnScalarValue()
				d.memo.popped(ends)
				d.memo.OnKeyDecoded()
			}
		}

	case token.Array:
		// a new array for this, or some earlier collection.
		if q := val.(rune); q != runes.ArrayOpen {
			err = charm.InvalidRune(q)
		} else {
			if ends, e := d.out.popToIndent(at.X); e != nil {
				err = e
			} else {
				d.memo.popped(ends)
				p := d.collector.newArray(d.memo.newComments())
				d.out.push(at, p)
				d.state = d.waitForFirstEl
			}
		}

	case token.Bool, token.Number, token.String:
		// a value for this, or some earlier collection.
		// the value detected is for this collection
		if ends, e := d.out.popToIndent(at.X); e != nil {
			err = e
		} else if e := d.out.setValue(val); e != nil {
			err = e
		} else {
			d.memo.popped(ends)
			d.memo.OnScalarValue()
			d.state = d.waitForKey
		}

	default:
		panic("unknown token")
	}
	return
}

// waiting for an array separator, or close.
// [ 1, 2 .... <-ex. here ]
func (d *Decoder) waitForSep(at token.Pos, tokenType token.Type, val any) (err error) {
	if q, ok := val.(rune); !ok {
		err = fmt.Errorf("%s unexpected", tokenType)
	} else {
		switch q {
		case runes.ArrayClose:
			d.onArrayClose()
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
		err = d.onArrayClose()
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
		err = d.onArrayValue(val)

	case token.Array:
		switch q := val.(rune); q {
		case runes.ArraySeparator:
			if e := d.onArrayValue(nil); e != nil {
				err = e
			} else {
				err = d.out.setKey(at.Y, "")
			}

		case runes.ArrayClose:
			if e := d.onArrayValue(nil); e != nil {
				err = e
			} else if e := d.onArrayClose(); e != nil {
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

func (d *Decoder) onArrayValue(val any) (err error) {
	if e := d.out.setValue(val); e != nil {
		err = e
	} else {
		d.state = d.waitForSep
	}
	return
}

func (d *Decoder) onArrayClose() (err error) {
	if e := d.out.popTop(); e != nil {
		err = e
	} else if len(d.out.stack) == 0 {
		d.state = d.docFooter
	} else {
		d.memo.OnScalarValue() // report the completed array.
		d.state = d.waitForKey
	}
	return
}

func (d *Decoder) onComment(at token.Pos, tokenType token.Type, str string) (err error) {
	// is the (valid) comment for an earlier collection:
	// see? you can have ternaries in go.... O_o
	if ends, e := func() (ret int, err error) {
		if len(str) > 0 {
			ret, err = d.out.uncheckedPop(at.X)
		}
		return
	}(); e != nil {
		err = e
	} else {
		d.memo.popped(ends)
		d.memo.noteAt(d.out.pos, at, str)
	}
	return
}
