package main

import (
	"net"

	"github.com/adityathebe/protohackers"
)

func main() {
	protohackers.StartTCPServer(handleConn)
}

func handleConn(conn *net.TCPConn) {
	conn.ReadFrom(conn)
}
