package protohackers

import (
	"log"
	"net"
)

const (
	addr    = ":3723"
	network = "tcp"
)

type ConnectionHandler func(conn *net.TCPConn)

func StartTCPServer(handler ConnectionHandler) {
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
			log.Printf("listener.AcceptTCP(); %v\n", err)
			continue
		}

		go func() {
			defer conn.Close()
			handler(conn)
		}()
	}
}
