package main

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
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
	Code func() string // ฟังก์ชันที่จะถูกเรียก
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

	// สร้าง Smart Contract
	randomNumberContract := SmartContract{
		Name: "RandomNumberGenerator",
		Code: generateRandomNumber,
	}

	// เรียกใช้ Smart Contract ทุกๆ 5 วินาที
	go executeSmartContract(randomNumberContract)

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
	fmt.Println("New connection:", conn.RemoteAddr().String())
	mu.Lock()
	peers = append(peers, conn)
	mu.Unlock()
	conn.Write([]byte("Welcome to the P2P Blockchain Server!\n"))
	for {
		var msg string
		_, err := fmt.Fscan(conn, &msg)
		if err != nil {
			break
		}
		processMessage(msg)

		response := fmt.Sprintf("Received: %s\n", msg)
		conn.Write([]byte(response))
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

func executeSmartContract(contract SmartContract) {
	for {
		data := contract.Code() // เรียกใช้ฟังก์ชันของ Smart Contract
		addBlock(data)
		time.Sleep(5 * time.Second)
	}
}

func generateRandomNumber() string {
	randomNumber := rand.Intn(100) + 1
	return fmt.Sprintf("Random Number: %d", randomNumber)
}
