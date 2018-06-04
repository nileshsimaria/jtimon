package main

import (
	"fmt"
	"log"
	"net"
	"net/http"

	"google.golang.org/grpc"
)

func startGtrace(gtrace bool) {
	if gtrace {
		/*
		 * Turn on grpc trace - for the same turn on http server to
		 * serve /debug/requests and /debug/events.
		 */
		grpc.EnableTracing = true
		go func() {
			lis, err := net.Listen("tcp", ":0")
			if err != nil {
				log.Fatalf("Failed to listen: %v", err)
			}
			fmt.Println("Client profiling address: ", lis.Addr().String())
			if err := http.Serve(lis, nil); err != nil {
				log.Fatalf("Failed to serve: %v", err)
			}
		}()
	}

}
