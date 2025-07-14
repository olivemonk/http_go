package main

import (
	"io"
	"log"
	"net"
)

func handleConnection(conn net.Conn) {
	defer conn.Close() // ensure the connection is closed when done
	data, err := io.ReadAll(conn)
	if err != nil {
		log.Printf("Read error: %v", err)
		return
	}
	log.Printf("Received data: %s", data)
	_, err = conn.Write(data) // echo back
	if err != nil {
		log.Printf("Write error: %v", err)
	}
}

func main() {
	listener, err := net.Listen("tcp", ":8080")
	if err != nil {
		log.Fatal("Listen failed:", err)
	}

	defer listener.Close() // ensure the listener is closed when done
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