package decode

import (
	"github.com/ionous/tell/charm"
	"github.com/ionous/tell/runes"
)

// represents a member of a collection
type tellEntry struct {
	doc          *Document
	depth        int
	pendingValue pendingValue
	addsValue    func(any) error
}

// pop parser states up to the current indentation level
func (ent *tellEntry) popToIndent() charm.State {
	return ent.doc.popToIndent()
}

// called when the indentation level is popped.
func (ent *tellEntry) finalizeEntry() (err error) {
	// finalizeEntry is the one moment common to all values being finished
	// including nil values...
	if _, isCollection := ent.pendingValue.(entryDecoder); !isCollection {
		ent.doc.notes.OnScalarValue()
	}
	if val, e := ent.pendingValue.FinalizeValue(); e != nil {
		err = e
	} else {
		err = ent.addsValue(val)
	}
	return
}

type entryDecoder interface{ EntryDecoder() charm.State }

var _ entryDecoder = (*Document)(nil)
var _ entryDecoder = (*Sequence)(nil)
var _ entryDecoder = (*Mapping)(nil)

// immediately after the key has been decoded:
// parses contents and loops (by popping) after its done
func StartContentDecoding(ent *tellEntry) charm.State {
	ent.doc.notes.OnKeyDecoded() // kind of an ugly place... but oh well.
	return charm.Step(ContentDecoder(ent),
		charm.Self("after entry", func(afterEntry charm.State, r rune) (ret charm.State) {
			switch r {
			case runes.Newline: // pop to find an appropriate next state
				//ent.doc.notes.OnTermDecoded() // MOVED
				ret = NextIndent(ent.doc, nil)
			}
			return
		}))
}

// Content appears after a collection marker:
func ContentDecoder(ent *tellEntry) charm.State {
	return charm.Self("contents", func(contents charm.State, r rune) (ret charm.State) {
		switch r {
		case runes.Space:
			ret = contents
		case runes.Newline: // a blank line with no contents is the header.
			ret = NextIndent(ent.doc, func(at int) (ret charm.State) {
				if at >= ent.depth {
					ret = HeaderDecoder(ent, at, LineValueDecoder(ent))
				}
				return
			})
		case runes.Hash: // a hash starts the key comment
			if at := ent.doc.Col; at >= ent.depth {
				ret = KeyCommentDecoder(ent, at)
			}
		default:
			if ent.doc.Col >= ent.depth {
				ret = DecodeLineValue(ent, r)
			}
		}
		return
	})
}

// starts reading just after the comment hash following a collection key.
func KeyCommentDecoder(ent *tellEntry, depth int) charm.State {
	return CommentDecoder(ent.doc.notes,
		NextIndent(ent.doc, func(at int) (ret charm.State) {
			switch {
			case at == depth:
				// the same indent means switch to split
				ret = HeaderDecoder(ent, depth, LineValueDecoder(ent))
			case at > depth:
				// a deeper indent means nesting
				// ( after nesting, the comment may appear at the original depth )
				ent.doc.Push(depth, HeaderDecoder(ent, depth, LineValueDecoder(ent)))
				ret = NestedCommentDecoder(ent.doc)
			}
			return
		}))
}

// starts reading a section that might be a key comment or might be an element header.
// the type depends on the eventual value of the entry
func HeaderDecoder(ent *tellEntry, depth int, next charm.State) charm.State {
	return charm.Self("header", func(header charm.State, r rune) (ret charm.State) {
		switch r {
		default:
			ret = next.NewRune(r)
		case runes.Hash:
			ret = CommentDecoder(ent.doc.notes.OnParagraph(), header)

		case runes.Newline:
			ret = NextIndent(ent.doc, func(at int) (ret charm.State) {
				switch {
				case at == depth:
					ret = SubheaderDecoder(ent, depth)

				case at > depth:
					ent.doc.Push(depth, LineValueDecoder(ent))
					ret = NestedCommentDecoder(ent.doc)
				}
				return
			})
		}
		return
	})
}

// subsequent lines of the header are all at the value's indent
// keep reading comments at that indent until there is a value.
func SubheaderDecoder(ent *tellEntry, depth int) charm.State {
	return charm.Self("second header", func(header charm.State, r rune) (ret charm.State) {
		switch r {
		default:
			ret = DecodeLineValue(ent, r)
		case runes.Hash:
			ret = CommentDecoder(ent.doc.notes.OnParagraph(), header)
		case runes.Newline:
			ret = MaintainIndent(ent.doc, header, depth)
		}
		return
	})
}

// expects the passed rune is the first rune of a value
// reads that value (if any) and any inline comment describing it.
func DecodeLineValue(ent *tellEntry, r rune) (ret charm.State) {
	// dont bother trying to read a value if it wasn't meant to be.
	if r != runes.Newline && r != runes.Space {
		ret = charm.RunState(r, LineValueDecoder(ent))
	}
	return
}

// expects the *next* rune is the first rune of a value
// reads that value (if any) and any inline comment describing it.
func LineValueDecoder(ent *tellEntry) (ret charm.State) {
	return charm.Step(ValueDecoder(ent), InlineCommentDecoder(ent))
}

// these are comments to the right of a known value.
func InlineCommentDecoder(ent *tellEntry) (ret charm.State) {
	inlineIndent := -1
	return charm.Self("inline comment", func(loop charm.State, r rune) (ret charm.State) {
		switch r {
		case runes.Space: // eat spaces on the line after the value
			ret = loop

		case runes.Hash: // an inline comment? read it; loop back to handle the newline.
			inlineIndent = ent.doc.Col
			ret = CommentDecoder(ent.doc.notes, loop)

		case runes.Newline: // a newline ( regardless of whether there was a comment )
			ret = NextIndent(ent.doc, func(at int) (ret charm.State) {
				if at == inlineIndent { // inline comments are all left aligned
					ret = loop
				} else { // a footer lives between the term and less than any inline comments
					if (at >= ent.depth) && (inlineIndent < 0 || at < inlineIndent) {
						ret = FooterDecoder(ent, at)
					}
				}
				return
			})
		}
		return
	})
}

// an optional comment can appear on the first line after a value
// starts on something other than whitespace
// at the indent we want to stick with.
func FooterDecoder(ent *tellEntry, wantIndent int) charm.State {
	return charm.Self("trailing comments", func(loop charm.State, r rune) (ret charm.State) {
		switch r {
		case runes.Hash:
			ret = CommentDecoder(ent.doc.notes.OnFootnote(), loop)
		case runes.Newline:
			ret = MaintainIndent(ent.doc, loop, wantIndent)
		}
		return
	})
}
