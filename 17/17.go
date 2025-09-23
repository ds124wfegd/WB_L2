package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	timeout := flag.Int("timeout", 10, "Connection timeout in seconds")
	flag.Parse()

	args := flag.Args()
	if len(args) < 2 {
		fmt.Fprintf(os.Stderr, "Usage: %s [--timeout=10] host port\n", os.Args[0])
		os.Exit(1)
	}

	host := args[0]
	port := args[1]
	address := net.JoinHostPort(host, port)

	connTimeout := time.Duration(*timeout) * time.Second

	conn, err := net.DialTimeout("tcp", address, connTimeout)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Connection error: %v\n", err)
		os.Exit(1)
	}
	defer conn.Close()

	done := make(chan struct{})

	// Горутина для чтения из сокета и вывода в STDOUT
	go func() {
		defer close(done)
		reader := bufio.NewReader(conn)
		buffer := make([]byte, 1024)

		for {
			n, err := reader.Read(buffer)
			if err != nil {
				if err != io.EOF {
					// Вывод ошибок в STDERR, а не в STDOUT
					fmt.Fprintf(os.Stderr, "Read error: %v\n", err)
				}
				return
			}
			if n > 0 {
				os.Stdout.Write(buffer[:n])
			}
		}
	}()

	// Горутина для чтения из STDIN и отправки в сокет
	go func() {
		reader := bufio.NewReader(os.Stdin)
		writer := bufio.NewWriter(conn)

		for {
			line, err := reader.ReadString('\n')
			if err != nil {
				if err == io.EOF {
					conn.Close()
					return
				}
				fmt.Fprintf(os.Stderr, "StdIn read error: %v\n", err)
				conn.Close()
				return
			}

			_, err = writer.WriteString(line)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Write error: %v\n", err)
				conn.Close()
				return
			}
			err = writer.Flush()
			if err != nil {
				fmt.Fprintf(os.Stderr, "Flush error: %v\n", err)
				conn.Close()
				return
			}
		}
	}()

	// Обработка сигналов (Ctrl+C)
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)

	select {
	case <-sigCh:
		conn.Close()
	case <-done:
	}
}
