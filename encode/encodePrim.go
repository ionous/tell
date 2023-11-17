package encode

import (
	r "reflect"
	"strconv"
)

func formatBool(v r.Value) string {
	return strconv.FormatBool(v.Bool())
}

func formatInt(v r.Value) string {
	ofs := v.Kind() - intKind
	width := intBase[ofs]
	prefix := intPrefix[ofs]
	return prefix + strconv.FormatInt(v.Int(), width)
}

func formatUint(v r.Value) string {
	ofs := v.Kind() - uintKind
	width := intBase[ofs]
	prefix := intPrefix[ofs]
	return prefix + strconv.FormatUint(v.Uint(), width)
}

func formatFloat(v r.Value) string {
	width := floatWidth[v.Kind()-floatKind]
	return strconv.FormatFloat(v.Float(), 'g', -1, width)
	// fix: handle infinity, etc?
}

var intKind = r.Int // Int, Int8,	Int16, Int32,	Int64
var uintKind = r.Uint
var intBase = []int{10, 16, 16, 10, 10}
var intPrefix = []string{"", "0x", "0x", "", ""}

var floatKind = r.Float32
var floatWidth = []int{32, 64}
