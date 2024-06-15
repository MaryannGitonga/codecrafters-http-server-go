package main

import (
	"fmt"
	"net"
	"os"
)

func main() {
	l, err := net.Listen("tcp", "0.0.0.0:4221")
	if err != nil {
		fmt.Println("Failed to bind to port 4221")
		os.Exit(1)
	}

	defer l.Close()

	fmt.Println("Server is listening on port 4221")

	for {
		// accept an incoming connection
		conn, err := l.Accept()
		if err != nil {
			fmt.Println("Error accepting connection:", err.Error())
			continue
		}

		// goroutine to handle connection
		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	defer conn.Close()

	response := "HTTP/1.1 200 OK\r\n\r\n"
	_, err := conn.Write([]byte(response))

	if err != nil {
		fmt.Println("Error writing response:", err.Error())
		return
	}
}
