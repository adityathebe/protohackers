package queue

import (
	"sync"
)

type subscriberRecord struct {
	id int
	ch chan struct{}
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
	for i, s := range t.subscribers {
		if s.id == id {
			return t.subscribers[i].ch
		}
	}

	ch := make(chan struct{}, 100000) // Buffered channel to avoid race condition.
	t.subscribers = append(t.subscribers, subscriberRecord{id: id, ch: ch})
	return ch
}

func (t *subscription) Unsubscribe(id int) {
	t.m.Lock()
	defer t.m.Unlock()

	for i, s := range t.subscribers {
		if s.id == id {
			t.subscribers = append(t.subscribers[:i], t.subscribers[i+1:]...)
			return
		}
	}
}

func (t *subscription) announce() {
	t.m.Lock()
	defer t.m.Unlock()

	if len(t.subscribers) == 0 {
		return
	}

	for i := range t.subscribers {
		t.subscribers[i].ch <- struct{}{}
	}
}
