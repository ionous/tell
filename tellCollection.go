package tell

import "github.com/ionous/tell/charm"

func StartSequence(c *Sequence) charm.State {
	return c.doc.Push(c.depth, charm.Statement("start sequence", func(r rune) charm.State {
		return c.NewEntry().NewRune(r)
	}))
}

func StartMapping(c *Mapping) charm.State {
	return c.doc.Push(c.depth, charm.Statement("start mapping", func(r rune) charm.State {
		return c.NewEntry().NewRune(r)
	}))
}
