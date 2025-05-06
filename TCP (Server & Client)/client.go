//go:build client

// Jesse - TCP Client
package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run -tags client client.go <username>")
		return
	}

	conn, err := net.Dial("tcp", "localhost:8080")
	if err != nil {
		log.Fatal("Connection failed:", err)
	}
	defer conn.Close()

	// Register name
	_, err = fmt.Fprintf(conn, "NAME:%s\n", os.Args[1])
	if err != nil {
		log.Fatal("Name registration failed:", err)
	}

	// Graceful shutdown
	quit := make(chan struct{})
	go func() {
		sigchan := make(chan os.Signal, 1)
		signal.Notify(sigchan, syscall.SIGINT, syscall.SIGTERM)
		<-sigchan
		fmt.Println("\nDisconnecting...")
		fmt.Fprintf(conn, "QUIT\n")
		close(quit)
		os.Exit(0)
	}()

	// Message receiver
	go func() {
		scanner := bufio.NewScanner(conn)
		for scanner.Scan() {
			select {
			case <-quit:
				return
			default:
				fmt.Println(scanner.Text())
			}
		}
	}()

	// User input
	fmt.Printf("Connected as %s. Type messages or QUIT to exit.\n", os.Args[1])
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		select {
		case <-quit:
			return
		default:
			msg := scanner.Text()
			if msg == "QUIT" {
				fmt.Fprintf(conn, "QUIT\n")
				return
			}
			_, err := fmt.Fprintf(conn, "%s\n", msg)
			if err != nil {
				log.Println("Send failed:", err)
				return
			}
		}
	}
}
