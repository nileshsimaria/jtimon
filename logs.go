package main

import (
	"log"
	"os"
)

func emitLog(s string) {
	if *logFile != "" || *print == true {
		log.Output(1, s)
	}
}

func logInit(jctx *jcontext, logFile string) {
	if logFile != "" {
		f, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_EXCL, 0666)
		if err != nil {
			log.Fatalf("Could not create log file(%s): %v\n", logFile, err)
		}
		log.SetOutput(f)
		jctx.cfg.Log.LogFileName = logFile
	} else if *print == true {
		log.SetOutput(os.Stdout)
	}
}
