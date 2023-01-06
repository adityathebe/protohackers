package main

import (
	"bytes"
	"encoding/binary"
	"log"
	"net"
	"testing"
)

func TestMain(t *testing.T) {
	addr := ":3723"
	network := "tcp"
	laddr, err := net.ResolveTCPAddr(network, addr)
	if err != nil {
		log.Fatalf("net.ResolveTCPAddr(); %v", err)
	}

	conn, err := net.DialTCP(network, nil, laddr)
	if err != nil {
		log.Fatalf("net.ResolveTCPAddr(); %v", err)
	}

	conn.Write([]byte{73, 0, 0, 48, 57, 0, 0, 0, 10})
	conn.Write([]byte{73, 0, 0, 48, 58, 0, 0, 0, 12})
	conn.Write([]byte{81, 0, 0, 0, 0, 0, 0, 50, 50})

	var resp = make([]byte, 9)
	conn.Read(resp)

	var x int32
	buf := bytes.NewBuffer(resp)
	if err := binary.Read(buf, binary.BigEndian, &x); err != nil {
		t.Fatalf("%v", err)
	}

	if x != 11 {
		t.Fatalf("Expected 11 got %d", x)
	}
}
