package encode

import (
	"fmt"
	r "reflect"
)

type SequenceTransform struct {
	commentFactory CommentFactory
	loc            CommentLocation
}

// return a factory function for the encoder
func (n *SequenceTransform) Sequencer() SequenceFactory {
	return n.makeSequence
}

type CommentLocation int

const (
	NoComments CommentLocation = iota
	CommentsAtFront
	CommentsAtBack
)

// todo? sort values; by default sequences are not sorted
// func (m *SequenceTransform) Sort(t func(a, b r.Value) bool) {
// 	m.sort = t
// }

// the factory is handed the sequence element at the CommentLocation.
// the default factory assumes a tell standard comment block,
// and errors if the the value isn't an interface with an underlying string value.
// ie. it matches []any{"comment"}
func (n *SequenceTransform) CommentFactory(fn func(r.Value) (CommentIter, error)) *SequenceTransform {
	n.commentFactory = fn
	return n
}

// by default comments come from the first element
func (n *SequenceTransform) CommentLocation(where CommentLocation) *SequenceTransform {
	n.loc = where
	return n
}

func (n *SequenceTransform) makeSequence(src r.Value) (ret SequenceIter, err error) {
	if e := validateSeq(src); e != nil {
		err = e
	} else {
		newComments := n.commentFactory
		if newComments == nil {
			newComments = DiscardComments
		}
		var cit CommentIter
		if cnt := src.Len(); cnt > 0 {
			var cmt r.Value
			switch n.loc {
			case CommentsAtFront:
				cmt, src = src.Index(0), src.Slice(1, cnt)
			case CommentsAtBack:
				cmt, src = src.Index(cnt-1), src.Slice(0, cnt-1)
			default:
				cmt = blank
			}
			cit, err = newComments(cmt)
		}
		if err == nil {
			ret = &rseq{slice: src, comments: cit}
		}
	}
	return
}

func validateSeq(src r.Value) (err error) {
	if k := src.Kind(); k != r.Slice && k != r.Array {
		err = fmt.Errorf("slices must be of interface type, have %s(%s)", k, src.Type())
	}
	return
}

type rseq struct {
	slice    r.Value
	next     int
	comments CommentIter
}

func (m *rseq) Next() (okay bool) {
	if okay = m.next < m.slice.Len(); okay {
		// advance comments, but dont force them to have the same number of elements
		if m.comments != nil {
			m.comments.Next() // alt: could swap to emptyComments when done.
		}
		m.next++
	}
	return
}

func (m *rseq) GetValue() any {
	return m.GetReflectedValue().Interface()
}

func (m *rseq) GetReflectedValue() r.Value {
	at := m.next - 1
	return m.slice.Index(at)
}

func (m *rseq) GetComment() (ret Comment) {
	if m.comments != nil {
		ret = m.comments.GetComment()
	}
	return
}
