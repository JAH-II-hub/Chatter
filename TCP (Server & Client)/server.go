//Jesse - server.go
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

type Client struct {
	conn net.Conn
	name string
}

var (
	clients    = make(map[*Client]bool)
	clientsMux = sync.Mutex{}
	shutdown   = make(chan struct{})
)

func broadcast(msg string) {
	clientsMux.Lock()
	defer clientsMux.Unlock()

	for client := range clients {
		_, err := client.conn.Write([]byte(msg))
		if err != nil {
			log.Printf("TCP: Client %s failed - %v", client.name, err)
			delete(clients, client)
			client.conn.Close()
		}
	}
}

func handleClient(conn net.Conn) {
	// Connection demo
	log.Printf("TCP: New connection from %s (awaiting name)...", conn.RemoteAddr())

	buf := make([]byte, 1024)
	n, err := conn.Read(buf)
	if err != nil {
		log.Printf("TCP: Failed to get name from %s", conn.RemoteAddr())
		conn.Close()
		return
	}

	msg := strings.TrimSpace(string(buf[:n]))
	if !strings.HasPrefix(msg, "NAME:") {
		conn.Write([]byte("Please register with NAME:yourname\n"))
		conn.Close()
		return
	}

	name := msg[5:]
	client := &Client{conn, name}

	clientsMux.Lock()
	clients[client] = true
	clientsMux.Unlock()

	log.Printf("TCP: Connection established with %s (%s)", conn.RemoteAddr(), name)
	broadcast(name + " joined the chat\n")

	defer func() {
		clientsMux.Lock()
		delete(clients, client)
		clientsMux.Unlock()
		conn.Close()
		log.Printf("TCP: %s disconnected", name)
		broadcast(name + " left the chat\n")
	}()

	for {
		conn.SetReadDeadline(time.Now().Add(1 * time.Second))

		n, err := conn.Read(buf)
		if err != nil {
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				select {
				case <-shutdown:
					return
				default:
					continue
				}
			}
			log.Printf("TCP: Connection with %s terminated (%v)", name, err)
			return
		}

		msg := strings.TrimSpace(string(buf[:n]))
		switch {
		case msg == "/quit":
			return
		case msg == "/demo":
			conn.Write([]byte("TCP Demo Commands:\n"))
			conn.Write([]byte("TEST RELIABILITY - Show guaranteed delivery\n"))
			conn.Write([]byte("TEST ORDER - Show in-order delivery\n"))
		default:
			broadcast("[" + name + "]: " + msg + "\n")
		}
	}
}

func main() {
	go func() {
		sigchan := make(chan os.Signal, 1)
		signal.Notify(sigchan, syscall.SIGINT, syscall.SIGTERM)
		<-sigchan
		log.Println("TCP: Shutting down server gracefully...")
		close(shutdown)
		os.Exit(0)
	}()

	listener, err := net.Listen("tcp", ":8080")
	if err != nil {
		log.Fatal("TCP Server failed: ", err)
	}
	defer listener.Close()
	log.Println("TCP Server running on :8080 (Press Ctrl+C to shutdown)")

	for {
		select {
		case <-shutdown:
			return
		default:
			listener.(*net.TCPListener).SetDeadline(time.Now().Add(1 * time.Second))
			conn, err := listener.Accept()
			if err != nil {
				if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
					continue
				}
				log.Printf("TCP: Connection failed: %v", err)
				continue
			}

			go handleClient(conn)
		}
	}
}
