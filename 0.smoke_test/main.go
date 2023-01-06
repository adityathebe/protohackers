package main

import (
	"log"
	"net"
)

func main() {
	addr := ":3723"
	network := "tcp"
	laddr, err := net.ResolveTCPAddr(network, addr)
	if err != nil {
		log.Fatalf("net.ResolveTCPAddr(); %v", err)
	}

	listener, err := net.ListenTCP(network, laddr)
	if err != nil {
		log.Fatalf("net.ListenTCP(); %v", err)
	}
	defer listener.Close()

	for {
		conn, err := listener.AcceptTCP()
		if err != nil {
			log.Printf("listener.AcceptTCP(); %v", err)
			continue
		}

		go handleConn(conn)
	}
}

func handleConn(conn *net.TCPConn) {
	defer conn.Close()

	var b = make([]byte, 1024)
	for {
		read, err := conn.Read(b)
		if err != nil {
			log.Printf("conn.Read(); %v", err)
			return
		}

		conn.Write(b[:read])
	}
}
