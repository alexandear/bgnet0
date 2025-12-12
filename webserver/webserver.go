package main

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"math/rand/v2"
	"net"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"syscall"
	"time"
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

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	ln, err := net.Listen("tcp", net.JoinHostPort("", port))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to listen on port %s: %v\n", port, err)
		os.Exit(1)
	}
	defer ln.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		<-sigs
		fmt.Println("Shutting down server...")
		cancel()
		ln.Close()
	}()

	var wg sync.WaitGroup

	for {
		conn, err := ln.Accept()
		if err != nil {
			if errors.Is(ctx.Err(), context.Canceled) {
				// Server is shutting down
				break
			}
			fmt.Fprintf(os.Stderr, "Failed to accept connection: %v\n", err)
			continue
		}
		fmt.Printf("Accepted connection from %s\n", conn.RemoteAddr().String())

		wg.Go(func() {
			if err := handleConn(conn, "Hello, World!"); err != nil {
				fmt.Fprintf(os.Stderr, "Error handling connection: %v\n", err)
			}
		})
	}

	wg.Wait()
	fmt.Println("Server stopped")
}

func handleConn(conn net.Conn, body string) error {
	defer conn.Close()

	r := bufio.NewReader(conn)
	var buf bytes.Buffer

	for {
		time.Sleep(time.Duration(rand.N(5)) * time.Second) // Simulate processing delay

		line, err := r.ReadBytes('\n')
		if err != nil {
			if errors.Is(err, io.EOF) {
				return nil
			}
			return err
		}
		buf.Write(line)
		if bytes.HasSuffix(buf.Bytes(), []byte("\r\n\r\n")) {
			break
		}
	}

	response := "HTTP/1.1 200 OK\r\n" +
		"Content-Type: text/plain\r\n" +
		fmt.Sprintf("Content-Length: %d\r\n", len(body)) +
		"Connection: close\r\n" +
		"\r\n" +
		body

	_, err := conn.Write([]byte(response))
	return err
}
