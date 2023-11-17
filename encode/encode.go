package encode

import (
	"bytes"
	"errors"
	"fmt"
	"math"
	r "reflect"
	"strconv"
	"unicode"

	"github.com/ionous/tell/runes"
)

func Encode(v any) (ret []byte, err error) {
	var out bytes.Buffer
	tab := TabWriter{Writer: &out}
	if e := WriteValue(&tab, r.ValueOf(v), false); e != nil {
		err = e
	} else {
		// end with an artificial newline?
		tab.WriteRune(runes.Newline)
		ret = out.Bytes()
	}
	return
}

func WriteValue(tab *TabWriter, v r.Value, indent bool) (err error) {
	switch v.Kind() {
	// write structs as maps?
	// should struct names be used as part of the signature?
	// how about package?
	// case r.Struct;

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
		// fix: determine wrapping based on settings
		// and write long strings as heredocs?
		// select raw strings based on the presence of escapes?
		str := strconv.Quote(v.String())
		tab.WriteString(str)

	case r.Pointer:
		err = WriteValue(tab, v.Elem(), indent)

	case r.Array, r.Slice:
		// tbd: look at tag for "want array"?
		if v.Len() > 0 {
			if indent {
				tab.Indent(true)
			}
			m := rseq{slice: v}
			err = writeSequence(tab, &m)
			if indent {
				tab.Indent(false)
			}
		}

	case r.Map:
		if t := v.Type(); t.Key().Kind() != r.String {
			err = fmt.Errorf("map keys must be string, have %T", t)
		} else {
			if v.Len() > 0 {
				if indent {
					tab.Indent(true)
				}
				m := makeSortedMap(v)
				err = writeMapping(tab, m)
				if indent {
					tab.Indent(false)
				}
			}
		}

	case r.Interface:
		if t := v.Type(); t.Implements(mappingType) {
			m := v.Interface().(Mapper)
			err = writeMapping(tab, m.TellMapping())

		} else if t.Implements(sequenceType) {
			m := v.Interface().(Sequencer)
			err = writeSequence(tab, m.TellSequence())

		} else {
			err = WriteValue(tab, v.Elem(), indent)
		}

	default:
		// tbd: Complex64&128?
		// others: Chan, Func, UnsafePointer
		err = fmt.Errorf("unexpected type %s(%T)", v.Kind(), v.Type())
	}
	if err == nil {
		tab.Newline()
	}
	return
}

var mappingType = r.TypeOf((*Mapper)(nil)).Elem()
var sequenceType = r.TypeOf((*Sequencer)(nil)).Elem()

func writeMapping(tab *TabWriter, it MappingIter) (err error) {
	for it.Next() {
		key, val := it.GetKey(), it.GetValue()
		if key = validateKey(key); len(key) == 0 {
			err = errors.New("invalid key")
			break
		} else {
			tab.Flush().WriteString(key)
			if key[len(key)-1] != runes.WordSep {
				tab.WriteRune(runes.WordSep)
			}
			tab.Space()
			if e := WriteValue(tab, r.ValueOf(val), true); e != nil {
				err = e
				break
			}
		}
	}
	return
}

func writeSequence(tab *TabWriter, it SequenceIter) (err error) {
	for it.Next() {
		val := it.GetValue()
		tab.Flush().WriteRune(runes.Dash)
		tab.Space()
		if e := WriteValue(tab, r.ValueOf(val), true); e != nil {
			err = e
			break
		} else {
			tab.Newline()
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
