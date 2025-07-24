# Basic HTTP Server in Go

This project demonstrates a simple HTTP server in Go, with branches for different HTTP versions.

## Branches

- `main` – HTTP/2 server
- `http/1` – Raw TCP implementation of HTTP/1.0
- `http/1.1` – HTTP/1.1 using `net/http`
- `http/2` – HTTP/2 with h2c (cleartext)

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
