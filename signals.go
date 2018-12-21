package main

import (
	"os"
	"os/signal"
	"sync"
	"syscall"
)

func signalHandler(configFileList string, wMap map[string]*workerCtx, wg *sync.WaitGroup) {
	sigchan := make(chan os.Signal, 10)
	// handle interrupt and sighup
	signal.Notify(sigchan, os.Interrupt, syscall.SIGHUP)
	for {
		s := <-sigchan
		switch s {
		case syscall.SIGHUP:
			// propagate the signal to workers and continue waiting for signals
			if len(configFileList) != 0 {
				handleConfigChanges(&configFileList, wMap, wg)
			}
		case os.Interrupt:
			// send the interrupt to the worker routines and return
			for _, wCtx := range wMap {
				wCtx.signalch <- s
			}
			return
		}
	}
}
