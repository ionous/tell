package tell

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	r "reflect"

	"github.com/ionous/tell/decode"
	"github.com/ionous/tell/maps"
	"github.com/ionous/tell/maps/stdmap"
	"github.com/ionous/tell/notes"
)

// Decoder - follows the pattern of encoding/json
type Decoder struct {
	src   io.RuneReader
	inner decode.Decoder
}

// NewDecoder -
func NewDecoder(src io.Reader) *Decoder {
	var rr io.RuneReader
	if qr, ok := src.(io.RuneReader); ok {
		rr = qr
	} else {
		rr = bufio.NewReader(src)
	}
	return &Decoder{src: rr, inner: decode.MakeDecoder(
		stdmap.Builder,
		notes.DiscardComments(),
	)}
}

// control the creation of mappings for the upcoming Decode.
// the default is to create native maps ( via stdmap.Builder )
func (d *Decoder) SetMapper(maps maps.BuilderFactory) {
	d.inner.SetMapper(maps)
}

// control the creation of comment blocks for the upcoming Decode.
// the default is to discard comments.
func (d *Decoder) UseNotes(comments notes.Commentator) {
	d.inner.UseNotes(comments)
}

// configure the upcoming Decode to produce only floating point numbers.
// otherwise it will produce int for integers, and unit for hex specifications.
func (d *Decoder) UseFloats() {
	d.inner.UseFloats = true
}

// read a tell document from the stream configured in NewDecoder,
// and store the result at the value pointed by pv.
func (dec *Decoder) Decode(pv any) (err error) {
	out := r.ValueOf(pv)
	if out.Kind() != r.Pointer || out.IsNil() {
		err = &InvalidUnmarshalError{r.TypeOf(pv)}
	} else if out := out.Elem(); !out.CanSet() {
		err = errors.New("expected a settable value")
	} else if raw, e := dec.inner.Decode(dec.src); e != nil {
		err = e
	} else if raw == nil {
		out.SetZero()
	} else {
		res := r.ValueOf(raw)
		if rt, ot := res.Type(), out.Type(); rt.AssignableTo(ot) {
			out.Set(res)
		} else if res.CanConvert(ot) {
			out.Set(res.Convert(ot))
		} else {
			err = fmt.Errorf("result of %q cant be written to a pointer of %q", rt, ot)
		}
	}
	return
}

// As per package encoding/json, describes an invalid argument passed to Unmarshal or Decode.
// Arguments must be non-nil pointers
type InvalidUnmarshalError struct {
	Type r.Type
}

func (e *InvalidUnmarshalError) Error() (ret string) {
	if e.Type == nil {
		ret = "tell: Unmarshal(nil)"
	} else if e.Type.Kind() != r.Pointer {
		ret = "tell: Unmarshal(non-pointer " + e.Type.String() + ")"
	} else {
		ret = "tell: Unmarshal(nil " + e.Type.String() + ")"
	}
	return
}
