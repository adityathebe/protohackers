package main

import (
	"encoding/json"
	"errors"
	"io"
	"log"
	"math/big"
	"net"
)

type Request struct {
	Method string `json:"method"`
	Number int64  `json:"number"`
}

type Response struct {
	Method string `json:"method"`
	Prime  bool   `json:"prime"`
}

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

	var b = make([]byte, 1024*16)
	for {
		read, err := conn.Read(b)
		if err != nil && !errors.Is(err, io.EOF) {
			log.Printf("conn.Read(); %v", err)
			return
		}

		var req Request
		if err := json.Unmarshal(b[:read], &req); err != nil {
			log.Printf("malformed request(); %v", err)
			continue
		}

		if req.Method == "isPrime" {
			isPrime := big.NewInt(req.Number).ProbablyPrime(0)
			resp := Response{
				Method: "isPrime",
				Prime:  isPrime,
			}
			encoder := json.NewEncoder(conn)
			if err := encoder.Encode(resp); err != nil {
				log.Printf("conn.Read(); %v", err)
				continue
			}
		}

		if errors.Is(err, io.EOF) {
			break
		}
	}
}
