package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"testing"

	js "github.com/nileshsimaria/jtisim"
	lps "github.com/nileshsimaria/lpserver"

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

	go func() {
		s := lps.NewLPServer("127.0.0.1", 50052)
		s.StartServer()
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

type storeType int

const (
	STOREOPEN = iota
	STORECLOSE
)

func influxStore(host string, port int, op storeType, name string) error {
	url := ""
	switch op {
	case STOREOPEN:
		url = fmt.Sprintf("http://%s:%d/store/open", host, port)
	case STORECLOSE:
		url = fmt.Sprintf("http://%s:%d/store/close", host, port)
	}

	s := fmt.Sprintf(`{"name":"%s"}`, name)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer([]byte(s)))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	ioutil.ReadAll(resp.Body)
	return nil
}

func prometheusCollect(host string, port int, jctx *JCtx) error {
	var f *os.File
	var err error

	if f, err = os.Create(jctx.file + ".testres"); err != nil {
		return err
	}
	defer f.Close()

	url := fmt.Sprintf("http://%s:%d/metrics", host, port)

	var resp *http.Response
	if resp, err = http.Get(url); err != nil {
		return err
	}
	defer resp.Body.Close()

	var body []byte
	if body, err = ioutil.ReadAll(resp.Body); err != nil {
		return err
	}

	strs := strings.Split(string(body), "\n")
	for _, str := range strs {
		if strings.HasPrefix(str, "_interfaces_interface") {
			f.WriteString(str + "\n")
		}
	}

	return nil
}
