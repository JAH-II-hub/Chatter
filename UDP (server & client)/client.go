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
	if len(os.Args) < 3 {
		fmt.Println("Usage: udpclient <server_ip> <username>")
		return
	}

	conn, err := net.ListenUDP("udp", nil)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	serverAddr, err := net.ResolveUDPAddr("udp", os.Args[1]+":8080")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("UDP: Registering with server...")
	_, err = conn.WriteToUDP([]byte("NAME:"+os.Args[2]), serverAddr)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("UDP: Connected as %s\n", os.Args[2])

	quit := make(chan struct{})
	go func() {
		sigchan := make(chan os.Signal, 1)
		signal.Notify(sigchan, syscall.SIGINT, syscall.SIGTERM)
		<-sigchan
		fmt.Println("\nUDP: Quitting...")
		conn.WriteToUDP([]byte("QUIT"), serverAddr)
		close(quit)
		os.Exit(0)
	}()

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
						log.Printf("UDP Receive error: %v", err)
					}
					continue
				}
				fmt.Printf("\n%s\n> ", buf[:n])
			}
		}
	}()

	fmt.Printf("Type messages or commands:\nTEST RELIABILITY - Test packet loss\nTEST ORDER - Test ordering\n> ")
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
			} else if msg == "TEST RELIABILITY" {
				fmt.Println("UDP: Sending 10 messages (some may get lost)...")
				for i := 0; i < 10; i++ {
					if i != 3 && i != 7 {
						conn.WriteToUDP([]byte(fmt.Sprintf("Message %d", i)), serverAddr)
					}
				}
			} else if msg == "TEST ORDER" {
				fmt.Println("UDP: Sending messages with random delays...")
				for i := 0; i < 5; i++ {
					conn.WriteToUDP([]byte(fmt.Sprintf("Message %d", i)), serverAddr)
					time.Sleep(time.Duration(i) * 100 * time.Millisecond)
				}
			} else {
				conn.WriteToUDP([]byte(msg), serverAddr)
			}
			fmt.Print("> ")
		}
	}
}
