package main

import (
	"bufio"
	"encoding/json"
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

	controller := pkg.NewJobController()
	var clientID int
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("listener.Accept(); %v", err)
			continue
		}

		clientID++
		go handleConn(conn, clientID, controller)
	}
}

func handleConn(conn net.Conn, clientID int, controller *pkg.JobController) {
	defer func() {
		log.Println("Cleanup", clientID)
		conn.Close()
		controller.ReleaseActiveJobs(clientID)
		controller.Leave(clientID)
	}()

	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		var r pkg.Request
		if err := json.Unmarshal([]byte(scanner.Text()), &r); err != nil || !r.IsValid() {
			response := pkg.Response{Status: "error"}
			conn.Write(response.Json())
			continue
		}

		response := handleCommand(clientID, controller, r)
		conn.Write(response.Json())
	}
}

func handleCommand(clientID int, controller *pkg.JobController, r pkg.Request) pkg.Response {
	switch r.Request {
	case "put":
		job := controller.Put(r.Queue, *r.Priority, r.Job)
		return pkg.Response{Status: "ok", ID: job.ID}

	case "get":
		job, qName := controller.GetWithWait(clientID, r.Queues, r.Wait)
		if job == nil {
			return pkg.Response{Status: "no-job"}
		}

		return pkg.Response{Status: "ok", Queue: qName, Priority: job.Priority, ID: job.ID, Job: job.Content}

	case "abort":
		aborted := controller.Abort(clientID, *r.ID)
		if !aborted {
			return pkg.Response{Status: "no-job"}
		}

		return pkg.Response{Status: "ok"}

	case "delete":
		deleted := controller.Delete(clientID, *r.ID)
		if !deleted {
			return pkg.Response{Status: "no-job"}
		}

		return pkg.Response{Status: "ok"}

	default:
		return pkg.Response{Status: "error"}
	}
}
