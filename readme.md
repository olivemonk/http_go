# Basic HTTP Server in Go

This project demonstrates simple HTTP server implementations in Go for different HTTP versions, implemented from scratch using raw `net.Conn` where applicable.

## Branches

- `main` – HTTP/2 server with h2c (cleartext upgrade from HTTP/1.1)
- `http/1` – Basic HTTP/1.0 using raw TCP and manual parsing
- `http/1.1` – HTTP/1.1 with keep-alive, headers, body parsing, and `Connection: close`
- `http/2` – HTTP/2 with focus on h2c upgrade and minimal HTTP/2 frame handling

## Run

```bash
go run main.go

```

## Tests

This project includes a test suite covering:

- HTTP/1.1 GET and POST requests
- Connection: close behavior
- Malformed request handling
- h2c (HTTP/2 cleartext) upgrade request
- Timeout handling
- Basic stress test using Go benchmarks

### Run tests

```bash
go test -v
```

### Benchmarking

```bash
go test -bench=.
```
