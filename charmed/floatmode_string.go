// Code generated by "stringer -type=FloatMode"; DO NOT EDIT.

package charmed

import "strconv"

func _() {
	// An "invalid array index" compiler error signifies that the constant values have changed.
	// Re-run the stringer command to generate them again.
	var x [1]struct{}
	_ = x[Pending-0]
	_ = x[Int10-1]
	_ = x[Int16-2]
	_ = x[Float64-3]
}

const _FloatMode_name = "PendingInt10Int16Float64"

var _FloatMode_index = [...]uint8{0, 7, 12, 17, 24}

func (i FloatMode) String() string {
	if i < 0 || i >= FloatMode(len(_FloatMode_index)-1) {
		return "FloatMode(" + strconv.FormatInt(int64(i), 10) + ")"
	}
	return _FloatMode_name[_FloatMode_index[i]:_FloatMode_index[i+1]]
}
