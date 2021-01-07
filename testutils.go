package main

import (
	"fmt"
	"log"
	"os"
	"sort"
	"strings"
	"unicode"
)

type testDataType int

const (
	// GENTESTEXPDATA to generate expected test data
	GENTESTEXPDATA = iota
	// GENTESTRESDATA to generate result test data
	GENTESTRESDATA
)

func testSetup(jctx *JCtx) error {
	file := jctx.file
	if *genTestData {
		var err error
		if jctx.testMeta, err = os.Create(file + ".testmeta"); err != nil {
			return err
		}
		if jctx.testBytes, err = os.Create(file + ".testbytes"); err != nil {
			return err
		}
		if jctx.testExp, err = os.Create(file + ".testexp"); err != nil {
			return err
		}
	}
	return nil
}

func testTearDown(jctx *JCtx) {
	if *genTestData {
		if jctx.testMeta != nil {
			jctx.testMeta.Close()
			jctx.testMeta = nil
		}
		if jctx.testBytes != nil {
			jctx.testBytes.Close()
			jctx.testBytes = nil
		}
		if jctx.testExp != nil {
			jctx.testExp.Close()
			jctx.testExp = nil
		}
	}
}

func generateTestData(jctx *JCtx, data []byte) {
	log.Printf("data len = %d", len(data))
	if jctx.testMeta != nil {
		d := fmt.Sprintf("%d:", len(data))
		jctx.testMeta.WriteString(d)
	}
	if jctx.testBytes != nil {
		jctx.testBytes.Write(data)
	}
}

func testDataPoints(jctx *JCtx, testType testDataType, tags map[string]string, fields map[string]interface{}) {
	var f *os.File

	switch testType {
	case GENTESTEXPDATA:
		if jctx.testExp == nil {
			return
		}
		f = jctx.testExp
	case GENTESTRESDATA:
		if jctx.testRes == nil {
			return
		}
		f = jctx.testRes
	default:
		return
	}

	f.WriteString("TAGS: [")
	var keys []string
	for k := range tags {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		v := tags[k]
		f.WriteString(fmt.Sprintf(" %s=%s ", k, v))
	}
	f.WriteString("]\n")

	f.WriteString("FIELDS: [")

	keys = nil
	for k := range fields {
		// Skip the fields that need not be compared
		if k == gDeviceTs {
			continue
		}
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		v := fields[k]
		f.WriteString(fmt.Sprintf("%s=%s", k, v))
	}
	f.WriteString("]\n")
	f.Sync()
}

func compareString(a string, b string) bool {
	filter := func(r rune) rune {
		if unicode.IsSpace(r) {
			return -1
		}
		return r
	}
	if strings.Compare(strings.Map(filter, a), strings.Map(filter, b)) == 0 {
		return true
	}
	return false
}
