package decode

import (
	"fmt"
	"io"
	"log"

	"github.com/ionous/tell/charm"
	"github.com/ionous/tell/charmed"
	"github.com/ionous/tell/maps"
	"github.com/ionous/tell/notes"
	"github.com/ionous/tell/runes"
	"github.com/ionous/tell/token"
)

func MakeDecoder(maps maps.BuilderFactory,
	comments notes.Commentator) Decoder {
	return Decoder{
		mapMaker: mapMaker{maps},
		memo:     makeMemo(comments),
	}
}

// configure the production of mappings
func (d *Decoder) SetMapper(maps maps.BuilderFactory) {
	d.mapMaker.create = maps
}

// configure the production of comment blocks
func (d *Decoder) UseNotes(comments notes.Commentator) {
	d.memo = makeMemo(comments)
}

// read a tell document from the passed stream
func (d *Decoder) Decode(src io.RuneReader) (ret any, err error) {
	var x, y int
	run := charm.Parallel("parallel",
		charmed.FilterControlCodes(),
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
		d.memo.OnEof() // fix; can this be removed?
	}
	return
}

type Decoder struct {
	out      output
	mapMaker mapMaker
	memo     memo
	state    func(token.Pos, token.Type, any) error
	// configure the tokenizer for the next decode
	UseFloats bool
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
	// tbd: change these into functions?
	switch tokenType {
	case token.Comment:
		var zeroPos token.Pos
		d.memo.noteAt(zeroPos, at, val.(string))

	case token.Key:
		key := val.(string)
		d.out.setPending(at, d.mapMaker.newCollection(key, d.memo.newComments()))
		d.state = d.waitForValue

	case token.Bool, token.Number, token.InterpretedString, token.RawString:
		d.out.setPending(at, newScalar(val))
		d.memo.OnDocScalar()
		d.state = d.docValue

	default:
		panic("unknown token")
	}
	return
}

// the document value was written:
func (d *Decoder) docValue(at token.Pos, tokenType token.Type, val any) (err error) {
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
	case token.Bool, token.Number, token.InterpretedString, token.RawString:
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
		if key := val.(string); at.X > d.out.pos.X {
			// a new collection
			p := d.mapMaker.newCollection(key, d.memo.newComments())
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

	case token.Bool, token.Number, token.InterpretedString, token.RawString:
		// the value detected is for this collection
		if at.X >= d.out.pos.X {
			if e := d.out.setValue(val); e != nil {
				err = e
			} else {
				d.memo.OnScalarValue()
				d.state = d.waitForKey
			}
		} else {
			// the value is for an earlier collection
			// mod: popToIndent sets nil, when it needs to.
			if ends, e := d.out.popToIndent(at.X); e != nil {
				err = e
			} else if e := d.out.setValue(val); e != nil {
				err = e
			} else {
				d.memo.popped(ends)
				d.memo.OnScalarValue()
				d.state = d.waitForKey
			}
		}
	default:
		panic("unknown token")
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
