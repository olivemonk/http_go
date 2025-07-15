package main

import (
	"bytes"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"testing"
	"time"
)

func startTestServer() net.Listener {
	ln, err := net.Listen("tcp", ":8080")
	if err != nil {
		panic(err)
	}
	go func() {
		for {
			conn, err := ln.Accept()
			if err != nil {
				continue
			}
			go handleConnection(conn)
		}
	}()
	return ln
}

func TestHTTP11GET(t *testing.T) {
	ln := startTestServer()
	defer ln.Close()

	resp, err := http.Get("http://localhost:8080/")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		t.Errorf("Expected 200, got %d", resp.StatusCode)
	}
	body, _ := io.ReadAll(resp.Body)
	if string(body) != "Hello, World!" {
		t.Errorf("Unexpected body: %s", body)
	}
}

func TestHTTP11POST(t *testing.T) {
	resp, err := http.Post("http://localhost:8080/", "text/plain", strings.NewReader("foo=bar"))
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if !strings.Contains(string(body), "Received") {
		t.Errorf("Unexpected POST response: %s", body)
	}
}

func TestHTTP11ConnectionClose(t *testing.T) {
	conn, err := net.Dial("tcp", "localhost:8080")
	if err != nil {
		t.Fatal(err)
	}
	fmt.Fprintf(conn, "GET / HTTP/1.1\r\nHost: localhost\r\nConnection: close\r\n\r\n")
	resp := make([]byte, 1024)
	n, err := conn.Read(resp)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Contains(resp[:n], []byte("Connection: close")) {
		t.Error("Expected Connection: close in response")
	}
}

func TestMalformedRequest(t *testing.T) {
	conn, err := net.Dial("tcp", "localhost:8080")
	if err != nil {
		t.Fatal(err)
	}
	fmt.Fprintf(conn, "BAD REQUEST\r\n\r\n")
	resp := make([]byte, 1024)
	n, _ := conn.Read(resp)
	if !bytes.Contains(resp[:n], []byte("400 Bad Request")) {
		t.Error("Expected 400 Bad Request")
	}
}

func TestTimeout(t *testing.T) {
	conn, err := net.Dial("tcp", "localhost:8080")
	if err != nil {
		t.Fatal(err)
	}
	time.Sleep(6 * time.Second)
	_, err = conn.Write([]byte("GET / HTTP/1.1\r\nHost: localhost\r\n\r\n"))
	if err == nil {
		t.Error("Expected connection timeout to close socket")
	}
}

func TestH2CUpgrade(t *testing.T) {
	conn, err := net.Dial("tcp", "localhost:8080")
	if err != nil {
		t.Fatal(err)
	}
	fmt.Fprintf(conn, "GET / HTTP/1.1\r\nHost: localhost\r\nConnection: Upgrade\r\nUpgrade: h2c\r\n\r\n")
	resp := make([]byte, 1024)
	n, err := conn.Read(resp)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Contains(resp[:n], []byte("101 Switching Protocols")) {
		t.Errorf("Expected 101 Switching Protocols, got: %s", resp[:n])
	}
}

// ----------------------
// Basic Stress Testing
// ----------------------
func BenchmarkGETRequests(b *testing.B) {
	for i := 0; i < b.N; i++ {
		http.Get("http://localhost:8080/")
	}
}
