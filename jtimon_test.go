package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"testing"

	js "github.com/nileshsimaria/jtisim"

	flag "github.com/spf13/pflag"
)

func TestMain(m *testing.M) {
	flag.Parse()

	go func() {
		host := "127.0.0.1"
		port := 50051

		jtisim := js.NewJTISim(host, int32(port), false, "tests/simulator/desc")
		if err := jtisim.Start(); err != nil {
			log.Printf("can not start jti simulator: %v", err)
		}
	}()

	os.Exit(m.Run())
}

func compareResults(jctx *JCtx) error {
	testRes := jctx.testRes

	if testRes == nil {
		return fmt.Errorf("testRes is nil")
	}
	ret, err := testRes.Seek(0, 0)
	if ret != 0 {
		return fmt.Errorf("can not seek testRes")
	}
	if err != nil {
		return err
	}

	testExp, err := os.Open(jctx.file + ".testexp")
	if err != nil {
		return err
	}
	defer testExp.Close()

	r, err := ioutil.ReadAll(testRes)
	if err != nil {
		return err
	}
	e, err := ioutil.ReadAll(testExp)
	if err != nil {
		return err
	}

	if !bytes.Equal(r, e) {
		return fmt.Errorf("test failed: content of %s does not match with content of %s",
			jctx.file+".testres", jctx.file+".testexp")
	}

	return nil
}
