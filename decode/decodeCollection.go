package decode

import "github.com/ionous/tell/charm"

func SequenceDecoder(c *Sequence) charm.State {
	return c.doc.Push(c.depth, charm.Statement("start sequence", func(r rune) charm.State {
		return c.EntryDecoder().NewRune(r)
	}))
}

func MappingDecoder(c *Mapping) charm.State {
	return c.doc.Push(c.depth, charm.Statement("start mapping", func(r rune) charm.State {
		return c.EntryDecoder().NewRune(r)
	}))
}
