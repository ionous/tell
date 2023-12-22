package encode

import (
	"fmt"
	"io"
	"math"
	r "reflect"
	"strings"

	"github.com/ionous/tell/runes"
)

// an encoder that expects no comments
func MakeEncoder(w io.Writer) Encoder {
	var m MapTransform
	var n SequenceTransform
	return Encoder{
		TabWriter: TabWriter{Writer: w},
		Mapper:    m.Mapper(),
		Sequencer: n.Sequencer(),
	}
}

// use the "CommentBlock" encoder
func MakeCommentEncoder(w io.Writer) Encoder {
	var m MapTransform
	var n SequenceTransform
	return Encoder{
		TabWriter: TabWriter{Writer: w},
		Mapper:    m.Mapper(),
		Sequencer: n.Sequencer(),
		Commenter: CommentBlock,
	}
}

type Encoder struct {
	TabWriter
	Mapper    MappingFactory
	Sequencer SequenceFactory
	Commenter CommentFactory
}

func (enc *Encoder) Encode(v any) (err error) {
	if e := enc.WriteValue(r.ValueOf(v), indentNone); e != nil {
		err = e
	} else {
		// ends with an artificial newline
		// fwiw: i guess go's json does too.
		enc.Newline()
		enc.pad()
	}
	return
}

type indentation int

const (
	indentNone indentation = iota
	inlineLine
	indentWithoutLine
)

// true if written as a heredoc
func (enc *Encoder) writeHere(str string) (okay bool) {
	if okay = len(str) > 23 && strings.ContainsRune(str, runes.Newline); okay {
		lines := strings.FieldsFunc(str, func(q rune) bool {
			return q == runes.Newline
		})
		enc.WriteString(`"""`)
		enc.Indent(true, true)
		for _, el := range lines {
			enc.Escape(el)
			enc.Newline()
		}
		enc.WriteString(`"""`)
		enc.Indent(false, false)
	}
	return
}

// writes a single value to the stream wrapped by tab writer
func (enc *Encoder) WriteValue(v r.Value, indent indentation) (err error) {
	// skips nil values; hrm.
	if v.IsValid() {
		if t := v.Type(); t.Implements(mappingType) {
			m := v.Interface().(TellMapping)
			err = enc.WriteMapping(m.TellMapping())

		} else if t.Implements(sequenceType) {
			m := v.Interface().(TellSequence)
			err = enc.WriteSequence(m.TellSequence())
		} else {
			switch k := v.Kind(); k {
			case r.Pointer, r.Interface:
				err = enc.WriteValue(v.Elem(), indent)

			case r.Bool:
				str := formatBool(v)
				enc.WriteString(str)

			case r.Int, r.Int8, r.Int16, r.Int32, r.Int64:
				str := formatInt(v)
				enc.WriteString(str)

			case r.Uint, r.Uint8, r.Uint16, r.Uint32, r.Uint64:
				// tbd: tag for format? ( hex, #, etc. )
				str := formatUint(v)
				enc.WriteString(str)

			case r.Float32, r.Float64:
				str := formatFloat(v)
				if f := v.Float(); math.IsInf(f, 0) || math.IsNaN(f) {
					err = fmt.Errorf("unsupported value %s", str)
				} else {
					enc.WriteString(str)
				}

			case r.String:
				// fix: determine wrapping based on settings?
				// select raw strings based on the presence of escapes?
				if str := v.String(); !enc.writeHere(str) {
					enc.Quote(str)
				}

			case r.Array, r.Slice:
				// tbd: look at tag for "want array"?
				if it, e := enc.Sequencer(v); e != nil {
					err = e
				} else if it == nil {
					enc.WriteRune(runes.ArrayOpen)
					enc.WriteRune(runes.ArrayClose)
				} else {
					if indent != indentNone {
						enc.Indent(true, indent != indentWithoutLine)
					}
					err = enc.WriteSequence(it)
					if indent != indentNone {
						enc.Indent(false, true)
					}
				}

			case r.Map:
				if v.Len() > 0 {
					if indent != indentNone {
						enc.Indent(true, indent != indentWithoutLine)
					}
					if it, e := enc.Mapper(v); e != nil {
						err = e
					} else if it != nil {
						err = enc.WriteMapping(it)
						if indent != indentNone {
							enc.Indent(false, true)
						}
					}
				}

			default:
				// others: Complex, Chan, Func, UnsafePointer
				err = fmt.Errorf("unexpected type %s %s", v.Kind(), v.Type())
			}
		}
	}
	return
}

