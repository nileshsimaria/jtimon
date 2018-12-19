package main

import (
	"os"
	"os/signal"
	"sync"
	"syscall"
)

func signalHandler(cfgFileList string, wMap map[string]*workerCtx, wg *sync.WaitGroup) {
	sigchan := make(chan os.Signal, 10)
	// Handling only Interrupt and SIGHUP signals
	signal.Notify(sigchan, os.Interrupt, syscall.SIGHUP)
	for {
		s := <-sigchan
		switch s {
		case syscall.SIGHUP:
			// Propagate the signal to workers
			// and continue waiting for signals
			if len(cfgFileList) != 0 {
				HandleConfigChanges(&cfgFileList, wMap, wg)
			}
		case os.Interrupt:
			// Send the interrupt to the worker routines and
			// return
			for _, wCtx := range wMap {
				wCtx.signalch <- s
			}
			return
		}
	}
}
