package main

import (
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"strconv"
)

func main() {
	if len(os.Args) == 1 {
		fmt.Fprintf(os.Stdout, "Usage: %s <hostname> [port]\n", os.Args[0])
		os.Exit(1)
	}

	hostname := os.Args[1]
	port := "80"
	if len(os.Args) == 3 {
		port = os.Args[2]
		if p, err := strconv.Atoi(os.Args[2]); err != nil || p <= 0 || p > 65535 {
			fmt.Fprintf(os.Stderr, "Invalid port number: %s\n", port)
			os.Exit(1)
		}
	}

	address := net.JoinHostPort(hostname, port)
	conn, err := net.Dial("tcp", address)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to connect: %v\n", err)
		os.Exit(1)
	}

	req := "GET / HTTP/1.1\r\n" +
		"Host: " + hostname + "\r\n" +
		"Connection: close\r\n\r\n"
	_, err = conn.Write([]byte(req))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to write request: %v\n", err)
		os.Exit(1)
	}

	buf := make([]byte, 4096)
	for {
		n, err := conn.Read(buf)
		if n > 0 {
			if _, err := os.Stdout.Write(buf[:n]); err != nil {
				fmt.Fprintf(os.Stderr, "Failed to write to stdout: %v\n", err)
				os.Exit(1)
			}
		}
		if err != nil {
			if !errors.Is(err, io.EOF) {
				fmt.Fprintf(os.Stderr, "Failed to read response: %v\n", err)
				os.Exit(1)
			}
			break
		}
	}

	err = conn.Close()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to close connection: %v\n", err)
		os.Exit(1)
	}
}
