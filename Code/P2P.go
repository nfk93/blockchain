package Code

import (
	"bytes"
	"encoding/binary"
	"encoding/gob"
	"fmt"
	"log"
	"net"
)

const (
	BROADCAST_BLOCK          uint16 = 44444
	BROADCAST_NEW_CONNECTION uint16 = 32000
)

var networkList map[string]bool
var peers *[]string

func serveOnPort(port string) error {
	ln, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Fatal(err)
	}
	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Fatal(err)
		}
		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	requestIDEncoded := make([]byte, 16)
	conn.Read(requestIDEncoded)
	requestId := binary.BigEndian.Uint16(requestIDEncoded)

	switch requestId {
	case BROADCAST_BLOCK:
		receivedBlock(conn)
	case BROADCAST_NEW_CONNECTION:
		receivedNewConnection(conn)
	default:
		fmt.Println("Invalid request ID: ", requestId)
	}
}

type newConnectionData struct {
	Address string
}

func receivedNewConnection(conn net.Conn) {
	var data newConnectionData
	decoder := gob.NewDecoder(conn)
	decoder.Decode(&data)

	if networkList[data.Address] {
		// Early exit, since we already know the address
		return
	}

	networkList[data.Address] = true
	for _, peer := range getPeers() {
		broadcastNewConnection(peer, data)
	}
}

func broadcastNewConnection(toAddr string, data newConnectionData) {
	conn := initateRequest(toAddr, BROADCAST_NEW_CONNECTION)
	defer conn.Close()

	encoder := gob.NewEncoder(conn)
	encoder.Encode(data)
}

func receivedBlock(conn net.Conn) {
	var block Block
	decoder := gob.NewDecoder(conn)
	decoder.Decode(&block)
	fmt.Println("Received Block with value: ", block.Value)
}

func broadcastBlock(addr string, block Block) {
	conn := initateRequest(addr, BROADCAST_BLOCK)
	defer conn.Close()

	encoder := gob.NewEncoder(conn)
	encoder.Encode(block)
}

func getPeers() []string {
	if peers == nil {
		determinePeers()
	}
	return *peers
}

func determinePeers() {
	// determines your peers. Is run every time you receive a new connection
	// TODO
}

func initateRequest(toAddr string, requestID uint16) net.Conn {
	conn, err := net.Dial("tcp", toAddr)
	if err != nil {
		log.Fatal(err)
	}
	buf := new(bytes.Buffer)
	binary.Write(buf, binary.BigEndian, requestID)
	conn.Write(buf.Bytes())

	return conn
}
