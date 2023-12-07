package runes

import (
	"fmt"
	"io"
	"unicode/utf8"
)

type RuneWriter interface {
	WriteRune(q rune) (n int, err error)
}

// write the passed rune to the passed writer
// checks for a RuneWriter interface, and if that fails, writes the rune as bytes.
func WriteRune(w io.Writer, q rune) (ret int, err error) {
	if rw, ok := w.(RuneWriter); ok {
		ret, err = rw.WriteRune(q)
	} else if !utf8.ValidRune(q) {
		err = fmt.Errorf("rune %d out of range", q)
	} else {
		var scratch [utf8.UTFMax]byte
		cnt := utf8.EncodeRune(scratch[:], q)
		ret, err = w.Write(scratch[:cnt])
	}
	return
}
