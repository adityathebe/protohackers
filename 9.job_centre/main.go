package main

import (
	"log"
	"net"

	"github.com/adityathebe/protohackers/9.job_centre/pkg"
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

	controller := pkg.NewJobController()
	var clientID int
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("listener.Accept(); %v", err)
			continue
		}

		clientID++
		client := newClient(clientID, conn, controller)
		go client.start()
	}
}
