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
	minTime := m.First
	maxTime := m.Second

	var recordsWithinRange []Record
	for _, r := range t.records {
		if r.Timestamp >= minTime && r.Timestamp <= maxTime {
			recordsWithinRange = append(recordsWithinRange, r)
		}
	}

	if len(recordsWithinRange) == 0 {
		return 0
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
		log.Println("Handling client", clientIDctr)
		go handleConn(conn, clientStore[clientIDctr])
	}
}

func handleConn(conn *net.TCPConn, clientStore *Store) {
	conn.SetDeadline(time.Now().Add(time.Minute * 5))
	defer conn.Close()

	var buff []byte
	for {
		var b = make([]byte, 9)
		read, err := conn.Read(b)
		if err != nil {
			log.Printf("conn.Read(); %v", err)
			return
		}

		buff = append(buff, b[:read]...)
		if len(buff) < 9 {
			continue
		}

		actualMsgBuff := buff[:9]
		if len(buff) == 9 {
			buff = make([]byte, 0)
		} else {
			buff = buff[10:]
		}

		var msg Msg
		if err := msg.Decode(actualMsgBuff); err != nil {
			log.Printf("could not decode [%v]; %v", b, err)
			return
		}

		if !msg.isTypeOk() {
			continue
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
