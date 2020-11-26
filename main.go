package main

import (
	"flag"
	"fmt"
	"net"
	"strings"
	"time"
)

var port string

type state struct {
	//lock  sync.Mutex
	peers []string
}

var peersState *state

func main() {
	flag.StringVar(&port, "port", "", "server port")
	var peersList string
	flag.StringVar(&peersList, "peers", "", "peers list")
	flag.Parse()

	peersState = &state{}
	peersState.peers = strings.Split(peersList, ",")
	timeGossiping(doGossip)
	recieveGossip(port)
}

func recieveGossip(port string) {
	ln, err := net.Listen("tcp", ":"+port)
	if err != nil {
		fmt.Println("Error: ", err.Error())
	}
	for {
		conn, err := ln.Accept()
		if err != nil {
			fmt.Println("Error: ", err.Error())
		}

		go handleGossip(conn)
		fmt.Printf("============\n")
		fmt.Printf("Current Peers: %s\n", peersState.peers)
		fmt.Printf("============\n")
	}
}

func handleGossip(conn net.Conn) {
	defer conn.Close()

	buf := make([]byte, 1024)

	_, err := conn.Read(buf)
	if err != nil {
		fmt.Println("Error: ", err.Error())
	}

	parseReceivedData(buf)
}

func parseReceivedData(dataBuffer []byte) {
	data := strings.Split(string(dataBuffer), ":")
	peer := data[0]
	peerList := strings.Split(data[1], ",")

	newPeers := cleanNullBytes(peerList)

	fmt.Printf("============\n")
	fmt.Printf("Received %s from %s\n", newPeers, peer)
	fmt.Printf("============\n")

	newPeers = append(newPeers, peer)

	//peersState.lock.Lock()
	//defer peersState.lock.Unlock()

	m := make(map[string]bool)
	for _, item := range peersState.peers {
		m[item] = true
	}

	peersState.peers = append(peersState.peers, updatePeersState(port, newPeers, m)...)
}

func cleanNullBytes(peerList []string) []string {
	peers := make([]string, 0)
	for _, p := range peerList {
		peers = append(peers, strings.Trim(p, "\x00"))
	}
	return peers
}

func updatePeersState(port string, newPeers []string, existingPeers map[string]bool) []string {
	currentPeers := make([]string, 0)
	for _, p := range newPeers {
		if p != port {
			if _, ok := existingPeers[p]; !ok {
				currentPeers = append(currentPeers, p)
			}
		}
	}
	return currentPeers
}

func timeGossiping(do func()) {
	ticker := time.NewTicker(5 * time.Second)
	quit := make(chan struct{})
	go func() {
		for {
			select {
			case <-ticker.C:
				do()
			case <-quit:
				ticker.Stop()
				return
			}
		}
	}()
}

func doGossip() {
	//peersState.lock.Lock()
	//defer peersState.lock.Unlock()
	for _, p := range peersState.peers {
		go send(p, peersState.peers)
	}
}

func send(peer string, peers []string) {
	conn, err := net.Dial("tcp", "localhost:"+peer)
	if err != nil {
		return
	}

	peersData := strings.Join(peers, ",")
	fmt.Fprintf(conn, port+":"+peersData)
}
