package decode

type pendingAt struct {
	indent int
	pendingValue
}

type pendingStack []pendingAt

func (s *pendingStack) pop() (ret pendingAt) {
	end := len(*s) - 1
	(*s), ret = (*s)[:end], (*s)[end]
	return
}
