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
	conn, err := net.ListenUDP("udp", &net.UDPAddr{Port: 8080})
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	log.Println("UDP Server started on :8080. Press Ctrl+C to shutdown.")

	var clients sync.Map
	shutdown := make(chan struct{})

	go func() {
		sigchan := make(chan os.Signal, 1)
		signal.Notify(sigchan, syscall.SIGINT, syscall.SIGTERM)
		<-sigchan
		log.Println("UDP: Shutting down server...")
		close(shutdown)
		conn.Close()
	}()

	// Client timeout checker
	go func() {
		ticker := time.NewTicker(10 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				clients.Range(func(key, value interface{}) bool {
					if time.Now().Unix()%3 == 0 {
						log.Printf("UDP: Client %s timed out", value.(string))
						clients.Delete(key)
					}
					return true
				})
			case <-shutdown:
				return
			}
		}
	}()

	buf := make([]byte, 1024)
	for {
		select {
		case <-shutdown:
			return
		default:
			conn.SetReadDeadline(time.Now().Add(1 * time.Second))
			n, addr, err := conn.ReadFromUDP(buf)
			if err != nil {
				if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
					continue
				}
				log.Printf("UDP Read error: %v", err)
				continue
			}

			msg := string(buf[:n])

			if msg == "QUIT" {
				if name, ok := clients.Load(addr.String()); ok {
					log.Printf("UDP: %s left the chat", name)
					clients.Delete(addr.String())
				}
				continue
			}

			if strings.HasPrefix(msg, "NAME:") {
				name := strings.TrimSpace(msg[5:])
				log.Printf("UDP: New client %s at %s", name, addr)
				clients.Store(addr.String(), name)
				continue
			}

			// Simulate UDP unreliability
			if time.Now().UnixNano()%5 == 0 {
				log.Printf("UDP: Simulating packet loss for message: %s", msg)
				continue
			}

			// Simulate out-of-order delivery
			if strings.HasPrefix(msg, "Message ") && time.Now().UnixNano()%3 == 0 {
				time.Sleep(500 * time.Millisecond)
			}

			senderName, _ := clients.Load(addr.String())
			log.Printf("UDP: Received from %s: %s", senderName, msg)

			clients.Range(func(key, value interface{}) bool {
				if key.(string) != addr.String() {
					fullMsg := senderName.(string) + ": " + msg
					clientAddr, _ := net.ResolveUDPAddr("udp", key.(string))
					_, err := conn.WriteToUDP([]byte(fullMsg), clientAddr)
					if err != nil {
						log.Printf("UDP Send error to %s: %v", key, err)
						clients.Delete(key)
					}
				}
				return true
			})
		}
	}
}
