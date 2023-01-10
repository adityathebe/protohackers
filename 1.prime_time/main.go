package main

import (
	"bufio"
	"encoding/json"
	"log"
	"math/big"
	"net"

	"github.com/adityathebe/protohackers"
)

type Request struct {
	Method string   `json:"method"`
	Number *float64 `json:"number"`
}

func (r Request) isMalformed() bool {
	if r.Number == nil {
		log.Println("Malformed request. No Number")
		return true
	}

	if r.Method != "isPrime" {
		return true
	}

	return false
}

type Response struct {
	Method string `json:"method"`
	Prime  bool   `json:"prime"`
}

func main() {
	protohackers.StartTCPServer(handleConn)
}

func handleConn(conn *net.TCPConn) {
	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		var req Request
		if err := json.Unmarshal([]byte(scanner.Text()), &req); err != nil {
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
