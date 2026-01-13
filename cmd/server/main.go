package main

import (
	"log"

	mylog "github.com/nuggetplum/MiniKafka/internal/log" // Renamed to avoid collision with std lib
	"github.com/nuggetplum/MiniKafka/internal/server"
)

func main() {
	// 1. Initialize the Storage
	srv, err := mylog.NewStore("log_store.bin")
	if err != nil {
		log.Fatal(err)
	}
	// Best practice: Close the file when main exits
	defer srv.Close()

	// 2. Initialize the Network Server
	httpSrv := server.NewHTTPServer(":8080", srv)

	// 3. Start
	log.Println("Starting server on :8080")
	if err := httpSrv.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}
