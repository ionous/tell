package decode

import "github.com/ionous/tell/charm"

type badIndent struct {
	have, want int // number of spaces
}

func (badIndent) Error() string {
	return "bad indent"
}

func BadIndent(have, want int) charm.State {
	return charm.Error(badIndent{have, want})
}
