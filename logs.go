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

func logInit(logFile string) {
	if logFile != "" {
		f, err := os.OpenFile(logFile, os.O_CREATE|os.O_RDWR, 0666)
		if err != nil {
			log.Fatalf("Could not create log file(%s): %v\n", logFile, err)
		}
		log.SetOutput(f)
	} else if *print == true {
		log.SetOutput(os.Stdout)
	}
}
