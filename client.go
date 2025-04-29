// Jesse - TCP Client
package main

import (
    "bufio"
    "fmt"
    "log"
    "net"
    "os"
)

func main() {
    // connect to server
    conn, err := net.Dial("tcp", "localhost:8080")
    if err != nil {
        log.Fatal("Can't connect: ", err)
    }
    defer conn.Close() // close when done
    fmt.Println("Connected. Type /quit to exit.")
    
    // goroutine to handle incoming messages
    go func() {
        buf := make([]byte, 1024) // message buffer
        for {
            n, err := conn.Read(buf) // read from server
            if err != nil {
                log.Println("Server kicked us out")
                os.Exit(0) // exit if connection stops
            }
            fmt.Print(string(buf[:n])) // print message
        }
    }()
    
    //read user input
    scanner := bufio.NewScanner(os.Stdin)
    for scanner.Scan() {
        msg := scanner.Text() // get user input
        if msg == "/quit" {
            fmt.Println("Bye!")
            return // exit w/ /quit
        }
        
        // send message to server
        _, err := conn.Write([]byte(msg + "\n"))
        if err != nil {
            log.Println("Failed to send:", err)
            return // exit if fail
        }
    }
}