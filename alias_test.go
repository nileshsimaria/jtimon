package main

import (
	"testing"
)

func TestNewAlias(t *testing.T) {
	tests := []struct {
		name  string
		file  string
		err   bool
		count int
		m     map[string]string
	}{
		{
			name: "invalid-file",
			file: "no-such-file.txt",
			err:  true,
		},
		{
			name: "valid-file-syntax error",
			file: "tests/data/juniper-junos/alias/alias-error.txt",
			err:  true,
		},
		{
			name:  "valid-file",
			file:  "tests/data/juniper-junos/alias/alias.txt",
			err:   false,
			count: 10,
			m: map[string]string{
				"no-does-not-exists-in-test-data":                                                 "no-does-not-exists-in-test-data",
				"/interfaces/interface/subinterfaces/subinterface/state/counters/in-unicast-pkts": "ifl-in-ucast-pkts",
				"/interfaces/interface/@name":                                                     "physical_interface",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			alias, err := NewAlias(test.file)
			if test.err && err == nil {
				t.Errorf("want error, got nil")
			}
			if !test.err && err != nil {
				t.Errorf("%v", err)
			}
			if !test.err {
				if test.count != len(alias.m) {
					t.Errorf("count: want %d, got %d", test.count, len(alias.m))
				}
				for k, v := range test.m {
					r := getAlias(alias, k)
					if r != v {
						t.Errorf("want %s, got %s", v, r)
					}
				}

				r := getAlias(nil, "empty-alias")
				if r != "empty-alias" {
					t.Errorf("nil-alias failed: want empty-alias, got %s", r)
				}
			}
		})
	}
}
