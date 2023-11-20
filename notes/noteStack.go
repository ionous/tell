package notes

type stack []*pendingBlock

func (s stack) top() *pendingBlock {
	return s[len(s)-1]
}

func (s *stack) create() *pendingBlock {
	*s = append(*s, new(pendingBlock))
	return s.top()
}

// returns the old top
func (s *stack) pop() *pendingBlock {
	out := s.top()
	*s = (*s)[:len(*s)-1]
	return out
}
