// Code generated by "stringer -type=blockStage"; DO NOT EDIT.

package notes

import "strconv"

func _() {
	// An "invalid array index" compiler error signifies that the constant values have changed.
	// Re-run the stringer command to generate them again.
	var x [1]struct{}
	_ = x[startStage-0]
	_ = x[keyStage-1]
	_ = x[valueStage-2]
	_ = x[footerStage-3]
}

const _blockStage_name = "startStagekeyStagevalueStagefooterStage"

var _blockStage_index = [...]uint8{0, 10, 18, 28, 39}

func (i blockStage) String() string {
	if i < 0 || i >= blockStage(len(_blockStage_index)-1) {
		return "blockStage(" + strconv.FormatInt(int64(i), 10) + ")"
	}
	return _blockStage_name[_blockStage_index[i]:_blockStage_index[i+1]]
}
