package main

import (
	"os"
	"time"
)

func maxRun(jctx *jcontext, maxRun int64) {
	if maxRun == 0 {
		return
	}
	tickChan := time.NewTimer(time.Second * time.Duration(maxRun)).C
	<-tickChan
	printSummary(jctx, *pstats)
	os.Exit(0)
}
