// Code generated by "stringer -type=boolValue -linecomment"; DO NOT EDIT.

package token

import "strconv"

func _() {
	// An "invalid array index" compiler error signifies that the constant values have changed.
	// Re-run the stringer command to generate them again.
	var x [1]struct{}
	_ = x[boolInvalid-0]
	_ = x[boolFalse-1]
	_ = x[boolTrue-2]
}

const _boolValue_name = "boolInvalidfalsetrue"

var _boolValue_index = [...]uint8{0, 11, 16, 20}

func (i boolValue) String() string {
	if i < 0 || i >= boolValue(len(_boolValue_index)-1) {
		return "boolValue(" + strconv.FormatInt(int64(i), 10) + ")"
	}
	return _boolValue_name[_boolValue_index[i]:_boolValue_index[i+1]]
}
