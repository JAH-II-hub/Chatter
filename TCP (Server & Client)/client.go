//Jesse - client.go
package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run tcpclient.go <username>")
		return
	}

	fmt.Println("TCP: Connecting to server...")
	conn, err := net.Dial("tcp", "localhost:8080")
	if err != nil {
		log.Fatal("TCP Connection failed:", err)
	}
	defer conn.Close()

	_, err = fmt.Fprintf(conn, "NAME:%s\n", os.Args[1])
	if err != nil {
		log.Fatal("TCP Name registration failed:", err)
	}
	fmt.Printf("TCP: Connected as %s\n", os.Args[1])

	quit := make(chan struct{})
	go func() {
		sigchan := make(chan os.Signal, 1)
		signal.Notify(sigchan, syscall.SIGINT, syscall.SIGTERM)
		<-sigchan
		fmt.Println("\nTCP: Disconnecting...")
		fmt.Fprintf(conn, "/quit\n")
		close(quit)
		os.Exit(0)
	}()

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

	fmt.Printf("Type messages or commands:\n/demo - Show demo commands\n> ")
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		select {
		case <-quit:
			return
		default:
			msg := scanner.Text()
			if msg == "quit" {
				fmt.Fprintf(conn, "/quit\n")
				return
			} else if msg == "TEST RELIABILITY" {
				fmt.Println("TCP: Sending 10 messages...")
				for i := 0; i < 100; i++ {
					fmt.Fprintf(conn, "Message %d\n", i)
				}
			} else if msg == "TEST ORDER" {
				fmt.Println("TCP: Sending ordered messages...")
				for i := 0; i < 50; i++ {
					fmt.Fprintf(conn, "Message %d\n", i)
					time.Sleep(time.Duration(5-i) * 100 * time.Millisecond)
				}
			} else {
				fmt.Fprintf(conn, "%s\n", msg)
			}
			fmt.Print("> ")
		}
	}
}