type sequenceAdapter struct{ SequenceIter }

func (sq sequenceAdapter) GetKey() string { return dashing }

const dashing = "-"

var mappingType = r.TypeOf((*TellMapping)(nil)).Elem()
var sequenceType = r.TypeOf((*TellSequence)(nil)).Elem()

// get the value of an iterator, ducking down to GetReflectedValue if it exists
func getValue(v interface{ GetValue() any }) (ret r.Value) {
	if i, ok := v.(GetReflectedValue); ok {
		ret = i.GetReflectedValue()
	} else {
		i := v.GetValue()
		ret = r.ValueOf(i)
	}
	return
}

func (enc *Encoder) WriteMapping(it MappingIter) (err error) {
	return enc.writeCollection(it, true)
}

func (enc *Encoder) WriteSequence(it SequenceIter) (err error) {
	a := sequenceAdapter{it}
	return enc.writeCollection(a, false)
}

func (enc *Encoder) writeCollection(it MappingIter, maps bool) (err error) {
	// determine indentation style
	var indent indentation
	if maps {
		indent = inlineLine
	} else {
		indent = indentWithoutLine
	}
	// setup a comment iterator:
	var cit CommentIter = noComments{} // expect none by default
	hasNext := it.Next()               // dance around the possibly blank first element
	if c := enc.Commenter; c != nil {
		key, val := it.GetKey(), getValue(it)
		if !maps || len(key) == 0 {
			cit, err = c(val)
			hasNext = err == nil && it.Next()
		}
	}
	if !hasNext && err == nil {
		if cit.Next() {
			cmt := cit.GetComment()
			enc.writeLines(cmt.Header, true)
		}
		if !maps {
			// this probably needs to be more sophisticated.
			// for prefix/suffix comments
			enc.WriteRune(runes.ArrayOpen)
			enc.WriteRune(runes.ArrayClose)
			enc.Newline()
		}
		return // early out.
	}
	//
	for hasNext {
		key, val := it.GetKey(), getValue(it)
		hasNext = it.Next()
		var cmt Comment
		if cit.Next() {
			cmt = cit.GetComment()
		}
		// header comment:
		enc.writeLines(cmt.Header, false)
		// key:
		enc.WriteString(key)
		// friendliness; write a separating colon if needed.
		if maps && key[len(key)-1] != runes.Colon {
			enc.WriteRune(runes.Colon)
		}
		// prefix comment:
		if prefix := cmt.Prefix; len(prefix) == 0 {
			enc.Space()
		} else {
			enc.writeLines(prefix, true)
		}
		// value:
		if e := enc.WriteValue(val, indent); e != nil {
			err = e
		} else {
			if suffix := cmt.Suffix; len(suffix) == 0 {
				enc.Newline()
			} else {
				enc.writeLines(suffix, true)
			}
		}
	}
	// write the footer ( if any )
	// ( it appears as a header comment for a non existent item )
	if err == nil && cit.Next() {
		cmt := cit.GetComment()
		enc.writeLines(cmt.Header, false)
	}
	return
}

func (enc *Encoder) writeLines(lines []string, allowInline bool) {
	for _, line := range lines {
		if len(line) > 0 {
			if allowInline {
				enc.WriteRune(runes.Space)
			}
			enc.WriteString(line)
		}
		enc.Newline()
		allowInline = false
	}
}
