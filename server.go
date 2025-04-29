 // Jesse - TCP
package main

import (
    "log"
    "net"
    "strings"
    "sync"
)

// Client struct because we need to track connections
type Client struct {
    conn net.Conn //actual connectino
}

// Global variables
var (
    clients    = make(map[*Client]bool) // global map connected clients
    clientsMux = sync.Mutex{}          
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
    // create new client object
    client := &Client{conn}
    
    // add to clients map
    clientsMux.Lock()
    clients[client] = true
    clientsMux.Unlock()
    
    //cleanup when client disconnect
    defer func() {
        clientsMux.Lock()
        delete(clients, client) // remove from map
        clientsMux.Unlock()
        conn.Close() // close connection
        log.Printf("Client %s left", conn.RemoteAddr())
    }()
    
    // buffer for incoming messages
    buf := make([]byte, 1024)
    
    //infinite loop to keep readign messages
    for {
        // read from connection
        n, err := conn.Read(buf)
        if err != nil {
            log.Printf("Client %s vanished", conn.RemoteAddr())
            return // exit if error
        }
        
        // process message
        msg := strings.TrimSpace(string(buf[:n]))
        if msg == "/quit" {
            log.Printf("Client %s quit properly", conn.RemoteAddr())
            return // exit on /quit command
        }
        
        // broadcast to everyone
        broadcast("[" + conn.RemoteAddr().String() + "]: " + msg + "\n")
    }
}

func main() {
    // start TCP server
    listener, err := net.Listen("tcp", ":8080")
    if err != nil {
        log.Fatal("Server go boom: ", err)
    }
    defer listener.Close() // close listener when done
    log.Println("Server running on :8080")
    
    // infinite connection seeker
    for {
        conn, err := listener.Accept()
        if err != nil {
            log.Printf("Connection failed: %v", err)
            continue // skip broken connections
        }
        
        log.Printf("New client: %s", conn.RemoteAddr())
        go handleClient(conn) // handle in goroutine
    }
}