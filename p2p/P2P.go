package p2p

import (
	"bytes"
	"encoding/binary"
	"encoding/gob"
	"fmt"
	. "github.com/nfk93/blockchain/objects"
	"log"
	"net"
	"sync"
)

const (
	BROADCAST_BLOCK          uint16 = 44444
	BROADCAST_NEW_CONNECTION uint16 = 32001
	CONNECT_TO_NETWORK       uint16 = 32000
	OK                       uint16 = 10000
)

var networkList map[string]bool
var networkListLock sync.RWMutex
var peers []string
var peersLock sync.RWMutex

func StartP2P(connectTo string, myHostPort string) {
	networkList = make(map[string]bool)
	peers = make([]string, 0)

	if connectTo == "" {
		fmt.Println("STARTING OWN NETWORK!")
		myAddr := "127.0.0.1" + ":" + myHostPort // TODO: get real ip
		addNewConnection(myAddr)
		serveOnPort(myHostPort)

	} else {
		fmt.Println("CONNECTING TO EXISTING NETWORK AT ", connectTo)
		connectToNetwork(connectTo, myHostPort)
		serveOnPort(myHostPort)
	}
}

func PrintNetworkList() {
	for _, k := range keyset(networkList) {
		fmt.Println(k)
	}
}

func serveOnPort(port string) (net.Listener, error) {
	ln, err := net.Listen("tcp", ":"+port)
	if err != nil {
		return nil, err
	}
	go func() {
		for {
			conn, err := ln.Accept()
			if err != nil {
				log.Println("Connection error: ", err)
			}
			go handleConnection(conn)
		}
	}()
	return ln, err
}

func handleConnection(conn net.Conn) {
	requestIDEncoded := make([]byte, 16)
	conn.Read(requestIDEncoded)
	requestId := binary.BigEndian.Uint16(requestIDEncoded)

	switch requestId {
	case BROADCAST_BLOCK:
		fmt.Print("\nReceived BROADCAST_BLOCK")
		receivedBlock(conn)
	case BROADCAST_NEW_CONNECTION:
		fmt.Print("\nReceived BROADCAST_NEW_CONNECTION")
		receivedNewConnection(conn)
	case CONNECT_TO_NETWORK:
		fmt.Print("\nReceived CONNECT_TO_NETWORK")
		receivedConnectToNetwork(conn)
	default:
		fmt.Print("\nInvalid request ID: ", requestId)
	}
}

// -----------------------------------------------------------
// NETWORK METHODS
// -----------------------------------------------------------

type newConnectionData struct {
	Address string
}

func receivedNewConnection(conn net.Conn) {
	var data newConnectionData
	decoder := gob.NewDecoder(conn)
	decoder.Decode(&data)
	fmt.Println("\n\treceived broadcast connection:", data.Address)

	// Check if we know the peer, and exit if we do.
	alreadyKnown := false
	func() {
		networkListLock.RLock()
		defer networkListLock.RUnlock()
		if networkList[data.Address] {
			alreadyKnown = true
		}
	}()
	if alreadyKnown {
		// Early exit
		return
	}

	// Add the connection to your network
	addNewConnection(data.Address)

	// Broadcast it to everyone else
	broadcastNewConnection(data)
}

func broadcastNewConnection(data newConnectionData) {
	peersLock.RLock()
	defer peersLock.RUnlock()
	for _, peer := range getPeers() {
		sendNewConnectionTo(peer, data)
	}
}

func sendNewConnectionTo(toAddr string, data newConnectionData) {
	conn := initateRequest(toAddr, BROADCAST_NEW_CONNECTION)
	defer conn.Close()

	encoder := gob.NewEncoder(conn)
	encoder.Encode(data)
}

func receivedBlock(conn net.Conn) {
	var block Block
	decoder := gob.NewDecoder(conn)
	decoder.Decode(&block)
	// TODO
}

func broadcastBlock(addr string, block Block) {
	conn := initateRequest(addr, BROADCAST_BLOCK)
	defer conn.Close()

	encoder := gob.NewEncoder(conn)
	encoder.Encode(block)
}

func connectToNetwork(addr string, myPort string) {
	conn := initateRequest(addr, CONNECT_TO_NETWORK)
	defer conn.Close()

	buf := new(bytes.Buffer)
	binary.Write(buf, binary.BigEndian, myPort)
	conn.Write(buf.Bytes())

	/*
		encoder := gob.NewEncoder(conn)
		encoder.Encode(myPort)

		time.Sleep(time.Second*2)

		var myAddr string
		decoder := gob.NewDecoder(conn)
		decoder.Decode(&myAddr)
		decoder.Decode(&networkList)
		fmt.Println("\n\t... connected!") */
}

func receivedConnectToNetwork(conn net.Conn) {
	decoder := gob.NewDecoder(conn)
	var port string
	decoder.Decode(&port)

	addr, _, _ := net.SplitHostPort(conn.LocalAddr().String())
	fmt.Println("\n\tconnecting", addr)
	addNewConnection(addr + ":" + port)

	encoder := gob.NewEncoder(conn)
	encoder.Encode(addr)
	encoder.Encode(networkList)

	data := newConnectionData{addr + ":" + port}
	go broadcastNewConnection(data)
}

// -----------------------------------------------------------
// INTERNAL METHODS
// -----------------------------------------------------------

func addNewConnection(addr string) {
	networkListLock.Lock()
	defer networkListLock.Unlock()
	peersLock.Lock()
	defer peersLock.Unlock()
	networkList[addr] = true
	determinePeers()
}

func getPeers() []string {
	if peers == nil {
		determinePeers()
	}
	return peers
}

func determinePeers() {
	// determines your peers. Is run every time you receive a new connection
	// TODO
	peers = keyset(networkList)
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

// This is a slightly hacky way to obtain your own IPv4 IP address
func getIP() net.IP {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		log.Println(err.Error())
		return nil
	}
	defer conn.Close()
	addr := conn.LocalAddr().(*net.UDPAddr).IP
	return addr
}
