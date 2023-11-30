package decode

import (
	"github.com/ionous/tell/charm"
	"github.com/ionous/tell/notes"
	"github.com/ionous/tell/runes"
)

// internal rune used to indicate the end of a comment
// to help share handling with Newline,NextIndent
// todo? could refactor to use charm.Step, tho that time might be better spent on unwinding next indent
const commentLine = '\v'

// interpret every up to, and including, the end of the line as a comment.
// sends the last rune -- the newline as CommentLine -- to the passed eol state.
func CommentDecoder(out notes.Commentator, eol charm.State) charm.State {
	out.WriteRune(runes.Hash) // ick.
	return charm.Self("decode comment", func(self charm.State, r rune) (ret charm.State) {
		out.WriteRune(r)
		if r == runes.Newline {
			ret = eol.NewRune(commentLine)
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
			ret = CommentDecoder(doc.notes.OnNestedComment(), self)
		case commentLine:
			// fix: this eats all newlines
			// but nested should be altogether
			ret = MaintainIndent(doc, depth, self)
		}
		return
	})
}
