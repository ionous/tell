package tell

import (
	"bytes"

	"github.com/ionous/tell/decode"
	"github.com/ionous/tell/encode"
	"github.com/ionous/tell/maps/stdmap"
	"github.com/ionous/tell/notes"
)

// Marshal returns a tell document representing the passed value.
//
// It traverses the passed type recursively to produce tell data.
// If a value implements encode.Mapper or encode.Sequencer,
// Marshal will use their iterators to serialize their contents.
//
// Otherwise, Marshal() uses the following rules:
//
// Boolean values are encoded as either 'true' or 'false'.
//
// Integer and floating point values are encoded as per go's
// strconv.FormatInt, strconv.FormatUnit, strconv.FormatFloat
// except int16 and uint16 are encoded as hex values starting with '0x'.
// NaN, infinities, and complex numbers will return an error.
//
// # Strings are encoded as per strconv.Quote
//
// Arrays and slice values are encoded as tell sequences.
// []byte is not handled in any special way. ( fix? )
//
// Maps with string keys are encoded as tell mappings; sorted by string.
// other key types return an error.
//
// Pointers and interface values are encoded in place as the value they represent.
// Cyclic data is not handled and will never return. ( fix? )
//
// Any other types will error ( ie. functions, channels, and structs )
//
// All documents end with a newline.
func Marshal(v any) (ret []byte, err error) {
	return encode.Encode(v)
}

// Unmarshal from a tell formatted document and store the result
// into the value pointed to by pv.
//
// Permissible values include:
// bool, floating point, signed and unsigned integers, maps and slices.
//
// For more flexibility, see package decode
func Unmarshal(in []byte, pv any) (err error) {
	dec := Decoder{
		src:   bytes.NewReader(in),
		inner: decode.MakeDecoder(stdmap.Builder, notes.DiscardComments()),
	}
	return dec.Decode(pv)
}
