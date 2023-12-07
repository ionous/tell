package decode

import "github.com/ionous/tell/token"

type pendingAt struct {
	pos token.Pos
	pendingValue
}

type pendingStack []pendingAt

func (s *pendingStack) pop() (ret pendingAt) {
	end := len(*s) - 1
	(*s), ret = (*s)[:end], (*s)[end]
	return
}
