package decode

import (
	"io"

	"github.com/ionous/tell/charm"
	"github.com/ionous/tell/maps"
	"github.com/ionous/tell/notes"
	"github.com/ionous/tell/runes"
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
	} else {
		// some values (ex. number) dont finish until whitespace
		// send eol to flush them out; pop alone doesnt place nice with notes
		// ( because for instance, it can send an "EndCollection" without an eol of the comment line;
		//   but cant simply send eol to notes alone, because then it would eol before the scalar flush. )
		if next := charm.RunState(runes.Eof, run); next != nil {
			if es, ok := next.(charm.Terminal); ok && es != charm.Error(nil) {
				err = es
			}
		}
		if err == nil {
			err = doc.PopAll()
		}
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
		pendingValue: scalarValue{emptyValue},
		addsValue: func(val any) (_ error) {
			doc.value = val // tbd: error if already written?
			return
		},
	}

	loop := charm.Self("doc", func(self charm.State, r rune) (ret charm.State) {
		if check, ok := ent.pendingValue.(scalarValue); ok && check.v != emptyValue {
			if e := ent.finalizeEntry(); e != nil {
				ret = charm.Error(e)
			}
		}
		if ret == nil {
			ret = charm.RunState(r, HeaderDecoder(ent, depth, charm.Statement(
				"after header", func(r rune) (ret charm.State) {
					return LineValueDecoder(ent).NewRune(r)
				})))
		}
		return
	})
	// previously returned header
	return doc.PushCallback(depth, loop, ent.finalizeEntry)
}

// pop parser states up to the current indentation level
func (doc *Document) popToIndent() charm.State {
	return doc.History.Pop(doc.Cursor.Col)
}
