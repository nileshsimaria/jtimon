package main

import (
	"fmt"
	"log"
	"os"
)

func jLog(jctx *JCtx, msg string) {
	if *logMux {
		log.Print(fmt.Sprintf("[%s]:%s", jctx.config.Host, msg))
		return
	}

	if jctx.config.Log.logger != nil {
		jctx.config.Log.logger.Printf(msg)
	}
}

func logStop(jctx *JCtx) {
	if jctx.config.Log.out != nil {
		if jctx.config.Log.out != os.Stdout {
			jctx.config.Log.out.Close()
		}
		jctx.config.Log.out = nil
		jctx.config.Log.logger = nil
	}
}
func logInit(jctx *JCtx) {
	if *logMux {
		return
	}

	file := jctx.config.Log.File
	var out *os.File

	if *print {
		out = os.Stdout
		if file != "" {
			log.Println("Both print and log options are used, ignoring log")
		}
	} else if file != "" {
		var err error
		out, err = os.OpenFile(file, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0600)
		if err != nil {
			log.Printf("Could not create log file(%s): %v\n", file, err)
		}
	}

	if out != nil {
		flags := 0

		jctx.config.Log.logger = log.New(out, "", flags)
		jctx.config.Log.out = out

		log.Printf("logging in %s for %s:%d [periodic stats every %d seconds]\n",
			jctx.config.Log.File, jctx.config.Host, jctx.config.Port, jctx.config.Log.PeriodicStats)
	}
}
