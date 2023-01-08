package queue

import (
	"fmt"
	"sync"
)

type subscriberRecord struct {
	id int
	ch chan<- struct{}
}

type subscription struct {
	m           *sync.Mutex
	subscribers []subscriberRecord
}

func newSubscription() *subscription {
	return &subscription{
		m: &sync.Mutex{},
	}
}

func (t *subscription) Subscribe(id int) <-chan struct{} {
	t.m.Lock()
	defer t.m.Unlock()

	// Sanity check
	for _, s := range t.subscribers {
		if s.id == id {
			panic(fmt.Sprintf("User [%d] is already subscribe", id))
		}
	}

	ch := make(chan struct{})
	t.subscribers = append(t.subscribers, subscriberRecord{id: id, ch: ch})
	return ch
}

func (t *subscription) Unsubscribe(token int) {
	t.m.Lock()
	defer t.m.Unlock()

	for i, s := range t.subscribers {
		if s.id == token {
			t.subscribers = append(t.subscribers[:i], t.subscribers[i+1:]...)
			return
		}
	}
}

func (t *subscription) announce() {
	t.m.Lock()
	defer t.m.Unlock()

	for i := range t.subscribers {
		t.subscribers[i].ch <- struct{}{}
	}
}
