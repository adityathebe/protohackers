package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"log"
	"net"
	"time"
)

type Record struct {
	Timestamp int32
	Price     int32
}

type Store struct {
	records []Record
}

func (t *Store) insert(m Msg) {
	var r Record
	r.Price = m.Second
	r.Timestamp = m.First
	t.records = append(t.records, r)
}

func (t *Store) query(m Msg) int32 {
	from := m.First
	to := m.Second
	var recordsWithinRange []Record
	for _, r := range t.records {
		if r.Timestamp >= from && r.Timestamp <= to {
			recordsWithinRange = append(recordsWithinRange, r)
		}
	}

	// calculate mean
	var total int64
	for _, r := range recordsWithinRange {
		total += int64(r.Price)
	}

	return int32(float64(total) / float64(len(recordsWithinRange)))
}

type Msg struct {
	Type   int8  // I or Q
	First  int32 // Timestamp for insert type & min time for query type
	Second int32 // Price for insert type & max time for query type
}

func (t *Msg) Decode(b []byte) error {
	if len(b) != 9 {
		return fmt.Errorf("incorrect length for Msg: expected 9, got %d", len(b))
	}

	buf := bytes.NewBuffer(b)
	if err := binary.Read(buf, binary.BigEndian, &t.Type); err != nil {
		return err
	}
	if err := binary.Read(buf, binary.BigEndian, &t.First); err != nil {
		return err
	}
	if err := binary.Read(buf, binary.BigEndian, &t.Second); err != nil {
		return err
	}

	return nil
}

func (t *Msg) isTypeOk() bool {
	msgType := t.Type
	return msgType == 'I' || msgType == 'Q'
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

	var clientStore map[int]*Store = make(map[int]*Store)
	var clientIDctr int

	for {
		conn, err := listener.AcceptTCP()
		if err != nil {
			log.Printf("listener.AcceptTCP(); %v", err)
			continue
		}

		clientIDctr++
		clientStore[clientIDctr] = &Store{}
		go handleConn(conn, clientStore[clientIDctr])
	}
}

func handleConn(conn *net.TCPConn, clientStore *Store) {
	conn.SetDeadline(time.Now().Add(time.Second * 10))
	defer conn.Close()

	var b = make([]byte, 9)
	for {
		read, err := conn.Read(b)
		if err != nil {
			log.Printf("conn.Read(); %v", err)
			return
		}

		var msg Msg
		if err := msg.Decode(b[:read]); err != nil {
			log.Println("Client sent invalid msg")
			return
		}

		if !msg.isTypeOk() {
			log.Println("Client sent invalid msg")
			return
		}

		switch msg.Type {
		case 'I':
			clientStore.insert(msg)
		case 'Q':
			mean := clientStore.query(msg)
			if err := binary.Write(conn, binary.BigEndian, mean); err != nil {
				log.Println("Error writing mean")
			}
		}
	}
}
