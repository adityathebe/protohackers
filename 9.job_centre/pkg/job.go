package pkg

import (
	"encoding/json"
	"sync"
)

type Job struct {
	ID       int
	Content  json.RawMessage
	Priority int
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
