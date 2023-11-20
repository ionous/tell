package notes

import (
	"io"
	"slices"
)

type Builder struct {
	blocks stack
	// headers need to be buffered
	// ( see also headerBuffering.md )
	// for headers:
	// 1. on collection:
	//    the new collection gets the buffered text as a heading;
	// 2. on scalar:
	//    the buffer goes to the current collection (as more padding)
	// 3. on end: ( get comments )
	//    the buffer goes to the parent collection as header.
	buf Lines
}

func (n *Builder) GetComments() string {
	top := n.blocks.pop() // returns the old top.
	n.flushBuffer(top)
	return top.GetComments()
}

func (n *Builder) GetAllComments() (ret []string) {
	for len(n.blocks) > 0 {
		ret = append(ret, n.GetComments())
	}
	slices.Reverse(ret)
	return ret
}

// start of a new collection
// create a new comment block ( each collection gets its own )
// and use current buffer as the header of the new block
func (n *Builder) OnBeginCollection() Commentator {
	top := n.blocks.create()
	if n.buf.NumLines() > 0 {
		top.startStage(headerStage)
		n.flushBuffer(top)
	}
	return n
}

// jump to the next term
func (n *Builder) OnTermDecoded() Commentator {
	top := n.blocks.top()
	top.terms++
	top.flags = 0
	top.stage = startingStage
	return n
}

func (n *Builder) OnBeginHeader() Commentator {
	top := n.blocks.top()
	top.advanceHeader()
	// hack to indicate a future newline
	// ( blocks manage this on their own;
	// buffered text is not a block )
	if top.stage == bufferStage {
		n.buf.newline = n.buf.NumLines() > 0
	}
	return n
}

func (n *Builder) OnKeyDecoded() Commentator {
	top := n.blocks.top()
	top.startStage(paddingStage)
	return n
}

func (n *Builder) OnScalarValue() Commentator {
	top := n.blocks.top()
	n.flushBuffer(top)
	top.startStage(inlineStage)
	return n
}

func (n *Builder) OnBeginFooter() Commentator {
	top := n.blocks.top()
	n.flushBuffer(top)
	top.startStage(footerStage)
	return n
}

func (n *Builder) WriteRune(r rune) (int, error) {
	var out RuneWriter
	if top := n.blocks.top(); top.stage != bufferStage {
		out = top
	} else {
		out = &n.buf
	}
	return out.WriteRune(r)
}

func (n *Builder) flushBuffer(top *pendingBlock) {
	if n.buf.NumLines() > 0 {
		// simulate writing a line ( or lines )
		top.startWriting()
		if str := n.buf.GetComments(); len(str) > 0 {
			io.WriteString(&top.lines.buf, str)
		}
		top.lines.writing = false
	}
}
