package main

import (
	"fmt"
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

func testSetup(jctx *JCtx) {
	file := jctx.file
	if *genTestData {
		var errf error
		jctx.testMeta, errf = os.Create(file + ".testmeta")
		if errf != nil {
			jLog(jctx, fmt.Sprintf("Could not create 'testmeta' for file %s\n", file+".testmeta"))
		}

		jctx.testBytes, errf = os.Create(file + ".testbytes")
		if errf != nil {
			jLog(jctx, fmt.Sprintf("Could not create 'testbytes' for file %s\n", file+".testbytes"))
		}

		jctx.testExp, errf = os.Create(file + ".testexp")
		if errf != nil {
			jLog(jctx, fmt.Sprintf("Could not create 'testexp' for file %s\n", file+".testexp"))
		}
	}
}

func testTearDown(jctx *JCtx) {
	if *genTestData {
		if jctx.testMeta != nil {
			jctx.testMeta.Close()
		}
		if jctx.testBytes != nil {
			jctx.testBytes.Close()
		}
		if jctx.testExp != nil {
			jctx.testExp.Close()
		}
	}
}

func generateTestData(jctx *JCtx, data []byte) {
	fmt.Printf("data len = %d\n", len(data))
	if jctx.testMeta != nil {
		dat := fmt.Sprintf("%d:", len(data))
		jctx.testMeta.WriteString(dat)
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
