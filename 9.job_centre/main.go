package main

import (
	"bufio"
	"encoding/json"
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

	controller := newJobController()
	var clientID int
	for {
		conn, err := listener.AcceptTCP()
		if err != nil {
			log.Printf("listener.AcceptTCP(); %v", err)
			continue
		}

		clientID++
		go handleConn(conn, clientID, controller)
	}
}

func handleConn(conn *net.TCPConn, clientID int, controller *JobController) {
	defer func() {
		conn.Close()
		controller.releaseActiveJobs(clientID)
	}()

	writer := json.NewEncoder(conn)
	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		var r Request
		if err := json.Unmarshal([]byte(scanner.Text()), &r); err != nil {
			response := Response{Status: "error"}
			if err := writer.Encode(response); err != nil {
				log.Println("error sending response", err)
			}
			continue
		}

		if !r.isValid() {
			log.Println("Invalid request")
			response := Response{Status: "error"}
			if err := writer.Encode(response); err != nil {
				log.Println("error sending response", err)
			}
			continue
		}

		response := handleCommand(clientID, controller, r)
		if err := writer.Encode(response); err != nil {
			log.Println("error sending response after handleCommand", err)
		}
	}
}

func handleCommand(clientID int, controller *JobController, r Request) Response {
	switch r.Request {
	case "put":
		job := controller.put(r.Queue, *r.Priority, r.Job)
		return Response{Status: "ok", ID: job.id}

	case "get":
		job, qName := controller.getWithWait(clientID, r.Queues, r.Wait)
		if job == nil {
			return Response{Status: "no-job"}
		}

		return Response{Status: "ok", Queue: qName, Priority: job.priority, ID: job.id, Job: job.content}

	case "abort":
		aborted := controller.abort(clientID, *r.ID)
		if !aborted {
			return Response{Status: "no-job"}
		}

		return Response{Status: "ok"}

	case "delete":
		deleted := controller.delete(*r.ID)
		if !deleted {
			return Response{Status: "no-job"}
		}

		return Response{Status: "ok"}

	default:
		return Response{Status: "error"}
	}
}