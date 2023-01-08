package main

import (
	"log"
	"net"

	"github.com/adityathebe/protohackers/jobcentre/client"
	"github.com/adityathebe/protohackers/jobcentre/queue"
)

func main() {
	addr := ":3723"
	network := "tcp"

	listener, err := net.Listen(network, addr)
	if err != nil {
		log.Fatalf("net.Listen(); %v", err)
	}
	defer listener.Close()

	log.Println("Started listening on", addr)

	queue := queue.NewQueue()
	var clientID int
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("listener.Accept(); %v", err)
			continue
		}

		clientID++
		client := client.NewClient(clientID, conn, queue)
		go client.Start()
	}
}
