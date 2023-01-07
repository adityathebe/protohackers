package pkg

import "fmt"

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

func (t *lock) leave(clientID int) {
	fmt.Println("broadcast:: leaving", clientID)
	defer fmt.Println("broadcast:: left", clientID)

	ch, ok := t.listeners[clientID]
	if !ok {
		return
	}

	ch <- struct{}{}
	delete(t.listeners, clientID)
}

func (t *lock) wait(clientID int) {
	fmt.Println("broadcast:: wait", clientID)
	defer fmt.Println("broadcast:: wait over", clientID)

	ch := make(chan struct{}, 1)
	t.listeners[clientID] = ch
	<-ch
	delete(t.listeners, clientID)
}
