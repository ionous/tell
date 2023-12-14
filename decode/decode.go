package decode

import (
	"errors"
	"fmt"
	"io"
	"log"

	"github.com/ionous/tell/charm"
	"github.com/ionous/tell/charmed"
	"github.com/ionous/tell/collect"
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
func (d *Decoder) UseNotes(w runes.RuneWriter) {
	d.collector.memo.doc = w // hrm.
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
	memoBlock memoBlock
	state     decoderState
	// configure the tokenizer for the next decode
	UseFloats bool
}

type decoderState func(token.Pos, token.Type, any) error

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

func (d *Decoder) writeComment(noteType noteType, str string) {
	d.collector.memo.Comment(d.out.comments(), noteType, str)
}

func (d *Decoder) docStart(at token.Pos, tokenType token.Type, val any) (err error) {
	switch tokenType {
	case token.Comment:
		str := val.(string)
		d.writeComment(NoteHeader, str)

	case token.Key:
		key := val.(string)
		d.out.setPending(at, d.collector.newCollection(key))
		d.state = d.waitForValue

	case token.Array:
		if q := val.(rune); q != runes.ArrayOpen {
			err = charm.InvalidRune(q)
		} else {
			d.out.setPending(at, d.collector.newArray())
			d.state = d.waitForFirstEl
		}

	case token.Bool, token.Number, token.String:
		d.out.setPending(at, makeDocScalar(val)) // sets doc scalar for "finalizeAll"
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
		d.writeComment(NoteFooter, val.(string))
	default:
		err = fmt.Errorf("unexpected %s", tokenType)
	}
	return
}

// a value had just been decoded, now we need a new key.
func (d *Decoder) waitForKey(at token.Pos, tokenType token.Type, val any) (err error) {
	switch tokenType {
	case token.Array, token.Bool, token.Number, token.String:
		err = fmt.Errorf("unexpected %s", tokenType)

	case token.Key:
		if key := val.(string); at.X > d.out.pos.X {
			err = fmt.Errorf("unexpected %s", tokenType)
		} else {
			err = d.newKey(at, key)
		}

	case token.Comment:
		if str := val.(string); len(str) > 0 {
			err = d.newComment(NoteSuffix, at, str)
		}
	default:
		panic("unknown token")
	}
	return
}

// a key has just been decoded, now we need a value.
func (d *Decoder) waitForValue(at token.Pos, tokenType token.Type, val any) (err error) {
	switch tokenType {
	case token.Array:
		if at.X < d.out.pos.X {
			err = InvalidIndent(d.out.pos, at)
		} else {
			p := d.collector.newArray()
			d.out.push(at, p)
			d.state = d.waitForFirstEl
		}

	case token.Bool, token.Number, token.String:
		if at.X < d.out.pos.X {
			err = InvalidIndent(d.out.pos, at)
		} else if e := d.out.setValue(val); e != nil {
			err = e
		} else {
			d.state = d.waitForKey
		}

	case token.Key:
		// a new collection, or a key for some earlier one:
		if key := val.(string); at.X > d.out.pos.X {
			p := d.collector.newCollection(key)
			d.out.push(at, p)
		} else {
			err = d.newKey(at, key)
		}

	case token.Comment:
		if str := val.(string); len(str) > 0 {
			err = d.newComment(NotePrefix, at, str)
		}

	default:
		panic("unknown token")
	}
	return
}

func (d *Decoder) newKey(at token.Pos, key string) (err error) {
	// the key is for the same or an earlier collection
	// write a nil value, and go find the right collection
	if e := d.out.popToIndent(at.X); e != nil {
		err = e
	} else if e := d.out.setKey(at.Y, key); e != nil {
		err = e
	} else {
		d.state = d.waitForValue // same as current state.
	}
	return
}

func (d *Decoder) newComment(defaultType noteType, at token.Pos, str string) (err error) {
	// eat blank lines: they don't change the interpretation here.
	if at.X > d.out.pos.X {
		noteType := defaultType
		if at.Y == d.out.pos.Y {
			noteType++
		}
		d.writeComment(noteType, str)
	} else {
		if e := d.out.popToIndent(at.X); e != nil {
			err = e
		} else {
			// interkey whether for this collection (ends==0) or a parent
			d.writeComment(NoteInterKey, str)
			d.state = d.waitForKey
		}
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
