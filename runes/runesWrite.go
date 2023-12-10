package runes

import (
	"io"
	"unicode/utf8"
)

type RuneWriter interface {
	WriteRune(q rune) (n int, err error)
}

// turn a writer into a rune writer
// first attempts to cast, otherwise builds an adapter for the output
func WriterToRunes(w io.Writer) (ret RuneWriter) {
	if rw, ok := w.(RuneWriter); ok {
		ret = rw
	} else {
		ret = runeWrapper{w}
	}
	return
}

type runeWrapper struct {
	io.Writer
}

func (rw runeWrapper) WriteRune(q rune) (n int, err error) {
	return WriteRune(rw.Writer, q)
}

func WriteRune(w io.Writer, q rune) (n int, err error) {
	var scratch [utf8.UTFMax]byte
	cnt := utf8.EncodeRune(scratch[:], q)
	return w.Write(scratch[:cnt])
}
