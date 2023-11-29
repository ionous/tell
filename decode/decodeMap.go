package decode

import (
	"errors"
	"strings"

	"github.com/ionous/tell/charm"
	"github.com/ionous/tell/maps"
	"github.com/ionous/tell/notes"
	"github.com/ionous/tell/runes"
)

type Mapping struct {
	doc          *Document
	depth, count int
	key          Signature
	values       maps.Builder
	comments     strings.Builder // for comments
}

// maybe doc is a factory even?
func NewMapping(doc *Document, depth int) *Mapping {
	keepComments := !notes.IsNothing(doc.notes)
	m := &Mapping{doc: doc, depth: depth, values: doc.makeMap(keepComments)}
	if keepComments {
		m.doc.notes.BeginCollection(&m.comments)
	}
	return m
}

// a state that can parse one key-value pair
// caller uses doc.Push() to loop at a given indent.
func (c *Mapping) EntryDecoder() charm.State {
	ent := tellEntry{
		doc:          c.doc,
		depth:        c.depth + 2,
		count:        c.count,
		pendingValue: scalarValue{}, // unlike seq, maps can set the nil value by default
		addsValue: func(val any) (err error) {
			if c.key.IsKeyPending() {
				err = errors.New("signature must end with a colon, did you forget to quote a value?")
			} else if key, e := c.key.GetKey(); e != nil {
				err = e
			} else {
				c.values = c.values.Add(key, val)
				c.count++
			}
			return
		},
	}
	next := charm.Self("map entry", func(self charm.State, r rune) (ret charm.State) {
		switch r {
		case runes.Hash:
			ret = charm.RunState(r, HeaderDecoder(&ent, c.depth, self))
		default:
			// key and after key:
			ret = charm.RunStep(r, &c.key, charm.Statement("after key", func(r rune) charm.State {
				// unlike sequence, we need to hand off the first character that isnt the key
				return StartContentDecoding(&ent).NewRune(r)
			}))
		}
		return
	})
	return c.doc.PushCallback(ent.depth, next, ent.finalizeEntry)
}

// used by parent collections to read the completed collection
func (c *Mapping) FinalizeValue() (ret any, err error) {
	if c.key.IsKeyPending() {
		err = errors.New("signature must end with a colon, did you forget to quote a value?")
	} else {
		if !notes.IsNothing(c.doc.notes) {
			str := c.comments.String()
			c.values = c.values.Add("", str)
		}
		ret, c.values = c.values.Map(), nil
	}
	return
}
