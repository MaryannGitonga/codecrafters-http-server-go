package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strings"
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

	// read incoming request
	reader := bufio.NewReader(conn)
	requestLine, err := reader.ReadString('\n')

	if err != nil {
		fmt.Println("Error reading request: ", err.Error())
		return
	}

	// parse request line
	requestParts := strings.Fields(requestLine)

	if len(requestParts) < 3 {
		fmt.Println("Invalid request line.")
		return
	}

	path := requestParts[1]

	var response string
	if path == "/" {
		response = "HTTP/1.1 200 OK\r\n\r\n"
	} else {
		response = "HTTP/1.1 404 Not Found\r\n\r\n"
	}

	// write the response
	_, err = conn.Write([]byte(response))

	if err != nil {
		fmt.Println("Error writing response:", err.Error())
		return
	}
}
