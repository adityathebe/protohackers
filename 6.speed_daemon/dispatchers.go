package main

import (
	"log"
	"net"
	"sync"
)

type DispatcherClient struct {
	id     int
	conn   net.Conn
	closed bool
}

type DispatcherStore struct {
	mu           *sync.Mutex
	dispatcher   map[int]*DispatcherClient
	roadMaps     map[uint16][]*DispatcherClient
	issued       map[TicketHash]struct{}
	issuedPerDay map[string]map[int]struct{}
	pending      map[uint16]map[TicketHash]TicketMsg // pending stores all tickets to be dispatched per road
}

func newDispatcherStore() *DispatcherStore {
	return &DispatcherStore{
		mu:           &sync.Mutex{},
		dispatcher:   make(map[int]*DispatcherClient),
		roadMaps:     make(map[uint16][]*DispatcherClient),
		issued:       make(map[TicketHash]struct{}),
		issuedPerDay: make(map[string]map[int]struct{}),
		pending:      make(map[uint16]map[TicketHash]TicketMsg),
	}
}

func (t *DispatcherStore) Register(id int, conn net.Conn, roads []uint16) {
	t.mu.Lock()
	defer t.mu.Unlock()

	dispatcher := &DispatcherClient{id: id, conn: conn}
	t.dispatcher[id] = dispatcher
	for _, r := range roads {
		t.roadMaps[r] = append(t.roadMaps[r], dispatcher)
	}

	t.publishPending(roads...)
}

func (t *DispatcherStore) Unregister(id int) {
	t.dispatcher[id].closed = true
}

// Dispatch dispatches tickets if there are dispatches available
// for the road.
//
// Else it will simply store the tickets until it finds a dispatcher.
func (t *DispatcherStore) Dispatch(tickets []TicketMsg) {
	t.mu.Lock()
	defer t.mu.Unlock()

	for _, ticket := range tickets {
		hash := ticket.Hash()

		if _, ok := t.issued[hash]; ok {
			log.Printf("Ticket has already been dispatched [%s] [%v]\n", hash, ticket)
			return
		}

		if _, ok := t.pending[ticket.Road]; !ok {
			t.pending[ticket.Road] = make(map[TicketHash]TicketMsg, 1)
		}
		t.pending[ticket.Road][hash] = ticket

		t.publishPending(ticket.Road)
	}
}

func (t *DispatcherStore) getActiveDispatchersOfRoad(road uint16) []*DispatcherClient {
	activesOnes := make([]*DispatcherClient, 0, len(t.roadMaps[road]))
	for _, dispatcher := range t.roadMaps[road] {
		if !dispatcher.closed {
			activesOnes = append(activesOnes, dispatcher)
		}
	}

	return activesOnes
}

func (t *DispatcherStore) publishPending(roads ...uint16) {
	for _, r := range roads {
		activeDispatchers := t.getActiveDispatchersOfRoad(r)
		if len(activeDispatchers) == 0 {
			continue
		}

		chosen := activeDispatchers[0]
		var justIssued []TicketHash
		for hash, tbTicket := range t.pending[r] {
			if _, ok := t.issuedPerDay[tbTicket.Plate][int(tbTicket.Timestamp1)/86400]; ok {
				log.Printf("This car has already received a ticket for day [%d]\n", int(tbTicket.Timestamp1)/86400)
				continue
			}

			if _, ok := t.issuedPerDay[tbTicket.Plate][int(tbTicket.Timestamp2)/86400]; ok {
				log.Printf("This car has already received a ticket for day [%d]\n", int(tbTicket.Timestamp2)/86400)
				continue
			}

			// Publish to the dispatcher
			res := Response{Type: Ticket, Ticket: tbTicket}
			if _, err := chosen.conn.Write(res.Encode()); err != nil {
				log.Printf("Error dispatching to %d\n", chosen.id)
				continue
			}

			log.Printf("Dispatched ticket [%s] [%v]\n", hash, tbTicket)

			// Mark the ticket as published
			t.issued[hash] = struct{}{}

			// Store the day of ticket for the car
			if _, ok := t.issuedPerDay[tbTicket.Plate]; !ok {
				t.issuedPerDay[tbTicket.Plate] = make(map[int]struct{})
			}
			t.issuedPerDay[tbTicket.Plate][int(tbTicket.Timestamp1)/86400] = struct{}{}
			t.issuedPerDay[tbTicket.Plate][int(tbTicket.Timestamp2)/86400] = struct{}{}

			justIssued = append(justIssued, hash)
		}

		// The issued tickets should be removed from pending
		for _, hash := range justIssued {
			delete(t.pending[r], hash)
		}
	}
}
