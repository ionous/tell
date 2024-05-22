package encode

import (
	"errors"
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
	return Encoder{
		Tabs:      TabWriter{Writer: w},
		Mapper:    m.Mapper(),
		Sequencer: MakeSequence,
	}
}

// use the "CommentBlock" encoder
func MakeCommentEncoder(w io.Writer) Encoder {
	var m MapTransform
	return Encoder{
		Tabs:             TabWriter{Writer: w},
		Mapper:           m.Mapper(),
		Sequencer:        MakeSequence,
		MapComments:      CommentBlock,
		SequenceComments: CommentBlock,
	}
}

type Encoder struct {
	Tabs              TabWriter
	Mapper, Sequencer StartCollection
	MapComments       Commenting
	SequenceComments  Commenting
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

// fix? determine quote style based on some sort of heuristic....
func (enc *Encoder) encodeQuotes(str string) {
	if !strings.ContainsRune(str, runes.Newline) {
		enc.Tabs.Quote(str)
	} else {
		// note: using strings.FieldsFunc isnt enough
		// by creating left and right parts; it eats trailing newlines
		var lines []string
		var prev int
		for i, q := range str {
			if q == runes.Newline {
				lines = append(lines, str[prev:i])
				prev = i + 1
			}
		}
		lines = append(lines, str[prev:])
		writeHere(&enc.Tabs, lines)
	}
}

func writeHere(tab *TabWriter, lines []string) {
	if len(lines) == 0 {
		panic("heredocs should have lines")
	}
	tab.WriteString(`|`)
	tab.Indent(true)
	var escaped, emptyLine bool
	for _, el := range lines {
		tab.Nextline()
		if tab.Escape(el) {
			escaped = true
		}
		emptyLine = len(el) == 0
	}
	if !emptyLine {
		// there was content in the final line,
		// so the heredoc should trim the final line.
		// if we're escaping, we have to do that with backslash.
		if escaped {
			tab.WriteRune('\\')
		}
		tab.Nextline()
	}
	// if we're escaping we have to write the double quotes
	// if we're not escaping we can choose to write it if the final line was empty
	if escaped || emptyLine {
		tab.WriteString(`"""`)
	} else {
		tab.WriteString(`'''`)
	}
	tab.Indent(false)
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
				enc.encodeQuotes(v.String())

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

func (enc *Encoder) WriteMapping(it Iterator, wasMaps bool) (err error) {
	return enc.writeCollection(it, enc.MapComments, wasMaps, true)
}

func (enc *Encoder) WriteSequence(it Iterator, wasMaps bool) (err error) {
	return enc.writeCollection(it, enc.SequenceComments, wasMaps, false)
}

func (enc *Encoder) writeCollection(it Iterator, cmts Commenting, wasMaps, maps bool) (err error) {
	tab := &enc.Tabs
	hasNext := it.Next() // dance around the possibly blank first element
	if !hasNext {
		return
	}

	// setup a comment iterator:
	var cit Comments = noComments{} // expect none by default
	if cmts != nil {
		key, val := it.GetKey(), getValue(it)
		if !maps || len(key) == 0 {
			cit, err = cmts(val)
			hasNext = err == nil && it.Next() // skip this comment value.
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
	if wasMaps {
		tab.Softline()
	}
	//
	for hasNext {
		key, val := it.GetKey(), getValue(it)
		if len(key) == 0 {
			err = errors.New("can't encode empty keys; maybe you meant to encode with comments?")
			break
		}
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
