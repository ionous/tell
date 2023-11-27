package notes

import "strings"

type stack []*pendingBlock

type pendingBlock struct {
	strings.Builder
	terms int // count empty terms
}

func (s stack) top() *pendingBlock {
	return s[len(s)-1]
}

func (s *stack) push(prev *pendingBlock) {
	*s = append(*s, prev)
}

// returns the old top
func (s *stack) pop() *pendingBlock {
	out := s.top()
	*s = (*s)[:len(*s)-1]
	return out
}
