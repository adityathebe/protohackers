package main

import "net"

type store struct {
	sessions map[int]*session
}

func newStore() *store {
	return &store{
		sessions: make(map[int]*session),
	}
}

func (t *store) OpenSession(sessID int, addr net.Addr) *session {
	if s, ok := t.sessions[sessID]; ok {
		return s
	}

	s := &session{
		id:   sessID,
		addr: addr,
		acks: make(map[int]struct{}),
	}
	t.sessions[sessID] = s
	return s
}

func (t *store) Session(sessID int) *session {
	s, ok := t.sessions[sessID]
	if !ok {
		return nil
	}

	return s
}

func (t *store) Close(sessID int) {
	delete(t.sessions, sessID)
}
