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
		tab := &enc.Tabs
		tab.Softline()
		tab.pad()
	}
	return
}

// true if written as a heredoc
func (enc *Encoder) writeHere(str string) (okay bool) {
	if okay = len(str) > 23 && strings.ContainsRune(str, runes.Newline); okay {
		lines := strings.FieldsFunc(str, func(q rune) bool {
			return q == runes.Newline
		})
		tab := &enc.Tabs
		tab.WriteString(`"""`)
		tab.Softline()
		for _, el := range lines {
			tab.Escape(el)
			tab.Softline()
			tab.newLines++ // not needed if using backtick literals... hrm.
		}
		tab.newLines--
		tab.WriteString(`"""`)
	}
	return
}

// writes a single value to the stream wrapped by tab writer
// if the parent was  map, and there is a new sequence;
// then we want a newline
func (enc *Encoder) WriteValue(v r.Value, wasMaps bool) (err error) {
	// skips nil values; hrm.
	if v.IsValid() {
		tab := &enc.Tabs

		if t := v.Type(); t.Implements(mappingType) {
			m := v.Interface().(TellMapping)
			err = enc.WriteMapping(m.TellMapping(), wasMaps)

		} else if t.Implements(sequenceType) {
			m := v.Interface().(TellSequence)
			err = enc.WriteSequence(m.TellSequence(), wasMaps)
		} else {
			switch k := v.Kind(); k {
			case r.Pointer, r.Interface:
				err = enc.WriteValue(v.Elem(), wasMaps)

			case r.Bool:
				str := formatBool(v)
				tab.WriteString(str)

			case r.Int, r.Int8, r.Int16, r.Int32, r.Int64:
				str := formatInt(v)
				tab.WriteString(str)

			case r.Uint, r.Uint8, r.Uint16, r.Uint32, r.Uint64:
				// tbd: tag for format? ( hex, #, etc. )
				str := formatUint(v)
				tab.WriteString(str)

			case r.Float32, r.Float64:
				str := formatFloat(v)
				if f := v.Float(); math.IsInf(f, 0) || math.IsNaN(f) {
					err = fmt.Errorf("unsupported value %s", str)
				} else {
					tab.WriteString(str)
				}

			case r.String:
				// fix: determine wrapping based on settings?
				// select raw strings based on the presence of escapes?
				if str := v.String(); !enc.writeHere(str) {
					tab.Quote(str)
				}

			case r.Array, r.Slice:
				// tbd: look at tag for "want array"?
				if it, e := enc.Sequencer(v); e != nil {
					err = e
				} else if it == nil {
					tab.WriteRune(runes.ArrayOpen)
					tab.WriteRune(runes.ArrayClose)
				} else {
					err = enc.WriteSequence(it, wasMaps)
				}

			case r.Map:
				if it, e := enc.Mapper(v); e != nil {
					err = e
				} else if it != nil {
					err = enc.WriteMapping(it, wasMaps)
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

func (enc *Encoder) WriteMapping(it MappingIter, wasMaps bool) (err error) {
	return enc.writeCollection(it, wasMaps, true)
}

func (enc *Encoder) WriteSequence(it SequenceIter, wasMaps bool) (err error) {
	a := sequenceAdapter{it}
	return enc.writeCollection(a, wasMaps, false)
}

func (enc *Encoder) writeCollection(it MappingIter, wasMaps, maps bool) (err error) {
	tab := &enc.Tabs
	hasNext := it.Next() // dance around the possibly blank first element
	if !hasNext {
		return
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
			tab.Softline()
		}
		return // early out.
	}
	tab.OptionalLine(wasMaps)
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
		// key; friendliness; write a separating colon if needed.
		tab.WriteString(key)
		if maps && key[len(key)-1] != runes.Colon {
			tab.WriteRune(runes.Colon)
		}
		tab.Indent(true)
		{
			// prefix comment:
			if prefix := cmt.Prefix; len(prefix) == 0 {
				tab.Space()
			} else {
				fixedWrite(tab, prefix)
			}
			// value: recursive!
			if e := enc.WriteValue(val, maps); e != nil {
				err = e
				break
			}
			if suffix := cmt.Suffix; len(suffix) > 0 {
				fixedWrite(tab, suffix)
			}
		}
		tab.Indent(false)
		tab.Softline()
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
