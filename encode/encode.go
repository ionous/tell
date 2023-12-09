package encode

import (
	"fmt"
	"io"
	"math"
	r "reflect"
	"strconv"
	"unicode"

	"github.com/ionous/tell/runes"
)

func MakeEncoder(w io.Writer) Encoder {
	var m MapTransform
	var n SequenceTransform
	return Encoder{
		TabWriter: TabWriter{Writer: w},
		Mapper:    m.makeMapping,
		Sequencer: n.makeSequence,
	}
}

type Encoder struct {
	TabWriter
	Mapper    MappingFactory
	Sequencer SequenceFactory
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

// writes a single value to the stream wrapped by tab writer
func (enc *Encoder) WriteValue(v r.Value, indent indentation) (err error) {
	switch v.Kind() {
	// write structs as maps?
	// should struct names be used as part of the signature?
	// how about package?
	// case r.Struct;

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
		// fix: determine wrapping based on settings
		// and write long strings as heredocs?
		// select raw strings based on the presence of escapes?
		str := strconv.Quote(v.String())
		enc.WriteString(str)

	case r.Pointer:
		err = enc.WriteValue(v.Elem(), indent)

	case r.Array, r.Slice:
		// tbd: look at tag for "want array"?
		if v.Len() > 0 {
			if indent != indentNone {
				enc.Indent(true, indent != indentWithoutLine)
			}
			if it, e := enc.Sequencer(v); e != nil {
				err = e
			} else if it != nil {
				err = enc.WriteSequence(it)
				if indent != indentNone {
					enc.Indent(false, true)
				}
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

	case r.Interface:
		if t := v.Type(); t.Implements(mappingType) {
			m := v.Interface().(Mapper)
			err = enc.WriteMapping(m.TellMapping(enc))

		} else if t.Implements(sequenceType) {
			m := v.Interface().(Sequencer)
			err = enc.WriteSequence(m.TellSequence(enc))

		} else {
			err = enc.WriteValue(v.Elem(), indent)
		}

	default:
		// others: Complex, Chan, Func, UnsafePointer
		err = fmt.Errorf("unexpected type %s(%T)", v.Kind(), v.Type())
	}
	return
}

var mappingType = r.TypeOf((*Mapper)(nil)).Elem()
var sequenceType = r.TypeOf((*Sequencer)(nil)).Elem()

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
	for it.Next() {
		raw, val := it.GetKey(), getValue(it)
		if key := validateKey(raw); len(key) == 0 {
			err = fmt.Errorf("invalid key %q", raw)
			break
		} else {
			cmt := it.GetComment()
			// header comment:
			enc.writeComment(cmt.Header)
			// key:
			enc.WriteString(key)
			// friendliness; write a separating colon if needed.
			if key[len(key)-1] != runes.Colon {
				enc.WriteRune(runes.Colon)
			}
			// prefix comment:
			if prefix := cmt.Prefix; len(prefix) == 0 {
				enc.Space()
			} else {
				enc.WriteRune(runes.Space)
				enc.writeComment(prefix)
			}
			// value:
			if e := enc.WriteValue(val, inlineLine); e != nil {
				err = e
				break
			}
			if suffix := cmt.Suffix; len(suffix) == 0 {
				enc.Newline()
			} else {
				enc.writeComment(suffix)
			}
		}
	}
	return
}

func (enc *Encoder) writeComment(lines []string) {
	for _, line := range lines {
		if len(line) > 0 {
			if line[0] == runes.HTab {
				enc.Tab()
			}
			enc.WriteString(line)
		}
		enc.Newline()
	}
}

func (enc *Encoder) WriteSequence(it SequenceIter) (err error) {
	for it.Next() {
		val := getValue(it)
		enc.WriteRune(runes.Dash)
		enc.Space()
		if e := enc.WriteValue(val, indentWithoutLine); e != nil {
			err = e
			break
		} else {
			enc.Newline()
		}
	}
	return
}

// minimal check that the first element is a letter
// customize whether things are validated?
// and auto colon-ized.
func validateKey(key string) (ret string) {
	for _, first := range key {
		if unicode.IsLetter(first) {
			ret = key
		}
		break
	}
	return
}
