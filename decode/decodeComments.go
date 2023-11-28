package decode

import (
	"github.com/ionous/tell/charm"
	"github.com/ionous/tell/notes"
	"github.com/ionous/tell/runes"
)

// interpret every up to, and including, the end of the line as a comment.
// sends the last rune -- the newline -- to the passed eol state.
func CommentDecoder(out notes.RuneWriter, eol charm.State) charm.State {
	out.WriteRune(runes.Hash) // ick.
	return charm.Self("decode comment", func(self charm.State, r rune) (ret charm.State) {
		out.WriteRune(r)
		if r == runes.Newline {
			ret = eol.NewRune(r)
		} else {
			ret = self
		}
		return
	})
}

// expects a series of comments all at the same ( current ) depth.
// returns unhandled the first time it can't find a comment hash.
// ( and, maybe there is no comment nesting at all. )
func NestedCommentDecoder(doc *Document) charm.State {
	depth := doc.Col
	return charm.Self("nested comment", func(self charm.State, r rune) (ret charm.State) {
		switch r {
		case runes.Hash:
			doc.notes.OnNestedComment()
			ret = CommentDecoder(doc.notes, self)
		case runes.Newline:
			ret = MaintainIndent(doc, self, depth)
		}
		return
	})
}
