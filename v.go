package main

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"sync"
	"time"
)

type Block struct {
	Index        int    `json:"index"`
	Timestamp    string `json:"timestamp"`
	Data         string `json:"data"`
	PreviousHash string `json:"previous_hash"`
	Hash         string `json:"hash"`
}

type Blockchain struct {
	Blocks []Block
	mu     sync.Mutex
}

type SmartContract struct {
	Name string
	Code string
}

var blockchain Blockchain
var peers []net.Conn
var mu sync.Mutex

func main() {
	port := "8080"
	ln, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Fatal(err)
	}
	defer ln.Close()

	go handleIncomingConnections(ln)
	createGenesisBlock()

	fmt.Println("Blockchain server started on port", port)
	select {}
}

func createGenesisBlock() {
	genesisBlock := Block{
		Index:        0,
		Timestamp:    time.Now().String(),
		Data:         "Genesis Block",
		PreviousHash: "",
		Hash:         calculateHash(0, time.Now().String(), "Genesis Block", ""),
	}
	blockchain.Blocks = append(blockchain.Blocks, genesisBlock)
}

func calculateHash(index int, timestamp string, data string, previousHash string) string {
	record := fmt.Sprintf("%d%s%s%s", index, timestamp, data, previousHash)
	hash := sha256.New()
	hash.Write([]byte(record))
	return fmt.Sprintf("%x", hash.Sum(nil))
}

func handleIncomingConnections(ln net.Listener) {
	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Println("Error accepting connection:", err)
			continue
		}
		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	defer conn.Close()
	mu.Lock()
	peers = append(peers, conn)
	mu.Unlock()

	for {
		var msg string
		_, err := fmt.Fscan(conn, &msg)
		if err != nil {
			break
		}
		processMessage(msg)
	}
}

func processMessage(msg string) {
	var block Block
	if err := json.Unmarshal([]byte(msg), &block); err == nil {
		mu.Lock()
		blockchain.Blocks = append(blockchain.Blocks, block)
		mu.Unlock()
		fmt.Printf("New block added: %+v\n", block)
	}
}

func addBlock(data string) {
	mu.Lock()
	defer mu.Unlock()

	lastBlock := blockchain.Blocks[len(blockchain.Blocks)-1]
	newBlock := Block{
		Index:        lastBlock.Index + 1,
		Timestamp:    time.Now().String(),
		Data:         data,
		PreviousHash: lastBlock.Hash,
		Hash:         calculateHash(lastBlock.Index+1, time.Now().String(), data, lastBlock.Hash),
	}
	blockchain.Blocks = append(blockchain.Blocks, newBlock)
	broadcastBlock(newBlock)
}

func broadcastBlock(block Block) {
	mu.Lock()
	defer mu.Unlock()
	blockJSON, _ := json.Marshal(block)
	for _, peer := range peers {
		peer.Write(blockJSON)
	}
}

func deploySmartContract(name string, code string) SmartContract {
	return SmartContract{Name: name, Code: code}
}
