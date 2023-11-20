package notes

import (
	"github.com/ionous/tell/runes"
)

// contains a string builder so has to be on the stack or new'd
type pendingBlock struct {
	lines      Lines
	stage      blockStage
	stageStart int        // used to detect undesirable nesting
	terms      int        // ever term indicates a line feed.
	flags      stageFlags // tracks which stages have lines.
}

func (n *pendingBlock) GetComments() string {
	return n.lines.GetComments()
}

func (n *pendingBlock) startStage(stage blockStage) {
	n.stage.set(stage)
	n.stageStart = n.lines.NumLines()
}

func (n *pendingBlock) advanceHeader() {
	var next blockStage
	switch n.stage {
	case paddingStage, bufferStage:
		next = bufferStage
	case emptyStage, startingStage:
		next = headerStage
	case headerStage:
		if n.stageLines() > 1 {
			panic("can't extend the header after having written multiple lines")
		}
		fallthrough
	case subheaderStage:
		next = subheaderStage
	default:
		panic("invalid state")
	}
	n.startStage(next)
}

func (n *pendingBlock) stageLines() int {
	currLines := n.lines.NumLines()
	return currLines - n.stageStart
}

func (n *pendingBlock) WriteRune(r rune) (_ int, _ error) {
	n.startWriting()
	return n.lines.WriteRune(r)
}

func (n *pendingBlock) startWriting() {
	// starting a new line? check whether to write previous info.
	if !n.lines.writing {
		// sanity checks
		if n.stage <= startingStage {
			panic("comment doesnt allow writing " + n.stage.String())
		}
		stageLines := n.stageLines()
		if stageLines > 0 && !n.stage.allowNesting() {
			panic("comment doesnt allow nesting " + n.stage.String())
		}
		// first rune of one or more empty terms?
		if n.terms > 0 {
			dupe(&n.lines.buf, n.terms, runes.Record)
			n.terms = 0
		}
		// first rune of this term?
		if n.stage >= paddingStage && n.flags.update(paddingStage) {
			n.lines.buf.WriteRune(runes.CollectionMark)
		}
		if n.stage >= inlineStage && n.flags.update(inlineStage) {
			n.lines.buf.WriteRune(runes.CollectionMark)
		}
		if n.stage.allowMultiple() {
			// hackish: the subheading and footers always are preceded by newline
			n.lines.buf.WriteRune(runes.Newline)
		} else if stageLines > 0 {
			// everything else can nest... and newlines need indentation.
			n.lines.buf.WriteRune(runes.Newline)
			n.lines.buf.WriteRune(runes.HTab)
		}
		// enter the new line
		n.lines.writing = true
	}
}

func dupe(out RuneWriter, cnt int, r rune) {
	for i := 0; i < cnt; i++ {
		out.WriteRune(r)
	}
}
