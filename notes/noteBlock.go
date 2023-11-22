package notes

import (
	"fmt"
	"io"

	"github.com/ionous/tell/runes"
)

// comment block for a collection
// contains a string builder so has to be on the stack or new'd
type pendingBlock struct {
	lines      Lines      // string buffer style output
	stage      blockStage //
	stageStart int        // number of lines at stage start
	terms      int        // counts empty collection entries
	flags      stageFlags // helper for record writing
}

// return and reset the buffer
func (n *pendingBlock) GetComments() string {
	return n.lines.GetComments()
}

func (n *pendingBlock) startStage(next blockStage) {
	prev := n.stage.set(next)
	// loop
	if prev >= valueStage && next <= valueStage {
		n.terms++
		n.flags = 0
	}
	n.stageStart = n.lines.Len()
}

// number of lines written for the current stag
func (n *pendingBlock) stageLines() int {
	currLines := n.lines.Len()
	return currLines - n.stageStart
}

func (n *pendingBlock) WriteRune(r rune) (_ int, _ error) {
	if !n.lines.writing {
		n.flushPending()
		if n.stage == footerStage {
			n.lines.Break()
		}
	}
	return n.lines.WriteRune(r)
}

func (n *pendingBlock) merge(src *Lines, useNewLine bool) {
	special := src.special // record before
	if str := src.GetComments(); len(str) > 0 {
		// yuck. determine whether to write a newline based on what's already there.
		// maybe more robust states could handle this? not sure.
		if n.lines.Len() > 0 && (useNewLine || !n.lines.special) {
			n.lines.Break()
		}
		io.WriteString(&n.lines.buf, str)
		n.lines.special = special
	}
}

// writes any and all pending form and record separators
func (n *pendingBlock) flushPending() {
	stageLines := n.stageLines()
	// sanity check for footer
	if stageLines > 0 && !n.stage.allowNesting() {
		msg := fmt.Sprintf("%s doesnt allow nesting", n.stage)
		panic(msg)
	}
	out := &n.lines.buf
	cnt := out.Len()
	// first rune after one or more empty terms?
	// write all those separators.
	if n.terms > 0 {
		dupe(out, n.terms, runes.Record)
		n.terms = 0
	}
	if n.stage >= keyStage && n.flags.set(keyStage) {
		out.WriteRune(runes.CollectionMark)
	}
	if n.stage >= valueStage && n.flags.set(valueStage) {
		out.WriteRune(runes.CollectionMark)
	}
	n.lines.special = out.Len() > cnt
}

func dupe(out RuneWriter, cnt int, r rune) {
	for i := 0; i < cnt; i++ {
		out.WriteRune(r)
	}
}
