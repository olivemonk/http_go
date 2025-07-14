package main

import (
	"log"
	"net"
	"bufio"
	"fmt"
	"strings"
	"strconv"
	"time"
)

func handleConnection(conn net.Conn) {
	defer conn.Close() // ensure the connection is closed when done

	reader := bufio.NewReader(conn)
	conn.SetReadDeadline(time.Now().Add(5 * time.Second)) // set a read deadline, Idle timeout

	for {
		reqLine, err := reader.ReadString('\n')
		if err != nil {
			return // close on error or timeout
		}

		reqLine = strings.TrimSpace(reqLine)
		parts := strings.Split(reqLine, " ")
		if len(parts) != 3 || parts[2] != "HTTP/1.1" {
			conn.Write([]byte("HTTP/1.1 400 Bad Request\r\n\r\nInvalid request"))
			return
		}
		method, path := parts[0], parts[1]

		// read headers
		headers := make(map[string]string)
		for {
			line, err := reader.ReadString('\n')
			if err != nil {
				return
			}
			line = strings.TrimSpace(line)
			if line == "" {
				break
			}
			hParts := strings.SplitN(line, ":", 2)
			if len(hParts) == 2 {
				headers[strings.TrimSpace(hParts[0])] = strings.TrimSpace(hParts[1])
			}
		}

		// read body
		body := ""
		if lengthStr, ok := headers["Content-Length"]; ok && method == "POST" {
			length, err := strconv.Atoi(lengthStr)
			if err != nil {
				conn.Write([]byte("HTTP/1.1 400 Bad Request\r\n\r\nInvalid Content-Length"))
				return
			}
			bodyBytes := make([]byte, length)
			_, err = reader.Read(bodyBytes)
			if err != nil {
				return
			}
			body = string(bodyBytes)
		}

		// handle by method paths
		respBody := ""
		status := "200 OK"

		if method == "GET" && path == "/" {
			respBody = "Hello, World!"
		} else if method == "POST" && path == "/" {
			respBody = fmt.Sprintf("Received: %s", body)
		} else {
			status = "404 Not Found"
			respBody = "Not Found"
		}

		// construct response
		connection := "keep-alive"
		if headers["Connection"] == "close" {
			connection = "close"
		}

		response := fmt.Sprintf("HTTP/1.1 %s\r\nContent-Length: %d\r\nConnection: %s\r\n\r\n%s", status, len(respBody), connection, respBody)
		conn.Write([]byte(response))
		if connection == "close" {
			log.Println("Closing connection as per request")
			return // close the connection if requested
		}
		conn.SetReadDeadline(time.Now().Add(5 * time.Second)) // reset read deadline for keep-alive
	}
}

func main() {
	listener, err := net.Listen("tcp", ":8080")
	if err != nil {
		log.Fatal("Listen failed:", err)
	}

	defer listener.Close() // ensure the listener is closed when done
	log.Println("HTTP/1.0 server listening on :8080")

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("Accept error: %v", err)
			continue // continue to accept next connection
		}
		go handleConnection(conn) // handle the connection in a goroutine
		log.Println("Accepted new connection")
		log.Printf("Connection from %s", conn.RemoteAddr())
	}
}