package decode

import (
	"fmt"
	"io"

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
func (d *Decoder) UseNotes(b *CommentBlock) {
	d.docBlock = b
	d.collector.memo.keepComments = b != nil
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
		err = ErrorAt(y, x, e)
	} else if next := charm.RunState(runes.Eof, run); next != nil {
		if es, ok := next.(charm.Terminal); ok && es != charm.Error(nil) {
			err = ErrorAt(y, x, es)
		}
		if err == nil {
			if res, e := d.out.finalizeAll(); e != nil {
				err = e
			} else {
				// a completely empty document wont have a value
				// tbd: can it produce this ( and nil ) more cleanly?
				// ex. actually push a block up front, and setValue instead of setPending
				if d.out.pendingValue == nil && d.docBlock != nil {
					d.docBlock.End()
				}
				ret = res
			}
		}
	}
	return
}

// the decoder is *not* ready to use by default
// the mapper, sequencer, and notes need to be set.
type Decoder struct {
	out       output
	collector collector
	docBlock  *CommentBlock
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
	d.collector.memo.Begin(d.docBlock)
	t := token.Tokenizer{
		Notifier:  dispatcher{d},
		UseFloats: d.UseFloats,
	}
	return t.Decode()
}

func (d *Decoder) docStart(at token.Pos, tokenType token.Type, val any) (err error) {
	switch tokenType {
	case token.Comment:
		if str := val.(string); len(str) > 0 {
			d.docBlock.Comment(NoteHeader, str)
		}

	case token.Key:
		key := val.(string)
		d.out.setPending(at, d.collector.newCollection(key))
		d.out.waitingForValue = true
		d.state = d.waitForValue

	case token.Array:
		if q := val.(rune); q != runes.ArrayOpen {
			err = charm.InvalidRune(q)
		} else {
			d.out.setPending(at, d.collector.newArray())
			d.out.waitingForValue = true
			d.state = d.waitForFirstEl
		}

	case token.Bool, token.Number, token.String:
		scalar := pendingScalar{value: val, block: d.docBlock}
		d.out.setPending(at, scalar) // sets doc scalar for "finalizeAll"
		d.state = d.waitForFooter

	default:
		panic("unknown token")
	}
	return
}

// the document value was written:
func (d *Decoder) docFooter(at token.Pos, tokenType token.Type, val any) (err error) {
	switch tokenType {
	case token.Comment:
		if str := val.(string); len(str) > 0 {
			d.docBlock.Comment(NoteFooter, str)
		}
	default:
		err = fmt.Errorf("unexpected %s", tokenType)
	}
	return
}

// a value had just been decoded, now we need a new key.
func (d *Decoder) waitForFooter(at token.Pos, tokenType token.Type, val any) (err error) {
	switch tokenType {
	case token.Comment:
		if str := val.(string); len(str) > 0 {
			err = d.newComment(NoteSuffix, at, str)
		}
	default:
		err = fmt.Errorf("unexpected %s", tokenType)
	}
	return
}

// a value had just been decoded, now we need a new key.
func (d *Decoder) waitForKey(at token.Pos, tokenType token.Type, val any) (err error) {
	switch tokenType {
	default:
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
		// tbd: error if not +2 spaces?
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
		if n := d.out.comments(); n != nil {
			n.NextKey()
		}
		d.state = d.waitForValue // same as current state.
	}
	return
}

func (d *Decoder) newComment(defaultType noteType, at token.Pos, str string) (err error) {
	// eat blank lines: they don't change the interpretation here.
	if n := d.out.comments(); n != nil {
		if at.X > d.out.pos.X {
			noteType := defaultType
			if at.Y == d.out.pos.Y {
				noteType++
			}
			n.Comment(noteType, str)
		} else {
			if e := d.out.popToIndent(at.X); e != nil {
				err = e
			} else {
				var noteType noteType
				if len(d.out.stack) == 0 {
					d.state = d.docFooter
					noteType = NoteFooter
				} else {
					d.state = d.waitForKey
					noteType = NoteHeader
				}
				n.Comment(noteType, str)
			}
		}
	}
	return
}
