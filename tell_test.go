package tell

import "testing"

// minimal testing of the simplified Marshal function.
// ( more extensive testing of decode exists in TestFiles and package decode )
func TestUnmarshal(t *testing.T) {
	var b bool
	if e := Unmarshal([]byte(`true`), &b); e != nil {
		t.Fatal(e)
	} else if !b {
		t.Fatal("expected true")
	}
	var m map[string]any
	if e := Unmarshal([]byte(`hello: "there"`), &m); e != nil {
		t.Fatal(e)
	} else if hello, ok := m["hello:"]; !ok {
		t.Fatal("expected key")
	} else if hello != "there" {
		t.Fatal("expected value")
	}
}
