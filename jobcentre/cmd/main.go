package main

import (
	"context"
	"log"
	"net"
	"os"
	"os/signal"

	"github.com/adityathebe/protohackers/jobcentre/client"
	"github.com/adityathebe/protohackers/jobcentre/queue"
)

func main() {
	addr := ":3723"
	network := "tcp"

	// we need a webserver to get the pprof webserver
	// f, err := os.Create("cmd-profile.prof")
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// pprof.StartCPUProfile(f)

	// go func() {
	// 	ctx := newCancelableContext()
	// 	<-ctx.Done()

	// 	fmt.Println("Done. Creating profile")

	// 	pprof.StopCPUProfile()

	// 	os.Exit(1)
	// }()

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

func newCancelableContext() context.Context {
	doneCh := make(chan os.Signal, 1)
	signal.Notify(doneCh, os.Interrupt)

	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		<-doneCh
		cancel()
	}()

	return ctx
}
