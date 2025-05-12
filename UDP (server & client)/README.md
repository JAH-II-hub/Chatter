# UDP Chat Server & Client in Go

A simple demonstration of UDP networking in Go, featuring a chat server and client that showcase UDP's connectionless nature and performance characteristics.

## Features

- **UDP-based communication**
- Multi-client chat room
- Message broadcasting
- Graceful shutdown handling
- Demo commands to test UDP features

## Quick Start

1. **Start the server**:
   go run server.go

2. **Start the client**:
   go run client.go 127.0.0.1 [userName]

## Commands

    TEST RELIABILITY
    Sends messages to test delivery (simulates 20% packet loss)

    TEST ORDER
    Sends messages with random delays to demonstrate ordering

    quit
    Leaves the chat
