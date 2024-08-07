package decode

import (
	"fmt"
	"io"

	"github.com/ionous/tell/charm"
	"github.com/ionous/tell/charmed"
	"github.com/ionous/tell/collect"
	"github.com/ionous/tell/note"
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

// pass a valid target for document level comments
// a nil disables comment collection
// ( comments are disabled by default )
func (d *Decoder) UseNotes(b *note.Book) {
	d.docBlock = b
	d.collector.keepComments = b != nil
}

// read a tell document from the passed stream
func (d *Decoder) Decode(src io.RuneReader) (ret any, err error) {
	if d.docBlock == nil {
		d.docBlock = note.Nothing{}
	}
	var x, y int
	run := charm.Parallel("parallel",
		charmed.FilterInvalidRunes(),
		d.decodeDoc(), // tbd: wrap with charmed.UnhandledError()? why/why not.
		charmed.DecodePos(&y, &x),
	)
	p := charm.MakeParser(src)
	if e := p.ParseEof(run); e != nil {
		err = ErrorAt(y, x, e)
	} else {
		ret, err = d.out.finalizeAll()
	}
	return
}

// the decoder is *not* ready to use by default
// the mapper, sequencer, and notes need to be set.
type Decoder struct {
	out       output
	collector collector
	docBlock  note.Taker
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
	d.docBlock.BeginCollection(&d.collector.commentContext)
	t := token.Tokenizer{
		Notifier:  dispatcher{d},
		UseFloats: d.UseFloats,
	}
	return t.Decode()
}

func (d *Decoder) docStart(at token.Pos, tokenType token.Type, val any) (err error) {
	switch tokenType {
	case token.Comment:
		str := val.(string)
		err = d.docBlock.Comment(note.Header, str)

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
		scalar := pendingScalar{value: val, Taker: d.docBlock}
		d.out.setPending(at, scalar) // sets doc scalar for "finalizeAll"
		d.state = d.docSuffix

	default:
		panic("unknown token")
	}
	return
}

// the document value was written:
func (d *Decoder) docFooter(at token.Pos, tokenType token.Type, val any) (err error) {
	switch tokenType {
	case token.Comment:
		str := val.(string)
		err = d.docBlock.Comment(note.Footer, str)
	default:
		err = fmt.Errorf("unexpected %s while reading document footer", tokenType)
	}
	return
}

// the document value has just been decoded, process the comments as a suffix:
func (d *Decoder) docSuffix(at token.Pos, tokenType token.Type, val any) (err error) {
	switch tokenType {
	case token.Comment:
		if str := val.(string); at.X > d.out.pos.X {
			err = d.out.addComment(note.Suffix, at, str)
		} else if e := d.out.Comment(note.Footer, str); e != nil {
			err = e
		} else { // doesn't pop; there's no map or sequence.
			d.state = d.docFooter
		}
	default:
		err = fmt.Errorf("unexpected %s while reading document suffix", tokenType)
	}
	return
}

// a value had just been decoded, now we need a new key.
func (d *Decoder) waitForKey(at token.Pos, tokenType token.Type, val any) (err error) {
	switch tokenType {
	default:
		err = fmt.Errorf("unexpected %s while waiting for key", tokenType)

	case token.Key:
		key := val.(string)
		diff := at.X - d.out.pos.X
		if diff > 0 {
			err = fmt.Errorf("unexpected %s", tokenType)
		} else {
			// First:
			// - "value"
			// Second:
			if diff == 0 && isSequence(d.out.pendingValue) && len(key) > 0 {
				d.out.popTop()
			}
			if e := d.out.newKey(at, key); e != nil {
				err = e
			} else {
				d.state = d.waitForValue
			}
		}
	case token.Comment:
		err = d.onComment(at, note.Suffix, val.(string))
	}
	return
}

// a key has just been decoded, now we need a value.
func (d *Decoder) waitForValue(at token.Pos, tokenType token.Type, val any) (err error) {
	switch tokenType {
	default:
		err = fmt.Errorf("unexpected %s while waiting for value", tokenType)

	case token.Key:
		key := val.(string)
		diff := at.X - d.out.pos.X // sequences can act as mapping values
		keyAsValue := isMapping(d.out.pendingValue) && len(key) == 0
		//
		if diff > 0 || (diff == 0 && keyAsValue) {
			p := d.collector.newCollection(key)
			d.out.push(at, p)
		} else {
			err = d.out.newKey(at, key)
		}

	case token.Comment:
		err = d.onComment(at, note.Prefix, val.(string))

	case token.Array:
		if at.X < d.out.pos.X {
			err = InvalidIndent(d.out.pos, at)
		} else {
			p := d.collector.newArray()
			d.out.push(at, p)
			d.state = d.waitForFirstEl
		}

	case token.Bool, token.Number, token.String:
		if at.X <= d.out.pos.X {
			err = InvalidIndent(d.out.pos, at)
		} else if e := d.out.setValue(val); e != nil {
			err = e
		} else {
			d.state = d.waitForKey
		}
	}
	return
}

func (d *Decoder) onComment(at token.Pos, kind note.Type, str string) (err error) {
	// a prefix for the still yet to be found value
	if at.X > d.out.pos.X {
		err = d.out.addComment(kind, at, str)
	} else if e := d.out.newHeader(at, str); e != nil {
		err = e
	} else {
		// added a header for the collection ( or popped and added to a parent )
		// generates an implicit nil for the value we never encountered.
		d.state = d.waitForKey
	}
	return
}
