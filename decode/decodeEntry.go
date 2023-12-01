package decode

import (
	"errors"

	"github.com/ionous/tell/charm"
	"github.com/ionous/tell/runes"
)

// represents a member of a collection
type tellEntry struct {
	doc          *Document
	depth, count int
	pendingValue pendingValue
	addsValue    func(any) error
}

// unparsed values are guarded by the empty value.
var emptyValue = errors.New("empty value")

// called when the indentation level is popped.
func (ent *tellEntry) finalizeEntry() (err error) {
	// protects against double call
	// ( ex. finalization in LineValueDecoder, vs pop due to eof )
	if ent.pendingValue != nil {
		if val, e := ent.pendingValue.FinalizeValue(); e != nil {
			err = e
		} else if val != emptyValue {
			if !isPendingCollection(ent.pendingValue) {
				ent.doc.notes.OnScalarValue()
			} else {
				ent.doc.notes.OnCollectionEnded()
			}
			err = ent.addsValue(val)
		}
		ent.pendingValue = nil
	}
	return
}

// immediately after the key has been decoded:
// parses contents and loops (by popping) after its done
func StartContentDecoding(ent *tellEntry) charm.State {
	if ent.count > 0 {
		// kind of ugly... but oh well.
		// might have been better to have notes eat the first key after collection.
		ent.doc.notes.OnKeyDecoded()
	}
	return charm.Step(ContentDecoder(ent),
		charm.Self("after entry", func(afterEntry charm.State, r rune) (ret charm.State) {
			switch r {
			case runes.Eof:
				ret = charm.Error(nil)
			case runes.Newline: // pop to find an appropriate next state
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
		case runes.Hash: // a hash starts the key comment
			if at := ent.doc.Col; at >= ent.depth {
				ret = KeyCommentDecoder(ent, at)
			}
		case runes.Eof:
			ret = charm.Error(nil)
		case runes.Newline:
			ent.doc.notes.WriteRune(r)
			fallthrough
		case commentLine: // a blank line with no contents is the header.
			ret = NextIndent(ent.doc, func(at int) (ret charm.State) {
				if at >= ent.depth {
					ret = HeaderDecoder(ent, at, LineValueDecoder(ent))
				}
				return
			})
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
			ret = CommentDecoder(ent.doc.notes, header)
		case runes.Eof:
			ret = charm.Error(nil)
		case runes.Newline:
			ent.doc.notes.WriteRune(r)
			fallthrough
		case commentLine:
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
			ret = CommentDecoder(ent.doc.notes, header)
		case runes.Eof:
			ret = charm.Error(nil)
		case runes.Newline:
			ent.doc.notes.WriteRune(r)
			fallthrough
		case commentLine:
			ret = MaintainIndent(ent.doc, depth, header)
		}
		return
	})
}

// expects the passed rune is the first rune of a value
// reads that value (if any) and any inline comment describing it.
func DecodeLineValue(ent *tellEntry, r rune) (ret charm.State) {
	// dont bother trying to read a value if it wasn't meant to be.
	if !isWhitespace(r) {
		ret = charm.RunState(r, LineValueDecoder(ent))
	}
	return
}

// expects the *next* rune is the first rune of a value
// reads that value (if any) and any inline comment describing it.
func LineValueDecoder(ent *tellEntry) (ret charm.State) {
	return charm.Step(ValueDecoder(ent),
		charm.MakeState(func() charm.State {
			ent.finalizeEntry()
			return PostValueDecoder(ent)
		}))
}

// the space to the right of a non-nil value
// ( nil-value comments look like key comments )
func PostValueDecoder(ent *tellEntry) (ret charm.State) {
	inlineIndent := -1
	return charm.Self("inline comment", func(self charm.State, r rune) (ret charm.State) {
		switch r {
		case runes.Space: // eat spaces on the line after the value
			ret = self
		case runes.Hash:
			inlineIndent = ent.doc.Col
			ret = CommentDecoder(ent.doc.notes, self)
		case runes.Eof:
			ret = charm.Error(nil)
		case runes.Newline:
			// the newline distinguishes trailing inline from block comments
			// the problem here is that --- we know there should be a value
			// but its not written yet.
			ent.doc.notes.WriteRune(r)
			fallthrough
		case commentLine: // a newline ( regardless of whether there was a comment )
			ret = NextIndent(ent.doc, func(at int) (ret charm.State) {
				// ent.depth is the indentation of the value; not the key/dash.
				// except for document level... :/
				var nested bool
				if cnt := ent.doc.History.Len(); cnt == 1 {
					nested = at > ent.depth
				} else if cnt > 1 {
					nested = (at >= ent.depth)
				}
				if (at == inlineIndent) ||
					(nested && (inlineIndent < 0 || at < inlineIndent)) {
					ret = NestedCommentDecoder(ent.doc)
				}
				return
			})
		}
		return
	})
}
