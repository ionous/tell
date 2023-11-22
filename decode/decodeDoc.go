package decode

import (
	"io"

	"github.com/ionous/tell/charm"
	"github.com/ionous/tell/maps"
	"github.com/ionous/tell/notes"
)

// document decoder
type Document struct {
	History
	Cursor
	value   any
	notes   notes.Commentator
	makeMap maps.BuilderFactory
}

func NewDocument(mapMaker maps.BuilderFactory, comments notes.Commentator) *Document {
	return &Document{makeMap: mapMaker, notes: comments}
}

// has incorrect behavior if called multiple times
func (doc *Document) ReadDoc(src io.RuneReader) (ret any, err error) {
	if e := doc.ReadLines(src, doc.EntryDecoder()); e != nil {
		err = e
	} else {
		ret, err = doc.Finalize()
	}
	return
}

// slightly lower level access for reading explicit kinds of values
// calling this multiple times leads to undefined results (fix?)
func (doc *Document) ReadLines(src io.RuneReader, start charm.State) (err error) {
	run := charm.Parallel("parse lines", FilterControlCodes(), UnhandledError(start), &doc.Cursor)
	if e := charm.Read(src, run); e != nil {
		err = e
	} else if e := doc.PopAll(); e != nil {
		err = e
	}
	return
}

// ugly: if preserve comments is true,
// { value, comment, error }
func (doc *Document) Finalize() (ret any, err error) {
	ret, doc.value = doc.value, nil
	return
}

// create an initial reader state
func (doc *Document) EntryDecoder() charm.State {
	depth := 0
	// fix call case here? [ and move case into NewCommentBlock ]
	ent := &tellEntry{
		doc:          doc,
		depth:        depth,
		pendingValue: scalarValue{},
		addsValue: func(val any) (_ error) {
			doc.value = val // tbd: error if already written?
			return
		},
	}
	next := HeaderDecoder(ent, depth, charm.Statement(
		"after header", func(r rune) (ret charm.State) {
			// doc.notes.OnKeyDecoded()
			return LineValueDecoder(ent).NewRune(r)
		}))
	return doc.PushCallback(depth, next, ent.finalizeEntry)
}

// pop parser states up to the current indentation level
func (doc *Document) popToIndent() charm.State {
	return doc.History.Pop(doc.Cursor.Col)
}
