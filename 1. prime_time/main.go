package main

import (
	"encoding/json"
	"errors"
	"io"
	"log"
	"math/big"
	"net"
	"strings"
)

type Request struct {
	Method *string  `json:"method"`
	Number *float64 `json:"number"`
}

func (r Request) isMalformed() bool {
	if r.Method == nil {
		log.Println("Malformed request. No method")
		return true
	}

	if r.Number == nil {
		log.Println("Malformed request. No Number")
		return true
	}

	if *r.Method != "isPrime" {
		return true
	}

	return false
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
	defer func() {
		log.Println("Closing connection")
		conn.Close()
	}()

	var b = make([]byte, 256)
	var msgBuffer string
	for {
		len, err := conn.Read(b)
		if err != nil {
			if errors.Is(err, io.EOF) {
				log.Println("client connection closed")
				return
			}

			log.Println("Something went wrong")
			return
		}

		msgBuffer += string(b[:len])
		idx := strings.Index(msgBuffer, "\n")
		if idx < 0 {
			continue
		}

		jsonMsg := msgBuffer[:idx]
		newmsgbuffer := msgBuffer[idx+1:]
		msgBuffer = newmsgbuffer

		var req Request
		if err := json.Unmarshal([]byte(jsonMsg), &req); err != nil {
			log.Println("malformed request();", err)
			conn.Write([]byte("nonsense"))
			return
		}

		if req.isMalformed() {
			log.Println("malformed request()")
			conn.Write([]byte("nonsense"))
			return
		}

		var isPrime bool
		if float64(int(*req.Number)) == *req.Number { // only check for prime if the number is an integer
			isPrime = big.NewInt(int64(*req.Number)).ProbablyPrime(0)
		}

		resp := Response{
			Method: "isPrime",
			Prime:  isPrime,
		}
		encoder := json.NewEncoder(conn)
		if err := encoder.Encode(resp); err != nil {
			log.Printf("conn.Read(); %v\n", err)
			continue
		}
	}
}
