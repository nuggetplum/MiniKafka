package main

import (
	"log"

	mylog "github.com/nuggetplum/MiniKafka/internal/log" // Renamed to avoid collision with std lib
	"github.com/nuggetplum/MiniKafka/internal/server"
)

func main() {
	// 1. Initialize the registry
	registry, err := mylog.NewRegistry("./prog_log")
	if err != nil {
		log.Fatal(err)
	}
	defer registry.CloseAll()

	// 2. Initialize the Network Server
	httpSrv := server.NewHTTPServer(":8080", registry)

	// 3. Start
	log.Println("Starting server on :8080")
	if err := httpSrv.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}
