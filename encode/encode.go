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
	if e := enc.WriteValue(r.ValueOf(v), false); e != nil {
		err = e
	} else {
		// ends with an artificial newline
		// fwiw: i guess go's json does too.
		tabs := &enc.Tabs
		tabs.Softline()
		tabs.pad()
	}
	return
}

// true if written as a heredoc
func (enc *Encoder) writeHere(str string) (okay bool) {
	if okay = len(str) > 23 && strings.ContainsRune(str, runes.Newline); okay {
		lines := strings.FieldsFunc(str, func(q rune) bool {
			return q == runes.Newline
		})
		tabs := &enc.Tabs
		tabs.WriteString(`"""`)
		tabs.Softline()
		for _, el := range lines {
			tabs.Escape(el)
			tabs.Softline()
		}
		tabs.WriteString(`"""`)
	}
	return
}

// writes a single value to the stream wrapped by tab writer
// if the parent was  map, and there is a new sequence;
// then we want a newline
func (enc *Encoder) WriteValue(v r.Value, wasMaps bool) (err error) {
	// skips nil values; hrm.
	if v.IsValid() {
		tabs := &enc.Tabs

		if t := v.Type(); t.Implements(mappingType) {
			m := v.Interface().(TellMapping)
			tabs.OptionalLine(wasMaps)
			err = enc.WriteMapping(m.TellMapping())

		} else if t.Implements(sequenceType) {
			m := v.Interface().(TellSequence)
			tabs.OptionalLine(wasMaps)
			err = enc.WriteSequence(m.TellSequence())
		} else {
			switch k := v.Kind(); k {
			case r.Pointer, r.Interface:
				err = enc.WriteValue(v.Elem(), wasMaps)

			case r.Bool:
				str := formatBool(v)
				tabs.WriteString(str)

			case r.Int, r.Int8, r.Int16, r.Int32, r.Int64:
				str := formatInt(v)
				tabs.WriteString(str)

			case r.Uint, r.Uint8, r.Uint16, r.Uint32, r.Uint64:
				// tbd: tag for format? ( hex, #, etc. )
				str := formatUint(v)
				tabs.WriteString(str)

			case r.Float32, r.Float64:
				str := formatFloat(v)
				if f := v.Float(); math.IsInf(f, 0) || math.IsNaN(f) {
					err = fmt.Errorf("unsupported value %s", str)
				} else {
					tabs.WriteString(str)
				}

			case r.String:
				// fix: determine wrapping based on settings?
				// select raw strings based on the presence of escapes?
				if str := v.String(); !enc.writeHere(str) {
					tabs.Quote(str)
				}

			case r.Array, r.Slice:
				// tbd: look at tag for "want array"?
				if it, e := enc.Sequencer(v); e != nil {
					err = e
				} else if it == nil {
					tabs.WriteRune(runes.ArrayOpen)
					tabs.WriteRune(runes.ArrayClose)
				} else {
					tabs.OptionalLine(wasMaps)
					err = enc.WriteSequence(it)
				}

			case r.Map:
				if it, e := enc.Mapper(v); e != nil {
					err = e
				} else if it != nil {
					tabs.OptionalLine(wasMaps)
					err = enc.WriteMapping(it)
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
	tab := &enc.Tabs
	hasNext := it.Next() // dance around the possibly blank first element
	if !hasNext {
		return
	}

	// determine indentation style

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
			tab.Softline()
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
		/// * INDENT?
		tab.Indent(true, false) // nextIndent != IndentWithoutLine)

		// prefix comment:
		if prefix := cmt.Prefix; len(prefix) == 0 {
			tab.Space()
		} else {
			fixedWrite(tab, prefix)
		}
		// var nextIndent Indent
		// if maps {
		// 	nextIndent = IndentInline
		// } else {
		// 	nextIndent = IndentWithoutLine
		// }
		// value: recursive!
		if e := enc.WriteValue(val, maps); e != nil {
			err = e
			break
		}

		if suffix := cmt.Suffix; len(suffix) == 0 {
			tab.Softline()
		} else {
			fixedWrite(tab, suffix)
		}

		/// UNINDENT +/- yhr etiyr
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
