package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

type ResponseType uint8

const (
	Error     ResponseType = 0x10
	Ticket    ResponseType = 0x21
	Heartbeat ResponseType = 0x41
)

type TicketMsg struct {
	Plate      string
	Road       uint16
	Mile1      uint16
	Timestamp1 uint32
	Mile2      uint16
	Timestamp2 uint32
	Speed      uint16
}

type TicketHash string

func (t TicketMsg) Hash() TicketHash {
	s := fmt.Sprintf("%s%d%d%d%d%d%d", t.Plate, t.Road, t.Mile1, t.Timestamp1, t.Mile2, t.Timestamp2, t.Speed)
	return TicketHash(s)
}

type Response struct {
	Type   ResponseType
	ErrMsg string
	Ticket TicketMsg
}

func (t *Response) Encode() []byte {
	var b []byte
	buf := bytes.NewBuffer(b)

	// Write the type
	if err := binary.Write(buf, binary.BigEndian, t.Type); err != nil {
		panic(err)
	}

	switch t.Type {
	case Error:
		if err := binary.Write(buf, binary.BigEndian, stringEncoder(t.ErrMsg)); err != nil {
			panic(err)
		}

	case Ticket:
		if err := binary.Write(buf, binary.BigEndian, stringEncoder(t.Ticket.Plate)); err != nil {
			panic(err)
		}

		if err := binary.Write(buf, binary.BigEndian, t.Ticket.Road); err != nil {
			panic(err)
		}

		if err := binary.Write(buf, binary.BigEndian, t.Ticket.Mile1); err != nil {
			panic(err)
		}

		if err := binary.Write(buf, binary.BigEndian, t.Ticket.Timestamp1); err != nil {
			panic(err)
		}

		if err := binary.Write(buf, binary.BigEndian, t.Ticket.Mile2); err != nil {
			panic(err)
		}

		if err := binary.Write(buf, binary.BigEndian, t.Ticket.Timestamp2); err != nil {
			panic(err)
		}

		if err := binary.Write(buf, binary.BigEndian, t.Ticket.Speed); err != nil {
			panic(err)
		}

	case Heartbeat:
		// No fields
	}

	return buf.Bytes()
}

func stringEncoder(s string) []byte {
	return append([]byte{byte(len(s))}, s...)
}
