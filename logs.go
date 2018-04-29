package main

import (
	"fmt"
	"log"
	"os"
)

func emitLog(jctx *JCtx, s string) {
	if *print {
		fmt.Printf(s)
	} else if jctx.cfg.Log.Logger != nil {
		jctx.cfg.Log.Logger.Printf(s)
	}
}

func logInit(jctx *JCtx) {
	file := jctx.cfg.Log.File
	if file != "" {
		f, err := os.OpenFile(file, os.O_CREATE|os.O_WRONLY|os.O_EXCL, 0666)
		if err != nil {
			fmt.Printf("Could not create log file(%s): %v\n", file, err)
		} else {
			jctx.cfg.Log.Logger = log.New(f, "", log.LstdFlags|log.Lshortfile)
			jctx.cfg.Log.FileHandle = f
			fmt.Printf("\nlogging Telemetry data from %s:%d in file %s\n", jctx.cfg.Host, jctx.cfg.Port, jctx.cfg.Log.File)
		}
	} else if *print == true {
		log.SetOutput(os.Stdout)
	}
}
