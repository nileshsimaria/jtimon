package main

import (
	"os"
	"time"
)

func maxRunHandler(maxRunTime int64, wMap map[string]*workerCtx) {
	if maxRunTime > 0 {
		// mr - Max run time in seconds
		// Subscription is configured for a certain time period
		// Once the time expires, interrupt worker go routines.
		tickChan := time.NewTimer(time.Second * time.Duration(maxRunTime)).C
		<-tickChan
		for _, worker := range wMap {
			worker.signalch <- os.Interrupt
		}
	}
}
