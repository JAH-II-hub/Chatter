package main

import (
	"log"
	"net"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"
)

func main() {
	// Create UDP server listening on port 8080
	conn, err := net.ListenUDP("udp", &net.UDPAddr{Port: 8080})
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close() // Ensure connection closes when main() exits

	log.Println("Server started on :8080. Press Ctrl+C to shutdown.")

	// Thread-safe map to store connected clients
	var clients sync.Map

	// Channel for shutdown signal
	shutdown := make(chan struct{})

	// Handle graceful shutdown
	go func() {
		sigchan := make(chan os.Signal, 1)
		signal.Notify(sigchan, syscall.SIGINT, syscall.SIGTERM)
		<-sigchan
		log.Println("Shutting down server...")
		close(shutdown)
		conn.Close() // This will unblock ReadFromUDP
	}()

	// Buffer for incoming messages
	buf := make([]byte, 1024)
	for {
		select {
		case <-shutdown:
			return
		default:
			// Set timeout to periodically check for shutdown
			conn.SetReadDeadline(time.Now().Add(1 * time.Second))
			// Read incoming UDP packet
			n, addr, err := conn.ReadFromUDP(buf)
			if err != nil {
				if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
					continue
				}
				log.Printf("Read error: %v", err)
				continue
			}

			// Convert received bytes to string
			msg := string(buf[:n])

			// Handle QUIT command from clients
			if msg == "QUIT" {
				if name, ok := clients.Load(addr.String()); ok {
					log.Printf("%s has left the chat", name)
					clients.Delete(addr.String())
				}
				continue
			}

			// Handle new client registration
			if strings.HasPrefix(msg, "NAME:") {
				name := strings.TrimSpace(msg[5:])
				clients.Store(addr.String(), name)
				log.Printf("%s joined from %s", name, addr)
				continue
			}

			// Handle normal chat messages
			senderName, _ := clients.Load(addr.String())

			// Broadcast to all other clients
			clients.Range(func(key, value interface{}) bool {
				if key.(string) != addr.String() {
					fullMsg := senderName.(string) + ": " + msg
					clientAddr, _ := net.ResolveUDPAddr("udp", key.(string))
					_, err := conn.WriteToUDP([]byte(fullMsg), clientAddr)
					if err != nil {
						log.Printf("Send error to %s: %v", key, err)
						clients.Delete(key)
					}
				}
				return true
			})
		}
	}
}
