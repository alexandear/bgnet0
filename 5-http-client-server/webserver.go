package main

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"strconv"
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

		go func(conn net.Conn) {
			defer func() {
				if err := conn.Close(); err != nil {
					fmt.Fprintf(os.Stderr, "Failed to close connection %s: %v\n", conn.RemoteAddr(), err)
					return
				}
				fmt.Printf("Closed connection from %s\n", conn.RemoteAddr())
			}()

			r := bufio.NewReader(conn)
			var buf bytes.Buffer

			for {
				line, err := r.ReadBytes('\n')
				if err != nil {
					if !errors.Is(err, io.EOF) {
						fmt.Fprintf(os.Stderr, "Failed to read bytes for connection %s: %v\n", conn.RemoteAddr(), err)
					}
					return
				}
				buf.Write(line)
				if bytes.HasSuffix(buf.Bytes(), []byte("\r\n\r\n")) {
					break
				}
			}

			body := "Hello, World!\n"

			response := "HTTP/1.1 200 OK\r\n" +
				"Content-Type: text/plain\r\n" +
				fmt.Sprintf("Content-Length: %d\r\n", len(body)) +
				"Connection: close\r\n" +
				"\r\n" +
				body

			_, err := conn.Write([]byte(response))
			if err != nil {
				fmt.Fprintf(os.Stderr, "Failed to write response for connection %s: %v\n", conn.RemoteAddr(), err)
				return
			}
		}(conn)
	}
}
