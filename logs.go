package main

import (
	"log"
	"os"
)

func emitLog(s string) {
	if *logFile != "" || *print == true {
		log.Output(0, s)
	}
}

func logInit(jctx *jcontext, logFile string) {
	if logFile != "" {
		f, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_EXCL, 0666)
		if err != nil {
			log.Fatalf("Could not create log file(%s): %v\n", logFile, err)
		}
		log.SetOutput(f)
		log.SetFlags(0)
		jctx.cfg.Log.LogFileName = logFile
		jctx.cfg.Log.FileHandle = f
	} else if *print == true {
		log.SetOutput(os.Stdout)
	}
}
