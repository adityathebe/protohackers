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

	t.Parallel()

	t.Run("all positive", func(t *testing.T) {
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
	})

	t.Run("example from protohackers", func(t *testing.T) {
		conn, err := net.DialTCP(network, nil, laddr)
		if err != nil {
			log.Fatalf("net.ResolveTCPAddr(); %v", err)
		}

		//     Hexadecimal:                 Decoded:
		// <-- 49 00 00 30 39 00 00 00 65   I 12345 101
		// <-- 49 00 00 30 3a 00 00 00 66   I 12346 102
		// <-- 49 00 00 30 3b 00 00 00 64   I 12347 100
		// <-- 49 00 00 a0 00 00 00 00 05   I 40960 5
		// <-- 51 00 00 30 00 00 00 40 00   Q 12288 16384
		// --> 00 00 00 65                  101
		conn.Write([]byte{0x49, 0, 0, 0x30, 0x39, 0, 0, 0, 0x65})
		conn.Write([]byte{0x49, 0, 0, 0x30, 0x3a, 0, 0, 0, 0x66})
		conn.Write([]byte{0x49, 0, 0, 0x30, 0x3b, 0, 0, 0, 0x64})
		conn.Write([]byte{0x49, 0, 0, 0xa0, 0, 0, 0, 0, 0xff})
		conn.Write([]byte{0x51, 0, 0, 0x30, 0, 0, 0, 0x40, 0})

		var resp = make([]byte, 4)
		conn.Read(resp)

		var x int32
		buf := bytes.NewBuffer(resp)
		if err := binary.Read(buf, binary.BigEndian, &x); err != nil {
			t.Fatalf("%v", err)
		}

		if x != 101 {
			t.Fatalf("Expected 11 got %d", x)
		}
	})
}
