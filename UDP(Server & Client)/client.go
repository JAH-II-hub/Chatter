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
	// Check command-line arguments
	if len(os.Args) < 3 {
		fmt.Println("Usage: client <server_ip> <username>")
		return
	}

	// Create UDP socket
	conn, err := net.ListenUDP("udp", nil)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	// Resolve server address
	serverAddr, err := net.ResolveUDPAddr("udp", os.Args[1]+":8080")
	if err != nil {
		log.Fatal(err)
	}

	// Register with server
	_, err = conn.WriteToUDP([]byte("NAME:"+os.Args[2]), serverAddr)
	if err != nil {
		log.Fatal(err)
	}

	// Handle graceful exit
	quit := make(chan struct{})
	go func() {
		sigchan := make(chan os.Signal, 1)
		signal.Notify(sigchan, syscall.SIGINT, syscall.SIGTERM)
		<-sigchan
		fmt.Println("\nQuitting...")
		conn.WriteToUDP([]byte("QUIT"), serverAddr)
		close(quit)
	}()

	// Message receiver
	go func() {
		buf := make([]byte, 1024)
		for {
			select {
			case <-quit:
				return
			default:
				n, _, err := conn.ReadFromUDP(buf)
				if err != nil {
					select {
					case <-quit:
						return
					default:
						log.Printf("Receive error: %v", err)
					}
					continue
				}
				fmt.Printf("\n%s\n> ", buf[:n])
			}
		}
	}()

	fmt.Printf("Connected as %s. Type messages (or 'quit' to exit):\n> ", os.Args[2])
	// Read user input from console
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		select {
		case <-quit:
			return
		default:
			msg := scanner.Text()
			if msg == "quit" {
				conn.WriteToUDP([]byte("QUIT"), serverAddr)
				return
			}

			// Send message to server
			_, err := conn.WriteToUDP([]byte(msg), serverAddr)
			if err != nil {
				log.Printf("Send error: %v", err)
			}
			fmt.Print("> ")
		}
	}
}
