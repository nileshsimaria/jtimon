package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"testing"
)

func TestCompareStrings(t *testing.T) {
	tests := []struct {
		name string
		a    string
		b    string
		res  bool
	}{
		{
			name: "space-equal",
			a:    "foo",
			b:    "  foo",
			res:  true,
		},
		{
			name: "space-not-equal",
			a:    "foo",
			b:    "  foox",
			res:  false,
		},
		{
			name: "multiline-space-equal",
			a: `	openconfig-bgp:bgp
						neighbors
							neighbor	
								neighbor-	address[key]
								afi-safis
									afi-safi
										afi-safi-name[key]`,
			b: `openconfig-bgp:bgp
			neighbors
				neighbor
					neighbor-address[key]
					afi-safis
						afi-safi
							afi-safi-name[key]`,
			res: true,
		},
		{
			name: "multiline-space-not-equal",
			a: `	openconfig-bgp:bgp
						neighbors
							neighbor
								neighbor-address
								afi-safis
									afi-safi
										afi-safi-name[key]`,
			b: `openconfig-bgp:bgp
			neighbors
				neighbor
					neighbor-address[key]
					afi-safis
						afi-safi
							afi-safi-name[key]`,
			res: false,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			res := compareString(test.a, test.b)
			if res != test.res {
				t.Errorf("want %v, got %v", test.res, res)
			}
		})
	}
}

func TestTestUtils(t *testing.T) {
	*genTestData = true
	jctx := &JCtx{
		file: "tests/data/empty.json",
	}

	if err := testSetup(jctx); err != nil {
		t.Errorf("%v", err)
	} else {
		data := "this is test data"
		b := []byte(data)
		generateTestData(jctx, b)

		ret, err := jctx.testMeta.Seek(0, 0)
		if ret != 0 {
			t.Errorf("can not seek testMeta")
		}
		if err != nil {
			t.Errorf("%v", err)
		}
		r, err := ioutil.ReadAll(jctx.testMeta)
		if err != nil {
			t.Errorf("%v", err)
		}
		expStr := fmt.Sprintf("%d:", len(string(b)))
		if expStr != string(r) {
			t.Errorf("testMeta: want (%s), got (%s)", expStr, string(r))
		}

		ret, err = jctx.testBytes.Seek(0, 0)
		if ret != 0 {
			t.Errorf("can not seek testBytes")
		}
		if err != nil {
			t.Errorf("%v", err)
		}

		r, err = ioutil.ReadAll(jctx.testBytes)
		if err != nil {
			t.Errorf("%v", err)
		}

		if len(r) != len(b) {
			t.Errorf("testBytes: data len mismatch, want (%d), got (%d)", len(b), len(r))
		}

		if !bytes.Equal(r, b) {
			t.Errorf("testBytes: data content mismatch, want (%v), got (%v)", b, r)
		}

		testTearDown(jctx)
		if jctx.testMeta != nil {
			t.Errorf("got non null jctx.testMeta, expected it to be null after tear down call")
		}
		if jctx.testBytes != nil {
			t.Errorf("got non null jctx.testBytes, expected it to be null after tear down call")
		}
		if jctx.testExp != nil {
			t.Errorf("got non null jctx.testExp, expected it to be null after tear down call")
		}
	}
	*genTestData = false
}
