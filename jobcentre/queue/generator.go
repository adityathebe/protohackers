package queue

import "sync"

type generator struct {
	id int
	m  *sync.Mutex
}

func newIDGenerator() *generator {
	return &generator{
		id: 1,
		m:  &sync.Mutex{},
	}
}

func (t *generator) Gen() int {
	t.m.Lock()
	id := t.id
	t.id++
	t.m.Unlock()
	return id
}
