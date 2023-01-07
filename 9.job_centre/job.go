package main

import (
	"encoding/json"
	"sync"
)

type job struct {
	id              int
	content         json.RawMessage
	priority        int
	isBeingWorkedOn bool
}

type jobIDGenerator struct {
	id int
	m  *sync.Mutex
}

func newJobIDGenerator() *jobIDGenerator {
	return &jobIDGenerator{
		id: 1,
		m:  &sync.Mutex{},
	}
}

func (t *jobIDGenerator) gen() int {
	t.m.Lock()
	id := t.id
	t.id++
	t.m.Unlock()
	return id
}
