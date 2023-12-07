package tell

import (
	"io"

	"github.com/ionous/tell/encode"
)

// Encoder - follows the pattern of encoding/json
type Encoder encode.Encoder

// NewEncoder -
func NewEncoder(w io.Writer) *Encoder {
	enc := encode.MakeEncoder(w)
	return (*Encoder)(&enc)
}

// Encode - serializes the passed document to the encoder's stream
// followed by a newline character.
// tell doesnt support multiple documents in the same file,
// but this interface doesn't stop callers from trying
func (enc *Encoder) Encode(v any) (err error) {
	inner := (*encode.Encoder)(enc)
	return inner.Encode(v)
}

func (enc *Encoder) SetMapper(n encode.MappingFactory) {
	inner := (*encode.Encoder)(enc)
	inner.Mapper = n
}

func (enc *Encoder) SetSequencer(n encode.SequenceFactory) {
	inner := (*encode.Encoder)(enc)
	inner.Sequencer = n
}
