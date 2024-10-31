package main

import (
	"fmt"
	"net"
	"os"
	"sync"
)

type Peer struct {
	address string
	conn    net.Conn
}

var peers []Peer
var mu sync.Mutex

func main() {
	port := "8080"
	if len(os.Args) > 1 {
		port = os.Args[1]
	}

	ln, err := net.Listen("tcp", ":"+port)
	if err != nil {
		fmt.Println("Error starting server:", err)
		return
	}
	defer ln.Close()

	go handleIncomingConnections(ln)

	fmt.Println("Server started on port", port)
	select {} // Keep the server running
}

func handleIncomingConnections(ln net.Listener) {
	for {
		conn, err := ln.Accept()
		if err != nil {
			fmt.Println("Error accepting connection:", err)
			continue
		}
		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	defer conn.Close()

	var address string
	fmt.Fscan(conn, &address)

	mu.Lock()
	peers = append(peers, Peer{address, conn})
	mu.Unlock()

	fmt.Println("Connected to peer:", address)

	for {
		var msg string
		_, err := fmt.Fscan(conn, &msg)
		if err != nil {
			fmt.Println("Peer disconnected:", address)
			break
		}
		fmt.Printf("Message from %s: %s\n", address, msg)
		broadcastMessage(msg, address)
	}
}

func broadcastMessage(msg, sender string) {
	mu.Lock()
	defer mu.Unlock()
	for _, peer := range peers {
		if peer.address != sender {
			_, err := fmt.Fprintln(peer.conn, msg)
			if err != nil {
				fmt.Println("Error sending message to", peer.address)
			}
		}
	}
}
