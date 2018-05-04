package main

import (
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

func apiInit(jctx *JCtx) {

	if jctx.config.API.Port == 0 {
		return
	}

	portstr := fmt.Sprintf(":%v", jctx.config.API.Port)

	jctx.pause.pch = make(chan int64)
	jctx.pause.upch = make(chan struct{})

	router := mux.NewRouter().StrictSlash(true)
	router.HandleFunc("/api/pause/{pauseTime}", jctx.pauseHandler)
	router.HandleFunc("/api/unpause", jctx.upauseHandler)

	log.Fatal(http.ListenAndServe(portstr, router))
}

func (jctx *JCtx) pauseHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	pauseTime := vars["pauseTime"]
	fmt.Printf("Pause Time: %v\n", pauseTime)

	if pauseV, err := strconv.ParseInt(pauseTime, 10, 64); err != nil {
		fmt.Printf("/api/pause - invalid input\n")
	} else {
		jctx.pause.pch <- pauseV
	}
}

func (jctx *JCtx) upauseHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("Request for unpause\n")
	jctx.pause.upch <- struct{}{}
}
