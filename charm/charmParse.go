package charm

import (
	"errors"
	"fmt"
	"io"
	"strings"
)

const Eof = rune(-1)

// Parse sends each rune of string to the passed state chart,
// Returns the error underlying error states,
// or the last returned state if there was no error.
func Parse(str string, first State) (ret State, err error) {
	return innerParse(first, strings.NewReader(str))
}

func Read(in io.RuneReader, first State) (err error) {
	_, err = innerParse(first, in)
	return
}

func innerParse(first State, in io.RuneReader) (ret State, err error) {
	try := first
	for i := 0; ; i++ {
		if r, _, e := in.ReadRune(); e != nil {
			if e != io.EOF {
				err = errors.Join(e, EndpointError{errContext(r, in), i, try, e.Error()})
			}
			break
		} else {
			if next := try.NewRune(r); next == nil {
				// no states left to parse remaining input
				err = EndpointError{errContext(r, in), i, try, "unhandled rune"}
				break
			} else if es, ok := next.(Terminal); ok {
				err = EndpointError{errContext(r, in), i, try, es.err.Error()}
				break
			} else {
				try = next
			}
		}
	}
	if err == nil {
		ret = try
	}
	return
}

// on error, provide a bit of the input remaining
// so that the user has an idea of where the error occurred
func errContext(r rune, in io.RuneReader) (ret string) {
	const size = 25
	var b strings.Builder
	b.WriteRune(r)
	for i := 0; i < size; i++ {
		if r, _, e := in.ReadRune(); e != nil {
			break
		} else {
			b.WriteRune(r)
		}
	}
	return b.String()
}

// ParseEof sends each rune of string to the passed state chart;
// after its done with the string, it sends an eof(-1) to flush any remaining data.
// see also Parse() which does not send the eof.
func ParseEof(str string, first State) (err error) {
	if last, e := innerParse(first, strings.NewReader(str)); e != nil {
		err = e
	} else if last != nil {
		if fini := last.NewRune(Eof); fini != nil {
			if es, ok := fini.(Terminal); ok && es.err != nil {
				err = fmt.Errorf("%s handling eof for %q", es.err, str)
			} else {
				// and if we are passing eof, shouldnt the states check for it and return nil?
				// err = EndpointError{str, len(str), fini}
			}
		}
	}
	return
}

// ended before the whole input was parsed.
type EndpointError struct {
	str    string
	end    int
	last   State
	reason string
}

// index of the failure point in the input
func (e EndpointError) End() int {
	return e.end
}

func (e EndpointError) Error() (ret string) {
	return fmt.Sprintf("%s %q (%q ended at index %d)",
		e.reason, e.str, StateName(e.last), e.end)
}
