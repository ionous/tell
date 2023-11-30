package notes

// counts runes
type Lines struct {
	w     RuneWriter
	total int
}

// approximate count of runes in the buffer.
func (n *Lines) Len() int {
	return n.total
}

// for now assume that s is trimmed
func (n *Lines) WriteString(str string) (ret int, err error) {
	ret, err = writeString(n.w, str)
	n.total += ret
	return
}

func (n *Lines) WriteRune(q rune) (_ int, _ error) {
	n.w.WriteRune(q)
	n.total++
	return
}
