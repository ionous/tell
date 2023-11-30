package notes

import (
	"github.com/ionous/tell/runes"
)

// uses concrete instances:
// because RuneWriter is an interface; pendingBlock doesnt need a pointer.
type stack []pendingBlock

// an in-progress comment block
type pendingBlock struct {
	Lines     // rune counting writer
	terms int // count empty terms
}

func makeBlock(w RuneWriter) pendingBlock {
	return pendingBlock{Lines: Lines{w: w}}
}

// write passed runes, and then the buffer, to out
func (p *pendingBlock) writeTerms() {
	if cnt := p.terms; cnt > 0 {
		for i := 0; i < cnt; i++ {
			p.WriteRune(runes.NextRecord)
		}
		p.terms = 0
	}
}

func (s stack) top() pendingBlock {
	return s[len(s)-1]
}

func (s *stack) push(prev pendingBlock) {
	*s = append(*s, prev)
}

// returns the old top
func (s *stack) pop() pendingBlock {
	out := s.top()
	*s = (*s)[:len(*s)-1]
	return out
}
