package main

import (
	"net"
)

type session struct {
	id   int
	addr net.Addr

	maxAck int32
	// acks is the list of all lengths sent by the client
	// via ack messages
	acks map[int]struct{}

	bSent     int32
	bReceived int32
}

func (t *session) Ack(sessID int, len int32) {
	t.acks[sessID] = struct{}{}

	if len > t.maxAck {
		t.maxAck = len
	}
}

func (t *session) RegisterPayload(pos int, data string) {
	t.bSent += int32(len(data)) + 1
}
