package main

import (
	"bufio"
	"compress/gzip"
	"flag"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

var filesDirectory string

func main() {
	// Parse command-line arguments
	flag.StringVar(&filesDirectory, "directory", ".", "directory to serve files from")
	flag.Parse()

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

	method := requestParts[0]
	path := requestParts[1]

	var response string

	if method == "GET" {
		if strings.HasPrefix(path, "/echo/") {
			// handle /echo/{string}
			content := strings.TrimPrefix(path, "/echo/")

			var contentEncoding string

			// Read headers
			for {
				line, err := reader.ReadString('\n')
				if err != nil {
					fmt.Println("Error reading header:", err.Error())
					return
				}
				if line == "\r\n" {
					break
				}
				headerParts := strings.SplitN(line, ":", 2)
				if len(headerParts) == 2 && strings.TrimSpace(strings.ToLower(headerParts[0])) == "accept-encoding" {
					encodings := strings.Split(strings.TrimSpace(headerParts[1]), ",")
					for _, encoding := range encodings {
						if strings.TrimSpace(encoding) == "gzip" {
							contentEncoding = "gzip"
							break
						}
					}
				}
			}

			if contentEncoding == "gzip" {
				var gzippedContent []byte
				{
					var b strings.Builder
					gz := gzip.NewWriter(&b)
					_, err := gz.Write([]byte(content))

					if err != nil {
						fmt.Println("Error compressing content:", err.Error())
						return
					}

					err = gz.Close()
					if err != nil {
						fmt.Println("Error closing gzip writer:", err.Error())
						return
					}

					gzippedContent = []byte(b.String())
				}

				contentLength := len(gzippedContent)
				response = fmt.Sprintf(
					"HTTP/1.1 200 OK\r\nContent-Encoding: gzip\r\nContent-Type: text/plain\r\nContent-Length: %d\r\n\r\n%s",
					contentLength, gzippedContent,
				)
			} else {
				contentLength := len(content)
				response = fmt.Sprintf(
					"HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Length: %d\r\n\r\n%s",
					contentLength, content,
				)
			}
		} else if strings.HasPrefix(path, "/files/") {
			// handle /files/{filename}
			fileName := strings.TrimPrefix(path, "/files/")
			filePath := filepath.Join(filesDirectory, fileName)

			// read file
			fileContent, err := os.ReadFile(filePath)

			if err != nil {
				response = "HTTP/1.1 404 Not Found\r\n\r\n"
			} else {
				contentLength := len(fileContent)
				response = fmt.Sprintf(
					"HTTP/1.1 200 OK\r\nContent-Type: application/octet-stream\r\nContent-Length: %d\r\n\r\n%s",
					contentLength, string(fileContent),
				)
			}

		} else if path == "/" {
			response = "HTTP/1.1 200 OK\r\n\r\n"
		} else if path == "/user-agent" {
			// Read headers and look for User-Agent
			var userAgent string
			for {
				line, err := reader.ReadString('\n')
				if err != nil {
					fmt.Println("Error reading header:", err.Error())
					return
				}
				if line == "\r\n" {
					break
				}
				headerParts := strings.SplitN(line, ":", 2)
				if len(headerParts) == 2 && strings.TrimSpace(strings.ToLower(headerParts[0])) == "user-agent" {
					userAgent = strings.TrimSpace(headerParts[1])
				}
			}

			if userAgent != "" {
				contentLength := len(userAgent)
				response = fmt.Sprintf(
					"HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Length: %d\r\n\r\n%s",
					contentLength, userAgent,
				)
			} else {
				response = "HTTP/1.1 400 Bad Request\r\n\r\n"
			}
		} else {
			response = "HTTP/1.1 404 Not Found\r\n\r\n"
		}
	} else if method == "POST" {
		if strings.HasPrefix(path, "/files/") {
			// handle /files/{filename}
			fileName := strings.TrimPrefix(path, "/files/")
			filePath := filepath.Join(filesDirectory, fileName)

			// read headers and look for Content-Length
			var contentLength int
			for {
				line, err := reader.ReadString('\n')
				if err != nil {
					fmt.Println("Error reading header:", err.Error())
					return
				}

				if line == "\r\n" {
					break
				}

				headerParts := strings.SplitN(line, ":", 2)
				if len(headerParts) == 2 && strings.TrimSpace(strings.ToLower(headerParts[0])) == "content-length" {
					contentLength, err = strconv.Atoi(strings.TrimSpace(headerParts[1]))
					if err != nil {
						fmt.Println("Invalid Content-Length header.")
						return
					}
				}
			}

			// read body
			body := make([]byte, contentLength)
			_, err := reader.Read(body)

			if err != nil {
				fmt.Println("Error reading body:", err.Error())
				return
			}

			// write body to file
			err = os.WriteFile(filePath, body, 0644)
			if err != nil {
				fmt.Println("Error writing file:", err.Error())
				return
			}

			response = "HTTP/1.1 201 Created\r\n\r\n"
		}
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
