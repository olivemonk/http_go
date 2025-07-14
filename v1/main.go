package main

import (
	"log"
	"net"
	"bufio"
	"fmt"
	"strings"
)

func handleConnection(conn net.Conn) {
	defer conn.Close() // ensure the connection is closed when done

	reader := bufio.NewReader(conn)

	// read req line
	reqLine, err := reader.ReadString('\n')
	if err != nil {
		log.Printf("Read error: %v", err)
		return
	}
	reqLine = strings.TrimSpace(reqLine)
	parts := strings.Split(reqLine, " ")
	if len(parts) != 3 || parts[0] != "GET" || parts[1] != "/" || parts[2] != "HTTP/1.0" {
		conn.Write([]byte("HTTP/1.0 400 Bad Request\r\n\r\nInvalid request"))
		return
	}
	
	// read headers 
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			log.Printf("Header read error: %v", err)
			return
		}
		line = strings.TrimSpace(line)
		if line == "" { // end of headers
			break
		}
	}

	// construct and send the response
	body := "Hello, World!\n"
	response := fmt.Sprintf("HTTP/1.0 200 OK\r\nContent-Length: %d\r\n\r\n%s", len(body), body)
	_, err = conn.Write([]byte(response))
	if err != nil {
		log.Printf("Write error: %v", err)
		return
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