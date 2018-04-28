package main

import (
	"time"
)

func maxRun(jctx *JCtx, maxRun int64) {
	if maxRun == 0 {
		return
	}
	tickChan := time.NewTimer(time.Second * time.Duration(maxRun)).C
	<-tickChan
	printSummary(jctx, *pstats)
	jctx.wg.Done()
}
