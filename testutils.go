package main

import (
	"log"
	"os"
)

func testSetup(jctx *JCtx) {
	file := jctx.file
	if *genTestData {
		var errf error
		jctx.testMeta, errf = os.Create(file + ".testmeta")
		if errf != nil {
			log.Printf("Could not create 'testmeta' for file %s\n", file+".testmeta")
		}

		jctx.testBytes, errf = os.Create(file + ".testbytes")
		if errf != nil {
			log.Printf("Could not create 'testbytes' for file %s\n", file+".testbytes")
		}

		jctx.testExp, errf = os.Create(file + ".testexp")
		if errf != nil {
			log.Printf("Could not create 'testexp' for file %s\n", file+".testexp")
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
