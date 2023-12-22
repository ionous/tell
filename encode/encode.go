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
		Tabs:      TabWriter{Writer: w},
		Mapper:    m.Mapper(),
		Sequencer: n.Sequencer(),
	}
}

// use the "CommentBlock" encoder
func MakeCommentEncoder(w io.Writer) Encoder {
	var m MapTransform
	var n SequenceTransform
	return Encoder{
		Tabs:      TabWriter{Writer: w},
		Mapper:    m.Mapper(),
		Sequencer: n.Sequencer(),
		Commenter: CommentBlock,
	}
}

type Encoder struct {
	Tabs      TabWriter
	Mapper    MappingFactory
	Sequencer SequenceFactory
	Commenter CommentFactory
}

func (enc *Encoder) Encode(v any) (err error) {
	if e := enc.WriteValue(r.ValueOf(v), IndentNone); e != nil {
		err = e
	} else {
		// ends with an artificial newline
		// fwiw: i guess go's json does too.
		enc.Tabs.Newline()
		enc.Tabs.pad()
	}
	return
}

type Indent int

const (
	IndentNone Indent = iota
	IndentInline
	IndentWithoutLine
)

// true if written as a heredoc
func (enc *Encoder) writeHere(str string) (okay bool) {
	if okay = len(str) > 23 && strings.ContainsRune(str, runes.Newline); okay {
		lines := strings.FieldsFunc(str, func(q rune) bool {
			return q == runes.Newline
		})
		tabs := &enc.Tabs
		tabs.WriteString(`"""`)
		tabs.Indent(true, true)
		for _, el := range lines {
			tabs.Escape(el)
			tabs.Newline()
		}
		tabs.WriteString(`"""`)
		tabs.Indent(false, false)
	}
	return
}

// writes a single value to the stream wrapped by tab writer
func (enc *Encoder) WriteValue(v r.Value, indent Indent) (err error) {
	// skips nil values; hrm.
	if v.IsValid() {
		if t := v.Type(); t.Implements(mappingType) {
			m := v.Interface().(TellMapping)
			err = enc.WriteMapping(m.TellMapping(), indent)

		} else if t.Implements(sequenceType) {
			m := v.Interface().(TellSequence)
			err = enc.WriteSequence(m.TellSequence(), indent)
		} else {
			switch k := v.Kind(); k {
			case r.Pointer, r.Interface:
				err = enc.WriteValue(v.Elem(), indent)

			case r.Bool:
				str := formatBool(v)
				enc.Tabs.WriteString(str)

			case r.Int, r.Int8, r.Int16, r.Int32, r.Int64:
				str := formatInt(v)
				enc.Tabs.WriteString(str)

			case r.Uint, r.Uint8, r.Uint16, r.Uint32, r.Uint64:
				// tbd: tag for format? ( hex, #, etc. )
				str := formatUint(v)
				enc.Tabs.WriteString(str)

			case r.Float32, r.Float64:
				str := formatFloat(v)
				if f := v.Float(); math.IsInf(f, 0) || math.IsNaN(f) {
					err = fmt.Errorf("unsupported value %s", str)
				} else {
					enc.Tabs.WriteString(str)
				}

			case r.String:
				// fix: determine wrapping based on settings?
				// select raw strings based on the presence of escapes?
				if str := v.String(); !enc.writeHere(str) {
					enc.Tabs.Quote(str)
				}

			case r.Array, r.Slice:
				// tbd: look at tag for "want array"?
				if it, e := enc.Sequencer(v); e != nil {
					err = e
				} else if it == nil {
					enc.Tabs.WriteRune(runes.ArrayOpen)
					enc.Tabs.WriteRune(runes.ArrayClose)
				} else {
					err = enc.WriteSequence(it, indent)
				}

			case r.Map:
				if it, e := enc.Mapper(v); e != nil {
					err = e
				} else if it != nil {
					err = enc.WriteMapping(it, indent)
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

func (enc *Encoder) WriteMapping(it MappingIter, indent Indent) (err error) {
	return enc.writeCollection(it, indent, true)
}

func (enc *Encoder) WriteSequence(it SequenceIter, indent Indent) (err error) {
	a := sequenceAdapter{it}
	return enc.writeCollection(a, indent, false)
}

func (enc *Encoder) writeCollection(it MappingIter, indent Indent, maps bool) (err error) {
	tab := &enc.Tabs
	hasNext := it.Next() // dance around the possibly blank first element
	if !hasNext {
		return
	}
	if indent != IndentNone {
		tab.Indent(true, indent != IndentWithoutLine)
	}

	// determine indentation style
	var nextIndent Indent
	if maps {
		nextIndent = IndentInline
	} else {
		nextIndent = IndentWithoutLine
	}
	// setup a comment iterator:
	var cit CommentIter = noComments{} // expect none by default
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
			tab.writeInline(cmt.Header)
		}
		if !maps {
			// this probably needs to be more sophisticated.
			// for prefix/suffix comments
			tab.WriteRune(runes.ArrayOpen)
			tab.WriteRune(runes.ArrayClose)
			tab.Newline()
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
		tab.writeLines(cmt.Header)
		// key:
		tab.WriteString(key)
		// friendliness; write a separating colon if needed.
		if maps && key[len(key)-1] != runes.Colon {
			tab.WriteRune(runes.Colon)
		}
		// prefix comment:
		if prefix := cmt.Prefix; len(prefix) == 0 {
			tab.Space()
		} else {
			fixedWrite(tab, prefix)
		}
		// value: recursive!
		if e := enc.WriteValue(val, nextIndent); e != nil {
			err = e
		} else if suffix := cmt.Suffix; len(suffix) == 0 {
			tab.Newline()
		} else {
			fixedWrite(tab, suffix)
		}
	}
	if indent != IndentNone {
		tab.Indent(false, true)
	}

	// write the footer ( if any )
	// ( it appears as a header comment for a non existent item )
	if err == nil && cit.Next() {
		cmt := cit.GetComment()
		tab.writeLines(cmt.Header)
	}
	return
}

// fix
func (tab *TabWriter) writeInline(lines []string) {
	for i, line := range lines {
		if len(line) > 0 && i == 0 {
			tab.WriteRune(runes.Space)
		}
		tab.writeLine(line)
	}
}
