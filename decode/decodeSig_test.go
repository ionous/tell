package tell_test

import (
	"fmt"
	"testing"

	"github.com/ionous/tell"
	"github.com/ionous/tell/charm"
)

func TestSig(t *testing.T) {
	// returns point of failure
	test := func(str string) (ret string, err error) {
		var sig tell.Signature
		if e := charm.Parse(str, &sig); e != nil {
			err = e
		} else if str, e := sig.GetKey(); e != nil {
			err = e
		} else {
			ret = str
		}
		return
	}
	fails := func(str string) (err error) {
		if v, e := test(str); e != nil {
			t.Log("ok failure:", str, e)
		} else {
			err = fmt.Errorf("%s expected error %v", str, v)
		}
		return
	}
	succeeds := func(str string) (err error) {
		if res, e := test(str); e != nil {
			err = fmt.Errorf("%w for: %q", e, str)
		} else if str != res {
			err = fmt.Errorf("%q unexpected result %v", str, res)
		} else {
			t.Log("ok success:", str)
		}
		return
	}
	if e := fails("a"); e != nil {
		t.Fatal(e)
	}
	if e := fails(" a"); e != nil {
		t.Fatal(e)
	}
	if e := fails("b "); e != nil {
		t.Fatal(e)
	}
	if e := fails("1a"); e != nil {
		t.Fatal(e)
	}
	if e := succeeds("a:"); e != nil {
		t.Fatal(e)
	}
	if e := succeeds("a:b:c:"); e != nil {
		t.Fatal(e)
	}
	if e := succeeds("and:more complex:keys_like_this:"); e != nil {
		t.Fatal(e)
	}
	if e := fails("a:b::c:"); e != nil {
		t.Fatal(e)
	}
}
