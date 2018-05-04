package main

import (
	"fmt"
	"log"
	"os"
)

/*
	Log it, each go routine collecting JTI data has its own log file but user can
	specify common log file to log for all devices so we need protection.
	We also need protection for printing on terminal. Caller can control it using
	boolean safe argument.
*/
func l(safe bool, jctx *JCtx, s string) {
	if *print {
		switch safe {
		case true:
			gmutex.Lock()
			fmt.Printf(s)
			gmutex.Unlock()
		case false:
			fmt.Printf(s)
		}
	} else if jctx.config.Log.loger != nil {
		switch safe {
		case true:
			gmutex.Lock()
			jctx.config.Log.loger.Printf(s)
			gmutex.Unlock()

		case false:
			jctx.config.Log.loger.Printf(s)
		}
	}
}

func logInit(jctx *JCtx) {
	file := jctx.config.Log.File
	if file != "" {
		if *print {
			fmt.Println("Both print and log options are specified, ignoring log")
		} else {
			f, err := os.OpenFile(file, os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0666)
			if err != nil {
				fmt.Printf("Could not create log file(%s): %v\n", file, err)
			} else {
				flags := 0
				if !jctx.config.Log.CSVStats {
					flags = log.LstdFlags
				}
				jctx.config.Log.loger = log.New(f, "", flags)
				jctx.config.Log.handle = f
				fmt.Printf("logging in %s for %s:%d [periodic stats every %d seconds]\n",
					jctx.config.Log.File, jctx.config.Host, jctx.config.Port, jctx.config.Log.PeriodicStats)
			}
		}
	}
}
