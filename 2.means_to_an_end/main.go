package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"net"

	"github.com/adityathebe/protohackers"
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

	var n int
	var total int
	for _, r := range t.records {
		if r.Timestamp >= minTime && r.Timestamp <= maxTime {
			n++
			total += int(r.Price)
		}
	}

	if n == 0 {
		return 0
	}

	// calculate mean
	return int32(float64(total) / float64(n))
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
	protohackers.StartTCPServer(handleConn)
}

func handleConn(conn *net.TCPConn) {
	clientStore := &Store{}

	var b = make([]byte, 9)
	for {
		_, err := io.ReadAtLeast(conn, b, 9)
		if err != nil {
			log.Printf("conn.Read(); %v", err)
			return
		}

		var msg Msg
		if err := msg.Decode(b); err != nil {
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
				log.Println("Error writing mean", err)
			}
		}
	}
}
