package main

import (
	"bufio"
	"encoding/json"
	"log"
	"net"

	"github.com/adityathebe/protohackers/9.job_centre/pkg"
)

type client struct {
	clientID     int
	conn         net.Conn
	controller   *pkg.JobController
	requestChan  chan pkg.Request
	responseChan chan *pkg.Response
	doneCh       chan struct{}
}

func newClient(clientID int, conn net.Conn, controller *pkg.JobController) *client {
	return &client{
		clientID:     clientID,
		conn:         conn,
		controller:   controller,
		requestChan:  make(chan pkg.Request, 10000), // buffered because we still want to process other requests
		responseChan: make(chan *pkg.Response),      // unbuffered because we want response in order
		doneCh:       make(chan struct{}),
	}
}

func (t *client) start() {
	log.Println("Starting client", t.clientID)
	go t.startReqHandler()
	go t.startResHandler()
	t.handleConn()
}

func (t *client) startReqHandler() {
	for {
		select {
		case <-t.doneCh:
			return

		case r := <-t.requestChan:
			res := t.handleCommand(r)
			t.responseChan <- res
		}
	}
}

func (t *client) startResHandler() {
	for {
		select {
		case <-t.doneCh:
			return

		case res := <-t.responseChan:
			t.conn.Write(res.Json())
		}
	}
}

func (t *client) Close() {
	log.Println("Cleanup", t.clientID)
	t.conn.Close()
	t.controller.ReleaseActiveJobs(t.clientID)
	t.controller.Leave(t.clientID)
	t.doneCh <- struct{}{}
	t.doneCh <- struct{}{}
}

func (t *client) handleConn() {
	defer t.Close()

	scanner := bufio.NewScanner(t.conn)
	for scanner.Scan() {
		var r pkg.Request
		if err := json.Unmarshal([]byte(scanner.Text()), &r); err != nil || !r.IsValid() {
			response := pkg.Response{Status: "error"}
			t.conn.Write(response.Json())
			continue
		}

		t.requestChan <- r
	}
}

func (t *client) handleCommand(r pkg.Request) *pkg.Response {
	switch r.Request {
	case "put":
		job := t.controller.Put(r.Queue, *r.Priority, *r.Job)
		return &pkg.Response{Status: "ok", ID: job.ID}

	case "get":
		job, qName := t.controller.GetWithWait(t.clientID, r.Queues, r.Wait)
		if job == nil {
			return &pkg.Response{Status: "no-job"}
		}

		return &pkg.Response{Status: "ok", Queue: qName, Priority: job.Priority, ID: job.ID, Job: job.Content}

	case "abort":
		aborted, unauthorized := t.controller.Abort(t.clientID, *r.ID)
		if unauthorized {
			return &pkg.Response{Status: "error"}
		}

		if !aborted {
			return &pkg.Response{Status: "no-job"}
		}

		return &pkg.Response{Status: "ok"}

	case "delete":
		deleted := t.controller.Delete(t.clientID, *r.ID)
		if !deleted {
			return &pkg.Response{Status: "no-job"}
		}

		return &pkg.Response{Status: "ok"}

	default:
		return &pkg.Response{Status: "error"}
	}
}
