package client

import (
	"bufio"
	"context"
	"encoding/json"
	"log"
	"net"

	jobcentre "github.com/adityathebe/protohackers/jobcentre"
	"github.com/adityathebe/protohackers/jobcentre/queue"
)

type client struct {
	clientID     int
	conn         net.Conn
	queue        *queue.Queue
	requestChan  chan jobcentre.Request
	responseChan chan *jobcentre.Response
	ctx          context.Context
	done         context.CancelFunc
}

func NewClient(clientID int, conn net.Conn, queue *queue.Queue) *client {
	ctx, cancel := context.WithCancel(context.Background())
	return &client{
		clientID:     clientID,
		conn:         conn,
		queue:        queue,
		requestChan:  make(chan jobcentre.Request, 70000), // buffered because we still want to process other requests
		responseChan: make(chan *jobcentre.Response),      // unbuffered because we want response in order
		ctx:          ctx,
		done:         cancel,
	}
}

func (t *client) Start() {
	log.Println("Starting client", t.clientID)
	go t.startReqHandler()
	go t.startResHandler()
	t.handleConn()
}

func (t *client) Close() {
	log.Println("Closing client", t.clientID)
	t.queue.ReleaseActiveJobs(t.clientID)
	t.conn.Close()
	t.done()
}

func (t *client) startReqHandler() {
	for {
		select {
		case <-t.ctx.Done():
			log.Println("stopping request handler [top level]", t.clientID)
			return

		case r := <-t.requestChan:
			ch := make(chan *jobcentre.Response)
			go t.handleRequest(r, ch)

			select {
			case res := <-ch:
				if res != nil {
					t.responseChan <- res
				}

			case <-t.ctx.Done():
				log.Println("stopping request handler [down level]", t.clientID)
				return
			}
		}
	}
}

func (t *client) startResHandler() {
	defer log.Println("stopping response handler", t.clientID)

	for {
		select {
		case <-t.ctx.Done():
			return

		case res := <-t.responseChan:
			t.conn.Write(res.Json())
		}
	}
}

func (t *client) handleConn() {
	defer t.Close()

	scanner := bufio.NewScanner(t.conn)
	for scanner.Scan() {
		var r jobcentre.Request
		if err := json.Unmarshal([]byte(scanner.Text()), &r); err != nil || !r.IsValid() {
			response := jobcentre.Response{Status: "error"}
			t.conn.Write(response.Json())
			continue
		}

		t.requestChan <- r
	}
}

func (t *client) handleRequest(r jobcentre.Request, ch chan<- *jobcentre.Response) {
	defer close(ch)

	switch r.Request {
	case "put":
		job := t.queue.Put(r.Queue, *r.Priority, *r.Job)
		ch <- &jobcentre.Response{Status: "ok", ID: job.ID}
		return

	case "get":
		record := t.queue.Get(t.clientID, r.Queues)
		if record != nil {
			ch <- &jobcentre.Response{Status: "ok", Queue: record.QName, Priority: record.Job.Priority, ID: record.Job.ID, Job: record.Job.Content}
			return
		} else {
			if !r.Wait {
				ch <- &jobcentre.Response{Status: "no-job"}
				return
			}

			qChan := t.queue.Subscribe(t.clientID)
			for {
				select {
				case <-t.ctx.Done():
					t.queue.Unsubscribe(t.clientID)
					log.Println("Aborting listening on queue")
					ch <- nil
					return

				case <-qChan:
					record := t.queue.Get(t.clientID, r.Queues)
					if record != nil {
						t.queue.Unsubscribe(t.clientID)
						ch <- &jobcentre.Response{Status: "ok", Queue: record.QName, Priority: record.Job.Priority, ID: record.Job.ID, Job: record.Job.Content}
						return
					}
				}
			}
		}

	case "abort":
		aborted, err := t.queue.Abort(t.clientID, *r.ID, true)
		if err != nil {
			ch <- &jobcentre.Response{Status: "error"}
			return
		}

		if !aborted {
			ch <- &jobcentre.Response{Status: "no-job"}
			return
		}

		ch <- &jobcentre.Response{Status: "ok"}
		return

	case "delete":
		deleted := t.queue.Delete(*r.ID)
		if !deleted {
			ch <- &jobcentre.Response{Status: "no-job"}
			return
		}

		ch <- &jobcentre.Response{Status: "ok"}
		return

	default:
		ch <- &jobcentre.Response{Status: "error"}
	}
}
