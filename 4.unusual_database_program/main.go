package main

import (
	"errors"
	"fmt"
	"log"
	"net"
	"strings"
	"sync"
)

type store struct {
	m       *sync.Mutex
	records map[string]string
}

func newStore() *store {
	return &store{
		m:       &sync.Mutex{},
		records: make(map[string]string),
	}
}

func (t *store) insert(key, val string) error {
	t.m.Lock()
	defer t.m.Unlock()

	if key == "version" {
		return nil
	}

	t.records[key] = val
	return nil
}

func (t *store) retrieve(key string) (string, error) {
	t.m.Lock()
	defer t.m.Unlock()

	if key == "version" {
		return "Aditya's Key-Value Store 1.0", nil
	}

	val, ok := t.records[key]
	if !ok {
		return "", errors.New("not_found")
	}

	return val, nil
}

func main() {
	network := "udp"
	addr := ":3723"
	laddr, err := net.ResolveUDPAddr(network, addr)
	if err != nil {
		log.Fatalf("net.ResolveUDPAddr(); %v", err)
	}

	conn, err := net.ListenUDP(network, laddr)
	if err != nil {
		log.Fatalf("net.ResolveUDPAddr(); %v", err)
	}

	defer conn.Close()

	var b = make([]byte, 1024)
	var store = newStore()
	for {
		read, addr, err := conn.ReadFromUDP(b)
		if err != nil {
			log.Printf("Err:: conn.ReadFromUDP(); %v", err)
			continue
		}

		go sendResponse(store, conn, addr, string(b[:read]))
	}
}

func sendResponse(store *store, conn *net.UDPConn, remoteAddr *net.UDPAddr, msg string) {
	// Handle insert request
	if strings.Contains(msg, "=") {
		splitted := strings.SplitN(msg, "=", 2)
		key, val := splitted[0], splitted[1]
		store.insert(key, val)
		return
	}

	// Handle retrieve request
	val, err := store.retrieve(msg)
	if err != nil {
		return
	}

	resp := fmt.Sprintf("%s=%v", msg, val)
	if _, err := conn.WriteToUDP([]byte(resp), remoteAddr); err != nil {
		log.Printf("Err:: conn.WriteToUDP(); %v", err)
		return
	}
}
