package main

import (
	"io"
	"log"
	"os"
)

func jLog(jctx *JCtx, msg string) {
	jctx.config.Log.Logger.Printf(msg)
}

func logInit(jctx *JCtx) {
	file := jctx.config.Log.File
	var out io.Writer

	if *print {
		out = os.Stdout
		if file != "" {
			log.Println("Both print and log options are used, ignoring log")
		}
	} else if file != "" {
		var err error
		out, err = os.OpenFile(file, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0666)
		if err != nil {
			log.Printf("Could not create log file(%s): %v\n", file, err)
		}
	}

	flags := 0
	if !jctx.config.Log.CSVStats {
		flags = log.LstdFlags
	}

	jctx.config.Log.Logger = log.New(out, "", flags)

	log.Printf("logging in %s for %s:%d [periodic stats every %d seconds]\n",
		jctx.config.Log.File, jctx.config.Host, jctx.config.Port, jctx.config.Log.PeriodicStats)
}
