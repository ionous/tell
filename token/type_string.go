// Code generated by "stringer -type=Type"; DO NOT EDIT.

package token

import "strconv"

func _() {
	// An "invalid array index" compiler error signifies that the constant values have changed.
	// Re-run the stringer command to generate them again.
	var x [1]struct{}
	_ = x[Invalid-0]
	_ = x[Array-1]
	_ = x[Bool-2]
	_ = x[Comment-3]
	_ = x[Key-4]
	_ = x[Number-5]
	_ = x[String-6]
}

const _Type_name = "InvalidArrayBoolCommentKeyNumberString"

var _Type_index = [...]uint8{0, 7, 12, 16, 23, 26, 32, 38}

func (i Type) String() string {
	if i < 0 || i >= Type(len(_Type_index)-1) {
		return "Type(" + strconv.FormatInt(int64(i), 10) + ")"
	}
	return _Type_name[_Type_index[i]:_Type_index[i+1]]
}
