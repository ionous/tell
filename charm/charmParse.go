package charm

import (
	"fmt"
	"io"
	"strings"
)

const Eof = rune(-1)

// utility function which creates a string reader,
// a parser, and calls Parser.ParseEof()
func ParseEof(str string, first State) error {
	p := MakeParser(strings.NewReader(str))
	return p.ParseEof(first)
}

type Parser struct {
	in  io.RuneReader
	err error
	ofs int
}

func MakeParser(in io.RuneReader) Parser {
	return Parser{in: in}
}

func (p *Parser) Error() error {
	return p.err
}

// number of runes read from the input
func (p *Parser) Offset() int {
	return p.ofs
}

// consumes ~25 of the remaining runes for error reporting
func (p *Parser) Remaining() string {
	const size = 25
	var b strings.Builder
	if r, ok := p.err.(UnhandledRune); ok {
		b.WriteRune(rune(r))
	}
	for i := 0; i < size; i++ {
		if r, _, e := p.in.ReadRune(); e != nil {
			break
		} else {
			b.WriteRune(r)
		}
	}
	return b.String()
}

// run Parse() and send the final state an explicit Eof rune.
// unlike parse, only returns an error if there was an error.
// this also unwraps Finished and all terminal errors,
// returning the underlying error ( if any. )
func (p *Parser) ParseEof(first State) (err error) {
	if last, e := p.Parse(first); e != io.EOF {
		// return unrecognized errors as is.
		if es, ok := e.(Terminal); !ok {
			err = e
		} else if !es.Finished() {
			// honor a state if it finished ( by not returning error )
			// while unwrapping all other reported errors.
			err = es.Unwrap()
		}
	} else if fini := last.NewRune(Eof); fini != nil {
		// if there's *still* a state and its not a terminal state... report that
		if es, ok := fini.(Terminal); !ok {
			err = fmt.Errorf("unfinished states remain after end of file %s", fini)
		} else if !es.Finished() {
			// otherwise return the wrapped error
			err = es.Unwrap()
		}
	}
	return
}

// always returns an error, and the final state.
// ex. if the last rune was unhandled, then this returns an
// UnhandledRune error and the state that failed to handle it.
func (p *Parser) Parse(first State) (ret State, err error) {
	try := first
	for {
		if r, _, e := p.in.ReadRune(); e != nil {
			err = e // ex. io.Eof
			break
		} else if next := try.NewRune(r); next == nil {
			err = UnhandledRune(r)
			break
		} else if es, ok := next.(Terminal); ok {
			err = es // a wrapped error: ex. ErrFinished or otherwise.
			break
		} else {
			try = next // keep going.
			p.ofs++
		}
	}
	ret = try
	return
}
