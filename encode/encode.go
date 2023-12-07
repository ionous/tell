package encode

import (
	"bytes"
	"fmt"
	"io"
	"math"
	r "reflect"
	"strconv"
	"unicode"

	"github.com/ionous/tell/runes"
)

// creates a tab writer, writes to a local buffer, and returns the result.
// see WriteDocument.
func Encode(v any) (ret []byte, err error) {
	var out bytes.Buffer
	enc := MakeEncoder(&out)
	if e := enc.Encode(v); e != nil {
		err = e
	} else {
		ret = out.Bytes()
	}
	return
}

func MakeEncoder(w io.Writer) Encoder {
	var m MapTransform
	var n SequenceTransform
	return Encoder{
		tabs:      TabWriter{Writer: w},
		Mapper:    m.makeMapping,
		Sequencer: n.makeSequence,
	}
}

type Encoder struct {
	tabs      TabWriter
	Mapper    MappingFactory
	Sequencer SequenceFactory
}

func (enc *Encoder) Encode(v any) (err error) {
	if e := enc.WriteValue(r.ValueOf(v), indentNone); e != nil {
		err = e
	} else {
		// ends with an artificial newline
		// fwiw: i guess go's json does too.
		enc.tabs.Flush()
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
		enc.tabs.WriteString(str)

	case r.Int, r.Int8, r.Int16, r.Int32, r.Int64:
		str := formatInt(v)
		enc.tabs.WriteString(str)

	case r.Uint, r.Uint8, r.Uint16, r.Uint32, r.Uint64:
		// tbd: tag for format? ( hex, #, etc. )
		str := formatUint(v)
		enc.tabs.WriteString(str)

	case r.Float32, r.Float64:
		str := formatFloat(v)
		if f := v.Float(); math.IsInf(f, 0) || math.IsNaN(f) {
			err = fmt.Errorf("unsupported value %s", str)
		} else {
			enc.tabs.WriteString(str)
		}

	case r.String:
		// fix: determine wrapping based on settings
		// and write long strings as heredocs?
		// select raw strings based on the presence of escapes?
		str := strconv.Quote(v.String())
		enc.tabs.WriteString(str)

	case r.Pointer:
		err = enc.WriteValue(v.Elem(), indent)

	case r.Array, r.Slice:
		// tbd: look at tag for "want array"?
		if v.Len() > 0 {
			if indent != indentNone {
				enc.tabs.Indent(true, indent != indentWithoutLine)
			}
			if it, e := enc.Sequencer(v); e != nil {
				err = e
			} else if it != nil {
				err = enc.WriteSequence(it)
				if indent != indentNone {
					enc.tabs.Indent(false, true)
				}
			}
		}

	case r.Map:
		if v.Len() > 0 {
			if indent != indentNone {
				enc.tabs.Indent(true, indent != indentWithoutLine)
			}
			if it, e := enc.Mapper(v); e != nil {
				err = e
			} else if it != nil {
				err = enc.WriteMapping(it)
				if indent != indentNone {
					enc.tabs.Indent(false, true)
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
	if err == nil {
		enc.tabs.Newline()
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
			enc.tabs.Flush().WriteString(key)
			if key[len(key)-1] != runes.Colon {
				enc.tabs.WriteRune(runes.Colon)
			}
			enc.tabs.Space()
			if e := enc.WriteValue(val, inlineLine); e != nil {
				err = e
				break
			}
		}
	}
	return
}

func (enc *Encoder) WriteSequence(it SequenceIter) (err error) {
	for it.Next() {
		val := getValue(it)
		enc.tabs.Flush().WriteRune(runes.Dash)
		enc.tabs.Space()
		if e := enc.WriteValue(val, indentWithoutLine); e != nil {
			err = e
			break
		} else {
			enc.tabs.Newline()
		}
	}
	return
}

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
