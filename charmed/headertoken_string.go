// Code generated by "stringer -type=headerToken"; DO NOT EDIT.

package charmed

import "strconv"

func _() {
	// An "invalid array index" compiler error signifies that the constant values have changed.
	// Re-run the stringer command to generate them again.
	var x [1]struct{}
	_ = x[headerSpaces-0]
	_ = x[headerWord-1]
	_ = x[headerRedirect-2]
}

const _headerToken_name = "headerSpacesheaderWordheaderRedirect"

var _headerToken_index = [...]uint8{0, 12, 22, 36}

func (i headerToken) String() string {
	if i < 0 || i >= headerToken(len(_headerToken_index)-1) {
		return "headerToken(" + strconv.FormatInt(int64(i), 10) + ")"
	}
	return _headerToken_name[_headerToken_index[i]:_headerToken_index[i+1]]
}
