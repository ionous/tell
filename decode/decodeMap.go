package decode

import (
	"errors"
	"strings"
	"unicode"

	"github.com/ionous/tell/charm"
	"github.com/ionous/tell/maps"
	"github.com/ionous/tell/runes"
)

type Mapping struct {
	doc    *Document
	depth  int
	key    Signature
	values maps.Builder
	CommentBlock
}

// maybe doc is a factory even?
func NewMapping(doc *Document, header string, depth int) *Mapping {
	c := &Mapping{doc: doc, depth: depth, values: doc.MakeMap(doc.keepComments)}
	if doc.keepComments {
		c.keepComments = true
		c.comments.WriteString(header)
	}
	return c
}

// a state that can parse one key:value pair
// maybe push the returned thingy
// return doc.PushCallback(depth, STATE, ent.finalizeEntry)
func (c *Mapping) NewEntry() charm.State {
	ent := tellEntry{
		doc:          c.doc,
		depth:        c.depth + 2,
		pendingValue: computedValue{},
		addsValue: func(val any, comment string) (err error) {
			if key, e := c.key.GetKey(); e != nil {
				err = e
			} else {
				c.values = c.values.Add(key, val)
				c.comments.WriteString(comment)
				c.comments.WriteRune(runes.Record)
			}
			return
		},
	}
	next := charm.Self("map entry", func(self charm.State, r rune) (ret charm.State) {
		switch r {
		case runes.Hash:
			ret = charm.RunState(r, HeaderRegion(&ent, c.depth, self))
		default:
			// key and after key:
			ret = charm.RunStep(r, &c.key, charm.Statement("after key", func(r rune) charm.State {
				// unlike sequence, we need to hand off the first character that isnt the key
				return ContentsLoop(&ent).NewRune(r)
			}))
		}
		return
	})
	return c.doc.PushCallback(ent.depth, next, ent.finalizeEntry)
}

// used by parent collections to read the completed collection
func (c *Mapping) FinalizeValue() (ret any, err error) {
	if c.key.IsKeyPending() {
		err = errors.New("signature must end with a colon")
	} else {
		// write the comment block
		if c.keepComments {
			comment := strings.TrimRightFunc(c.comments.String(), unicode.IsSpace)
			c.values = c.values.Add("", comment)
		}
		ret, c.values = c.values.Map(), nil
	}
	return
}
