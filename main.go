package main

import (
	"flag"
	"fmt"
	"net"
	"strings"
	"sync"
	"time"
)

var port string

type state struct {
	lock  sync.Mutex
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
	}
}

func handleGossip(conn net.Conn) {
	defer conn.Close()

	buf := make([]byte, 1024)

	_, err := conn.Read(buf)
	if err != nil {
		fmt.Println("Error: ", err.Error())
	}

	data := strings.Split(string(buf), ":")
	peer := data[0]
	peerList := strings.Split(data[1], ",")

	newPeers := make([]string, 0)
	for _, p := range peerList {
		newPeers = append(newPeers, strings.Trim(p, "\x00"))
	}

	fmt.Printf("============\n")
	fmt.Printf("Current Peers: %s\n", peersState.peers)
	fmt.Printf("============\n")

	newPeers = append(newPeers, peer)

	m := make(map[string]bool)
	for _, item := range peersState.peers {
		m[item] = true
	}

	for _, p := range newPeers {
		if p != port {
			if _, ok := m[p]; !ok {
				peersState.peers = append(peersState.peers, p)
			}
		}
	}

	fmt.Printf("============\n")
	fmt.Printf("Current Peers: %s\n", peersState.peers)
	fmt.Printf("============\n")
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
