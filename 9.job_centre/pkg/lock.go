package pkg

import "log"

type lock struct {
	listeners map[int]chan struct{}
}

func newLock() *lock {
	return &lock{
		listeners: make(map[int]chan struct{}),
	}
}

func (t *lock) announce() {
	for clientID := range t.listeners {
		t.listeners[clientID] <- struct{}{}
	}
}

func (t *lock) isWaiting(clientID int) bool {
	_, ok := t.listeners[clientID]
	return ok
}

func (t *lock) leave(clientID int) {
	log.Println("loc:: leaving", clientID)
	defer log.Println("loc:: left", clientID)

	ch, ok := t.listeners[clientID]
	if !ok {
		return
	}

	ch <- struct{}{}
	delete(t.listeners, clientID)
}

func (t *lock) wait(clientID int) {
	log.Println("loc:: wait", clientID)
	defer log.Println("loc:: wait over", clientID)

	ch := make(chan struct{}, 1)
	t.listeners[clientID] = ch
	<-ch
	delete(t.listeners, clientID)
}
