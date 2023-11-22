package decode

import "github.com/ionous/tell/charmed"

// used by tellEntry to read values when the entry is finished.
// implemented by the collection types directly.
type pendingValue interface {
	FinalizeValue() (any, error)
}

// a final value, ex. from a boolean.
type scalarValue struct{ v any }

func (v scalarValue) FinalizeValue() (any, error) {
	return v.v, nil
}

// number values implement pendingValue
// because there's no explicit value for it
// ( ideally would be space or newline )
type numValue struct{ charmed.NumParser }

// fix? returns float64 because json does
// could also return int64 when its int like
func (v *numValue) FinalizeValue() (ret any, err error) {
	ret, err = v.GetFloat()
	return
}
