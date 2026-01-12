package main

import (
	"fmt"
	"net"
)

func main() {

	ln, _ := net.Listen("tcp", ":3000")
	fmt.Println("Broker is listening on :3000")

	for {
		conn, _ := ln.Accept()
		go handleConnection(conn) // Concurrency in action
	}
}
