# TCP Chat Server & Client in Go

A simple demonstration of TCP networking in Go, featuring a chat server and client that showcase TCP's reliability and ordered delivery.

## Features

- **TCP-based communication**
- Multi client chat room
- Message broadcasting
- Graceful shutdown handling
- Demo commands to test TCP features

## Quick Start

1. **Start the server**:
   ```bash
   go run -tags server.go
2. **Start client**:
   ```bash
   go run -tags client.go userName
## Commands 

  /demo
    - Shows you command list
    
  TEST RELIABILITY
    - Sends messages to test delivery

  TEST ORDER
    - Sends and tests to see if messages are in order
