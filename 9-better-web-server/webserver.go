package main

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

func main() {
	port := "28333"
	if len(os.Args) == 2 {
		port = os.Args[1]
		if p, err := strconv.Atoi(os.Args[1]); err != nil || p <= 0 || p > 65535 {
			fmt.Fprintf(os.Stderr, "Invalid port number: %s\n", port)
			os.Exit(1)
		}
	}

	root, err := os.OpenRoot(".")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to open root directory: %v\n", err)
		return
	}
	defer root.Close()

	ln, err := net.Listen("tcp", net.JoinHostPort("", port))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to listen on port %s: %v\n", port, err)
		os.Exit(1)
	}
	defer ln.Close()

	fmt.Println("Server started listening on port", port)

	for {
		conn, err := ln.Accept()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to accept connection: %v\n", err)
			continue
		}
		fmt.Printf("Accepted connection from %s\n", conn.RemoteAddr())

		go handleConnection(conn, root)
	}
}

func handleConnection(conn net.Conn, root *os.Root) {
	defer func() {
		if err := conn.Close(); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to close connection %s: %v\n", conn.RemoteAddr(), err)
			return
		}
		fmt.Printf("Closed connection from %s\n", conn.RemoteAddr())
	}()

	r := bufio.NewReader(conn)
	var req bytes.Buffer

	for {
		line, err := r.ReadBytes('\n')
		if err != nil {
			if !errors.Is(err, io.EOF) {
				fmt.Fprintf(os.Stderr, "Failed to read bytes for connection %s: %v\n", conn.RemoteAddr(), err)
			}
			return
		}
		req.Write(line)
		if bytes.HasSuffix(req.Bytes(), []byte("\r\n\r\n")) {
			break
		}
	}

	fullpath, err := pathFromRequest(&req)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to parse request from connection %s: %v\n", conn.RemoteAddr(), err)
		return
	}

	filename := filepath.Base(fullpath)
	f, err := root.Open(filename)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			response := httpResponse(404, "Not Found", "text/plain", []byte("404 not found"))
			_, writeErr := conn.Write([]byte(response))
			if writeErr != nil {
				fmt.Fprintf(os.Stderr, "Failed to write 404 response for connection %s: %v\n", conn.RemoteAddr(), writeErr)
			}
			return
		}
		fmt.Fprintf(os.Stderr, "Failed to open file %s for connection %s: %v\n", filename, conn.RemoteAddr(), err)
		return
	}
	defer f.Close()

	var contentType string
	switch filepath.Ext(filename) {
	case ".txt":
		contentType = "text/plain"
	case ".html":
		contentType = "text/html"
	default:
	}

	body, err := io.ReadAll(f)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to read file %s for connection %s: %v\n", filename, conn.RemoteAddr(), err)
		return
	}

	resp := httpResponse(200, "OK", contentType, body)

	_, err = conn.Write([]byte(resp))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to write response for connection %s: %v\n", conn.RemoteAddr(), err)
		return
	}
}

func httpResponse(statusCode int, statusText, contentType string, body []byte) string {
	fmt.Fprintf(os.Stdout, "Sending response: %d %s, Content-Type: %s, Content-Length: %d\n", statusCode, statusText, contentType, len(body))
	return fmt.Sprintf("HTTP/1.1 %d %s\r\nContent-Type: %s\r\nContent-Length: %d\r\nConnection: close\r\n\r\n%s",
		statusCode, statusText, contentType, len(body), body)
}

func pathFromRequest(req *bytes.Buffer) (string, error) {
	lines := strings.Split(req.String(), "\r\n")
	if len(lines) == 0 {
		return "", errors.New("empty request")
	}

	parts := strings.Split(lines[0], " ")
	if len(parts) < 2 {
		return "", errors.New("malformed request line")
	}

	return parts[1], nil
}
