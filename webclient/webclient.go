package main

import (
	"context"
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

	if err := get(context.Background(), hostname, port); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func get(ctx context.Context, hostname, port string) error {
	conn, err := (&net.Dialer{}).DialContext(ctx, "tcp", net.JoinHostPort(hostname, port))
	if err != nil {
		return fmt.Errorf("failed to connect to %s:%s: %w", hostname, port, err)
	}
	defer func() {
		err := conn.Close()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to close connection: %v\n", err)
		}
	}()

	if err := httpGET(conn, hostname); err != nil {
		return fmt.Errorf("failed to send HTTP GET request: %w", err)
	}

	return nil
}

func httpGET(conn net.Conn, hostname string) error {
	req := "GET / HTTP/1.1\r\n" +
		"Host: " + hostname + "\r\n" +
		"Connection: close\r\n\r\n"
	_, err := conn.Write([]byte(req))
	if err != nil {
		return fmt.Errorf("failed to write request: %w", err)
	}

	buf := make([]byte, 4096)
	for {
		n, err := conn.Read(buf)
		if n > 0 {
			if _, err := os.Stdout.Write(buf[:n]); err != nil {
				return fmt.Errorf("failed to write to stdout: %w", err)
			}
		}
		if err != nil {
			if !errors.Is(err, io.EOF) {
				return fmt.Errorf("failed to read response: %w", err)
			}
			break
		}
	}

	return nil
}
