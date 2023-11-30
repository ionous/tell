package decode

import (
	"strings"

	"github.com/ionous/tell/charm"
	"github.com/ionous/tell/notes"
	"github.com/ionous/tell/runes"
)

// a sequence of array values are specified with:
// a dash, whitespace, the value, trailing whitespace.
// then loops back to itself to handle the next dash.
type Sequence struct {
	doc          *Document
	depth, count int
	values       []any
	comments     strings.Builder // for comments
}

// re: depth value decoding must first discover whether the dash is part of a number
// so the doc position isnt necessarily the real position.
func NewSequence(doc *Document, depth int) *Sequence {
	c := &Sequence{doc: doc, depth: depth}
	if keepComments := !notes.IsNothing(doc.notes); keepComments {
		c.values = make([]any, 1)
		c.doc.notes.BeginCollection(&c.comments)
	}
	return c
}

// a state that can parse one key:value pair
// intended to be used with doc.Push() to loop at a given indent.
func (c *Sequence) EntryDecoder() charm.State {
	ent := tellEntry{
		doc:          c.doc,
		depth:        c.depth + 2,
		count:        c.count,
		pendingValue: scalarValue{emptyValue},
		addsValue: func(val any) (_ error) {
			c.values = append(c.values, val)
			c.count++
			return
		},
	}
	next := charm.Self("sequence", func(self charm.State, r rune) (ret charm.State) {
		switch r {
		case runes.Hash:
			// this is in between sequence entries
			// potentially, its a header comment for the next element
			// if there is no element, it could be considered a tail
			// of the parent container; it can have nesting.
			ret = charm.RunState(r, HeaderDecoder(&ent, c.depth, self))

		case runes.Dash:
			// we dont need to hand off the dash rune itself
			ent.pendingValue = scalarValue{}
			ret = StartContentDecoding(&ent)
		}
		return
	})
	return c.doc.PushCallback(ent.depth, next, ent.finalizeEntry)
}

// used by parent collections to read the completed collection
func (c *Sequence) FinalizeValue() (ret any, err error) {
	if !notes.IsNothing(c.doc.notes) {
		c.values[0] = c.comments.String()
	}
	ret, c.values = c.values, nil
	return
}
