package notes

import (
	"slices"
)

type Builder struct {
	blocks stack
	// headers need to be buffered
	// ( see also headerBuffering.md )
	// for headers, approximately:
	// 1. on collection event:
	//    the new collection gets the buffer as a heading;
	// 2. on scalar event:
	//    the buffer goes to the current collection (as a key comment)
	// 3. on key event:
	//    this is the header of the new element in the current collection
	//    ( happens between elements )
	// 4. on end event: ( get comments )
	//    the buffer goes to the parent collection as header.
	// 5. on footer event
	//    panic, shouldnt be able to move from header to footer.
	// -  on header event:
	//    this is a new line, assuming there's something written
	// -  on rune
	//    add, and if necessary nest.
	buf Lines
}

func (n *Builder) init() *Builder {
	n.blocks.create()
	return n
}

func (n *Builder) GetComments() string {
	prev := n.blocks.pop() // returns the old top.
	if tgt := prev; n.buf.buf.Len() > 0 {
		if len(n.blocks) > 0 {
			tgt = n.blocks.top() // buffer goes to the parent collection as footer.
		}
		tgt.flushPending()
		tgt.merge(&n.buf, true)
	}
	return prev.GetComments()
}

func (n *Builder) GetAllComments() (ret []string) {
	for len(n.blocks) > 0 {
		ret = append(ret, n.GetComments())
	}
	slices.Reverse(ret)
	return ret
}

// start of a new collection
func (n *Builder) OnBeginCollection() Commentator {
	top := n.blocks.top()

	// if the container block has no comments
	// use the most recent as a header for the container
	// if it already has some comments, use it as a header
	// for incoming subcollection
	if top.stageLines() == 0 {
		n.flushBuffer(top)
	}

	top.startStage(valueStage)
	next := n.blocks.create() // create a new comment block ( each collection gets its own )
	n.flushBuffer(next)
	return n
}

func (n *Builder) OnParagraph() Commentator {
	top := n.blocks.top()
	if !top.stage.buffers() {
		top.startStage(startStage)
	} else {
		// stop auto-nesting
		n.buf.skipNest = true
		// only the last comment of the buffered region
		// should be kept as a header for a new collection
		n.flushBuffer(top)
	}
	return n
}

func (n *Builder) OnKeyDecoded() Commentator {
	top := n.blocks.top()
	n.flushBuffer(top) // inter-element buffering; header of the new element in the current collection
	top.startStage(keyStage)
	return n
}

func (n *Builder) OnScalarValue() Commentator {
	top := n.blocks.top()
	n.flushBuffer(top) // the buffer goes to the key comment
	top.startStage(valueStage)
	return n
}

func (n *Builder) OnFootnote() Commentator {
	top := n.blocks.top()
	top.startStage(footerStage)
	if n.buf.NumLines() > 0 {
		panic("footer shouldnt be buffered, or have any buffer")
	}
	return n
}

func (n *Builder) WriteRune(r rune) (int, error) {
	top := n.blocks.top()
	var out RuneWriter
	if !top.stage.buffers() {
		out = top
	} else {
		out = &n.buf
	}
	return out.WriteRune(r)
}

// fix? technically i think it should many queue a flush --
// just in case nothing actually gets written from Paragraph
func (n *Builder) flushBuffer(top *pendingBlock) {
	if n.buf.NumLines() > 0 {
		// write any empty records, etc.
		top.flushPending()
		top.merge(&n.buf, false)
	}
}
