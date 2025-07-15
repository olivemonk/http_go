package main

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"fmt"
	"log"
	"net"
	"strconv"
	"strings"
	"time"
)

const (
	frameSettings = 0x4
	frameHeaders  = 0x1
	frameData     = 0x0
)

func handleConnection(conn net.Conn) {
	defer conn.Close()
	reader := bufio.NewReader(conn)
	conn.SetReadDeadline(time.Now().Add(5 * time.Second)) // idle timeout

	for {
		reqLine, err := reader.ReadString('\n')
		if err != nil {
			return
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

		// upgrade to HTTP/2 if requested
		if strings.EqualFold(headers["Upgrade"], "h2c") && strings.Contains(headers["Connection"], "Upgrade") {
			conn.Write([]byte("HTTP/1.1 101 Switching Protocols\r\nConnection: Upgrade\r\nUpgrade: h2c\r\n\r\n"))
			handleHTTP2(conn, reader, method, path)
			return
		}

		// read body if POST
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

		// handle request
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

		connection := "keep-alive"
		if strings.EqualFold(headers["Connection"], "close") {
			connection = "close"
		}

		response := fmt.Sprintf("HTTP/1.1 %s\r\nContent-Length: %d\r\nConnection: %s\r\n\r\n%s",
			status, len(respBody), connection, respBody)
		conn.Write([]byte(response))

		if connection == "close" {
			log.Println("Closing connection as per request")
			return
		}
		conn.SetReadDeadline(time.Now().Add(5 * time.Second))
	}
}

func handleHTTP2(conn net.Conn, reader *bufio.Reader, method, path string) {
	// expect client preface
	preface := make([]byte, 24)
	_, err := reader.Read(preface)
	if err != nil || !bytes.Equal(preface, []byte("PRI * HTTP/2.0\r\n\r\nSM\r\n\r\n")) {
		log.Println("Invalid HTTP/2 preface")
		return
	}

	log.Println("HTTP/2 upgrade successful")

	// send SETTINGS frame (empty payload)
	sendFrame(conn, frameSettings, 0x0, 0, []byte{})

	// for demo, skip real HEADERS parsing. send fixed HEADERS and DATA
	headersPayload := []byte{0x88} // :status: 200 (compressed with static table)
	sendFrame(conn, frameHeaders, 0x5, 1, headersPayload) // END_HEADERS | END_STREAM

	dataPayload := []byte("Hello, HTTP/2!")
	sendFrame(conn, frameData, 0x1, 1, dataPayload) // END_STREAM

	// no further multiplexing for simplicity
}

func sendFrame(conn net.Conn, typ uint8, flags uint8, streamID uint32, payload []byte) {
	length := len(payload)
	frame := make([]byte, 9+length)
	frame[0] = byte((length >> 16) & 0xff)
	frame[1] = byte((length >> 8) & 0xff)
	frame[2] = byte(length & 0xff)
	frame[3] = typ
	frame[4] = flags
	binary.BigEndian.PutUint32(frame[5:9], streamID&0x7fffffff) // clear reserved bit
	copy(frame[9:], payload)
	conn.Write(frame)
}

func main() {
	listener, err := net.Listen("tcp", ":8080")
	if err != nil {
		log.Fatal("Listen failed:", err)
	}
	defer listener.Close()
	log.Println("HTTP server (1.1 + h2c) listening on :8080")

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("Accept error: %v", err)
			continue
		}
		go handleConnection(conn)
		log.Println("Accepted connection from", conn.RemoteAddr())
	}
}
