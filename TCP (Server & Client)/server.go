// Jesse - Enhanced TCP Server with UDP Features
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

// Client struct because we need to track connections
type Client struct {
    conn   net.Conn // actual connection
    name   string   // added client name tracking
}

// Global variables
var (
    clients    = make(map[*Client]bool) // global map connected clients
    clientsMux = sync.Mutex{}
    shutdown   = make(chan struct{})    // added graceful shutdown channel
)

// broadcast sends message to all connected clients
func broadcast(msg string) {
    clientsMux.Lock() // lock because we're touching shared data
    defer clientsMux.Unlock() // unlock when done. prevent deadlocks
    
    // loop through all clients
    for client := range clients {
        // try to send message
        _, err := client.conn.Write([]byte(msg))
        if err != nil {
            // if fail remove client
            log.Printf("Client %s failed", client.conn.RemoteAddr())
            delete(clients, client)
            client.conn.Close() // close connection after fail
        }
    }
}

// handleClient single client connection
func handleClient(conn net.Conn) {
    // First get client name (expecting NAME:username)
    buf := make([]byte, 1024)
    n, err := conn.Read(buf)
    if err != nil {
        log.Printf("Failed to get name from %s", conn.RemoteAddr())
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
    client := &Client{conn, name} // create new client obj with name
    
    // add to clients map
    clientsMux.Lock()
    clients[client] = true
    clientsMux.Unlock()
    
    log.Printf("%s joined from %s", name, conn.RemoteAddr())
    broadcast(name + " joined the chat\n")
    
    // cleanup when client disconnect
    defer func() {
        clientsMux.Lock()
        delete(clients, client) // remove from map
        clientsMux.Unlock()
        conn.Close() // close connection
        log.Printf("%s left", name)
        broadcast(name + " left the chat\n")
    }()
    
    // infinite loop to keep reading messages
    for {
        // Set read timeout for shutdown checks
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
            log.Printf("%s vanished", name)
            return // exit if error
        }
        
        // process message
        msg := strings.TrimSpace(string(buf[:n]))
        switch {
        case msg == "/quit":
            log.Printf("%s quit properly", name)
            return
        case strings.HasPrefix(msg, "/"):
            // Handle other commands here if needed
            conn.Write([]byte("Unknown command\n"))
        default:
            // broadcast to everyone with name instead of address
            broadcast("[" + name + "]: " + msg + "\n")
        }
    }
}

func main() {
    // Set up graceful shutdown handler
    go func() {
        sigchan := make(chan os.Signal, 1)
        signal.Notify(sigchan, syscall.SIGINT, syscall.SIGTERM)
        <-sigchan
        log.Println("Shutting down server gracefully...")
        close(shutdown)
        os.Exit(0)
    }()

    // start TCP server
    listener, err := net.Listen("tcp", ":8080")
    if err != nil {
        log.Fatal("Server go boom: ", err)
    }
    defer listener.Close() // close listener when done
    log.Println("Server running on :8080 (Press Ctrl+C to shutdown)")
    
    // infinite connection seeker
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
                log.Printf("Connection failed: %v", err)
                continue // skip broken connections
            }
            
            go handleClient(conn) // handle in goroutine
        }
    }
}
